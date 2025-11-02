package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/lib/pq"
)

// -----------------------------------------------------------------------------
//  STRUCT + CONSTRUCTOR
// -----------------------------------------------------------------------------

type postgresPharmacyRepository struct {
	Conn *sql.DB
}

func NewPostgresPharmacyRepository(conn *sql.DB) entities.PharmacyRepository {
	return &postgresPharmacyRepository{Conn: conn}
}

// -----------------------------------------------------------------------------
//  HELPERS (label builders; keep in repo so FE gets nice strings)
// -----------------------------------------------------------------------------

func displayStrength(p entities.DrugPresentation) string {
	if p.StrengthDen == nil || p.StrengthUnitDen == nil {
		// solids (no denominator), e.g. "500 mg TAB"
		if p.StrengthNum != nil && p.StrengthUnitNum != nil {
			return fmt.Sprintf("%d %s %s", *p.StrengthNum, *p.StrengthUnitNum, p.DosageFormCode)
		}
		return p.DosageFormCode
	}
	// liquids/creams, e.g. "250 mg/5 mL SYR"
	return fmt.Sprintf("%d %s/%d %s %s",
		derefInt(p.StrengthNum), derefStr(p.StrengthUnitNum),
		derefInt(p.StrengthDen), derefStr(p.StrengthUnitDen),
		p.DosageFormCode)
}

func displayLabel(drugName string, p entities.DrugPresentation) string {
	base := fmt.Sprintf("%s %s (%s)", drugName, displayStrength(p), p.RouteCode)
	if p.DispenseUnit == "bottle" && p.PieceContentAmount != nil && p.PieceContentUnit != nil {
		return fmt.Sprintf("%s - bottle %d %s", base, *p.PieceContentAmount, *p.PieceContentUnit)
	}
	return base
}

func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// -----------------------------------------------------------------------------
//  DRUGS
// -----------------------------------------------------------------------------

const qDrugsList = `
  SELECT id, generic_name, brand_name, atc_code, notes, is_active, created_at, updated_at
  FROM drugs
  /* optional WHERE added dynamically */
  ORDER BY generic_name, COALESCE(brand_name,'')`

func (r *postgresPharmacyRepository) ListDrugs(ctx context.Context, q *string) ([]entities.Drug, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	query := qDrugsList
	args := []any{}
	if q != nil && *q != "" {
		query = `
		  SELECT id, generic_name, brand_name, atc_code, notes, is_active, created_at, updated_at
		  FROM drugs
		  WHERE generic_name ILIKE $1 OR COALESCE(brand_name,'') ILIKE $1
		  ORDER BY generic_name, COALESCE(brand_name,'')`
		args = append(args, "%"+*q+"%")
	}

	rows, err := dbx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Drug{}
	for rows.Next() {
		var d entities.Drug
		if err := rows.Scan(&d.ID, &d.GenericName, &d.BrandName, &d.ATCCode, &d.Notes, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

const qDrugCreate = `
  INSERT INTO drugs (generic_name, brand_name, atc_code, notes, is_active)
  VALUES ($1,$2,$3,$4,COALESCE($5,TRUE))
  RETURNING id`

func (r *postgresPharmacyRepository) CreateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var id int64
	if err := dbx.QueryRowContext(ctx, qDrugCreate,
		d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive,
	).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetDrug(ctx, id)
}

const qDrugGet = `
  SELECT id, generic_name, brand_name, atc_code, notes, is_active, created_at, updated_at
  FROM drugs WHERE id=$1`

func (r *postgresPharmacyRepository) GetDrug(ctx context.Context, id int64) (*entities.Drug, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var d entities.Drug
	err := dbx.QueryRowContext(ctx, qDrugGet, id).
		Scan(&d.ID, &d.GenericName, &d.BrandName, &d.ATCCode, &d.Notes, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("drug not found")
	}
	return &d, err
}

const qDrugUpdate = `
  UPDATE drugs SET generic_name=$2, brand_name=$3, atc_code=$4, notes=$5, is_active=$6, updated_at=NOW()
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qDrugUpdate,
		d.ID, d.GenericName, d.BrandName, d.ATCCode, d.Notes, d.IsActive)
	if err != nil {
		return nil, err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return nil, errors.New("drug not found")
	}
	return r.GetDrug(ctx, d.ID)
}

const qDrugDelete = `DELETE FROM drugs WHERE id=$1`

func (r *postgresPharmacyRepository) DeleteDrug(ctx context.Context, id int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qDrugDelete, id)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return errors.New("drug not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
//  PRESENTATIONS
// -----------------------------------------------------------------------------

const qPresList = `
  SELECT
    id, drug_id, dosage_form_code, route_code,
    strength_num, strength_unit_num,
    strength_den, strength_unit_den,
    dispense_unit, piece_content_amount, piece_content_unit,
    is_fractional_allowed, barcode, notes, created_at, updated_at
  FROM drug_presentations
  WHERE drug_id=$1
  ORDER BY dosage_form_code, route_code, dispense_unit, id`

func (r *postgresPharmacyRepository) ListPresentations(ctx context.Context, drugID int64) ([]entities.DrugPresentationView, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	drug, err := r.GetDrug(ctx, drugID)
	if err != nil {
		return nil, err
	}

	rows, err := dbx.QueryContext(ctx, qPresList, drugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.DrugPresentationView{}
	for rows.Next() {
		var p entities.DrugPresentation
		if err := rows.Scan(
			&p.ID, &p.DrugID, &p.DosageFormCode, &p.RouteCode,
			&p.StrengthNum, &p.StrengthUnitNum,
			&p.StrengthDen, &p.StrengthUnitDen,
			&p.DispenseUnit, &p.PieceContentAmount, &p.PieceContentUnit,
			&p.IsFractionalAllowed, &p.Barcode, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, entities.DrugPresentationView{
			DrugPresentation: p,
			DrugName:         drug.GenericName,
			DisplayStrength:  displayStrength(p),
			DisplayRoute:     p.RouteCode,
			DisplayLabel:     displayLabel(drug.GenericName, p),
		})
	}
	return out, rows.Err()
}

const qPresGet = `
  SELECT
    id, drug_id, dosage_form_code, route_code,
    strength_num, strength_unit_num,
    strength_den, strength_unit_den,
    dispense_unit, piece_content_amount, piece_content_unit,
    is_fractional_allowed, barcode, notes, created_at, updated_at
  FROM drug_presentations
  WHERE id=$1`

func (r *postgresPharmacyRepository) GetPresentation(ctx context.Context, id int64) (*entities.DrugPresentationView, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	var p entities.DrugPresentation
	err := dbx.QueryRowContext(ctx, qPresGet, id).Scan(
		&p.ID, &p.DrugID, &p.DosageFormCode, &p.RouteCode,
		&p.StrengthNum, &p.StrengthUnitNum,
		&p.StrengthDen, &p.StrengthUnitDen,
		&p.DispenseUnit, &p.PieceContentAmount, &p.PieceContentUnit,
		&p.IsFractionalAllowed, &p.Barcode, &p.Notes, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("presentation not found")
	}
	if err != nil {
		return nil, err
	}

	drug, err := r.GetDrug(ctx, p.DrugID)
	if err != nil {
		return nil, err
	}

	v := entities.DrugPresentationView{
		DrugPresentation: p,
		DrugName:         drug.GenericName,
		DisplayStrength:  displayStrength(p),
		DisplayRoute:     p.RouteCode,
		DisplayLabel:     displayLabel(drug.GenericName, p),
	}
	return &v, nil
}

const qPresCreate = `
  INSERT INTO drug_presentations (
    drug_id, dosage_form_code, route_code,
    strength_num, strength_unit_num,
    strength_den, strength_unit_den,
    dispense_unit, piece_content_amount, piece_content_unit,
    is_fractional_allowed, barcode, notes
  ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,COALESCE($11,FALSE),$12,$13)
  RETURNING id`

func (r *postgresPharmacyRepository) CreatePresentation(ctx context.Context, p *entities.DrugPresentation) (*entities.DrugPresentationView, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var id int64
	err := dbx.QueryRowContext(ctx, qPresCreate,
		p.DrugID, p.DosageFormCode, p.RouteCode,
		p.StrengthNum, p.StrengthUnitNum,
		p.StrengthDen, p.StrengthUnitDen,
		p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
		p.IsFractionalAllowed, p.Barcode, p.Notes,
	).Scan(&id)
	if err != nil {
		return nil, err
	}
	// hydrate
	v, err := r.GetPresentation(ctx, id)
	if err != nil {
		return nil, err
	}
	return v, nil
}

const qPresUpdate = `
  UPDATE drug_presentations SET
    drug_id=$2, dosage_form_code=$3, route_code=$4,
    strength_num=$5, strength_unit_num=$6,
    strength_den=$7, strength_unit_den=$8,
    dispense_unit=$9, piece_content_amount=$10, piece_content_unit=$11,
    is_fractional_allowed=$12, barcode=$13, notes=$14, updated_at=NOW()
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdatePresentation(ctx context.Context, p *entities.DrugPresentation) (*entities.DrugPresentationView, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qPresUpdate,
		p.ID, p.DrugID, p.DosageFormCode, p.RouteCode,
		p.StrengthNum, p.StrengthUnitNum,
		p.StrengthDen, p.StrengthUnitDen,
		p.DispenseUnit, p.PieceContentAmount, p.PieceContentUnit,
		p.IsFractionalAllowed, p.Barcode, p.Notes,
	)
	if err != nil {
		return nil, err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return nil, errors.New("presentation not found")
	}
	v, err := r.GetPresentation(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return v, nil
}

const qPresDelete = `DELETE FROM drug_presentations WHERE id=$1`

func (r *postgresPharmacyRepository) DeletePresentation(ctx context.Context, id int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qPresDelete, id)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return errors.New("presentation not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
//  BATCHES & LOCATIONS (quantities in DispenseUnit)
// -----------------------------------------------------------------------------

const qBatchList = `
  SELECT id, presentation_id, batch_number, expiry_date, supplier, quantity, created_at, updated_at
  FROM drug_batches
  WHERE presentation_id=$1
  ORDER BY expiry_date NULLS LAST, batch_number, id`

func (r *postgresPharmacyRepository) ListBatches(ctx context.Context, presentationID int64) ([]entities.BatchDetail, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	rows, err := dbx.QueryContext(ctx, qBatchList, presentationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := make([]entities.DrugBatch, 0, 64)
	batchIDs := make([]int64, 0, 64)
	for rows.Next() {
		var b entities.DrugBatch
		if err := rows.Scan(&b.ID, &b.PresentationID, &b.BatchNumber, &b.ExpiryDate, &b.Supplier, &b.Quantity, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		batches = append(batches, b)
		batchIDs = append(batchIDs, b.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(batches) == 0 {
		return []entities.BatchDetail{}, nil
	}

	locRows, err := dbx.QueryContext(ctx, `
		SELECT id, batch_id, location, quantity, created_at, updated_at
		FROM batch_locations
		WHERE batch_id = ANY($1)
		ORDER BY batch_id, location, id`, pq.Array(batchIDs))
	if err != nil {
		return nil, err
	}
	defer locRows.Close()

	locsByBatch := make(map[int64][]entities.DrugBatchLocation, len(batchIDs))
	for locRows.Next() {
		var l entities.DrugBatchLocation
		if err := locRows.Scan(&l.ID, &l.BatchID, &l.Location, &l.Quantity, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		locsByBatch[l.BatchID] = append(locsByBatch[l.BatchID], l)
	}
	if err := locRows.Err(); err != nil {
		return nil, err
	}

	out := make([]entities.BatchDetail, 0, len(batches))
	for _, b := range batches {
		expKey := b.ExpiryDate
		out = append(out, entities.BatchDetail{
			DrugBatch:      b,
			DispenseUnit:   "", // optional to fill if you join presentation for UI; FE usually knows it
			ExpirySortKey:  expKey,
			BatchLocations: locsByBatch[b.ID],
		})
	}
	return out, nil
}

const qBatchGet = `
  SELECT id, presentation_id, batch_number, expiry_date, supplier, quantity, created_at, updated_at
  FROM drug_batches WHERE id=$1`

func (r *postgresPharmacyRepository) GetBatch(ctx context.Context, batchID int64) (*entities.BatchDetail, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var b entities.DrugBatch
	err := dbx.QueryRowContext(ctx, qBatchGet, batchID).
		Scan(&b.ID, &b.PresentationID, &b.BatchNumber, &b.ExpiryDate, &b.Supplier, &b.Quantity, &b.CreatedAt, &b.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("batch not found")
	}
	if err != nil {
		return nil, err
	}

	locs, err := r.ListBatchLocations(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	return &entities.BatchDetail{
		DrugBatch:      b,
		ExpirySortKey:  b.ExpiryDate,
		BatchLocations: locs,
	}, nil
}

const qBatchCreate = `
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  VALUES ($1,$2,$3,$4,COALESCE($5,0))
  RETURNING id`

func (r *postgresPharmacyRepository) CreateBatch(ctx context.Context, b *entities.DrugBatch, locations []entities.DrugBatchLocation) (*entities.BatchDetail, error) {
	tx, ok := TxFromCtx(ctx)
	own := false
	var err error
	if !ok {
		tx, err = r.Conn.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		own = true
		defer tx.Rollback()
	}

	var id int64
	if err := tx.QueryRowContext(ctx, qBatchCreate,
		b.PresentationID, b.BatchNumber, b.ExpiryDate, b.Supplier, b.Quantity,
	).Scan(&id); err != nil {
		return nil, err
	}

	if len(locations) > 0 {
		stmt, err := tx.PrepareContext(ctx, `
		  INSERT INTO batch_locations (batch_id, location, quantity)
		  VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		for i := range locations {
			if locations[i].Quantity < 0 {
				return nil, fmt.Errorf("location %q has negative quantity", locations[i].Location)
			}
			var locID int64
			var ca, ua time.Time
			if err := stmt.QueryRowContext(ctx, id, locations[i].Location, locations[i].Quantity).Scan(&locID, &ca, &ua); err != nil {
				return nil, err
			}
			locations[i].ID = locID
			locations[i].BatchID = id
			locations[i].CreatedAt = ca
			locations[i].UpdatedAt = ua
		}
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetBatch(ctx, id)
}

const qBatchUpdate = `
  UPDATE drug_batches
  SET presentation_id=$2, batch_number=$3, expiry_date=$4, supplier=$5, quantity=$6, updated_at=NOW()
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdateBatch(ctx context.Context, b *entities.DrugBatch) (*entities.BatchDetail, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qBatchUpdate,
		b.ID, b.PresentationID, b.BatchNumber, b.ExpiryDate, b.Supplier, b.Quantity)
	if err != nil {
		return nil, err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return nil, errors.New("batch not found")
	}
	return r.GetBatch(ctx, b.ID)
}

const qBatchDelete = `DELETE FROM drug_batches WHERE id=$1`

func (r *postgresPharmacyRepository) DeleteBatch(ctx context.Context, batchID int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qBatchDelete, batchID)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return errors.New("batch not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
//  LOCATIONS
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) ListBatchLocations(ctx context.Context, batchID int64) ([]entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	rows, err := dbx.QueryContext(ctx, `
	  SELECT id, batch_id, location, quantity, created_at, updated_at
	  FROM batch_locations
	  WHERE batch_id=$1
	  ORDER BY location, id`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entities.DrugBatchLocation
	for rows.Next() {
		var l entities.DrugBatchLocation
		if err := rows.Scan(&l.ID, &l.BatchID, &l.Location, &l.Quantity, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

const qLocCreate = `
  INSERT INTO batch_locations (batch_id, location, quantity)
  VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`

func (r *postgresPharmacyRepository) CreateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var id int64
	var ca, ua time.Time
	if err := dbx.QueryRowContext(ctx, qLocCreate, loc.BatchID, loc.Location, loc.Quantity).
		Scan(&id, &ca, &ua); err != nil {
		return nil, err
	}
	return r.GetBatchLocation(ctx, id)
}

const qLocGet = `
  SELECT id, batch_id, location, quantity, created_at, updated_at
  FROM batch_locations WHERE id=$1`

func (r *postgresPharmacyRepository) GetBatchLocation(ctx context.Context, id int64) (*entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var l entities.DrugBatchLocation
	err := dbx.QueryRowContext(ctx, qLocGet, id).
		Scan(&l.ID, &l.BatchID, &l.Location, &l.Quantity, &l.CreatedAt, &l.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("batch location not found")
	}
	return &l, err
}

const qLocUpdate = `
  UPDATE batch_locations SET batch_id=$2, location=$3, quantity=$4, updated_at=NOW()
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qLocUpdate, loc.ID, loc.BatchID, loc.Location, loc.Quantity)
	if err != nil {
		return nil, err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return nil, errors.New("batch location not found")
	}
	return r.GetBatchLocation(ctx, loc.ID)
}

const qLocDelete = `DELETE FROM batch_locations WHERE id=$1`

func (r *postgresPharmacyRepository) DeleteBatchLocation(ctx context.Context, id int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qLocDelete, id)
	if err != nil {
		return err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return errors.New("location not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
//  STOCK VIEW (FEFO summary for a presentation)
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) GetPresentationStock(ctx context.Context, presentationID int64) (*entities.PresentationStock, error) {
	// 1) Presentation view (for labels)
	pv, err := r.GetPresentation(ctx, presentationID)
	if err != nil {
		return nil, err
	}

	// 2) Batches & locations
	batches, err := r.ListBatches(ctx, presentationID)
	if err != nil {
		return nil, err
	}

	// 3) FEFO order and totals
	sort.SliceStable(batches, func(i, j int) bool {
		// nil expiry → last
		ie, je := batches[i].ExpiryDate, batches[j].ExpiryDate
		if ie == nil && je != nil {
			return false
		}
		if ie != nil && je == nil {
			return true
		}
		if ie == nil && je == nil {
			return batches[i].BatchNumber < batches[j].BatchNumber
		}
		if ie.Equal(*je) {
			return batches[i].BatchNumber < batches[j].BatchNumber
		}
		return ie.Before(*je)
	})
	total := 0
	for _, b := range batches {
		total += b.Quantity
	}

	return &entities.PresentationStock{
		Presentation: *pv,
		Batches:      batches,
		TotalQty:     total,
	}, nil
}
