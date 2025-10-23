package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jieqiboh/sothea_backend/entities"
)

type postgresPrescriptionRepository struct {
	Conn *sql.DB
}

func NewPostgresPrescriptionRepository(conn *sql.DB) entities.PrescriptionRepository {
	return &postgresPrescriptionRepository{Conn: conn}
}

// -----------------------------------------------------------------------------
// PRESCRIPTIONS
// -----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) CreatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	tx, ok := TxFromCtx(ctx)
	ownTx := false

	if !ok {
		var err error
		tx, err = r.Conn.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		ownTx = true
		defer tx.Rollback()
	}

	err := tx.QueryRowContext(ctx, `
		INSERT INTO prescriptions (patient_id, vid, notes)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, p.PatientID, p.VID, p.Notes).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for i := range p.PrescribedDrugs {
		d := &p.PrescribedDrugs[i]

		err = tx.QueryRowContext(ctx, `
			INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks, quantity_requested)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at
		`, p.ID, d.DrugID, d.Remarks, d.RequestedQty).
			Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}
		for j := range d.Batches {
			b := &d.Batches[j]

			err = tx.QueryRowContext(ctx, `
				INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
				VALUES ($1, $2, $3)
				RETURNING id, created_at, updated_at
			`, d.ID, b.BatchLocationId, b.Quantity).
				Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
			if err != nil {
				return nil, err
			}

			res, err := tx.ExecContext(ctx, `
				UPDATE batch_locations
				SET quantity = quantity - $1
				WHERE id = $2 AND quantity >= $1
			`, b.Quantity, b.BatchLocationId)
			if err != nil {
				return nil, err
			}

			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("insufficient stock in batch-location %d", b.BatchLocationId)
			}

		}
	}

	if ownTx {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	var p entities.Prescription
	dbx := DBFromCtx(ctx, r.Conn)

	// parent
	if err := dbx.QueryRowContext(ctx, `
        SELECT id, patient_id, vid, notes, created_by,
               is_dispensed, dispensed_by, dispensed_at,
               created_at, updated_at
        FROM prescriptions
        WHERE id = $1
    `, id).Scan(
		&p.ID, &p.PatientID, &p.VID, &p.Notes, &p.CreatedBy,
		&p.IsDispensed, &p.DispensedBy, &p.DispensedAt,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}

	// phase 1: load all drug_prescriptions (NO inner queries here)
	drugRows, err := dbx.QueryContext(ctx, `
        SELECT id, prescription_id, drug_id,
               remarks, quantity_requested,
               is_packed, packed_by, packed_at,
               created_at, updated_at
        FROM drug_prescriptions
        WHERE prescription_id = $1
        ORDER BY id
    `, id)
	if err != nil {
		return nil, err
	}

	var (
		drugs        = []entities.DrugPrescription{}
		packerIDsSet = map[int64]struct{}{}
	)
	for drugRows.Next() {
		var d entities.DrugPrescription
		if err := drugRows.Scan(&d.ID, &d.PrescriptionID, &d.DrugID,
			&d.Remarks, &d.RequestedQty,
			&d.IsPacked, &d.PackedBy, &d.PackedAt,
			&d.CreatedAt, &d.UpdatedAt); err != nil {
			drugRows.Close()
			return nil, err
		}
		if d.IsPacked && d.PackedBy != nil {
			packerIDsSet[*d.PackedBy] = struct{}{} // collect, don't query yet
		}
		drugs = append(drugs, d)
	}
	if err := drugRows.Err(); err != nil {
		drugRows.Close()
		return nil, err
	}
	drugRows.Close()

	// phase 2: load batches for each drug (still no inner queries inside loops)
	for i := range drugs {
		d := &drugs[i]
		d.Batches = []entities.PrescriptionBatchItem{}
		batchRows, err := dbx.QueryContext(ctx, `
            SELECT id, drug_prescription_id, drug_batch_location_id, quantity, created_at, updated_at
            FROM prescription_batch_items
            WHERE drug_prescription_id = $1
            ORDER BY id
        `, d.ID)
		if err != nil {
			return nil, err
		}

		for batchRows.Next() {
			var b entities.PrescriptionBatchItem
			var locID int64
			if err := batchRows.Scan(&b.ID, &b.DrugPrescriptionID, &locID, &b.Quantity, &b.CreatedAt, &b.UpdatedAt); err != nil {
				batchRows.Close()
				return nil, err
			}
			b.BatchLocationId = locID
			d.Batches = append(d.Batches, b)
		}
		if err := batchRows.Err(); err != nil {
			batchRows.Close()
			return nil, err
		}
		batchRows.Close()
	}

	// phase 3: resolve user names AFTER all rows are closed
	// You may use dbx (tx) here since no rows are open now.
	if p.CreatedBy != nil {
		u, err := getUserByID(dbx, *p.CreatedBy)
		if err != nil {
			return nil, err
		}
		p.CreatorName = &u.Name
	}
	if p.IsDispensed && p.DispensedBy != nil {
		u, err := getUserByID(dbx, *p.DispensedBy)
		if err != nil {
			return nil, err
		}
		p.DispenserName = &u.Name
	}

	if len(packerIDsSet) > 0 {
		// (Optional optimisation) fetch all packers in one query.
		// Or simpler: loop and call getUserByID sequentially.
		for i := range drugs {
			if drugs[i].IsPacked && drugs[i].PackedBy != nil {
				u, err := getUserByID(dbx, *drugs[i].PackedBy)
				if err != nil {
					return nil, err
				}
				drugs[i].PackerName = &u.Name
			}
		}
	}

	p.PrescribedDrugs = drugs
	return &p, nil
}

func (r *postgresPrescriptionRepository) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	// -------------------------
	// Phase 1: load prescriptions
	// -------------------------
	base := `
		SELECT id, patient_id, vid, notes, created_by,
		       is_dispensed, dispensed_by, dispensed_at,
		       created_at, updated_at
		FROM prescriptions`

	var (
		rows *sql.Rows
		err  error
	)
	switch {
	case patientID != nil && vid != nil:
		rows, err = dbx.QueryContext(ctx, base+` WHERE patient_id = $1 AND vid = $2 ORDER BY created_at DESC`, *patientID, *vid)
	case patientID != nil:
		rows, err = dbx.QueryContext(ctx, base+` WHERE patient_id = $1 ORDER BY created_at DESC`, *patientID)
	default:
		rows, err = dbx.QueryContext(ctx, base+` ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, err
	}

	var (
		result    = []*entities.Prescription{}
		prescIDs  = []int64{}
		prescByID = make(map[int64]*entities.Prescription)
	)
	for rows.Next() {
		p := new(entities.Prescription)
		if err := rows.Scan(
			&p.ID, &p.PatientID, &p.VID, &p.Notes, &p.CreatedBy,
			&p.IsDispensed, &p.DispensedBy, &p.DispensedAt,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}
		p.PrescribedDrugs = []entities.DrugPrescription{}
		result = append(result, p)
		prescIDs = append(prescIDs, p.ID)
		prescByID[p.ID] = p
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	if len(result) == 0 {
		return result, nil
	}

	// util: ($1,$2,...)
	inPlaceholders := func(n int, startIdx int) string {
		var b strings.Builder
		b.WriteByte('(')
		for i := 0; i < n; i++ {
			if i > 0 {
				b.WriteByte(',')
			}
			fmt.Fprintf(&b, "$%d", startIdx+i)
		}
		b.WriteByte(')')
		return b.String()
	}

	// -------------------------
	// Phase 2: load drug_prescriptions in one go (no inner queries!)
	// -------------------------
	dpArgs := make([]any, len(prescIDs))
	for i, id := range prescIDs {
		dpArgs[i] = id
	}
	dpQuery := `
		SELECT id, prescription_id, drug_id,
		       remarks, quantity_requested,
		       is_packed, packed_by, packed_at,
		       created_at, updated_at
		FROM drug_prescriptions
		WHERE prescription_id IN ` + inPlaceholders(len(prescIDs), 1) + `
		ORDER BY prescription_id, id`

	dpRows, err := dbx.QueryContext(ctx, dpQuery, dpArgs...)
	if err != nil {
		return nil, err
	}

	// collect for phase 3/4
	dpByID := make(map[int64]*entities.DrugPrescription)
	packerIDsSet := make(map[int64]struct{})

	for dpRows.Next() {
		var d entities.DrugPrescription
		if err := dpRows.Scan(
			&d.ID, &d.PrescriptionID, &d.DrugID,
			&d.Remarks, &d.RequestedQty,
			&d.IsPacked, &d.PackedBy, &d.PackedAt,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			dpRows.Close()
			return nil, err
		}
		d.Batches = []entities.PrescriptionBatchItem{}
		parent := prescByID[d.PrescriptionID]
		if parent != nil {
			parent.PrescribedDrugs = append(parent.PrescribedDrugs, d)
			dpByID[d.ID] = &parent.PrescribedDrugs[len(parent.PrescribedDrugs)-1]
			if d.IsPacked && d.PackedBy != nil {
				packerIDsSet[*d.PackedBy] = struct{}{} // collect; don't query yet
			}
		}
	}
	if err := dpRows.Err(); err != nil {
		dpRows.Close()
		return nil, err
	}
	dpRows.Close()

	if len(dpByID) == 0 {
		// No line items; resolve parent user names and return.
		for _, p := range result {
			if p.CreatedBy != nil {
				u, err := getUserByID(dbx, *p.CreatedBy)
				if err != nil {
					return nil, err
				}
				p.CreatorName = &u.Name
			}
			if p.DispensedBy != nil {
				u, err := getUserByID(dbx, *p.DispensedBy)
				if err != nil {
					return nil, err
				}
				p.DispenserName = &u.Name
			}
		}
		return result, nil
	}

	// -------------------------
	// Phase 3: load all batch items in one go
	// -------------------------
	dpIDs := make([]int64, 0, len(dpByID))
	for id := range dpByID {
		dpIDs = append(dpIDs, id)
	}
	biArgs := make([]any, len(dpIDs))
	for i, id := range dpIDs {
		biArgs[i] = id
	}
	biQuery := `
		SELECT id, drug_prescription_id, drug_batch_location_id, quantity, created_at, updated_at
		FROM prescription_batch_items
		WHERE drug_prescription_id IN ` + inPlaceholders(len(dpIDs), 1) + `
		ORDER BY drug_prescription_id, id`

	biRows, err := dbx.QueryContext(ctx, biQuery, biArgs...)
	if err != nil {
		return nil, err
	}
	for biRows.Next() {
		var b entities.PrescriptionBatchItem
		var locID int64
		if err := biRows.Scan(&b.ID, &b.DrugPrescriptionID, &locID, &b.Quantity, &b.CreatedAt, &b.UpdatedAt); err != nil {
			biRows.Close()
			return nil, err
		}
		b.BatchLocationId = locID
		if dp := dpByID[b.DrugPrescriptionID]; dp != nil {
			dp.Batches = append(dp.Batches, b)
		}
	}
	if err := biRows.Err(); err != nil {
		biRows.Close()
		return nil, err
	}
	biRows.Close()

	// -------------------------
	// Phase 4: resolve user names AFTER all rows are closed
	// -------------------------
	// parent creator/dispensers (as before)
	for _, p := range result {
		if p.CreatedBy != nil {
			u, err := getUserByID(dbx, *p.CreatedBy)
			if err != nil {
				return nil, err
			}
			p.CreatorName = &u.Name
		}
		if p.DispensedBy != nil {
			u, err := getUserByID(dbx, *p.DispensedBy)
			if err != nil {
				return nil, err
			}
			p.DispenserName = &u.Name
		}
	}

	// packers: bulk resolve to avoid N+1 (optional but nice)
	if len(packerIDsSet) > 0 {
		packerIDs := make([]int64, 0, len(packerIDsSet))
		for id := range packerIDsSet {
			packerIDs = append(packerIDs, id)
		}

		// build args
		args := make([]any, len(packerIDs))
		for i, id := range packerIDs {
			args[i] = id
		}

		// id -> name map
		packerByID := make(map[int64]string)
		q := `SELECT id, name FROM users WHERE id IN ` + inPlaceholders(len(packerIDs), 1)
		uRows, err := dbx.QueryContext(ctx, q, args...)
		if err != nil {
			return nil, err
		}
		for uRows.Next() {
			var id int64
			var name string
			if err := uRows.Scan(&id, &name); err != nil {
				uRows.Close()
				return nil, err
			}
			packerByID[id] = name
		}
		if err := uRows.Err(); err != nil {
			uRows.Close()
			return nil, err
		}
		uRows.Close()

		// assign names
		for _, p := range result {
			for i := range p.PrescribedDrugs {
				d := &p.PrescribedDrugs[i]
				if d.IsPacked && d.PackedBy != nil {
					if name, ok := packerByID[*d.PackedBy]; ok {
						d.PackerName = &name
					}
				}
			}
		}
	}

	return result, nil
}

func (r *postgresPrescriptionRepository) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	tx, ok := TxFromCtx(ctx)
	ownTx := false

	if !ok {
		var err error
		tx, err = r.Conn.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		ownTx = true
		defer tx.Rollback()
	}

	// --- Step 0: Check if the prescription has already been dispensed ---
	var isDispensed bool
	err := tx.QueryRowContext(ctx, `
		SELECT is_dispensed FROM prescriptions
		WHERE id = $1
	`, p.ID).Scan(&isDispensed)
	if err != nil {
		return nil, err
	}

	// Record the old workflow states of the prescription and drug batches to update
	dispenseNow := (!isDispensed) && p.IsDispensed // transition false->true?
	prevPackedByDrug := map[int64]bool{}
	rowsPrev, err := tx.QueryContext(ctx, `
		SELECT dp.drug_id, bool_or(dp.is_packed) AS was_packed
		FROM drug_prescriptions dp
		WHERE dp.prescription_id = $1
		GROUP BY dp.drug_id
	`, p.ID)
	if err != nil {
		return nil, err
	}
	for rowsPrev.Next() {
		var drugID int64
		var wasPacked bool
		if err := rowsPrev.Scan(&drugID, &wasPacked); err != nil {
			return nil, err
		}
		prevPackedByDrug[drugID] = wasPacked
	}
	if err := rowsPrev.Err(); err != nil {
		return nil, err
	}
	rowsPrev.Close()
	if isDispensed {
		return nil, errors.New("cannot update a dispensed prescription")
	}

	// --- Step 1: Load existing per-batch totals INSIDE this txn ---
	oldTotals := make(map[int64]int64)
	rows, err := tx.QueryContext(ctx, `
		SELECT pbi.drug_batch_location_id, COALESCE(SUM(pbi.quantity), 0)
		FROM drug_prescriptions dp
		JOIN prescription_batch_items pbi ON pbi.drug_prescription_id = dp.id
		WHERE dp.prescription_id = $1
		GROUP BY pbi.drug_batch_location_id
	`, p.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var batchLocationID, qty int64
		if err := rows.Scan(&batchLocationID, &qty); err != nil {
			return nil, err
		}
		oldTotals[batchLocationID] = qty
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	// --- Step 2: Build new per-batchLocation totals from payload ---
	newTotals := make(map[int64]int64)
	for i := range p.PrescribedDrugs {
		for j := range p.PrescribedDrugs[i].Batches {
			b := p.PrescribedDrugs[i].Batches[j]
			newTotals[b.BatchLocationId] += int64(b.Quantity)
		}
	}

	// --- Step 3: Apply per-batch delta (new - old). Positive = take stock; Negative = return stock ---
	keys := map[int64]struct{}{}
	for k := range oldTotals {
		keys[k] = struct{}{}
	}
	for k := range newTotals {
		keys[k] = struct{}{}
	}

	for batchLocationID := range keys {
		oldQ := oldTotals[batchLocationID]
		newQ := newTotals[batchLocationID]
		delta := newQ - oldQ
		if delta == 0 {
			continue
		}

		if delta > 0 {
			// Need more from this batch
			res, err := tx.ExecContext(ctx, `
				UPDATE batch_locations
				SET quantity = quantity - $1
				WHERE id = $2 AND quantity >= $1
			`, delta, batchLocationID)
			if err != nil {
				return nil, err
			}
			if rows, _ := res.RowsAffected(); rows == 0 {
				return nil, fmt.Errorf("insufficient stock in batch %d", batchLocationID)
			}
		} else {
			// Return surplus to this batch
			if _, err := tx.ExecContext(ctx, `
				UPDATE batch_locations
				SET quantity = quantity + $1
				WHERE id = $2
			`, -delta, batchLocationID); err != nil {
				return nil, err
			}
		}
	}

	// --- Step 4: Delete old batch items and drug prescriptions (no stock math here) ---
	_, err = tx.ExecContext(ctx, `
		DELETE FROM prescription_batch_items
		WHERE drug_prescription_id IN (SELECT id FROM drug_prescriptions WHERE prescription_id = $1)
	`, p.ID)
	if err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM drug_prescriptions WHERE prescription_id = $1`, p.ID)
	if err != nil {
		return nil, err
	}

	// --- Step 5: Update prescription metadata ---
	if _, err := tx.ExecContext(ctx, `
		UPDATE prescriptions pr
		SET
			patient_id     = $2,
			vid            = $3,
			notes          = $4,
			updated_at     = now(),
			is_dispensed   = $5,
			dispensed_by   = CASE
				WHEN $6 THEN current_setting('sothea.user_id')::BIGINT
				ELSE pr.dispensed_by
			END,
			dispensed_at   = CASE
				WHEN $6 THEN now()
				ELSE pr.dispensed_at
			END
		WHERE pr.id = $1
	`, p.ID, p.PatientID, p.VID, p.Notes, p.IsDispensed, dispenseNow); err != nil {
		return nil, err
	}

	// --- Step 6: Insert new prescribed drugs and batch items (DO NOT touch stock here) ---
	for i := range p.PrescribedDrugs {
		d := &p.PrescribedDrugs[i]

		wasPacked := prevPackedByDrug[d.DrugID]
		stampPack := (!wasPacked) && d.IsPacked

		// stampPack := (!wasPacked) && d.IsPacked
		err = tx.QueryRowContext(ctx, `
		INSERT INTO drug_prescriptions (
			prescription_id,
			drug_id,
			remarks,
			quantity_requested,
			is_packed,
			packed_by,
			packed_at
		)
		VALUES (
			$1, $2, $3, $4, $5,
			CASE WHEN $6 THEN current_setting('sothea.user_id')::BIGINT ELSE NULL END,
			CASE WHEN $6 THEN now() ELSE NULL END
		)
		RETURNING id, created_at, updated_at
		`, p.ID, d.DrugID, d.Remarks, d.RequestedQty, d.IsPacked, stampPack).
			Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}

		for j := range d.Batches {
			b := &d.Batches[j]

			err = tx.QueryRowContext(ctx, `
				INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
				VALUES ($1, $2, $3)
				RETURNING id, created_at, updated_at
			`, d.ID, b.BatchLocationId, b.Quantity).
				Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
			if err != nil {
				return nil, err
			}
		}
	}

	// --- Step 7: Commit ---
	if ownTx {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return r.GetPrescriptionByID(context.Background(), p.ID)
	}

	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) DeletePrescription(ctx context.Context, id int64) error {
	tx, ok := TxFromCtx(ctx)
	ownTx := false

	if !ok {
		var err error
		tx, err = r.Conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		ownTx = true
		defer tx.Rollback()
	}

	// 1) Collect current allocations (per batch) INSIDE this txn
	type alloc struct {
		batchID int64
		qty     int64
	}
	var allocs []alloc

	rows, err := tx.QueryContext(ctx, `
		SELECT pbi.drug_batch_location_id, COALESCE(SUM(pbi.quantity), 0) AS qty
		FROM drug_prescriptions dp
		JOIN prescription_batch_items pbi ON pbi.drug_prescription_id = dp.id
		WHERE dp.prescription_id = $1
		GROUP BY pbi.drug_batch_location_id
	`, id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var a alloc
		if err := rows.Scan(&a.batchID, &a.qty); err != nil {
			return err
		}
		allocs = append(allocs, a)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	// 2) Restock batches
	for _, a := range allocs {
		if a.qty == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE batch_locations
			SET quantity = quantity + $1
			WHERE id = $2
		`, a.qty, a.batchID); err != nil {
			return err
		}
	}

	// 3) Delete children, then parent
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM prescription_batch_items
		WHERE drug_prescription_id IN (
			SELECT id FROM drug_prescriptions WHERE prescription_id = $1
		)
	`, id); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM drug_prescriptions
		WHERE prescription_id = $1
	`, id); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM prescriptions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("prescription not found")
	}

	// 4) Commit
	if ownTx {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
