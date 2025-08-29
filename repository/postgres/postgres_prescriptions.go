package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	tx, err := r.Conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	fmt.Println("HELLO")
	fmt.Println(p.PatientID)
	fmt.Println(p.VID)

	err = tx.QueryRowContext(ctx, `
		INSERT INTO prescriptions (patient_id, vid, staff_id, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, p.PatientID, p.VID, p.StaffID, p.Notes).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	for i := range p.PrescribedDrugs {
		d := &p.PrescribedDrugs[i]

		err = tx.QueryRowContext(ctx, `
			INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at
		`, p.ID, d.DrugID, d.Quantity, d.Remarks).
			Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}

		for j := range d.Batches {
			b := &d.Batches[j]

			err = tx.QueryRowContext(ctx, `
				INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
				VALUES ($1, $2, $3)
				RETURNING id, created_at, updated_at
			`, d.ID, b.BatchId, b.Quantity).
				Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
			if err != nil {
				return nil, err
			}

			res, err := tx.ExecContext(ctx, `
				UPDATE drug_batches
				SET quantity = quantity - $1
				WHERE id = $2 AND quantity >= $1
			`, b.Quantity, b.BatchId)
			if err != nil {
				return nil, err
			}
			rows, _ := res.RowsAffected()
			if rows == 0 {
				return nil, fmt.Errorf("insufficient stock in batch %d", b.BatchId)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	var p entities.Prescription

	fmt.Println("HELLO1")
	fmt.Println(p.PrescribedDrugs)
	err := r.Conn.QueryRowContext(ctx, `
		SELECT id, patient_id, vid, staff_id, notes, created_at, updated_at
		FROM prescriptions
		WHERE id = $1
	`, id).Scan(&p.ID, &p.PatientID, &p.VID, &p.StaffID, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	fmt.Println("HELLO")
	fmt.Println(p.PrescribedDrugs)

	drugRows, err := r.Conn.QueryContext(ctx, `
		SELECT id, prescription_id, drug_id, quantity, remarks, created_at, updated_at
		FROM drug_prescriptions
		WHERE prescription_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer drugRows.Close()

	for drugRows.Next() {
		var d entities.DrugPrescription
		err := drugRows.Scan(&d.ID, &d.PrescriptionID, &d.DrugID, &d.Quantity, &d.Remarks, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}

		batchRows, err := r.Conn.QueryContext(ctx, `
			SELECT id, drug_prescription_id, drug_batch_id, quantity, created_at, updated_at
			FROM prescription_batch_items
			WHERE drug_prescription_id = $1
		`, d.ID)
		if err != nil {
			return nil, err
		}
		defer batchRows.Close()

		for batchRows.Next() {
			var b entities.PrescriptionBatchItem
			var drugBatchID int64
			err := batchRows.Scan(&b.ID, &b.DrugPrescriptionID, &drugBatchID, &b.Quantity, &b.CreatedAt, &b.UpdatedAt)
			if err != nil {
				return nil, err
			}
			b.BatchId = drugBatchID
			d.Batches = append(d.Batches, b)
		}
		p.PrescribedDrugs = append(p.PrescribedDrugs, d)
	}

	return &p, nil
}

func (r *postgresPrescriptionRepository) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	query := `
		SELECT id, patient_id, vid, staff_id, notes, created_at, updated_at
		FROM prescriptions`

	var rows *sql.Rows
	var err error

	switch {
	case patientID != nil && vid != nil:
		query += ` WHERE patient_id = $1 AND vid = $2 ORDER BY created_at DESC`
		rows, err = r.Conn.QueryContext(ctx, query, *patientID, *vid)
	case patientID != nil:
		query += ` WHERE patient_id = $1 ORDER BY created_at DESC`
		rows, err = r.Conn.QueryContext(ctx, query, *patientID)
	default:
		query += ` ORDER BY created_at DESC`
		rows, err = r.Conn.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*entities.Prescription
	for rows.Next() {
		var p entities.Prescription
		err := rows.Scan(&p.ID, &p.PatientID, &p.VID, &p.StaffID, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Hydrate prescribed drugs
		drugRows, err := r.Conn.QueryContext(ctx, `
			SELECT id, prescription_id, drug_id, quantity, remarks, created_at, updated_at
			FROM drug_prescriptions
			WHERE prescription_id = $1
		`, p.ID)
		if err != nil {
			return nil, err
		}

		var prescribedDrugs []entities.DrugPrescription
		for drugRows.Next() {
			var d entities.DrugPrescription
			err = drugRows.Scan(&d.ID, &d.PrescriptionID, &d.DrugID, &d.Quantity, &d.Remarks, &d.CreatedAt, &d.UpdatedAt)
			if err != nil {
				drugRows.Close()
				return nil, err
			}

			// Hydrate prescription batch items
			batchRows, err := r.Conn.QueryContext(ctx, `
				SELECT id, drug_prescription_id, drug_batch_id, quantity, created_at, updated_at
				FROM prescription_batch_items
				WHERE drug_prescription_id = $1
			`, d.ID)
			if err != nil {
				drugRows.Close()
				return nil, err
			}

			var batches []entities.PrescriptionBatchItem
			for batchRows.Next() {
				var b entities.PrescriptionBatchItem
				var drugBatchID int64
				err = batchRows.Scan(&b.ID, &b.DrugPrescriptionID, &drugBatchID, &b.Quantity, &b.CreatedAt, &b.UpdatedAt)
				if err != nil {
					batchRows.Close()
					drugRows.Close()
					return nil, err
				}
				b.BatchId = drugBatchID
				batches = append(batches, b)
			}
			batchRows.Close()

			d.Batches = batches
			prescribedDrugs = append(prescribedDrugs, d)
		}
		drugRows.Close()

		p.PrescribedDrugs = prescribedDrugs
		result = append(result, &p)
	}

	return result, nil
}

func (r *postgresPrescriptionRepository) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	tx, err := r.Conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// --- Step 1: Load existing per-batch totals INSIDE this txn ---
	oldTotals := make(map[int64]int64)
	rows, err := tx.QueryContext(ctx, `
		SELECT pbi.drug_batch_id, COALESCE(SUM(pbi.quantity), 0)
		FROM drug_prescriptions dp
		JOIN prescription_batch_items pbi ON pbi.drug_prescription_id = dp.id
		WHERE dp.prescription_id = $1
		GROUP BY pbi.drug_batch_id
	`, p.ID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var batchID, qty int64
		if err := rows.Scan(&batchID, &qty); err != nil {
			return nil, err
		}
		oldTotals[batchID] = qty
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	// --- Step 2: Build new per-batch totals from payload ---
	newTotals := make(map[int64]int64)
	for i := range p.PrescribedDrugs {
		for j := range p.PrescribedDrugs[i].Batches {
			b := p.PrescribedDrugs[i].Batches[j]
			newTotals[b.BatchId] += int64(b.Quantity)
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

	for batchID := range keys {
		oldQ := oldTotals[batchID]
		newQ := newTotals[batchID]
		delta := newQ - oldQ
		if delta == 0 {
			continue
		}

		if delta > 0 {
			// Need more from this batch
			res, err := tx.ExecContext(ctx, `
				UPDATE drug_batches
				SET quantity = quantity - $1
				WHERE id = $2 AND quantity >= $1
			`, delta, batchID)
			if err != nil {
				return nil, err
			}
			if rows, _ := res.RowsAffected(); rows == 0 {
				return nil, fmt.Errorf("insufficient stock in batch %d", batchID)
			}
		} else {
			// Return surplus to this batch
			if _, err := tx.ExecContext(ctx, `
				UPDATE drug_batches
				SET quantity = quantity + $1
				WHERE id = $2
			`, -delta, batchID); err != nil {
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
	_, err = tx.ExecContext(ctx, `
		UPDATE prescriptions
		SET patient_id = $2, vid = $3, staff_id = $4, notes = $5, updated_at = now()
		WHERE id = $1
	`, p.ID, p.PatientID, p.VID, p.StaffID, p.Notes)
	if err != nil {
		return nil, err
	}

	// --- Step 6: Insert new prescribed drugs and batch items (DO NOT touch stock here) ---
	for i := range p.PrescribedDrugs {
		d := &p.PrescribedDrugs[i]

		err = tx.QueryRowContext(ctx, `
			INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at
		`, p.ID, d.DrugID, d.Quantity, d.Remarks).
			Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, err
		}

		for j := range d.Batches {
			b := &d.Batches[j]

			err = tx.QueryRowContext(ctx, `
				INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
				VALUES ($1, $2, $3)
				RETURNING id, created_at, updated_at
			`, d.ID, b.BatchId, b.Quantity).
				Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
			if err != nil {
				return nil, err
			}
		}
	}

	// --- Step 7: Commit ---
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) DeletePrescription(ctx context.Context, id int64) error {
	tx, err := r.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1) Collect current allocations (per batch) INSIDE this txn
	type alloc struct {
		batchID int64
		qty     int64
	}
	var allocs []alloc

	rows, err := tx.QueryContext(ctx, `
		SELECT pbi.drug_batch_id, COALESCE(SUM(pbi.quantity), 0) AS qty
		FROM drug_prescriptions dp
		JOIN prescription_batch_items pbi ON pbi.drug_prescription_id = dp.id
		WHERE dp.prescription_id = $1
		GROUP BY pbi.drug_batch_id
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
			UPDATE drug_batches
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
	return tx.Commit()
}
