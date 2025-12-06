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

func displayStrength(d entities.Drug) string {
	if d.StrengthDen == nil || d.StrengthUnitDen == nil {
		// solids (no denominator), e.g. "500 mg TAB"
		if d.StrengthNum != nil && d.StrengthUnitNum != nil {
			return fmt.Sprintf("%g %s %s", *d.StrengthNum, *d.StrengthUnitNum, d.DosageFormCode)
		}
		// Unknown strength: show dosage form
		return d.DosageFormCode
	}
	// liquids/creams, e.g. "250 mg/5 mL SYR"
	if d.StrengthNum != nil && d.StrengthUnitNum != nil {
		numUnit := derefStr(d.StrengthUnitNum)
		denUnit := derefStr(d.StrengthUnitDen)
		numVal := derefFloat(d.StrengthNum)
		denVal := derefFloat(d.StrengthDen)

		// Check if we should display as percentage (based on drug setting)
		if d.DisplayAsPercentage && denVal > 0 {
			percentage := (numVal / denVal) * 100
			return fmt.Sprintf("%g%% %s", percentage, d.DosageFormCode)
		}

		return fmt.Sprintf("%g %s/%g %s %s",
			numVal, numUnit,
			denVal, denUnit,
			d.DosageFormCode)
	}
	// Unknown strength liquid: show dosage form
	return d.DosageFormCode
}

func displayLabel(d entities.Drug) string {
	strength := displayStrength(d)
	route := d.RouteCode

	// Build base label: "GenericName Strength (Route)"
	base := fmt.Sprintf("%s %s (%s)", d.GenericName, strength, route)

	// Add piece content info if applicable (bottles, tubes, inhalers)
	if d.PieceContentAmount != nil && d.PieceContentUnit != nil {
		switch d.DispenseUnit {
		case "bottle":
			base = fmt.Sprintf("%s - bottle %g %s", base, *d.PieceContentAmount, *d.PieceContentUnit)
		case "tube":
			base = fmt.Sprintf("%s - tube %g %s", base, *d.PieceContentAmount, *d.PieceContentUnit)
		case "inhaler":
			base = fmt.Sprintf("%s - inhaler %g %s", base, *d.PieceContentAmount, *d.PieceContentUnit)
		}
	}

	// Prepend ATC code with a dot if present
	if d.ATCCode != nil && *d.ATCCode != "" {
		return fmt.Sprintf("%s. %s", *d.ATCCode, base)
	}

	return base
}

func derefFloat(p *float64) float64 {
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

// Helper functions for comparing pointers
func equalFloatPtr(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalStrPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// -----------------------------------------------------------------------------
//  DRUGS
// -----------------------------------------------------------------------------

const qDrugsList = `
  SELECT id, generic_name, brand_name, atc_code, dosage_form_code, route_code,
    strength_num, strength_unit_num, strength_den, strength_unit_den,
    dispense_unit, piece_content_amount, piece_content_unit,
    is_fractional_allowed, display_as_percentage, barcode, notes, is_active, created_at, updated_at
  FROM drugs
  /* optional WHERE added dynamically */
  ORDER BY generic_name, COALESCE(brand_name,''), dosage_form_code, route_code`

func (r *postgresPharmacyRepository) ListDrugs(ctx context.Context, q *string) ([]entities.DrugView, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	query := qDrugsList
	args := []any{}
	if q != nil && *q != "" {
		query = `
		  SELECT id, generic_name, brand_name, atc_code, dosage_form_code, route_code,
		    strength_num, strength_unit_num, strength_den, strength_unit_den,
		    dispense_unit, piece_content_amount, piece_content_unit,
		    is_fractional_allowed, display_as_percentage, barcode, notes, is_active, created_at, updated_at
		  FROM drugs
		  WHERE generic_name ILIKE $1 OR COALESCE(brand_name,'') ILIKE $1
		  ORDER BY generic_name, COALESCE(brand_name,''), dosage_form_code, route_code`
		args = append(args, "%"+*q+"%")
	}

	rows, err := dbx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.DrugView{}
	for rows.Next() {
		var d entities.Drug
		if err := rows.Scan(
			&d.ID, &d.GenericName, &d.BrandName, &d.ATCCode,
			&d.DosageFormCode, &d.RouteCode,
			&d.StrengthNum, &d.StrengthUnitNum, &d.StrengthDen, &d.StrengthUnitDen,
			&d.DispenseUnit, &d.PieceContentAmount, &d.PieceContentUnit,
			&d.IsFractionalAllowed, &d.DisplayAsPercentage, &d.Barcode, &d.Notes, &d.IsActive,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		dv := entities.DrugView{
			Drug:            d,
			DisplayStrength: displayStrength(d),
			DisplayRoute:    d.RouteCode,
			DisplayLabel:    displayLabel(d),
		}
		out = append(out, dv)
	}
	return out, rows.Err()
}

const qDrugCreate = `
  INSERT INTO drugs (
    generic_name, brand_name, atc_code, dosage_form_code, route_code,
    strength_num, strength_unit_num, strength_den, strength_unit_den,
    dispense_unit, piece_content_amount, piece_content_unit,
    is_fractional_allowed, display_as_percentage, barcode, notes, is_active
  )
  VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,COALESCE($13,FALSE),COALESCE($14,FALSE),$15,$16,COALESCE($17,TRUE))
  RETURNING id`

func (r *postgresPharmacyRepository) CreateDrug(ctx context.Context, d *entities.Drug) (*entities.DrugView, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var id int64
	if err := dbx.QueryRowContext(ctx, qDrugCreate,
		d.GenericName, d.BrandName, d.ATCCode,
		d.DosageFormCode, d.RouteCode,
		d.StrengthNum, d.StrengthUnitNum, d.StrengthDen, d.StrengthUnitDen,
		d.DispenseUnit, d.PieceContentAmount, d.PieceContentUnit,
		d.IsFractionalAllowed, d.DisplayAsPercentage, d.Barcode, d.Notes, d.IsActive,
	).Scan(&id); err != nil {
		return nil, err
	}
	return r.GetDrug(ctx, id)
}

const qDrugGet = `
  SELECT id, generic_name, brand_name, atc_code, dosage_form_code, route_code,
    strength_num, strength_unit_num, strength_den, strength_unit_den,
    dispense_unit, piece_content_amount, piece_content_unit,
    is_fractional_allowed, display_as_percentage, barcode, notes, is_active, created_at, updated_at
  FROM drugs WHERE id=$1`

func (r *postgresPharmacyRepository) GetDrug(ctx context.Context, id int64) (*entities.DrugView, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var d entities.Drug
	err := dbx.QueryRowContext(ctx, qDrugGet, id).Scan(
		&d.ID, &d.GenericName, &d.BrandName, &d.ATCCode,
		&d.DosageFormCode, &d.RouteCode,
		&d.StrengthNum, &d.StrengthUnitNum, &d.StrengthDen, &d.StrengthUnitDen,
		&d.DispenseUnit, &d.PieceContentAmount, &d.PieceContentUnit,
		&d.IsFractionalAllowed, &d.DisplayAsPercentage, &d.Barcode, &d.Notes, &d.IsActive,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("drug not found")
	}
	if err != nil {
		return nil, err
	}
	dv := &entities.DrugView{
		Drug:            d,
		DisplayStrength: displayStrength(d),
		DisplayRoute:    d.RouteCode,
		DisplayLabel:    displayLabel(d),
	}
	return dv, nil
}

const qDrugUpdate = `
  UPDATE drugs SET
    generic_name=$2, brand_name=$3, atc_code=$4,
    dosage_form_code=$5, route_code=$6,
    strength_num=$7, strength_unit_num=$8, strength_den=$9, strength_unit_den=$10,
    dispense_unit=$11, piece_content_amount=$12, piece_content_unit=$13,
    is_fractional_allowed=$14, display_as_percentage=$15, barcode=$16, notes=$17, is_active=$18, updated_at=NOW()
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdateDrug(ctx context.Context, d *entities.Drug) (*entities.DrugView, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	// Get current drug to compare fields
	current, err := r.GetDrug(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	currentDrug := current.Drug

	// Check if any risky fields are being changed
	riskyFieldChanged :=
		!equalFloatPtr(currentDrug.StrengthNum, d.StrengthNum) ||
			!equalStrPtr(currentDrug.StrengthUnitNum, d.StrengthUnitNum) ||
			!equalFloatPtr(currentDrug.StrengthDen, d.StrengthDen) ||
			!equalStrPtr(currentDrug.StrengthUnitDen, d.StrengthUnitDen) ||
			currentDrug.DispenseUnit != d.DispenseUnit ||
			!equalFloatPtr(currentDrug.PieceContentAmount, d.PieceContentAmount) ||
			!equalStrPtr(currentDrug.PieceContentUnit, d.PieceContentUnit)

	// If risky fields changed, check if drug has prescriptions
	if riskyFieldChanged {
		var prescriptionCount int
		err := dbx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM prescription_lines WHERE drug_id=$1
		`, d.ID).Scan(&prescriptionCount)
		if err != nil {
			return nil, fmt.Errorf("failed to check prescriptions: %w", err)
		}

		if prescriptionCount > 0 {
			return nil, fmt.Errorf("cannot modify drug properties (strength, dispense unit, piece content) because this drug is already used in %d prescription(s). Only safe fields (name, ATC code, dosage form, route, notes, barcode, active status, display percentage) can be edited", prescriptionCount)
		}
	}

	res, err := dbx.ExecContext(ctx, qDrugUpdate,
		d.ID, d.GenericName, d.BrandName, d.ATCCode,
		d.DosageFormCode, d.RouteCode,
		d.StrengthNum, d.StrengthUnitNum, d.StrengthDen, d.StrengthUnitDen,
		d.DispenseUnit, d.PieceContentAmount, d.PieceContentUnit,
		d.IsFractionalAllowed, d.DisplayAsPercentage, d.Barcode, d.Notes, d.IsActive,
	)
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

	// Check if drug has any prescriptions
	var prescriptionCount int
	err := dbx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM prescription_lines WHERE drug_id=$1
	`, id).Scan(&prescriptionCount)
	if err != nil {
		return fmt.Errorf("failed to check prescriptions: %w", err)
	}

	if prescriptionCount > 0 {
		return fmt.Errorf("cannot delete drug because it is already used in %d prescription(s)", prescriptionCount)
	}

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
//  BATCHES & LOCATIONS (quantities in DispenseUnit)
// -----------------------------------------------------------------------------

const qBatchList = `
  SELECT id, drug_id, batch_number, expiry_date, supplier, quantity, created_at, updated_at
  FROM drug_batches
  WHERE drug_id=$1
  ORDER BY expiry_date NULLS LAST, batch_number, id`

func (r *postgresPharmacyRepository) ListBatches(ctx context.Context, drugID int64) ([]entities.BatchDetail, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	rows, err := dbx.QueryContext(ctx, qBatchList, drugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := make([]entities.DrugBatch, 0, 64)
	batchIDs := make([]int64, 0, 64)
	for rows.Next() {
		var b entities.DrugBatch
		if err := rows.Scan(&b.ID, &b.DrugID, &b.BatchNumber, &b.ExpiryDate, &b.Supplier, &b.Quantity, &b.CreatedAt, &b.UpdatedAt); err != nil {
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
  SELECT id, drug_id, batch_number, expiry_date, supplier, quantity, created_at, updated_at
  FROM drug_batches WHERE id=$1`

func (r *postgresPharmacyRepository) GetBatch(ctx context.Context, batchID int64) (*entities.BatchDetail, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	var b entities.DrugBatch
	err := dbx.QueryRowContext(ctx, qBatchGet, batchID).
		Scan(&b.ID, &b.DrugID, &b.BatchNumber, &b.ExpiryDate, &b.Supplier, &b.Quantity, &b.CreatedAt, &b.UpdatedAt)
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
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
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
		b.DrugID, b.BatchNumber, b.ExpiryDate, b.Supplier, b.Quantity,
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
  SET batch_number=$2, expiry_date=$3, supplier=$4
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdateBatch(ctx context.Context, b *entities.DrugBatch, locations []entities.DrugBatchLocation) (*entities.BatchDetail, error) {
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

	// Update batch itself (batch_number, expiry_date, supplier)
	// drug_id and quantity cannot be edited (drug_id is fixed, quantity is auto-synced from locations)
	res, err := tx.ExecContext(ctx, qBatchUpdate,
		b.ID, b.BatchNumber, b.ExpiryDate, b.Supplier)
	if err != nil {
		return nil, err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return nil, errors.New("batch not found")
	}

	// Get current batch locations from database
	currentLocs, err := r.listBatchLocationsTx(ctx, tx, b.ID)
	if err != nil {
		return nil, err
	}

	// Build a map of current locations by ID for quick lookup
	currentLocsMap := make(map[int64]entities.DrugBatchLocation)
	for _, loc := range currentLocs {
		currentLocsMap[loc.ID] = loc
	}

	// Build a map of locations to keep (from payload) by ID
	payloadLocsMap := make(map[int64]bool)
	for _, loc := range locations {
		if loc.ID > 0 {
			payloadLocsMap[loc.ID] = true
		}
	}

	// 1. Update or create locations from payload
	for i := range locations {
		loc := &locations[i]
		if loc.Quantity < 0 {
			return nil, fmt.Errorf("location %q has negative quantity", loc.Location)
		}

		if loc.ID > 0 {
			// Existing location - check if it belongs to this batch
			currentLoc, exists := currentLocsMap[loc.ID]
			if !exists {
				return nil, fmt.Errorf("batch location with id %d not found", loc.ID)
			}
			if currentLoc.BatchID != b.ID {
				return nil, fmt.Errorf("batch location %d does not belong to batch %d", loc.ID, b.ID)
			}

			// Update location if location name or quantity changed
			if currentLoc.Location != loc.Location || currentLoc.Quantity != loc.Quantity {
				_, err := tx.ExecContext(ctx, `
					UPDATE batch_locations 
					SET location=$2, quantity=$3, updated_at=NOW()
					WHERE id=$1`,
					loc.ID, loc.Location, loc.Quantity)
				if err != nil {
					return nil, fmt.Errorf("failed to update batch location %d: %w", loc.ID, err)
				}
			}
		} else {
			// New location - create it
			var locID int64
			var ca, ua time.Time
			err := tx.QueryRowContext(ctx, `
				INSERT INTO batch_locations (batch_id, location, quantity)
				VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`,
				b.ID, loc.Location, loc.Quantity).Scan(&locID, &ca, &ua)
			if err != nil {
				return nil, fmt.Errorf("failed to create batch location: %w", err)
			}
			loc.ID = locID
			loc.BatchID = b.ID
			loc.CreatedAt = ca
			loc.UpdatedAt = ua
		}
	}

	// 2. Delete locations that exist in DB but not in payload
	// First, check if any of these locations are referenced by prescriptions
	for _, currentLoc := range currentLocs {
		if !payloadLocsMap[currentLoc.ID] {
			// Location exists in DB but not in payload - check if referenced
			var referenceCount int
			err := tx.QueryRowContext(ctx, `
				SELECT COUNT(*) 
				FROM prescription_batch_items 
				WHERE batch_location_id = $1`,
				currentLoc.ID).Scan(&referenceCount)
			if err != nil {
				return nil, fmt.Errorf("failed to check prescription references: %w", err)
			}

			if referenceCount > 0 {
				return nil, fmt.Errorf("cannot delete batch location %d (location: %q) because it is allocated to %d prescription line(s)", currentLoc.ID, currentLoc.Location, referenceCount)
			}

			// Safe to delete
			_, err = tx.ExecContext(ctx, `DELETE FROM batch_locations WHERE id=$1`, currentLoc.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to delete batch location %d: %w", currentLoc.ID, err)
			}
		}
	}

	if own {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return r.GetBatch(ctx, b.ID)
}

// Helper function to list batch locations within a transaction
func (r *postgresPharmacyRepository) listBatchLocationsTx(ctx context.Context, tx *sql.Tx, batchID int64) ([]entities.DrugBatchLocation, error) {
	rows, err := tx.QueryContext(ctx, `
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

const qBatchDelete = `DELETE FROM drug_batches WHERE id=$1`

func (r *postgresPharmacyRepository) DeleteBatch(ctx context.Context, batchID int64) error {
	dbx := DBFromCtx(ctx, r.Conn)

	// Check if any batch locations belonging to this batch are referenced by prescriptions
	var referenceCount int
	err := dbx.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM prescription_batch_items pbi
		JOIN batch_locations bl ON bl.id = pbi.batch_location_id
		WHERE bl.batch_id = $1
	`, batchID).Scan(&referenceCount)
	if err != nil {
		return fmt.Errorf("failed to check prescription references: %w", err)
	}

	if referenceCount > 0 {
		return fmt.Errorf("cannot delete batch because its locations are allocated to %d prescription line(s)", referenceCount)
	}

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
  UPDATE batch_locations SET location=$2, quantity=$3, updated_at=NOW()
  WHERE id=$1`

func (r *postgresPharmacyRepository) UpdateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)

	// Only allow updating location name and quantity
	res, err := dbx.ExecContext(ctx, qLocUpdate, loc.ID, loc.Location, loc.Quantity)
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

	// Check if batch location is referenced by any prescriptions
	var referenceCount int
	err := dbx.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM prescription_batch_items 
		WHERE batch_location_id = $1
	`, id).Scan(&referenceCount)
	if err != nil {
		return fmt.Errorf("failed to check prescription references: %w", err)
	}

	if referenceCount > 0 {
		return fmt.Errorf("cannot delete batch location because it is allocated to %d prescription line(s)", referenceCount)
	}

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

func (r *postgresPharmacyRepository) GetDrugStock(ctx context.Context, drugID int64) (*entities.DrugStock, error) {
	// 1) Drug view (for labels)
	dv, err := r.GetDrug(ctx, drugID)
	if err != nil {
		return nil, err
	}

	// 2) Batches & locations
	batches, err := r.ListBatches(ctx, drugID)
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

	return &entities.DrugStock{
		Drug:     *dv,
		Batches:  batches,
		TotalQty: total,
	}, nil
}
