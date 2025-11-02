package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/stretchr/testify/assert"
)

// --- helpers ---------------------------------------------------------------

func sp(s string) *string { return &s }
func ip(i int64) *int64   { return &i }

func newRxRepo(t *testing.T) (*postgresPrescriptionRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	repo := &postgresPrescriptionRepository{Conn: db}
	cleanup := func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}
	return repo, mock, cleanup
}

// --- CreatePrescription ----------------------------------------------------

func Test_CreatePrescription_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`
		INSERT INTO prescriptions (patient_id, vid, notes)
		VALUES ($1,$2,$3)
		RETURNING id, created_at, updated_at`)).
		WithArgs(int64(99), int32(1), sp("note")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(5), now, now))
	mock.ExpectCommit()

	// hydrate via GetPrescriptionByID
	mock.ExpectQuery(qm(`
		SELECT id, patient_id, vid, notes,
		       created_by, created_at, updated_at,
		       is_dispensed, dispensed_by, dispensed_at
		FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(5), int64(99), int32(1), "note", ip(7), now, now, false, nil, nil))

	// lines (none)
	mock.ExpectQuery(qm(`
		SELECT
		  pl.id, pl.prescription_id, pl.presentation_id, pl.remarks,
		  pl.dose_amount, pl.dose_unit,
		  pl.schedule_kind, pl.every_n, pl.frequency_per_schedule, pl.duration,
		  pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
		  (SELECT generic_name FROM drugs d
		     JOIN drug_presentations dp ON dp.drug_id=d.id
		   WHERE dp.id=pl.presentation_id) AS drug_name,
		  (SELECT route_code FROM drug_presentations dp WHERE dp.id=pl.presentation_id) AS route_code,
		  (SELECT dispense_unit FROM drug_presentations dp WHERE dp.id=pl.presentation_id) AS dispense_unit,
		  (SELECT CASE
		            WHEN dp.strength_den IS NULL THEN
		              dp.strength_num::text || ' ' || dp.strength_unit_num || '/' || dp.dispense_unit
		            ELSE
		              dp.strength_num::text || ' ' || dp.strength_unit_num || '/' ||
		              dp.strength_den::text || ' ' || dp.strength_unit_den
		          END
		     FROM drug_presentations dp
		    WHERE dp.id=pl.presentation_id) AS display_strength
		FROM prescription_lines pl
		WHERE pl.prescription_id = $1
		ORDER BY pl.id`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at",
			"drug_name", "route_code", "dispense_unit", "display_strength",
		}))

	p := &entities.Prescription{PatientID: 99, VID: 1, Notes: sp("note")}
	out, err := repo.CreatePrescription(context.Background(), p)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), out.ID)
	assert.False(t, out.IsDispensed)
}

func Test_CreatePrescription_InsertErr_RollsBack(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`INSERT INTO prescriptions`)).
		WithArgs(int64(1), int32(2), (*string)(nil)).
		WillReturnError(errors.New("ins err"))
	mock.ExpectRollback()

	_, err := repo.CreatePrescription(context.Background(), &entities.Prescription{PatientID: 1, VID: 2})
	assert.Error(t, err)
}

func Test_CreatePrescription_CommitErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`INSERT INTO prescriptions`)).
		WithArgs(int64(1), int32(2), (*string)(nil)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(10), now, now))
	mock.ExpectCommit().WillReturnError(errors.New("commit fail"))

	_, err := repo.CreatePrescription(context.Background(), &entities.Prescription{PatientID: 1, VID: 2})
	assert.Error(t, err)
}

// --- GetPrescriptionByID ---------------------------------------------------

func Test_GetPrescriptionByID_WithLinesAndAllocations(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	// header
	mock.ExpectQuery(qm(`
		SELECT id, patient_id, vid, notes,
		       created_by, created_at, updated_at,
		       is_dispensed, dispensed_by, dispensed_at
		FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(55), int64(99), int32(3), "hdr", ip(7), now, now, false, nil, nil))

	// lines (2)
	lrows := sqlmock.NewRows([]string{
		"id", "prescription_id", "presentation_id", "remarks",
		"dose_amount", "dose_unit",
		"schedule_kind", "every_n", "frequency_per_schedule", "duration",
		"total_to_dispense", "is_packed", "packed_by", "packed_at",
		"drug_name", "route_code", "dispense_unit", "display_strength",
	}).AddRow(int64(1), int64(55), int64(101), "rmk", 500, "mg", "day", 1, 3.0, 5.0,
		9, false, nil, nil, "PCM", "PO", "tab", "500 mg/tab").
		AddRow(int64(2), int64(55), int64(102), nil, 5, "mL", "hour", 8, 1.0, 1.0,
			5, true, ip(42), now, "PCM", "PO", "mL", "250 mg/5 mL")
	mock.ExpectQuery(qm(`FROM prescription_lines pl
		WHERE pl.prescription_id = $1
		ORDER BY pl.id`)).
		WithArgs(int64(55)).
		WillReturnRows(lrows)

	// allocations for both lines
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
			FROM prescription_batch_items
			WHERE line_id IN ($1,$2)
			ORDER BY line_id, id`)).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at",
		}).AddRow(int64(11), int64(1), int64(9001), 4, now, now).
			AddRow(int64(21), int64(2), int64(9002), 5, now, now))

	p, err := repo.GetPrescriptionByID(context.Background(), 55)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(p.Lines))
	assert.Equal(t, 1, len(p.Lines[0].Allocations))
	assert.Equal(t, 1, len(p.Lines[1].Allocations))
}

func Test_GetPrescriptionByID_NotFound(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectQuery(qm(`FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(9)).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetPrescriptionByID(context.Background(), 9)
	assert.Error(t, err)
	assert.Nil(t, p)
	assert.Equal(t, "prescription not found", err.Error())
}

func Test_GetPrescriptionByID_LinesQueryErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	mock.ExpectQuery(qm(`FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(1), int64(2), int32(3), nil, nil, now, now, false, nil, nil))
	mock.ExpectQuery(qm(`FROM prescription_lines pl
		WHERE pl.prescription_id = $1
		ORDER BY pl.id`)).
		WithArgs(int64(1)).
		WillReturnError(errors.New("lines err"))

	_, err := repo.GetPrescriptionByID(context.Background(), 1)
	assert.Error(t, err)
}

func Test_GetPrescriptionByID_AllocQueryErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	mock.ExpectQuery(qm(`FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(1), int64(2), int32(3), nil, nil, now, now, false, nil, nil))
	lrows := sqlmock.NewRows([]string{
		"id", "prescription_id", "presentation_id", "remarks",
		"dose_amount", "dose_unit",
		"schedule_kind", "every_n", "frequency_per_schedule", "duration",
		"total_to_dispense", "is_packed", "packed_by", "packed_at",
		"drug_name", "route_code", "dispense_unit", "display_strength",
	}).AddRow(int64(7), int64(1), int64(50), nil, 1, "tab", "day", 1, 1.0, 1.0, 1, false, nil, nil, "X", "PO", "tab", "somestr")
	mock.ExpectQuery(qm(`FROM prescription_lines pl
		WHERE pl.prescription_id = $1
		ORDER BY pl.id`)).
		WithArgs(int64(1)).
		WillReturnRows(lrows)
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
			FROM prescription_batch_items
			WHERE line_id IN ($1)
			ORDER BY line_id, id`)).
		WithArgs(int64(7)).
		WillReturnError(errors.New("alloc err"))

	_, err := repo.GetPrescriptionByID(context.Background(), 1)
	assert.Error(t, err)
}

// --- ListPrescriptions -----------------------------------------------------

func Test_ListPrescriptions_All_And_Filters(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	// all
	mock.ExpectQuery(qm(`
	  SELECT id, patient_id, vid, notes,
	         created_by, created_at, updated_at,
	         is_dispensed, dispensed_by, dispensed_at
	  FROM prescriptions ORDER BY created_at DESC`)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(1), int64(10), int32(1), nil, nil, now, now, false, nil, nil).
			AddRow(int64(2), int64(11), int32(1), "n", ip(7), now, now, true, ip(8), tp(now)))

	all, err := repo.ListPrescriptions(context.Background(), nil, nil)
	assert.NoError(t, err)
	assert.Len(t, all, 2)

	// by patient
	pid := int64(10)
	mock.ExpectQuery(qm(`
	  SELECT id, patient_id, vid, notes,
	         created_by, created_at, updated_at,
	         is_dispensed, dispensed_by, dispensed_at
	  FROM prescriptions WHERE patient_id=$1 ORDER BY created_at DESC`)).
		WithArgs(pid).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(1), int64(10), int32(1), nil, nil, now, now, false, nil, nil))
	flt1, err := repo.ListPrescriptions(context.Background(), &pid, nil)
	assert.NoError(t, err)
	assert.Len(t, flt1, 1)

	// by patient + vid
	vid := int32(3)
	mock.ExpectQuery(qm(`
	  SELECT id, patient_id, vid, notes,
	         created_by, created_at, updated_at,
	         is_dispensed, dispensed_by, dispensed_at
	  FROM prescriptions WHERE patient_id=$1 AND vid=$2 ORDER BY created_at DESC`)).
		WithArgs(pid, vid).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(9), int64(10), int32(3), "x", nil, now, now, false, nil, nil))
	flt2, err := repo.ListPrescriptions(context.Background(), &pid, &vid)
	assert.NoError(t, err)
	assert.Len(t, flt2, 1)
}

func Test_ListPrescriptions_QueryErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectQuery(qm(`FROM prescriptions ORDER BY created_at DESC`)).
		WillReturnError(errors.New("q err"))

	_, err := repo.ListPrescriptions(context.Background(), nil, nil)
	assert.Error(t, err)
}

// --- UpdatePrescription ----------------------------------------------------

func Test_UpdatePrescription_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`
		UPDATE prescriptions
		SET patient_id=$2, vid=$3, notes=$4, updated_at=now()
		WHERE id=$1`)).
		WithArgs(int64(5), int64(100), int32(2), sp("n")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// hydrate
	mock.ExpectQuery(qm(`FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(5), int64(100), int32(2), "n", nil, now, now, false, nil, nil))
	mock.ExpectQuery(qm(`FROM prescription_lines pl
		WHERE pl.prescription_id = $1
		ORDER BY pl.id`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at",
			"drug_name", "route_code", "dispense_unit", "display_strength",
		}))

	out, err := repo.UpdatePrescription(context.Background(), &entities.Prescription{ID: 5, PatientID: 100, VID: 2, Notes: sp("n")})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), out.ID)
}

func Test_UpdatePrescription_DispensedGuard(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()

	_, err := repo.UpdatePrescription(context.Background(), &entities.Prescription{ID: 5})
	assert.Error(t, err)
	assert.Equal(t, "cannot modify a dispensed prescription", err.Error())
}

func Test_UpdatePrescription_SelectErr_ExecErr_CommitErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// select err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(1)).WillReturnError(errors.New("sel err"))
	mock.ExpectRollback()
	_, err := repo.UpdatePrescription(context.Background(), &entities.Prescription{ID: 1})
	assert.Error(t, err)

	// exec err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`UPDATE prescriptions`)).
		WithArgs(int64(2), int64(9), int32(1), (*string)(nil)).
		WillReturnError(errors.New("exec err"))
	mock.ExpectRollback()
	_, err = repo.UpdatePrescription(context.Background(), &entities.Prescription{ID: 2, PatientID: 9, VID: 1})
	assert.Error(t, err)

	// commit err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`UPDATE prescriptions`)).
		WithArgs(int64(3), int64(9), int32(1), (*string)(nil)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit err"))
	_, err = repo.UpdatePrescription(context.Background(), &entities.Prescription{ID: 3, PatientID: 9, VID: 1})
	assert.Error(t, err)
}

// --- DeletePrescription ----------------------------------------------------

func Test_DeletePrescription_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(77)).
		WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id IN (SELECT id FROM prescription_lines WHERE prescription_id=$1)`)).
		WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(qm(`DELETE FROM prescription_lines WHERE prescription_id=$1`)).
		WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(qm(`DELETE FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(77)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.DeletePrescription(context.Background(), 77)
	assert.NoError(t, err)
}

func Test_DeletePrescription_Guards_And_NotFound(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// dispensed guard
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	assert.EqualError(t, repo.DeletePrescription(context.Background(), 1), "cannot delete a dispensed prescription")

	// not found
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items`)).
		WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(qm(`DELETE FROM prescription_lines`)).
		WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(qm(`DELETE FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(2)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	assert.EqualError(t, repo.DeletePrescription(context.Background(), 2), "prescription not found")
}

func Test_DeletePrescription_Errors(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// select err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(3)).WillReturnError(errors.New("sel err"))
	mock.ExpectRollback()
	assert.Error(t, repo.DeletePrescription(context.Background(), 3))

	// child delete err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(4)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items`)).
		WithArgs(int64(4)).WillReturnError(errors.New("child err"))
	mock.ExpectRollback()
	assert.Error(t, repo.DeletePrescription(context.Background(), 4))
}

// --- AddLine ---------------------------------------------------------------

func Test_AddLine_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(55)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`
	  INSERT INTO prescription_lines (
	    prescription_id, presentation_id, remarks,
	    dose_amount, dose_unit,
	    schedule_kind, every_n, frequency_per_schedule, duration
	  )
	  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	  RETURNING id, total_to_dispense, is_packed`)).
		WithArgs(int64(55), int64(101), sp("r"), 500, "mg", "day", 1, 3.0, 5.0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "total_to_dispense", "is_packed"}).AddRow(int64(7), 9, false))
	mock.ExpectCommit()

	// hydrate GetLine
	mock.ExpectQuery(qm(`
	  SELECT
	    pl.id, pl.prescription_id, pl.presentation_id, pl.remarks,
	    pl.dose_amount, pl.dose_unit,
	    pl.schedule_kind, pl.every_n, pl.frequency_per_schedule, pl.duration,
	    pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
	    (SELECT dispense_unit FROM drug_presentations WHERE id=pl.presentation_id) AS du
	  FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at", "du",
		}).AddRow(int64(7), int64(55), int64(101), "r", 500, "mg", "day", 1, 3.0, 5.0, 9, false, nil, nil, "tab"))
	mock.ExpectQuery(qm(`FROM prescription_batch_items WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at"}))

	line := &entities.PrescriptionLine{
		PrescriptionID: 55, PresentationID: 101, Remarks: sp("r"),
		DoseAmount: 500, DoseUnit: "mg", ScheduleKind: "day", EveryN: 1, FrequencyPerSchedule: 3, Duration: 5,
	}
	out, err := repo.AddLine(context.Background(), line)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), out.ID)
	assert.Equal(t, "tab", out.DispenseUnit)
}

func Test_AddLine_DispensedGuard_InsertErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// dispensed
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	_, err := repo.AddLine(context.Background(), &entities.PrescriptionLine{PrescriptionID: 1})
	assert.EqualError(t, err, "cannot add line to a dispensed prescription")

	// insert err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`INSERT INTO prescription_lines`)).
		WithArgs(int64(2), int64(10), (*string)(nil), 1, "x", "day", 1, 1.0, 1.0).
		WillReturnError(errors.New("ins err"))
	mock.ExpectRollback()
	_, err = repo.AddLine(context.Background(), &entities.PrescriptionLine{PrescriptionID: 2, PresentationID: 10, DoseAmount: 1, DoseUnit: "x", ScheduleKind: "day", EveryN: 1, FrequencyPerSchedule: 1, Duration: 1})
	assert.Error(t, err)
}

// --- UpdateLine ------------------------------------------------------------

func Test_UpdateLine_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(55)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(55)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(9)).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery(qm(`UPDATE prescription_lines SET`)).
		WithArgs(int64(9), int64(101), (*string)(nil), 2, "tab", "day", 1, 1.5, 1.0).
		WillReturnRows(sqlmock.NewRows([]string{"total_to_dispense"}).AddRow(10))
	mock.ExpectCommit()

	// hydrate
	mock.ExpectQuery(qm(`FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at", "du",
		}).AddRow(int64(9), int64(55), int64(101), nil, 2, "tab", "day", 1, 1.5, 1.0, 10, false, nil, nil, "tab"))
	mock.ExpectQuery(qm(`FROM prescription_batch_items WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at"}))

	out, err := repo.UpdateLine(context.Background(), &entities.PrescriptionLine{ID: 9, PresentationID: 101, DoseAmount: 2, DoseUnit: "tab", ScheduleKind: "day", EveryN: 1, FrequencyPerSchedule: 1.5, Duration: 1})
	assert.NoError(t, err)
	assert.Equal(t, 10, out.TotalToDispense)
}

func Test_UpdateLine_GuardsAndErrors(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// select pid err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(1)).WillReturnError(errors.New("sel err"))
	mock.ExpectRollback()
	_, err := repo.UpdateLine(context.Background(), &entities.PrescriptionLine{ID: 1})
	assert.Error(t, err)

	// dispensed guard
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(20)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(20)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	_, err = repo.UpdateLine(context.Background(), &entities.PrescriptionLine{ID: 2})
	assert.EqualError(t, err, "cannot modify a line on a dispensed prescription")

	// delete allocs err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(21)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(21)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(3)).WillReturnError(errors.New("del err"))
	mock.ExpectRollback()
	_, err = repo.UpdateLine(context.Background(), &entities.PrescriptionLine{ID: 3})
	assert.Error(t, err)
}

// --- RemoveLine ------------------------------------------------------------

func Test_RemoveLine_Success_And_NotFound(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// success
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(40)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(40)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(qm(`DELETE FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	assert.NoError(t, repo.RemoveLine(context.Background(), 7))

	// not found
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(8)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(40)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(40)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(8)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(qm(`DELETE FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(8)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	assert.EqualError(t, repo.RemoveLine(context.Background(), 8), "line not found")
}

func Test_RemoveLine_Guard_And_Errors(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// dispensed guard
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(9)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(41)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(41)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	assert.EqualError(t, repo.RemoveLine(context.Background(), 9), "cannot remove a line from a dispensed prescription")

	// select pid err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(10)).WillReturnError(errors.New("sel err"))
	mock.ExpectRollback()
	assert.Error(t, repo.RemoveLine(context.Background(), 10))
}

// --- ListLineAllocations ---------------------------------------------------

func Test_ListLineAllocations_Success_And_Err(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	mock.ExpectQuery(qm(`
	  SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
	  FROM prescription_batch_items
	  WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at",
		}).AddRow(int64(1), int64(5), int64(100), 2, now, now))
	out, err := repo.ListLineAllocations(context.Background(), 5)
	assert.NoError(t, err)
	assert.Len(t, out, 1)

	mock.ExpectQuery(qm(`FROM prescription_batch_items
	  WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(6)).
		WillReturnError(errors.New("q err"))
	_, err = repo.ListLineAllocations(context.Background(), 6)
	assert.Error(t, err)
}

// --- SetLineAllocations ----------------------------------------------------

func Test_SetLineAllocations_Success_WithAndWithoutItems(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	// with items
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(50)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(77)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(77)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(50)).WillReturnResult(sqlmock.NewResult(0, 2))

	prep := mock.ExpectPrepare(qm(`
  	INSERT INTO prescription_batch_items (line_id, batch_location_id, quantity)
  	VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`))

	// 1st item
	prep.ExpectQuery().
		WithArgs(int64(50), int64(9001), 3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(1), now, now))

	// 2nd item
	prep.ExpectQuery().
		WithArgs(int64(50), int64(9002), 6).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(2), now, now))

	mock.ExpectCommit()

	// list after commit
	mock.ExpectQuery(qm(`FROM prescription_batch_items
	  WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at",
		}).AddRow(int64(1), int64(50), int64(9001), 3, now, now).
			AddRow(int64(2), int64(50), int64(9002), 6, now, now))

	items := []entities.LineAllocation{
		{BatchLocationID: 9001, Quantity: 3},
		{BatchLocationID: 9002, Quantity: 6},
	}
	out, err := repo.SetLineAllocations(context.Background(), 50, items)
	assert.NoError(t, err)
	assert.Len(t, out, 2)

	// without items (delete only)
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(51)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(77)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(77)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(51)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(qm(`FROM prescription_batch_items
	  WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(51)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at"}))
	empty, err := repo.SetLineAllocations(context.Background(), 51, nil)
	assert.NoError(t, err)
	assert.Len(t, empty, 0)
}

func Test_SetLineAllocations_GuardsAndErrors(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// dispensed guard
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(60)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(100)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(100)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	_, err := repo.SetLineAllocations(context.Background(), 60, nil)
	assert.EqualError(t, err, "cannot change allocations on a dispensed prescription")

	// delete err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(61)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(100)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(100)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(61)).WillReturnError(errors.New("del err"))
	mock.ExpectRollback()
	_, err = repo.SetLineAllocations(context.Background(), 61, nil)
	assert.Error(t, err)

	// prepare err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(62)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(100)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(100)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(62)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(qm(`INSERT INTO prescription_batch_items`)).WillReturnError(errors.New("prep err"))
	mock.ExpectRollback()
	_, err = repo.SetLineAllocations(context.Background(), 62, []entities.LineAllocation{{BatchLocationID: 1, Quantity: 1}})
	assert.Error(t, err)

	// insert err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(63)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(100)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(100)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(63)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectPrepare(qm(`INSERT INTO prescription_batch_items`)).
		ExpectQuery().WithArgs(int64(63), int64(1), 1).
		WillReturnError(errors.New("ins err"))
	mock.ExpectRollback()
	_, err = repo.SetLineAllocations(context.Background(), 63, []entities.LineAllocation{{BatchLocationID: 1, Quantity: 1}})
	assert.Error(t, err)

	// commit err
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(64)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(100)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(100)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`DELETE FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(64)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(errors.New("commit err"))
	_, err = repo.SetLineAllocations(context.Background(), 64, nil)
	assert.Error(t, err)
}

// --- MarkLinePacked --------------------------------------------------------

func Test_MarkLinePacked_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(55)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(55)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`SELECT total_to_dispense FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"total_to_dispense"}).AddRow(9))
	mock.ExpectQuery(qm(`SELECT COALESCE(SUM(quantity),0) FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(7)).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(9))
	mock.ExpectExec(qm(`
	  UPDATE prescription_lines
	  SET is_packed=TRUE, packed_by=$2, packed_at=NOW(), updated_at=NOW()
	  WHERE id=$1`)).
		WithArgs(int64(7), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// hydrate
	mock.ExpectQuery(qm(`FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at", "du",
		}).AddRow(int64(7), int64(55), int64(101), nil, 500, "mg", "day", 1, 3.0, 5.0, 9, true, ip(99), tp(now), "tab"))
	mock.ExpectQuery(qm(`FROM prescription_batch_items WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at"}))

	out, err := repo.MarkLinePacked(context.Background(), 7)
	assert.NoError(t, err)
	assert.True(t, out.IsPacked)
	assert.Equal(t, int64(99), *out.PackedBy)
}

func Test_MarkLinePacked_GuardsAndMismatch(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// dispensed
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(10)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(10)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	_, err := repo.MarkLinePacked(context.Background(), 1)
	assert.EqualError(t, err, "cannot pack a line on a dispensed prescription")

	// mismatch
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(10)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(10)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`SELECT total_to_dispense FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"total_to_dispense"}).AddRow(10))
	mock.ExpectQuery(qm(`SELECT COALESCE(SUM(quantity),0) FROM prescription_batch_items WHERE line_id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"sum"}).AddRow(9))
	mock.ExpectRollback()
	_, err = repo.MarkLinePacked(context.Background(), 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "allocation sum 9 does not equal total_to_dispense 10")
}

// --- UnpackLine ------------------------------------------------------------

func Test_UnpackLine_Success_And_Guard(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// success
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(5)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(55)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(55)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectExec(qm(`
	  UPDATE prescription_lines
	  SET is_packed=FALSE, packed_by=NULL, packed_at=NULL, updated_at=NOW()
	  WHERE id=$1`)).
		WithArgs(int64(5)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// hydrate
	mock.ExpectQuery(qm(`FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at", "du",
		}).AddRow(int64(5), int64(55), int64(101), nil, 500, "mg", "day", 1, 3.0, 5.0, 9, false, nil, nil, "tab"))
	mock.ExpectQuery(qm(`FROM prescription_batch_items WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at"}))

	l, err := repo.UnpackLine(context.Background(), 5)
	assert.NoError(t, err)
	assert.False(t, l.IsPacked)

	// guard
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT prescription_id FROM prescription_lines WHERE id=$1`)).
		WithArgs(int64(6)).WillReturnRows(sqlmock.NewRows([]string{"prescription_id"}).AddRow(int64(55)))
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(55)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	_, err = repo.UnpackLine(context.Background(), 6)
	assert.EqualError(t, err, "cannot unpack a line on a dispensed prescription")
}

// --- DispensePrescription --------------------------------------------------

func Test_DispensePrescription_Success(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1 FOR UPDATE`)).
		WithArgs(int64(99)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1`)).
		WithArgs(int64(99)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1 AND is_packed=TRUE`)).
		WithArgs(int64(99)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery(qm(`
	  SELECT COUNT(*) FROM prescription_lines l
	  WHERE l.prescription_id=$1 AND
	        l.total_to_dispense <> COALESCE((SELECT SUM(quantity) FROM prescription_batch_items i WHERE i.line_id=l.id),0)`)).
		WithArgs(int64(99)).WillReturnRows(sqlmock.NewRows([]string{"m"}).AddRow(0))
	mock.ExpectExec(qm(`
	  UPDATE prescriptions
	  SET is_dispensed=TRUE, dispensed_by=$2, dispensed_at=NOW(), updated_at=NOW()
	  WHERE id=$1`)).
		WithArgs(int64(99), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	// hydrate
	mock.ExpectQuery(qm(`FROM prescriptions WHERE id=$1`)).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "patient_id", "vid", "notes",
			"created_by", "created_at", "updated_at",
			"is_dispensed", "dispensed_by", "dispensed_at",
		}).AddRow(int64(99), int64(10), int32(1), nil, nil, now, now, true, ip(7), tp(now)))
	mock.ExpectQuery(qm(`FROM prescription_lines pl
		WHERE pl.prescription_id = $1
		ORDER BY pl.id`)).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at",
			"drug_name", "route_code", "dispense_unit", "display_strength",
		}))

	p, err := repo.DispensePrescription(context.Background(), 99)
	assert.NoError(t, err)
	assert.True(t, p.IsDispensed)
}

func Test_DispensePrescription_GuardsAndErrors(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	// already dispensed
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1 FOR UPDATE`)).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(true))
	mock.ExpectRollback()
	_, err := repo.DispensePrescription(context.Background(), 1)
	assert.EqualError(t, err, "prescription already dispensed")

	// no lines
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1 FOR UPDATE`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1`)).
		WithArgs(int64(2)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(0))
	mock.ExpectRollback()
	_, err = repo.DispensePrescription(context.Background(), 2)
	assert.EqualError(t, err, "no lines to dispense")

	// not all packed
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1 FOR UPDATE`)).
		WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1`)).
		WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1 AND is_packed=TRUE`)).
		WithArgs(int64(3)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(1))
	mock.ExpectRollback()
	_, err = repo.DispensePrescription(context.Background(), 3)
	assert.EqualError(t, err, "all lines must be packed before dispense")

	// allocation mismatch
	mock.ExpectBegin()
	mock.ExpectQuery(qm(`SELECT is_dispensed FROM prescriptions WHERE id=$1 FOR UPDATE`)).
		WithArgs(int64(4)).WillReturnRows(sqlmock.NewRows([]string{"is_dispensed"}).AddRow(false))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1`)).
		WithArgs(int64(4)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery(qm(`SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1 AND is_packed=TRUE`)).
		WithArgs(int64(4)).WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(2))
	mock.ExpectQuery(qm(`
	  SELECT COUNT(*) FROM prescription_lines l
	  WHERE l.prescription_id=$1 AND
	        l.total_to_dispense <> COALESCE((SELECT SUM(quantity) FROM prescription_batch_items i WHERE i.line_id=l.id),0)`)).
		WithArgs(int64(4)).WillReturnRows(sqlmock.NewRows([]string{"m"}).AddRow(1))
	mock.ExpectRollback()
	_, err = repo.DispensePrescription(context.Background(), 4)
	assert.EqualError(t, err, "allocation totals mismatch")
}

// --- GetLine ---------------------------------------------------------------

func Test_GetLine_Success_NotFound_AllocErr(t *testing.T) {
	repo, mock, done := newRxRepo(t)
	defer done()

	now := mustNow()
	// success
	mock.ExpectQuery(qm(`
	  SELECT
	    pl.id, pl.prescription_id, pl.presentation_id, pl.remarks,
	    pl.dose_amount, pl.dose_unit,
	    pl.schedule_kind, pl.every_n, pl.frequency_per_schedule, pl.duration,
	    pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
	    (SELECT dispense_unit FROM drug_presentations WHERE id=pl.presentation_id) AS du
	  FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at", "du",
		}).AddRow(int64(5), int64(55), int64(101), "r", 1, "tab", "day", 1, 1.0, 1.0, 1, false, nil, nil, "tab"))
	mock.ExpectQuery(qm(`FROM prescription_batch_items WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "line_id", "batch_location_id", "quantity", "created_at", "updated_at",
		}).AddRow(int64(1), int64(5), int64(900), 1, now, now))
	l, err := repo.GetLine(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, "tab", l.DispenseUnit)
	assert.Len(t, l.Allocations, 1)

	// not found
	mock.ExpectQuery(qm(`FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(6)).
		WillReturnError(sql.ErrNoRows)
	l, err = repo.GetLine(context.Background(), 6)
	assert.Error(t, err)
	assert.Nil(t, l)
	assert.Equal(t, "line not found", err.Error())

	// alloc err
	mock.ExpectQuery(qm(`FROM prescription_lines pl
	  WHERE pl.id=$1`)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "prescription_id", "presentation_id", "remarks",
			"dose_amount", "dose_unit",
			"schedule_kind", "every_n", "frequency_per_schedule", "duration",
			"total_to_dispense", "is_packed", "packed_by", "packed_at", "du",
		}).AddRow(int64(7), int64(55), int64(101), nil, 1, "tab", "day", 1, 1.0, 1.0, 1, false, nil, nil, "tab"))
	mock.ExpectQuery(qm(`FROM prescription_batch_items WHERE line_id=$1 ORDER BY id`)).
		WithArgs(int64(7)).
		WillReturnError(errors.New("alloc err"))
	l, err = repo.GetLine(context.Background(), 7)
	assert.Error(t, err)
	assert.Nil(t, l)
}
