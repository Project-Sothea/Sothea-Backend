package postgres

import (
	"context"
	"errors"
	"time"

	"sothea-backend/entities"
	db "sothea-backend/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ----------------------------------------------------------------------------
// REPO
// ----------------------------------------------------------------------------

type postgresPrescriptionRepository struct {
	Conn    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresPrescriptionRepository(conn *pgxpool.Pool) entities.PrescriptionRepository {
	return &postgresPrescriptionRepository{
		Conn:    conn,
		queries: db.New(conn),
	}
}

func (r *postgresPrescriptionRepository) q(ctx context.Context) *db.Queries {
	if tx, ok := TxFromCtx(ctx); ok && tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func withTx(ctx context.Context, pool *pgxpool.Pool) (pgx.Tx, bool, error) {
	if tx, ok := TxFromCtx(ctx); ok {
		return tx, false, nil
	}
	tx, err := pool.Begin(ctx)
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	if p == nil {
		return nil, errors.New("prescription payload required")
	}
	row, err := q.InsertPrescription(ctx, db.InsertPrescriptionParams{
		PatientID: p.PatientID,
		Vid:       p.Vid,
		Notes:     p.Notes,
	})
	if err != nil {
		return nil, err
	}
	p.ID = row.ID
	p.CreatedAt = row.CreatedAt
	p.UpdatedAt = row.UpdatedAt

	if own {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}
	return r.GetPrescriptionByID(ctx, p.ID)
}

func (r *postgresPrescriptionRepository) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	q := r.q(ctx)
	header, err := q.GetPrescriptionHeader(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("prescription not found")
	}
	if err != nil {
		return nil, err
	}

	p := entities.Prescription{
		Prescription: db.Prescription{
			ID:          header.ID,
			PatientID:   header.PatientID,
			Vid:         header.Vid,
			Notes:       header.Notes,
			CreatedBy:   header.CreatedBy,
			CreatedAt:   header.CreatedAt,
			UpdatedAt:   header.UpdatedAt,
			IsDispensed: header.IsDispensed,
			DispensedBy: header.DispensedBy,
			DispensedAt: header.DispensedAt,
		},
	}

	lineRows, err := q.ListPrescriptionLines(ctx, id)
	if err != nil {
		return nil, err
	}
	lines := make([]entities.PrescriptionLine, 0, len(lineRows))
	lineIDs := make([]int64, 0, len(lineRows))
	for _, row := range lineRows {
		l := entities.PrescriptionLine{
			PrescriptionLine: db.PrescriptionLine{
				ID:              row.ID,
				PrescriptionID:  row.PrescriptionID,
				DrugID:          row.DrugID,
				Remarks:         row.Remarks,
				Prn:             row.Prn,
				DoseAmount:      row.DoseAmount,
				DoseUnit:        row.DoseUnit,
				FrequencyCode:   row.FrequencyCode,
				Duration:        row.Duration,
				DurationUnit:    row.DurationUnit,
				TotalToDispense: row.TotalToDispense,
				IsPacked:        row.IsPacked,
				PackedBy:        row.PackedBy,
				PackedAt:        row.PackedAt,
			},
		}
		lines = append(lines, l)
		lineIDs = append(lineIDs, row.ID)
	}

	if len(lineIDs) > 0 {
		allocRows, err := q.ListAllocationsByLineIDs(ctx, lineIDs)
		if err != nil {
			return nil, err
		}
		index := make(map[int64]*entities.PrescriptionLine, len(lines))
		for i := range lines {
			index[lines[i].ID] = &lines[i]
		}
		for _, ar := range allocRows {
			a := db.PrescriptionBatchItem{
				ID:              ar.ID,
				LineID:          ar.LineID,
				BatchLocationID: ar.BatchLocationID,
				Quantity:        ar.Quantity,
				CreatedAt:       ar.CreatedAt,
				UpdatedAt:       ar.UpdatedAt,
			}
			if ln := index[a.LineID]; ln != nil {
				ln.Allocations = append(ln.Allocations, a)
			}
		}
	}

	p.Lines = lines
	return &p, nil
}

func (r *postgresPrescriptionRepository) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	q := r.q(ctx)
	switch {
	case patientID != nil && vid != nil:
		rows, err := q.ListPrescriptionsByPatientVisit(ctx, db.ListPrescriptionsByPatientVisitParams{PatientID: int32(*patientID), Vid: *vid})
		if err != nil {
			return nil, err
		}
		out := make([]*entities.Prescription, 0, len(rows))
		for _, row := range rows {
			out = append(out, toPrescriptionListEntity(row.ID, int64(row.PatientID), row.Vid, row.Notes, row.CreatedBy, row.CreatedAt, row.UpdatedAt, row.IsDispensed, row.DispensedBy, row.DispensedAt))
		}
		return out, nil
	case patientID != nil:
		rows, err := q.ListPrescriptionsByPatient(ctx, int32(*patientID))
		if err != nil {
			return nil, err
		}
		out := make([]*entities.Prescription, 0, len(rows))
		for _, row := range rows {
			out = append(out, toPrescriptionListEntity(row.ID, int64(row.PatientID), row.Vid, row.Notes, row.CreatedBy, row.CreatedAt, row.UpdatedAt, row.IsDispensed, row.DispensedBy, row.DispensedAt))
		}
		return out, nil
	default:
		rows, err := q.ListPrescriptionsAll(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]*entities.Prescription, 0, len(rows))
		for _, row := range rows {
			out = append(out, toPrescriptionListEntity(row.ID, int64(row.PatientID), row.Vid, row.Notes, row.CreatedBy, row.CreatedAt, row.UpdatedAt, row.IsDispensed, row.DispensedBy, row.DispensedAt))
		}
		return out, nil
	}
}

func (r *postgresPrescriptionRepository) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	isDispensed, err := q.GetPrescriptionDispensed(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	if isDispensed {
		// For dispensed prescriptions, only allow updating notes
		if err := q.UpdatePrescriptionNotes(ctx, db.UpdatePrescriptionNotesParams{
			ID:    p.ID,
			Notes: p.Notes,
		}); err != nil {
			return nil, err
		}
	} else {
		// For non-dispensed prescriptions, allow updating all fields
		if err := q.UpdatePrescriptionFull(ctx, db.UpdatePrescriptionFullParams{
			ID:        p.ID,
			PatientID: p.PatientID,
			Vid:       p.Vid,
			Notes:     p.Notes,
		}); err != nil {
			return nil, err
		}
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	isDispensed, err := q.GetPrescriptionDispensed(ctx, id)
	if err != nil {
		return err
	}
	if isDispensed {
		return errors.New("cannot delete a dispensed prescription")
	}

	// Delete children then parent.
	// NOTE: stock reservations are released by DB triggers on DELETE of prescription_batch_items.
	if err := q.DeleteAllocationsByPrescription(ctx, id); err != nil {
		return err
	}
	if err := q.DeleteLinesByPrescription(ctx, id); err != nil {
		return err
	}
	if err := q.DeletePrescription(ctx, id); err != nil {
		return err
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)

	var isDispensed bool
	if isDispensed, err = q.GetPrescriptionDispensed(ctx, line.PrescriptionID); err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot add line to a dispensed prescription")
	}

	row, err := q.InsertLine(ctx, db.InsertLineParams{
		PrescriptionID: line.PrescriptionID,
		DrugID:         line.DrugID,
		Remarks:        line.Remarks,
		Prn:            line.Prn,
		DoseAmount:     line.DoseAmount,
		DoseUnit:       line.DoseUnit,
		FrequencyCode:  line.FrequencyCode,
		Duration:       line.Duration,
		DurationUnit:   line.DurationUnit,
	})
	if err != nil {
		return nil, mapPrescriptionSQLError(err)
	}
	line.ID = row.ID
	line.TotalToDispense = row.TotalToDispense
	line.IsPacked = row.IsPacked

	// Return enriched line
	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)

	pid, err := q.GetPrescriptionIDForLine(ctx, line.ID)
	if err != nil {
		return nil, err
	}
	isDispensed, err := q.GetPrescriptionDispensed(ctx, pid)
	if err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot modify a line on a dispensed prescription")
	}

	curRow, err := q.GetLineGuardForUpdate(ctx, line.ID)
	if err != nil {
		return nil, err
	}
	cur := entities.PrescriptionLine{
		PrescriptionLine: db.PrescriptionLine{
			DrugID:        curRow.DrugID,
			DoseAmount:    curRow.DoseAmount,
			DoseUnit:      curRow.DoseUnit,
			FrequencyCode: curRow.FrequencyCode,
			Duration:      curRow.Duration,
			DurationUnit:  curRow.DurationUnit,
			Remarks:       curRow.Remarks,
			Prn:           curRow.Prn,
		},
	}

	presChanged := cur.DrugID != line.DrugID

	// 1) Only clear allocations if the presentation changed
	if presChanged {
		if err := q.DeleteAllocationsByLine(ctx, line.ID); err != nil {
			return nil, err
		}
	}

	td, err := q.UpdateLine(ctx, db.UpdateLineParams{
		ID:            line.ID,
		DrugID:        line.DrugID,
		Remarks:       line.Remarks,
		Prn:           line.Prn,
		DoseAmount:    line.DoseAmount,
		DoseUnit:      line.DoseUnit,
		FrequencyCode: line.FrequencyCode,
		Duration:      line.Duration,
		DurationUnit:  line.DurationUnit,
	})
	if err != nil {
		return nil, mapPrescriptionSQLError(err)
	}
	line.TotalToDispense = td

	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	pid, err := q.GetPrescriptionIDForLine(ctx, lineID)
	if err != nil {
		return err
	}
	isDispensed, err := q.GetPrescriptionDispensed(ctx, pid)
	if err != nil {
		return err
	}
	if isDispensed {
		return errors.New("cannot remove a line from a dispensed prescription")
	}

	// Triggers on DELETE of prescription_batch_items will release reserved stock.
	if err := q.DeleteAllocationsByLine(ctx, lineID); err != nil {
		return err
	}
	if err := q.DeleteLine(ctx, lineID); err != nil {
		return err
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Allocations (packing plan) - replace all
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) ListLineAllocations(ctx context.Context, lineID int64) ([]db.PrescriptionBatchItem, error) {
	rows, err := r.q(ctx).ListAllocationsByLine(ctx, lineID)
	if err != nil {
		return nil, err
	}
	out := make([]db.PrescriptionBatchItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, db.PrescriptionBatchItem{
			ID:              row.ID,
			LineID:          row.LineID,
			BatchLocationID: row.BatchLocationID,
			Quantity:        row.Quantity,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *postgresPrescriptionRepository) SetLineAllocations(ctx context.Context, lineID int64, allocs []db.PrescriptionBatchItem) ([]db.PrescriptionBatchItem, error) {
	tx, own, err := withTx(ctx, r.Conn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if own {
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	pid, err := q.GetPrescriptionIDForLine(ctx, lineID)
	if err != nil {
		return nil, err
	}
	isDispensed, err := q.GetPrescriptionDispensed(ctx, pid)
	if err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot change allocations on a dispensed prescription")
	}

	// Replace all; triggers will adjust reservations per-row (delete→return stock, insert→reserve stock)
	if err := q.DeleteAllocationsByLine(ctx, lineID); err != nil {
		return nil, err
	}
	if len(allocs) > 0 {
		for i := range allocs {
			row, err := q.InsertAllocation(ctx, db.InsertAllocationParams{
				LineID:          lineID,
				BatchLocationID: allocs[i].BatchLocationID,
				Quantity:        allocs[i].Quantity,
			})
			if err != nil {
				return nil, err
			}
			allocs[i].ID = row.ID
			allocs[i].LineID = lineID
			allocs[i].CreatedAt = row.CreatedAt
			allocs[i].UpdatedAt = row.UpdatedAt
		}
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	pid, err := q.GetPrescriptionIDForLine(ctx, lineID)
	if err != nil {
		return nil, err
	}
	isDispensed, err := q.GetPrescriptionDispensed(ctx, pid)
	if err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot pack a line on a dispensed prescription")
	}

	// Stamp packed fields
	if err := q.MarkLinePacked(ctx, lineID); err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	pid, err := q.GetPrescriptionIDForLine(ctx, lineID)
	if err != nil {
		return nil, err
	}
	isDispensed, err := q.GetPrescriptionDispensed(ctx, pid)
	if err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("cannot unpack a line on a dispensed prescription")
	}

	if err := q.UnpackLine(ctx, lineID); err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
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
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)
	isDispensed, err := q.GetPrescriptionDispensed(ctx, prescriptionID)
	if err != nil {
		return nil, err
	}
	if isDispensed {
		return nil, errors.New("prescription already dispensed")
	}

	// Must have lines and all must be packed
	totalLines, err := q.CountLinesForPrescription(ctx, prescriptionID)
	if err != nil {
		return nil, err
	}
	if totalLines == 0 {
		return nil, errors.New("no lines to dispense")
	}
	packedLines, err := q.CountPackedLinesForPrescription(ctx, prescriptionID)
	if err != nil {
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

	if err := q.DispensePrescription(ctx, prescriptionID); err != nil {
		return nil, err
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}
	return r.GetPrescriptionByID(ctx, prescriptionID)
}

// ----------------------------------------------------------------------------
// Utility used by FEFO/helper
// ----------------------------------------------------------------------------

func (r *postgresPrescriptionRepository) GetLine(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	q := r.q(ctx)
	row, err := q.GetLineWithDispenseUnit(ctx, lineID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("line not found")
	}
	if err != nil {
		return nil, err
	}

	l := entities.PrescriptionLine{
		PrescriptionLine: db.PrescriptionLine{
			ID:              row.ID,
			PrescriptionID:  row.PrescriptionID,
			DrugID:          row.DrugID,
			Remarks:         row.Remarks,
			Prn:             row.Prn,
			DoseAmount:      row.DoseAmount,
			DoseUnit:        row.DoseUnit,
			FrequencyCode:   row.FrequencyCode,
			Duration:        row.Duration,
			DurationUnit:    row.DurationUnit,
			TotalToDispense: row.TotalToDispense,
			IsPacked:        row.IsPacked,
			PackedBy:        row.PackedBy,
			PackedAt:        row.PackedAt,
		},
		DispenseUnit: row.Du,
	}

	allocRows, err := q.ListAllocationsByLine(ctx, lineID)
	if err != nil {
		return nil, err
	}
	for _, ar := range allocRows {
		l.Allocations = append(l.Allocations, db.PrescriptionBatchItem{
			ID:              ar.ID,
			LineID:          ar.LineID,
			BatchLocationID: ar.BatchLocationID,
			Quantity:        ar.Quantity,
			CreatedAt:       ar.CreatedAt,
			UpdatedAt:       ar.UpdatedAt,
		})
	}
	return &l, nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func toPrescriptionListEntity(id int64, patientID int64, vid int32, notes *string, createdBy *int64, createdAt time.Time, updatedAt *time.Time, isDispensed bool, dispensedBy *int64, dispensedAt *time.Time) *entities.Prescription {
	return &entities.Prescription{
		Prescription: db.Prescription{
			ID:          id,
			PatientID:   int32(patientID),
			Vid:         vid,
			Notes:       notes,
			CreatedBy:   createdBy,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			IsDispensed: isDispensed,
			DispensedBy: dispensedBy,
			DispensedAt: dispensedAt,
		},
	}
}
