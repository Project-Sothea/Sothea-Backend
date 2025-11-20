package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// ----------------------------------------------------------------------------
// REPO
// ----------------------------------------------------------------------------

type postgresPrescriptionRepository struct {
	Conn *sql.DB
}

func NewPostgresPrescriptionRepository(conn *sql.DB) entities.PrescriptionRepository {
	return &postgresPrescriptionRepository{Conn: conn}
}

func withTx(ctx context.Context, db *sql.DB) (*sql.Tx, bool, error) {
	if tx, ok := TxFromCtx(ctx); ok {
		return tx, false, nil
	}
	tx, err := db.BeginTx(ctx, nil)
	return tx, true, err
}

// ----------------------------------------------------------------------------
// PRESCRIPTIONS (header)
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) CreatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO prescriptions (patient_id, vid, notes)
		VALUES ($1,$2,$3)
		RETURNING id, created_at, updated_at
	`, p.PatientID, p.VID, p.Notes).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var p entities.Prescription
	var creatorName, dispenserName sql.NullString

	// Header
	if err := dbx.QueryRowContext(ctx, `
		SELECT p.id, p.patient_id, p.vid, p.notes,
				p.created_by, p.created_at, p.updated_at,
				p.is_dispensed, p.dispensed_by, p.dispensed_at,
				uc.name AS creator_name,
				ud.name AS dispenser_name
			FROM prescriptions p
			LEFT JOIN users uc ON uc.id = p.created_by
			LEFT JOIN users ud ON ud.id = p.dispensed_by
		WHERE p.id = $1
	`, id).Scan(
		&p.ID, &p.PatientID, &p.VID, &p.Notes,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
		&p.IsDispensed, &p.DispensedBy, &p.DispensedAt,
		&creatorName, &dispenserName,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("prescription not found")
		}
		return nil, err
	}

	if creatorName.Valid {
		p.CreatorName = &creatorName.String
	}
	if dispenserName.Valid {
		p.DispenserName = &dispenserName.String
	}

	// Lines (new schedule fields)
	rows, err := dbx.QueryContext(ctx, `
	SELECT
		pl.id, pl.prescription_id, pl.presentation_id, pl.remarks,
		pl.dose_amount, pl.dose_unit,
		pl.schedule_kind, pl.every_n, pl.frequency_per_schedule,
		pl.duration, pl.duration_unit,
		pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
		u1.name AS packer_name,
		u2.name AS updater_name,
		d.generic_name AS drug_name,
		dp.route_code AS route_code,
		dp.dispense_unit AS dispense_unit,
		CASE
		WHEN dp.strength_den IS NULL THEN
			dp.strength_num::text || ' ' || dp.strength_unit_num || '/' || dp.dispense_unit
		ELSE
			dp.strength_num::text || ' ' || dp.strength_unit_num || '/' ||
			dp.strength_den::text || ' ' || dp.strength_unit_den
		END AS display_strength
	FROM prescription_lines pl
	LEFT JOIN drug_presentations dp ON dp.id = pl.presentation_id
	LEFT JOIN drugs d ON d.id = dp.drug_id
	LEFT JOIN users u1 ON u1.id = pl.packed_by
	LEFT JOIN users u2 ON u2.id = pl.updated_by
	WHERE pl.prescription_id = $1
	ORDER BY pl.id
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := []entities.PrescriptionLine{}
	for rows.Next() {
		var l entities.PrescriptionLine
		var packerName sql.NullString
		var updaterName sql.NullString
		if err := rows.Scan(
			&l.ID, &l.PrescriptionID, &l.PresentationID, &l.Remarks,
			&l.DoseAmount, &l.DoseUnit,
			&l.ScheduleKind, &l.EveryN, &l.FrequencyPerSchedule,
			&l.Duration, &l.DurationUnit,
			&l.TotalToDispense, &l.IsPacked, &l.PackedBy, &l.PackedAt,
			&packerName, &updaterName,
			&l.DrugName, &l.DisplayRoute, &l.DispenseUnit, &l.DisplayStrength,
		); err != nil {
			return nil, err
		}
		if packerName.Valid {
			l.PackerName = &packerName.String
		}
		if updaterName.Valid {
			l.UpdaterName = &updaterName.String
		}
		lines = append(lines, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Allocations per line
	if len(lines) > 0 {
		lineIDs := make([]any, 0, len(lines))
		index := map[int64]*entities.PrescriptionLine{}
		for i := range lines {
			lineIDs = append(lineIDs, lines[i].ID)
			index[lines[i].ID] = &lines[i]
		}

		// Build IN list ($1,$2,...)
		in := "("
		for i := range lineIDs {
			if i > 0 {
				in += ","
			}
			in += fmt.Sprintf("$%d", i+1)
		}
		in += ")"

		allocRows, err := dbx.QueryContext(ctx, `
			SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
			FROM prescription_batch_items
			WHERE line_id IN `+in+`
			ORDER BY line_id, id
		`, lineIDs...)
		if err != nil {
			return nil, err
		}
		defer allocRows.Close()

		for allocRows.Next() {
			var a entities.LineAllocation
			if err := allocRows.Scan(&a.ID, &a.LineID, &a.BatchLocationID, &a.Quantity, &a.CreatedAt, &a.UpdatedAt); err != nil {
				return nil, err
			}
			if ln := index[a.LineID]; ln != nil {
				ln.Allocations = append(ln.Allocations, a)
			}
		}
		if err := allocRows.Err(); err != nil {
			return nil, err
		}
	}

	p.Lines = lines
	return &p, nil
}

func (r *postgresPrescriptionRepository) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	base := `
	  SELECT id, patient_id, vid, notes,
	         created_by, created_at, updated_at,
	         is_dispensed, dispensed_by, dispensed_at
	  FROM prescriptions`
	var (
		rows *sql.Rows
		err  error
	)
	switch {
	case patientID != nil && vid != nil:
		rows, err = dbx.QueryContext(ctx, base+` WHERE patient_id=$1 AND vid=$2 ORDER BY created_at DESC`, *patientID, *vid)
	case patientID != nil:
		rows, err = dbx.QueryContext(ctx, base+` WHERE patient_id=$1 ORDER BY created_at DESC`, *patientID)
	default:
		rows, err = dbx.QueryContext(ctx, base+` ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*entities.Prescription, 0)
	for rows.Next() {
		p := new(entities.Prescription)
		if err := rows.Scan(
			&p.ID, &p.PatientID, &p.VID, &p.Notes,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt,
			&p.IsDispensed, &p.DispensedBy, &p.DispensedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *postgresPrescriptionRepository) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, p.ID).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot modify a dispensed prescription")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE prescriptions
		SET patient_id=$2, vid=$3, notes=$4, updated_at=now()
		WHERE id=$1
	`, p.ID, p.PatientID, p.VID, p.Notes)
	if err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) DeletePrescription(ctx context.Context, id int64) error {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, id).Scan(&isDispensed); err != nil {
		return err
	}
	if isDispensed {
		return errors.New("cannot delete a dispensed prescription")
	}

	// Delete children then parent.
	// NOTE: stock reservations are released by DB triggers on DELETE of prescription_batch_items.
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM prescription_batch_items WHERE line_id IN (SELECT id FROM prescription_lines WHERE prescription_id=$1)
	`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM prescription_lines WHERE prescription_id=$1`, id); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM prescriptions WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return errors.New("prescription not found")
	}

	if own {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// LINES (one presentation per line)
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) AddLine(ctx context.Context, line *entities.PrescriptionLine) (*entities.PrescriptionLine, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	// Cannot add to dispensed Rx
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, line.PrescriptionID).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot add line to a dispensed prescription")
	}

	err = tx.QueryRowContext(ctx, `
	  INSERT INTO prescription_lines (
	    prescription_id, presentation_id, remarks,
	    dose_amount, dose_unit,
	    schedule_kind, every_n, frequency_per_schedule,
		duration, duration_unit
	  )
	  VALUES (
	  	$1,$2,$3,
		$4,$5,
		$6,$7,$8,
		$9,$10)
	  RETURNING id, total_to_dispense, is_packed
	`, line.PrescriptionID, line.PresentationID, line.Remarks,
		line.DoseAmount, line.DoseUnit,
		line.ScheduleKind, line.EveryN, line.FrequencyPerSchedule,
		line.Duration, line.DurationUnit).
		Scan(&line.ID, &line.TotalToDispense, &line.IsPacked)
	if err != nil {
		return nil, mapPrescriptionSQLError(err)
	}

	// Return enriched line
	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetLine(ctx, line.ID)
}

func (r *postgresPrescriptionRepository) UpdateLine(ctx context.Context, line *entities.PrescriptionLine) (*entities.PrescriptionLine, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	// Forbid updates on dispensed Rx
	var pid int64
	if err := tx.QueryRowContext(ctx, `SELECT prescription_id FROM prescription_lines WHERE id=$1`, line.ID).Scan(&pid); err != nil {
		return nil, err
	}
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, pid).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot modify a line on a dispensed prescription")
	}

	// Load current values to detect what changed (lock the row for the rest of the tx)
	var cur entities.PrescriptionLine
	if err := tx.QueryRowContext(ctx, `
		SELECT presentation_id,
		dose_amount, dose_unit,
		schedule_kind, every_n, frequency_per_schedule,
		duration, duration_unit,
		remarks
		FROM prescription_lines
		WHERE id=$1 FOR UPDATE
	`, line.ID).Scan(
		&cur.PresentationID,
		&cur.DoseAmount, &cur.DoseUnit,
		&cur.ScheduleKind, &cur.EveryN, &cur.FrequencyPerSchedule,
		&cur.Duration, &cur.DurationUnit,
		&cur.Remarks,
	); err != nil {
		return nil, err
	}

	presChanged := cur.PresentationID != line.PresentationID

	// 1) Only clear allocations if the presentation changed
	if presChanged {
		if _, err := tx.ExecContext(ctx, `DELETE FROM prescription_batch_items WHERE line_id=$1`, line.ID); err != nil {
			return nil, err
		}
	}

	err = tx.QueryRowContext(ctx, `
	  UPDATE prescription_lines SET
	    presentation_id=$2, remarks=$3,
	    dose_amount=$4, dose_unit=$5,
	    schedule_kind=$6, every_n=$7, frequency_per_schedule=$8,
		duration=$9, duration_unit=$10,
	    is_packed=FALSE, packed_by=NULL, packed_at=NULL,
	    updated_at=NOW()
	  WHERE id=$1
	  RETURNING total_to_dispense
	`, line.ID, line.PresentationID, line.Remarks,
		line.DoseAmount, line.DoseUnit,
		line.ScheduleKind, line.EveryN, line.FrequencyPerSchedule,
		line.Duration, line.DurationUnit).
		Scan(&line.TotalToDispense)
	if err != nil {
		return nil, mapPrescriptionSQLError(err)
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetLine(ctx, line.ID)
}

func (r *postgresPrescriptionRepository) RemoveLine(ctx context.Context, lineID int64) error {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	var pid int64
	if err := tx.QueryRowContext(ctx, `SELECT prescription_id FROM prescription_lines WHERE id=$1`, lineID).Scan(&pid); err != nil {
		return err
	}
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, pid).Scan(&isDispensed); err != nil {
		return err
	}
	if isDispensed {
		return errors.New("cannot remove a line from a dispensed prescription")
	}

	// Triggers on DELETE of prescription_batch_items will release reserved stock.
	if _, err := tx.ExecContext(ctx, `DELETE FROM prescription_batch_items WHERE line_id=$1`, lineID); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM prescription_lines WHERE id=$1`, lineID)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return errors.New("line not found")
	}

	if own {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Allocations (packing plan) - replace all
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) ListLineAllocations(ctx context.Context, lineID int64) ([]entities.LineAllocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	rows, err := dbx.QueryContext(ctx, `
	  SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
	  FROM prescription_batch_items
	  WHERE line_id=$1 ORDER BY id
	`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]entities.LineAllocation, 0)
	for rows.Next() {
		var a entities.LineAllocation
		if err := rows.Scan(&a.ID, &a.LineID, &a.BatchLocationID, &a.Quantity, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *postgresPrescriptionRepository) SetLineAllocations(ctx context.Context, lineID int64, allocs []entities.LineAllocation) ([]entities.LineAllocation, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	// Forbid if dispensed
	var pid int64
	if err := tx.QueryRowContext(ctx, `SELECT prescription_id FROM prescription_lines WHERE id=$1`, lineID).Scan(&pid); err != nil {
		return nil, err
	}
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, pid).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot change allocations on a dispensed prescription")
	}

	// Replace all; triggers will adjust reservations per-row (delete→return stock, insert→reserve stock)
	if _, err := tx.ExecContext(ctx, `DELETE FROM prescription_batch_items WHERE line_id=$1`, lineID); err != nil {
		return nil, err
	}
	if len(allocs) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
		  INSERT INTO prescription_batch_items (line_id, batch_location_id, quantity)
		  VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		for i := range allocs {
			var id int64
			var ca, ua time.Time
			if err := stmt.QueryRowContext(ctx, lineID, allocs[i].BatchLocationID, allocs[i].Quantity).Scan(&id, &ca, &ua); err != nil {
				return nil, err
			}
			allocs[i].ID = id
			allocs[i].LineID = lineID
			allocs[i].CreatedAt = ca
			allocs[i].UpdatedAt = ua
		}
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.ListLineAllocations(ctx, lineID)
}

// ----------------------------------------------------------------------------
// Pack / Unpack flags (no stock mutation here)
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) MarkLinePacked(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	// Forbid on dispensed
	var pid int64
	if err := tx.QueryRowContext(ctx, `SELECT prescription_id FROM prescription_lines WHERE id=$1`, lineID).Scan(&pid); err != nil {
		return nil, err
	}
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, pid).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot pack a line on a dispensed prescription")
	}

	// Stamp packed fields
	if _, err := tx.ExecContext(ctx, `
	  UPDATE prescription_lines
	  SET is_packed=TRUE,
	  packed_by = current_setting('sothea.user_id')::bigint,
	  packed_at=NOW()
	  WHERE id=$1
	`, lineID); err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetLine(ctx, lineID)
}

func (r *postgresPrescriptionRepository) UnpackLine(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	// Forbid on dispensed
	var pid int64
	if err := tx.QueryRowContext(ctx, `SELECT prescription_id FROM prescription_lines WHERE id=$1`, lineID).Scan(&pid); err != nil {
		return nil, err
	}
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1`, pid).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot unpack a line on a dispensed prescription")
	}

	if _, err := tx.ExecContext(ctx, `
	  UPDATE prescription_lines
	  SET is_packed=FALSE, packed_by=NULL, packed_at=NULL, updated_at=NOW()
	  WHERE id=$1
	`, lineID); err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetLine(ctx, lineID)
}

// ----------------------------------------------------------------------------
// DISPENSE (atomic: verify packed; stamp dispensed)
// Stock is already reserved by triggers on prescription_batch_items.
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) DispensePrescription(ctx context.Context, prescriptionID int64) (*entities.Prescription, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback()
		}
	}()

	// Guard: not already dispensed
	var isDispensed bool
	if err := tx.QueryRowContext(ctx, `SELECT is_dispensed FROM prescriptions WHERE id=$1 FOR UPDATE`, prescriptionID).Scan(&isDispensed); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("prescription already dispensed")
	}

	// Must have lines and all must be packed
	var totalLines, packedLines int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1`, prescriptionID).Scan(&totalLines); err != nil {
		return nil, err
	}
	if totalLines == 0 {
		return nil, errors.New("no lines to dispense")
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1 AND is_packed=TRUE`, prescriptionID).Scan(&packedLines); err != nil {
		return nil, err
	}
	if totalLines != packedLines {
		return nil, errors.New("all lines must be packed before dispense")
	}

	/*
		// Safety: totals vs allocations
		var mismatch int
		if err := tx.QueryRowContext(ctx, `
		  SELECT COUNT(*) FROM prescription_lines l
		  WHERE l.prescription_id=$1 AND
		        l.total_to_dispense <> COALESCE((SELECT SUM(quantity) FROM prescription_batch_items i WHERE i.line_id=l.id),0)
		`, prescriptionID).Scan(&mismatch); err != nil {
			return nil, err
		}
		if mismatch > 0 {
			return nil, errors.New("allocation totals mismatch")
		}
	*/

	// Stamp dispensed (no stock mutation here)
	if _, err := tx.ExecContext(ctx, `
	  UPDATE prescriptions
	  SET is_dispensed=TRUE,
	  dispensed_by = current_setting('sothea.user_id')::bigint,
	  dispensed_at=NOW(),
	  updated_at=NOW()
	  WHERE id=$1
	`, prescriptionID); err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetPrescriptionByID(ctx, prescriptionID)
}

// ----------------------------------------------------------------------------
// Utility used by FEFO/helper
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) GetLine(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var l entities.PrescriptionLine
	if err := dbx.QueryRowContext(ctx, `
	  SELECT
	    pl.id, pl.prescription_id, pl.presentation_id, pl.remarks,
	    pl.dose_amount, pl.dose_unit,
	    pl.schedule_kind, pl.every_n, pl.frequency_per_schedule, pl.duration,
	    pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
	    (SELECT dispense_unit FROM drug_presentations WHERE id=pl.presentation_id) AS du
	  FROM prescription_lines pl
	  WHERE pl.id=$1
	`, lineID).Scan(
		&l.ID, &l.PrescriptionID, &l.PresentationID, &l.Remarks,
		&l.DoseAmount, &l.DoseUnit,
		&l.ScheduleKind, &l.EveryN, &l.FrequencyPerSchedule, &l.Duration,
		&l.TotalToDispense, &l.IsPacked, &l.PackedBy, &l.PackedAt, &l.DispenseUnit,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("line not found")
		}
		return nil, err
	}

	// Allocations
	rows, err := dbx.QueryContext(ctx, `
	  SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
	  FROM prescription_batch_items WHERE line_id=$1 ORDER BY id
	`, lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a entities.LineAllocation
		if err := rows.Scan(&a.ID, &a.LineID, &a.BatchLocationID, &a.Quantity, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		l.Allocations = append(l.Allocations, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &l, nil
}
