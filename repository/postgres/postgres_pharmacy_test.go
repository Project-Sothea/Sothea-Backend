package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/stretchr/testify/assert"
)

func mustNow() time.Time {
	return time.Unix(1_700_000_000, 0).UTC()
}
func tp(t time.Time) *time.Time { return &t }
func strPtr(s string) *string   { return &s }

// quick regexp helper for exact SQL strings
func qm(s string) string { return regexp.QuoteMeta(s) }

func newRepoWithMock(t *testing.T) (*postgresPharmacyRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	repo := &postgresPharmacyRepository{Conn: db}
	cleanup := func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}
	return repo, mock, cleanup
}

// -----------------------------------------------------------------------------
// DRUGS
// -----------------------------------------------------------------------------

func Test_ListDrugs_NoQuery(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	rows := sqlmock.NewRows([]string{
		"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
	}).
		AddRow(int64(1), "Paracetamol", "Panadol", "N02BE01", "pain", true, now, now).
		AddRow(int64(2), "Ibuprofen", nil, "M01AE01", nil, true, now, now)

	mock.ExpectQuery(qm(qDrugsList)).WillReturnRows(rows)

	got, err := repo.ListDrugs(context.Background(), nil)
	assert.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "Paracetamol", got[0].GenericName)
	assert.Equal(t, "Ibuprofen", got[1].GenericName)
}

func Test_ListDrugs_WithQuery(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	q := "parac"
	rows := sqlmock.NewRows([]string{
		"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
	}).AddRow(int64(1), "Paracetamol", "Panadol", "N02BE01", "pain", true, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		  SELECT id, generic_name, brand_name, atc_code, notes, is_active, created_at, updated_at
		  FROM drugs
		  WHERE generic_name ILIKE $1 OR COALESCE(brand_name,'') ILIKE $1
		  ORDER BY generic_name, COALESCE(brand_name,'')`)).
		WithArgs("%" + q + "%").
		WillReturnRows(rows)

	got, err := repo.ListDrugs(context.Background(), &q)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "Paracetamol", got[0].GenericName)
}

func Test_ListDrugs_QueryErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectQuery(qm(qDrugsList)).WillReturnError(errors.New("query err"))
	got, err := repo.ListDrugs(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func Test_CreateDrug_Success(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	d := &entities.Drug{GenericName: "Paracetamol", IsActive: true}
	// INSERT ... RETURNING id
	mock.ExpectQuery(qm(qDrugCreate)).
		WithArgs(d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))
	// GetDrug(id)
	now := mustNow()
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(10), "Paracetamol", nil, nil, nil, true, now, now))

	got, err := repo.CreateDrug(context.Background(), d)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), got.ID)
}

func Test_CreateDrug_InsertErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	d := &entities.Drug{GenericName: "X"}
	mock.ExpectQuery(qm(qDrugCreate)).
		WithArgs(d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive).
		WillReturnError(errors.New("db error"))

	got, err := repo.CreateDrug(context.Background(), d)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func Test_GetDrug_Found_And_NotFound(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(7), "Aspirin", nil, nil, nil, true, now, now))
	got, err := repo.GetDrug(context.Background(), 7)
	assert.NoError(t, err)
	assert.Equal(t, int64(7), got.ID)

	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(8)).
		WillReturnError(sql.ErrNoRows)
	got, err = repo.GetDrug(context.Background(), 8)
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, "drug not found", err.Error())
}

func Test_UpdateDrug_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	d := &entities.Drug{ID: 5, GenericName: "Paracetamol", IsActive: true}
	// success
	mock.ExpectExec(qm(qDrugUpdate)).
		WithArgs(d.ID, d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive).
		WillReturnResult(sqlmock.NewResult(0, 1))
	now := mustNow()
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "Paracetamol", nil, nil, nil, true, now, now))
	got, err := repo.UpdateDrug(context.Background(), d)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), got.ID)

	// not found
	mock.ExpectExec(qm(qDrugUpdate)).
		WithArgs(d.ID, d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive).
		WillReturnResult(sqlmock.NewResult(0, 0))
	got, err = repo.UpdateDrug(context.Background(), d)
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, "drug not found", err.Error())

	// exec err
	mock.ExpectExec(qm(qDrugUpdate)).
		WithArgs(d.ID, d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive).
		WillReturnError(errors.New("exec err"))
	got, err = repo.UpdateDrug(context.Background(), d)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func Test_DeleteDrug_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectExec(qm(qDrugDelete)).
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteDrug(context.Background(), 9))

	mock.ExpectExec(qm(qDrugDelete)).
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.DeleteDrug(context.Background(), 10)
	assert.Error(t, err)
	assert.Equal(t, "drug not found", err.Error())

	mock.ExpectExec(qm(qDrugDelete)).
		WithArgs(int64(11)).
		WillReturnError(errors.New("exec err"))
	err = repo.DeleteDrug(context.Background(), 11)
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// PRESENTATIONS
// -----------------------------------------------------------------------------

func Test_ListPresentations_ComposesViewAndLabels(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	// GetDrug first
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "Paracetamol", "Panadol", "N02BE01", nil, true, now, now))

	// List presentations
	rows := sqlmock.NewRows([]string{
		"id", "drug_id", "dosage_form_code", "route_code",
		"strength_num", "strength_unit_num",
		"strength_den", "strength_unit_den",
		"dispense_unit", "piece_content_amount", "piece_content_unit",
		"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
	}).
		// solid: 500 mg TAB PO
		AddRow(int64(101), int64(5), "TAB", "PO",
			int64(500), "mg",
			nil, nil,
			"tab", nil, nil,
			false, "123", "solid", now, now).
		// liquid bottle: 250 mg/5 mL SYR PO, bottle 100 mL
		AddRow(int64(102), int64(5), "SYR", "PO",
			int64(250), "mg",
			int64(5), "mL",
			"bottle", int64(100), "mL",
			false, "456", "liquid", now, now)

	mock.ExpectQuery(qm(qPresList)).
		WithArgs(int64(5)).
		WillReturnRows(rows)

	out, err := repo.ListPresentations(context.Background(), 5)
	assert.NoError(t, err)
	assert.Len(t, out, 2)

	solid := out[0]
	assert.Equal(t, "Paracetamol", solid.DrugName)
	assert.Equal(t, "500 mg TAB", solid.DisplayStrength)
	assert.Equal(t, "PO", solid.DisplayRoute)
	assert.Contains(t, solid.DisplayLabel, "Paracetamol 500 mg TAB (PO)")

	liquid := out[1]
	assert.Equal(t, "250 mg/5 mL SYR", liquid.DisplayStrength)
	assert.Contains(t, liquid.DisplayLabel, "Paracetamol 250 mg/5 mL SYR (PO) - bottle 100 mL")
}

func Test_ListPresentations_GetDrugErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnError(sql.ErrNoRows)

	out, err := repo.ListPresentations(context.Background(), 5)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_ListPresentations_QueryErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "Paracetamol", nil, nil, nil, true, now, now))

	mock.ExpectQuery(qm(qPresList)).WithArgs(int64(5)).WillReturnError(errors.New("pres q err"))

	out, err := repo.ListPresentations(context.Background(), 5)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_GetPresentation_Success_NotFound_GetDrugErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	// found
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "drug_id", "dosage_form_code", "route_code",
			"strength_num", "strength_unit_num",
			"strength_den", "strength_unit_den",
			"dispense_unit", "piece_content_amount", "piece_content_unit",
			"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
		}).AddRow(int64(11), int64(2), "TAB", "PO",
			int64(200), "mg",
			nil, nil,
			"tab", nil, nil,
			false, "x", "n", now, now))

	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(2), "Ibuprofen", nil, nil, nil, true, now, now))

	v, err := repo.GetPresentation(context.Background(), 11)
	assert.NoError(t, err)
	assert.Equal(t, int64(11), v.ID)
	assert.Equal(t, "200 mg TAB", v.DisplayStrength)
	assert.Equal(t, "Ibuprofen 200 mg TAB (PO)", v.DisplayLabel)

	// not found
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)
	v, err = repo.GetPresentation(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, v)
	assert.Equal(t, "presentation not found", err.Error())

	// GetDrug err after reading presentation
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(12)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "drug_id", "dosage_form_code", "route_code",
			"strength_num", "strength_unit_num",
			"strength_den", "strength_unit_den",
			"dispense_unit", "piece_content_amount", "piece_content_unit",
			"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
		}).AddRow(int64(12), int64(5), "TAB", "PO", nil, nil, nil, nil, "tab", nil, nil, false, nil, nil, now, now))
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnError(errors.New("drug err"))
	v, err = repo.GetPresentation(context.Background(), 12)
	assert.Error(t, err)
	assert.Nil(t, v)
}

func Test_CreatePresentation_Success_InsertErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	p := &entities.DrugPresentation{DrugID: 5, DosageFormCode: "TAB", RouteCode: "PO"}

	// insert ok
	mock.ExpectQuery(qm(qPresCreate)).
		WithArgs(p.DrugID, p.DosageFormCode, p.RouteCode,
			p.StrengthNum, p.StrengthUnitNum,
			p.StrengthDen, p.StrengthUnitDen,
			p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
			p.IsFractionalAllowed, p.Barcode, p.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(123)))

	// hydrate via GetPresentation -> then GetDrug
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(123)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "drug_id", "dosage_form_code", "route_code",
			"strength_num", "strength_unit_num",
			"strength_den", "strength_unit_den",
			"dispense_unit", "piece_content_amount", "piece_content_unit",
			"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
		}).AddRow(int64(123), int64(5), "TAB", "PO",
			nil, nil, nil, nil, "tab", nil, nil, false, nil, nil, now, now))

	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "Paracetamol", nil, nil, nil, true, now, now))

	v, err := repo.CreatePresentation(context.Background(), p)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), v.ID)

	// insert error
	mock.ExpectQuery(qm(qPresCreate)).
		WithArgs(p.DrugID, p.DosageFormCode, p.RouteCode,
			p.StrengthNum, p.StrengthUnitNum,
			p.StrengthDen, p.StrengthUnitDen,
			p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
			p.IsFractionalAllowed, p.Barcode, p.Notes).
		WillReturnError(errors.New("ins err"))

	v, err = repo.CreatePresentation(context.Background(), p)
	assert.Error(t, err)
	assert.Nil(t, v)
}

func Test_UpdatePresentation_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	p := &entities.DrugPresentation{ID: 22, DrugID: 5}
	mock.ExpectExec(qm(qPresUpdate)).
		WithArgs(p.ID, p.DrugID, p.DosageFormCode, p.RouteCode,
			p.StrengthNum, p.StrengthUnitNum, p.StrengthDen, p.StrengthUnitDen,
			p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
			p.IsFractionalAllowed, p.Barcode, p.Notes).
		WillReturnResult(sqlmock.NewResult(0, 1))

	now := mustNow()
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(22)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "drug_id", "dosage_form_code", "route_code",
			"strength_num", "strength_unit_num",
			"strength_den", "strength_unit_den",
			"dispense_unit", "piece_content_amount", "piece_content_unit",
			"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
		}).AddRow(int64(22), int64(5), "TAB", "PO", nil, nil, nil, nil, "tab", nil, nil, false, nil, nil, now, now))
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "PCM", nil, nil, nil, true, now, now))

	v, err := repo.UpdatePresentation(context.Background(), p)
	assert.NoError(t, err)
	assert.Equal(t, int64(22), v.ID)

	// not found
	mock.ExpectExec(qm(qPresUpdate)).
		WithArgs(p.ID, p.DrugID, p.DosageFormCode, p.RouteCode,
			p.StrengthNum, p.StrengthUnitNum, p.StrengthDen, p.StrengthUnitDen,
			p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
			p.IsFractionalAllowed, p.Barcode, p.Notes).
		WillReturnResult(sqlmock.NewResult(0, 0))
	v, err = repo.UpdatePresentation(context.Background(), p)
	assert.Error(t, err)
	assert.Nil(t, v)
	assert.Equal(t, "presentation not found", err.Error())

	// exec err
	mock.ExpectExec(qm(qPresUpdate)).
		WithArgs(p.ID, p.DrugID, p.DosageFormCode, p.RouteCode,
			p.StrengthNum, p.StrengthUnitNum, p.StrengthDen, p.StrengthUnitDen,
			p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
			p.IsFractionalAllowed, p.Barcode, p.Notes).
		WillReturnError(errors.New("exec err"))
	v, err = repo.UpdatePresentation(context.Background(), p)
	assert.Error(t, err)
	assert.Nil(t, v)
}

func Test_DeletePresentation_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectExec(qm(qPresDelete)).
		WithArgs(int64(44)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeletePresentation(context.Background(), 44))

	mock.ExpectExec(qm(qPresDelete)).
		WithArgs(int64(45)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.DeletePresentation(context.Background(), 45)
	assert.Error(t, err)
	assert.Equal(t, "presentation not found", err.Error())

	mock.ExpectExec(qm(qPresDelete)).
		WithArgs(int64(46)).
		WillReturnError(errors.New("exec err"))
	err = repo.DeletePresentation(context.Background(), 46)
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// BATCHES
// -----------------------------------------------------------------------------

func Test_ListBatches_NoRows(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectQuery(qm(qBatchList)).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
		}))
	out, err := repo.ListBatches(context.Background(), 7)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Len(t, out, 0)
}

func Test_ListBatches_WithLocations(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	// batches
	brows := sqlmock.NewRows([]string{
		"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
	}).AddRow(int64(1), int64(5), "B1", tp(now), "ACME", 10, now, now).
		AddRow(int64(2), int64(5), "B2", nil, nil, 5, now, now)
	mock.ExpectQuery(qm(qBatchList)).
		WithArgs(int64(5)).
		WillReturnRows(brows)

	// locations for ANY($1)
	lrows := sqlmock.NewRows([]string{
		"id", "batch_id", "location", "quantity", "created_at", "updated_at",
	}).
		AddRow(int64(11), int64(1), "Main", 6, now, now).
		AddRow(int64(12), int64(1), "Cabinet", 4, now, now).
		AddRow(int64(21), int64(2), "Main", 5, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, batch_id, location, quantity, created_at, updated_at
		FROM batch_locations
		WHERE batch_id = ANY($1)
		ORDER BY batch_id, location, id`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(lrows)

	out, err := repo.ListBatches(context.Background(), 5)
	assert.NoError(t, err)
	assert.Len(t, out, 2)
	assert.Len(t, out[0].BatchLocations, 2)
	assert.Len(t, out[1].BatchLocations, 1)
}

func Test_ListBatches_QueryErr_And_LocQueryErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	// first query err
	mock.ExpectQuery(qm(qBatchList)).
		WithArgs(int64(5)).
		WillReturnError(errors.New("list err"))
	out, err := repo.ListBatches(context.Background(), 5)
	assert.Error(t, err)
	assert.Nil(t, out)

	// batches ok, loc query err
	now := mustNow()
	brows := sqlmock.NewRows([]string{
		"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
	}).AddRow(int64(1), int64(5), "B1", tp(now), nil, 1, now, now)
	mock.ExpectQuery(qm(qBatchList)).
		WithArgs(int64(5)).
		WillReturnRows(brows)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, batch_id, location, quantity, created_at, updated_at
		FROM batch_locations
		WHERE batch_id = ANY($1)
		ORDER BY batch_id, location, id`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(errors.New("loc err"))
	out, err = repo.ListBatches(context.Background(), 5)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_GetBatch_Success_NotFound_ListLocsErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	// Success
	mock.ExpectQuery(qm(qBatchGet)).
		WithArgs(int64(88)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
		}).AddRow(int64(88), int64(5), "B-001", tp(now), "ACME", 12, now, now))
	// ListBatchLocations called inside
	lrows := sqlmock.NewRows([]string{
		"id", "batch_id", "location", "quantity", "created_at", "updated_at",
	}).AddRow(int64(1), int64(88), "Main", 12, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`)).
		WithArgs(int64(88)).
		WillReturnRows(lrows)
	d, err := repo.GetBatch(context.Background(), 88)
	assert.NoError(t, err)
	assert.Equal(t, int64(88), d.ID)

	// Not Found
	mock.ExpectQuery(qm(qBatchGet)).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)
	d, err = repo.GetBatch(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Equal(t, "batch not found", err.Error())

	// List locations error
	mock.ExpectQuery(qm(qBatchGet)).
		WithArgs(int64(100)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
		}).AddRow(int64(100), int64(5), "B-002", nil, nil, 3, now, now))
	mock.ExpectQuery(regexp.QuoteMeta(`
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`)).
		WithArgs(int64(100)).
		WillReturnError(errors.New("loc list err"))
	d, err = repo.GetBatch(context.Background(), 100)
	assert.Error(t, err)
	assert.Nil(t, d)
}

func Test_CreateBatch_Tx_Success_WithLocations(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()

	mock.ExpectBegin()
	// insert batch
	mock.ExpectQuery(qm(qBatchCreate)).
		WithArgs(int64(5), "B-001", tp(now), "ACME", 100).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(123)))

	prep := mock.ExpectPrepare(regexp.QuoteMeta(`
	INSERT INTO batch_locations (batch_id, location, quantity)
	VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`))

	// 1st execution: ("Main", 60)
	prep.ExpectQuery().
		WithArgs(int64(123), "Main", 60).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(1), now, now))

	// 2nd execution: ("Cabinet A", 40)
	prep.ExpectQuery().
		WithArgs(int64(123), "Cabinet A", 40).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(2), now, now))

	mock.ExpectCommit()

	// After commit, GetBatch + locations
	mock.ExpectQuery(qm(qBatchGet)).
		WithArgs(int64(123)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
		}).AddRow(int64(123), int64(5), "B-001", tp(now), "ACME", 100, now, now))
	mock.ExpectQuery(regexp.QuoteMeta(`
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`)).
		WithArgs(int64(123)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "batch_id", "location", "quantity", "created_at", "updated_at",
		}).AddRow(int64(1), int64(123), "Main", 60, now, now).
			AddRow(int64(2), int64(123), "Cabinet A", 40, now, now))

	b := &entities.DrugBatch{
		PresentationID: 5, BatchNumber: "B-001", ExpiryDate: tp(now), Supplier: strPtr("ACME"), Quantity: 100,
	}
	locs := []entities.DrugBatchLocation{{Location: "Main", Quantity: 60}, {Location: "Cabinet A", Quantity: 40}}
	out, err := repo.CreateBatch(context.Background(), b, locs)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), out.ID)
	assert.Len(t, out.BatchLocations, 2)
}

func Test_CreateBatch_Tx_NegativeLocation_Rollback(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	mock.ExpectBegin()
	mock.ExpectQuery(qm(qBatchCreate)).
		WithArgs(int64(5), "B-001", tp(now), "ACME", 10).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(50)))
	// expect prepare; code prepares before checking per-location quantity
	mock.ExpectPrepare(regexp.QuoteMeta(`
		  INSERT INTO batch_locations (batch_id, location, quantity)
		  VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`))
	mock.ExpectRollback()

	b := &entities.DrugBatch{PresentationID: 5, BatchNumber: "B-001", ExpiryDate: tp(now), Supplier: strPtr("ACME"), Quantity: 10}
	locs := []entities.DrugBatchLocation{{Location: "Main", Quantity: -1}} // triggers error
	out, err := repo.CreateBatch(context.Background(), b, locs)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Contains(t, err.Error(), "negative quantity")
}

func Test_CreateBatch_Tx_InsertBatchErr_Rollback(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(qBatchCreate)).
		WithArgs(int64(5), "B-001", (*time.Time)(nil), (*string)(nil), 0).
		WillReturnError(errors.New("ins err"))
	mock.ExpectRollback()

	b := &entities.DrugBatch{PresentationID: 5, BatchNumber: "B-001"}
	out, err := repo.CreateBatch(context.Background(), b, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_CreateBatch_Tx_LocationInsertErr_Rollback(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(qBatchCreate)).
		WithArgs(int64(5), "B-001", (*time.Time)(nil), (*string)(nil), 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(90)))

	mock.ExpectPrepare(regexp.QuoteMeta(`
		  INSERT INTO batch_locations (batch_id, location, quantity)
		  VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`)).
		ExpectQuery().
		WithArgs(int64(90), "Main", 10).
		WillReturnError(errors.New("loc err"))

	mock.ExpectRollback()

	b := &entities.DrugBatch{PresentationID: 5, BatchNumber: "B-001"}
	locs := []entities.DrugBatchLocation{{Location: "Main", Quantity: 10}}
	out, err := repo.CreateBatch(context.Background(), b, locs)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_CreateBatch_Tx_CommitErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectBegin()
	mock.ExpectQuery(qm(qBatchCreate)).
		WithArgs(int64(5), "B-001", (*time.Time)(nil), (*string)(nil), 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	// no locations
	mock.ExpectCommit().WillReturnError(errors.New("commit fail"))

	b := &entities.DrugBatch{PresentationID: 5, BatchNumber: "B-001"}
	out, err := repo.CreateBatch(context.Background(), b, nil)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_UpdateBatch_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	b := &entities.DrugBatch{ID: 42, PresentationID: 5, BatchNumber: "B-001", Quantity: 12}
	mock.ExpectExec(qm(qBatchUpdate)).
		WithArgs(b.ID, b.PresentationID, b.BatchNumber, b.ExpiryDate, b.Supplier, b.Quantity).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// hydrate
	mock.ExpectQuery(qm(qBatchGet)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
		}).AddRow(int64(42), int64(5), "B-001", (*time.Time)(nil), (*string)(nil), 12, now, now))
	mock.ExpectQuery(regexp.QuoteMeta(`
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`)).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "batch_id", "location", "quantity", "created_at", "updated_at"}))

	out, err := repo.UpdateBatch(context.Background(), b)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), out.ID)

	// not found
	mock.ExpectExec(qm(qBatchUpdate)).
		WithArgs(b.ID, b.PresentationID, b.BatchNumber, b.ExpiryDate, b.Supplier, b.Quantity).
		WillReturnResult(sqlmock.NewResult(0, 0))
	out, err = repo.UpdateBatch(context.Background(), b)
	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Equal(t, "batch not found", err.Error())

	// exec err
	mock.ExpectExec(qm(qBatchUpdate)).
		WithArgs(b.ID, b.PresentationID, b.BatchNumber, b.ExpiryDate, b.Supplier, b.Quantity).
		WillReturnError(errors.New("exec err"))
	out, err = repo.UpdateBatch(context.Background(), b)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_DeleteBatch_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectExec(qm(qBatchDelete)).
		WithArgs(int64(55)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteBatch(context.Background(), 55))

	mock.ExpectExec(qm(qBatchDelete)).
		WithArgs(int64(56)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.DeleteBatch(context.Background(), 56)
	assert.Error(t, err)
	assert.Equal(t, "batch not found", err.Error())

	mock.ExpectExec(qm(qBatchDelete)).
		WithArgs(int64(57)).
		WillReturnError(errors.New("exec err"))
	err = repo.DeleteBatch(context.Background(), 57)
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// LOCATIONS
// -----------------------------------------------------------------------------

func Test_ListBatchLocations_Success_QueryErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	rows := sqlmock.NewRows([]string{
		"id", "batch_id", "location", "quantity", "created_at", "updated_at",
	}).AddRow(int64(1), int64(10), "Main", 7, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`)).
		WithArgs(int64(10)).WillReturnRows(rows)

	out, err := repo.ListBatchLocations(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, out, 1)

	// query err
	mock.ExpectQuery(regexp.QuoteMeta(`
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`)).
		WithArgs(int64(99)).WillReturnError(errors.New("q err"))
	out, err = repo.ListBatchLocations(context.Background(), 99)
	assert.Error(t, err)
	assert.Nil(t, out)
}

func Test_CreateBatchLocation_Success_InsertErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	loc := &entities.DrugBatchLocation{BatchID: 99, Location: "Main", Quantity: 3}
	mock.ExpectQuery(qm(qLocCreate)).
		WithArgs(loc.BatchID, loc.Location, loc.Quantity).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(500), now, now))
	// hydrate
	mock.ExpectQuery(qm(qLocGet)).
		WithArgs(int64(500)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "batch_id", "location", "quantity", "created_at", "updated_at",
		}).AddRow(int64(500), int64(99), "Main", 3, now, now))

	got, err := repo.CreateBatchLocation(context.Background(), loc)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), got.ID)

	// insert err
	mock.ExpectQuery(qm(qLocCreate)).
		WithArgs(loc.BatchID, loc.Location, loc.Quantity).
		WillReturnError(errors.New("loc ins err"))
	got, err = repo.CreateBatchLocation(context.Background(), loc)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func Test_GetBatchLocation_Found_NotFound(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	mock.ExpectQuery(qm(qLocGet)).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "batch_id", "location", "quantity", "created_at", "updated_at",
		}).AddRow(int64(1), int64(10), "Main", 2, now, now))
	l, err := repo.GetBatchLocation(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), l.ID)

	mock.ExpectQuery(qm(qLocGet)).
		WithArgs(int64(2)).
		WillReturnError(sql.ErrNoRows)
	l, err = repo.GetBatchLocation(context.Background(), 2)
	assert.Error(t, err)
	assert.Nil(t, l)
	assert.Equal(t, "batch location not found", err.Error())
}

func Test_UpdateBatchLocation_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	loc := &entities.DrugBatchLocation{ID: 77, BatchID: 10, Location: "A", Quantity: 5}
	mock.ExpectExec(qm(qLocUpdate)).
		WithArgs(loc.ID, loc.BatchID, loc.Location, loc.Quantity).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// hydrate
	now := mustNow()
	mock.ExpectQuery(qm(qLocGet)).
		WithArgs(int64(77)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "batch_id", "location", "quantity", "created_at", "updated_at",
		}).AddRow(int64(77), int64(10), "A", 5, now, now))

	got, err := repo.UpdateBatchLocation(context.Background(), loc)
	assert.NoError(t, err)
	assert.Equal(t, int64(77), got.ID)

	// not found
	mock.ExpectExec(qm(qLocUpdate)).
		WithArgs(loc.ID, loc.BatchID, loc.Location, loc.Quantity).
		WillReturnResult(sqlmock.NewResult(0, 0))
	got, err = repo.UpdateBatchLocation(context.Background(), loc)
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, "batch location not found", err.Error())

	// exec err
	mock.ExpectExec(qm(qLocUpdate)).
		WithArgs(loc.ID, loc.BatchID, loc.Location, loc.Quantity).
		WillReturnError(errors.New("exec err"))
	got, err = repo.UpdateBatchLocation(context.Background(), loc)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func Test_DeleteBatchLocation_Success_NotFound_ExecErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	mock.ExpectExec(qm(qLocDelete)).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteBatchLocation(context.Background(), 999))

	mock.ExpectExec(qm(qLocDelete)).
		WithArgs(int64(1000)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.DeleteBatchLocation(context.Background(), 1000)
	assert.Error(t, err)
	assert.Equal(t, "location not found", err.Error())

	mock.ExpectExec(qm(qLocDelete)).
		WithArgs(int64(1001)).
		WillReturnError(errors.New("exec err"))
	err = repo.DeleteBatchLocation(context.Background(), 1001)
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// STOCK VIEW (FEFO sorting + totals)
// -----------------------------------------------------------------------------

func Test_GetPresentationStock_Composes_And_Sorts_FEFO(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	now := mustNow()
	// GetPresentation(id) -> returns pres with drug_id=5
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(77)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "drug_id", "dosage_form_code", "route_code",
			"strength_num", "strength_unit_num",
			"strength_den", "strength_unit_den",
			"dispense_unit", "piece_content_amount", "piece_content_unit",
			"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
		}).AddRow(int64(77), int64(5), "TAB", "PO", nil, nil, nil, nil, "tab", nil, nil, false, nil, nil, now, now))
	// GetDrug for display fields
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "Paracetamol", nil, nil, nil, true, now, now))

	// ListBatches -> 3 batches: earlier, later, nil
	earlier := now.AddDate(0, 1, 0)
	later := now.AddDate(0, 2, 0)
	brows := sqlmock.NewRows([]string{
		"id", "presentation_id", "batch_number", "expiry_date", "supplier", "quantity", "created_at", "updated_at",
	}).
		AddRow(int64(1), int64(77), "A", tp(earlier), nil, 5, now, now).
		AddRow(int64(2), int64(77), "B", tp(later), nil, 7, now, now).
		AddRow(int64(3), int64(77), "C", nil, nil, 11, now, now)
	mock.ExpectQuery(qm(qBatchList)).
		WithArgs(int64(77)).
		WillReturnRows(brows)

	// Locations for ANY($1)
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, batch_id, location, quantity, created_at, updated_at
		FROM batch_locations
		WHERE batch_id = ANY($1)
		ORDER BY batch_id, location, id`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "batch_id", "location", "quantity", "created_at", "updated_at",
		}))

	ps, err := repo.GetPresentationStock(context.Background(), 77)
	assert.NoError(t, err)
	assert.Equal(t, "Paracetamol", ps.Presentation.DrugName)
	// FEFO: earlier expiry first, then later, then nil
	assert.Equal(t, "A", ps.Batches[0].BatchNumber)
	assert.Equal(t, "B", ps.Batches[1].BatchNumber)
	assert.Equal(t, "C", ps.Batches[2].BatchNumber)
	// Total
	assert.Equal(t, 5+7+11, ps.TotalQty)
}

func Test_GetPresentationStock_PresErr_And_BatchesErr(t *testing.T) {
	repo, mock, done := newRepoWithMock(t)
	defer done()

	// pres err
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(10)).
		WillReturnError(errors.New("pres err"))
	ps, err := repo.GetPresentationStock(context.Background(), 10)
	assert.Error(t, err)
	assert.Nil(t, ps)

	// pres ok, drug ok, batches err
	now := mustNow()
	mock.ExpectQuery(qm(qPresGet)).
		WithArgs(int64(11)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "drug_id", "dosage_form_code", "route_code",
			"strength_num", "strength_unit_num",
			"strength_den", "strength_unit_den",
			"dispense_unit", "piece_content_amount", "piece_content_unit",
			"is_fractional_allowed", "barcode", "notes", "created_at", "updated_at",
		}).AddRow(int64(11), int64(5), "TAB", "PO", nil, nil, nil, nil, "tab", nil, nil, false, nil, nil, now, now))
	mock.ExpectQuery(qm(qDrugGet)).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "generic_name", "brand_name", "atc_code", "notes", "is_active", "created_at", "updated_at",
		}).AddRow(int64(5), "X", nil, nil, nil, true, now, now))
	mock.ExpectQuery(qm(qBatchList)).
		WithArgs(int64(11)).
		WillReturnError(errors.New("batches err"))
	ps, err = repo.GetPresentationStock(context.Background(), 11)
	assert.Error(t, err)
	assert.Nil(t, ps)
}

// -----------------------------------------------------------------------------
// Display helpers (direct unit tests for edges)
// -----------------------------------------------------------------------------

func Test_displayHelpers(t *testing.T) {
	// solid with strength
	p1 := entities.DrugPresentation{
		DosageFormCode: "TAB", RouteCode: "PO",
		StrengthNum: floatPtr(500), StrengthUnitNum: strPtr("mg"),
		DispenseUnit: "tab",
	}
	assert.Equal(t, "500 mg TAB", displayStrength(p1))
	assert.Equal(t, "PCM 500 mg TAB (PO)", displayLabel("PCM", p1))

	// liquid with denominator
	p2 := entities.DrugPresentation{
		DosageFormCode: "SYR", RouteCode: "PO",
		StrengthNum: floatPtr(250), StrengthUnitNum: strPtr("mg"),
		StrengthDen: floatPtr(5), StrengthUnitDen: strPtr("mL"),
		DispenseUnit: "mL",
	}
	assert.Equal(t, "250 mg/5 mL SYR", displayStrength(p2))
	assert.Equal(t, "PCM 250 mg/5 mL SYR (PO)", displayLabel("PCM", p2))

	// bottle with piece content
	p3 := p2
	p3.DispenseUnit = "bottle"
	p3.PieceContentAmount = floatPtr(100)
	p3.PieceContentUnit = strPtr("mL")
	assert.Equal(t, "PCM 250 mg/5 mL SYR (PO) - bottle 100 mL", displayLabel("PCM", p3))

	// missing strength fields -> fallback to dosage form code
	p4 := entities.DrugPresentation{DosageFormCode: "TAB", RouteCode: "PO", DispenseUnit: "tab"}
	assert.Equal(t, "TAB", displayStrength(p4))
}

func intPtr(i int) *int           { return &i }
func floatPtr(i float64) *float64 { return &i }
