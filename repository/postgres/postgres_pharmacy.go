package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"sothea-backend/entities"
	db "sothea-backend/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -----------------------------------------------------------------------------
//  STRUCT + CONSTRUCTOR
// -----------------------------------------------------------------------------

type postgresPharmacyRepository struct {
	Conn    *pgxpool.Pool
	queries *db.Queries
}

func (r *postgresPharmacyRepository) q(ctx context.Context) *db.Queries {
	if tx, ok := TxFromCtx(ctx); ok && tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func NewPostgresPharmacyRepository(conn *pgxpool.Pool) entities.PharmacyRepository {
	return &postgresPharmacyRepository{
		Conn:    conn,
		queries: db.New(conn),
	}
}

// -----------------------------------------------------------------------------
//  HELPERS (label builders; keep in repo so FE gets nice strings)
// -----------------------------------------------------------------------------

func displayStrength(d db.Drug) string {
	// d.Strength* may be pgtype.Numeric and units as *string; guard nils
	hasNum := d.StrengthUnitNum != nil
	hasDen := d.StrengthUnitDen != nil

	if !hasDen {
		if hasNum {
			return fmt.Sprintf("%s %s", *d.StrengthUnitNum, d.DosageFormCode)
		}
		return d.DosageFormCode
	}

	if hasNum {
		numUnit := *d.StrengthUnitNum
		denUnit := *d.StrengthUnitDen
		return fmt.Sprintf("%s/%s %s", numUnit, denUnit, d.DosageFormCode)
	}
	return d.DosageFormCode
}

func displayLabel(d db.Drug) string {
	strength := displayStrength(d)
	route := d.RouteCode
	base := fmt.Sprintf("%s %s (%s)", d.GenericName, strength, route)
	if d.PieceContentUnit != nil {
		switch d.DispenseUnit {
		case "bottle":
			base = fmt.Sprintf("%s - bottle [?] %s", base, *d.PieceContentUnit)
		case "tube":
			base = fmt.Sprintf("%s - tube [?] %s", base, *d.PieceContentUnit)
		case "inhaler":
			base = fmt.Sprintf("%s - inhaler [?] %s", base, *d.PieceContentUnit)
		}
	}
	if d.DrugCode != nil {
		return fmt.Sprintf("%d. %s", *d.DrugCode, base)
	}
	return base
}

// findFirstEmptyBackwards finds the first empty spot going backwards from startCode down to minCode (inclusive).
// Returns the empty spot or minCode if no gap found.
func findFirstEmptyBackwards(ctx context.Context, tx pgx.Tx, startCode, minCode int64) (int64, error) {
	var firstEmpty *int64
	if err := tx.QueryRow(ctx, `
		SELECT MAX(series_value)
		FROM generate_series($1::INTEGER, $2::INTEGER) AS series_value
		WHERE NOT EXISTS (
			SELECT 1 FROM drugs WHERE drug_code = series_value
		)
	`, minCode, startCode).Scan(&firstEmpty); err != nil {
		return 0, err
	}
	if firstEmpty != nil {
		return *firstEmpty, nil
	}
	return minCode, nil
}

func findFirstEmptyForwards(ctx context.Context, tx pgx.Tx, startCode, maxCode int64) (int64, error) {
	var firstEmpty *int64
	if err := tx.QueryRow(ctx, `
		SELECT MIN(series_value)
		FROM generate_series($1::INTEGER, $2::INTEGER) AS series_value
		WHERE NOT EXISTS (
			SELECT 1 FROM drugs WHERE drug_code = series_value
		)
	`, startCode, maxCode).Scan(&firstEmpty); err != nil {
		return 0, err
	}
	if firstEmpty != nil {
		return *firstEmpty, nil
	}
	return maxCode, nil
}

func findFirstEmptyDrugCode(ctx context.Context, tx pgx.Tx, startCode int64) (int64, error) {
	var maxCode *int64
	if err := tx.QueryRow(ctx, `
		SELECT MAX(drug_code) FROM drugs WHERE drug_code IS NOT NULL
	`).Scan(&maxCode); err != nil {
		return 0, err
	}
	if maxCode == nil || startCode > *maxCode {
		return startCode, nil
	}
	return findFirstEmptyForwards(ctx, tx, startCode, *maxCode+1)
}

func equalPtrString(a, b *string) bool {
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

func (r *postgresPharmacyRepository) ListDrugs(ctx context.Context, q *string) ([]entities.DrugView, error) {
	qx := r.q(ctx)

	var rows []db.Drug
	var err error
	if q != nil && *q != "" {
		searchRows, e := qx.SearchDrugs(ctx, "%"+*q+"%")
		err = e
		for _, row := range searchRows {
			rows = append(rows, db.Drug{
				ID:                  row.ID,
				GenericName:         row.GenericName,
				BrandName:           row.BrandName,
				DrugCode:            row.DrugCode,
				DosageFormCode:      row.DosageFormCode,
				RouteCode:           row.RouteCode,
				StrengthNum:         row.StrengthNum,
				StrengthUnitNum:     row.StrengthUnitNum,
				StrengthDen:         row.StrengthDen,
				StrengthUnitDen:     row.StrengthUnitDen,
				DispenseUnit:        row.DispenseUnit,
				PieceContentAmount:  row.PieceContentAmount,
				PieceContentUnit:    row.PieceContentUnit,
				IsFractionalAllowed: row.IsFractionalAllowed,
				DisplayAsPercentage: row.DisplayAsPercentage,
				Barcode:             row.Barcode,
				Notes:               row.Notes,
				IsActive:            row.IsActive,
				CreatedAt:           row.CreatedAt,
				UpdatedAt:           row.UpdatedAt,
			})
		}
	} else {
		listRows, e := qx.ListDrugs(ctx)
		err = e
		for _, row := range listRows {
			rows = append(rows, db.Drug{
				ID:                  row.ID,
				GenericName:         row.GenericName,
				BrandName:           row.BrandName,
				DrugCode:            row.DrugCode,
				DosageFormCode:      row.DosageFormCode,
				RouteCode:           row.RouteCode,
				StrengthNum:         row.StrengthNum,
				StrengthUnitNum:     row.StrengthUnitNum,
				StrengthDen:         row.StrengthDen,
				StrengthUnitDen:     row.StrengthUnitDen,
				DispenseUnit:        row.DispenseUnit,
				PieceContentAmount:  row.PieceContentAmount,
				PieceContentUnit:    row.PieceContentUnit,
				IsFractionalAllowed: row.IsFractionalAllowed,
				DisplayAsPercentage: row.DisplayAsPercentage,
				Barcode:             row.Barcode,
				Notes:               row.Notes,
				IsActive:            row.IsActive,
				CreatedAt:           row.CreatedAt,
				UpdatedAt:           row.UpdatedAt,
			})
		}
	}
	if err != nil {
		return nil, err
	}

	out := make([]entities.DrugView, 0, len(rows))
	for _, row := range rows {
		out = append(out, entities.DrugView{
			Drug:            row,
			DisplayStrength: displayStrength(row),
			DisplayRoute:    row.RouteCode,
			DisplayLabel:    displayLabel(row),
		})
	}
	return out, nil
}

func (r *postgresPharmacyRepository) CreateDrug(ctx context.Context, d *db.Drug) (*entities.DrugView, error) {
	tx, ok := TxFromCtx(ctx)
	if !ok {
		return nil, errors.New("transaction not found")
	}
	q := r.queries.WithTx(tx)

	if d.DrugCode != nil {
		// Find the first empty spot starting from the requested code
		firstEmptySpot, err := findFirstEmptyDrugCode(ctx, tx, int64(*d.DrugCode))
		if err != nil {
			return nil, fmt.Errorf("failed to find empty drug code: %w", err)
		}

		// If the first empty spot is greater than requested code, shift codes to make room
		if _, err = tx.Exec(ctx, `
			UPDATE drugs
			SET drug_code = drug_code + 1
			WHERE drug_code >= $1 AND drug_code < $2
		`, *d.DrugCode, firstEmptySpot); err != nil {
			return nil, fmt.Errorf("failed to shift drug codes: %w", err)
		}
	}

	params := toInsertDrugParams(d)
	row, err := q.InsertDrug(ctx, params)
	if err != nil {
		return nil, err
	}
	return r.GetDrug(ctx, row)
}

func (r *postgresPharmacyRepository) GetDrug(ctx context.Context, id int64) (*entities.DrugView, error) {
	row, err := r.q(ctx).GetDrug(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("drug not found")
	}
	if err != nil {
		return nil, err
	}
	d := toDrugFromGetRow(row)
	return &entities.DrugView{
		Drug:            d,
		DisplayStrength: displayStrength(d),
		DisplayRoute:    d.RouteCode,
		DisplayLabel:    displayLabel(d),
	}, nil
}

func (r *postgresPharmacyRepository) UpdateDrug(ctx context.Context, d *db.Drug) (*entities.DrugView, error) {
	tx, ok := TxFromCtx(ctx)
	if !ok {
		return nil, errors.New("transaction not found")
	}
	q := r.queries.WithTx(tx)

	// Get current drug to compare fields
	current, err := r.GetDrug(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	currentDrug := current.Drug

	// Check if any risky fields are being changed
	riskyFieldChanged :=
		!equalPtrString(currentDrug.StrengthUnitNum, d.StrengthUnitNum) ||
			!equalPtrString(currentDrug.StrengthUnitDen, d.StrengthUnitDen) ||
			currentDrug.DispenseUnit != d.DispenseUnit ||
			!equalPtrString(currentDrug.PieceContentUnit, d.PieceContentUnit)

	// If risky fields changed, check if drug has prescriptions
	if riskyFieldChanged {
		var prescriptionCount int
		if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM prescription_lines WHERE drug_id=$1`, d.ID).Scan(&prescriptionCount); err != nil {
			return nil, fmt.Errorf("failed to check prescriptions: %w", err)
		}

		if prescriptionCount > 0 {
			return nil, fmt.Errorf("cannot modify drug properties (strength, dispense unit, piece content) because this drug is already used in %d prescription(s). Only safe fields (name, drug code, dosage form, route, notes, barcode, active status, display percentage) can be edited", prescriptionCount)
		}
	}

	// Handle drug_code shifting
	oldCode := currentDrug.DrugCode
	newCode := d.DrugCode

	if oldCode == nil && newCode != nil {
		// Case 1: Creating a code (previously no code) - follow create logic
		firstEmptySpot, err := findFirstEmptyDrugCode(ctx, tx, int64(*newCode))
		if err != nil {
			return nil, fmt.Errorf("failed to find empty drug code: %w", err)
		}

		if firstEmptySpot > int64(*newCode) {
			if _, err = tx.Exec(ctx, `
				UPDATE drugs
				SET drug_code = drug_code + 1
				WHERE drug_code >= $1 AND drug_code < $2
				AND id != $3
			`, *newCode, firstEmptySpot, d.ID); err != nil {
				return nil, fmt.Errorf("failed to shift drug codes: %w", err)
			}
		}

	} else if oldCode != nil && newCode == nil {
		// Case 2: Deleting a code - decrement all codes after old_code (pull up all till the end)
		if _, err := tx.Exec(ctx, `
			UPDATE drugs SET drug_code = drug_code - 1 WHERE drug_code > $1 AND id != $2
		`, *oldCode, d.ID); err != nil {
			return nil, fmt.Errorf("failed to shift drug codes: %w", err)
		}

	} else if oldCode != nil && *oldCode != *newCode {
		if *newCode > *oldCode {
			// Case 3: Increasing code - decrement codes from new_code backwards until empty spot or old_code
			// Find first empty spot going backwards from new_code to old_code
			firstEmpty, err := findFirstEmptyBackwards(ctx, tx, int64(*newCode), int64(*oldCode))
			if err != nil {
				return nil, fmt.Errorf("failed to find empty spot backwards: %w", err)
			}

			// Decrement all codes in range (firstEmpty, newCode] down by 1
			if firstEmpty < int64(*newCode) {
				if _, err = tx.Exec(ctx, `
					UPDATE drugs
					SET drug_code = drug_code - 1
					WHERE drug_code > $1 AND drug_code <= $2
					AND id != $3
				`, firstEmpty, *newCode, d.ID); err != nil {
					return nil, fmt.Errorf("failed to shift drug codes: %w", err)
				}
			}

		} else {
			// Case 2: Decreasing code - increment codes from new_code forwards until empty spot or old_code
			// Find first empty spot going forwards from new_code to old_code
			firstEmpty, err := findFirstEmptyForwards(ctx, tx, int64(*newCode), int64(*oldCode))
			if err != nil {
				return nil, fmt.Errorf("failed to find empty spot forwards: %w", err)
			}

			// Increment all codes in range [newCode, firstEmpty) up by 1
			if firstEmpty > int64(*newCode) {
				if _, err = tx.Exec(ctx, `
					UPDATE drugs
					SET drug_code = drug_code + 1
					WHERE drug_code >= $1 AND drug_code < $2
					AND id != $3
				`, *newCode, firstEmpty, d.ID); err != nil {
					return nil, fmt.Errorf("failed to shift drug codes: %w", err)
				}
			}
		}
		// If oldCode == newCode, no shifting needed
	}

	if err := q.UpdateDrug(ctx, toUpdateDrugParams(d)); err != nil {
		return nil, err
	}
	// Ensure row existed
	if _, err := q.GetDrug(ctx, d.ID); errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("drug not found")
	}
	return r.GetDrug(ctx, d.ID)
}

func (r *postgresPharmacyRepository) DeleteDrug(ctx context.Context, id int64) error {
	q := r.q(ctx)

	// Ensure exists
	if _, err := q.GetDrug(ctx, id); errors.Is(err, pgx.ErrNoRows) {
		return errors.New("drug not found")
	} else if err != nil {
		return err
	}

	// Check if drug has any prescriptions
	prescriptionCount, err := q.CountPrescriptionLinesForDrug(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check prescriptions: %w", err)
	}

	if prescriptionCount > 0 {
		return fmt.Errorf("cannot delete drug because it is already used in %d prescription(s)", prescriptionCount)
	}

	return q.DeleteDrug(ctx, id)
}

// -----------------------------------------------------------------------------
//  BATCHES & LOCATIONS (quantities in DispenseUnit)
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) ListBatches(ctx context.Context, drugID int64) ([]entities.BatchDetail, error) {
	q := r.q(ctx)

	batchRows, err := q.ListBatchesByDrug(ctx, drugID)
	if err != nil {
		return nil, err
	}
	if len(batchRows) == 0 {
		return []entities.BatchDetail{}, nil
	}

	batchIDs := make([]int64, 0, len(batchRows))
	batches := make([]db.DrugBatch, 0, len(batchRows))
	for _, row := range batchRows {
		batches = append(batches, toBatchFromListRow(row))
		batchIDs = append(batchIDs, row.ID)
	}

	locRows, err := q.ListBatchLocationsByBatchIDs(ctx, batchIDs)
	if err != nil {
		return nil, err
	}
	locsByBatch := make(map[int64][]db.BatchLocation, len(batchIDs))
	for _, row := range locRows {
		locsByBatch[row.BatchID] = append(locsByBatch[row.BatchID], toBatchLocationFromIDsRow(row))
	}

	out := make([]entities.BatchDetail, 0, len(batches))
	for _, b := range batches {
		out = append(out, entities.BatchDetail{
			DrugBatch:      b,
			DispenseUnit:   "", // optional to fill if you join presentation for UI; FE usually knows it
			ExpirySortKey:  b.ExpiryDate,
			BatchLocations: locsByBatch[b.ID],
		})
	}
	return out, nil
}

func (r *postgresPharmacyRepository) GetBatch(ctx context.Context, batchID int64) (*entities.BatchDetail, error) {
	q := r.q(ctx)
	bRow, err := q.GetBatch(ctx, batchID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("batch not found")
	}
	if err != nil {
		return nil, err
	}

	locRows, err := q.ListBatchLocationsByBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	locs := make([]db.BatchLocation, 0, len(locRows))
	for _, row := range locRows {
		locs = append(locs, toBatchLocationRow(row))
	}

	b := toBatchFromGetRow(bRow)
	return &entities.BatchDetail{
		DrugBatch:      b,
		ExpirySortKey:  b.ExpiryDate,
		BatchLocations: locs,
	}, nil
}

func (r *postgresPharmacyRepository) CreateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*entities.BatchDetail, error) {
	tx, ok := TxFromCtx(ctx)
	own := false
	var err error
	if !ok {
		tx, err = r.Conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		own = true
		defer func() { _ = tx.Rollback(ctx) }()
	}
	q := r.queries.WithTx(tx)

	params := db.InsertBatchParams{
		DrugID:      b.DrugID,
		BatchNumber: b.BatchNumber,
		ExpiryDate:  b.ExpiryDate,
		Supplier:    b.Supplier,
		Column5:     b.Quantity,
	}
	id, err := q.InsertBatch(ctx, params)
	if err != nil {
		return nil, err
	}

	if len(locations) > 0 {
		for i := range locations {
			if locations[i].Quantity < 0 {
				return nil, fmt.Errorf("location %q has negative quantity", locations[i].Location)
			}
			locRow, err := q.InsertBatchLocation(ctx, db.InsertBatchLocationParams{
				BatchID:  id,
				Location: locations[i].Location,
				Quantity: locations[i].Quantity,
			})
			if err != nil {
				return nil, err
			}
			locations[i].ID = locRow.ID
			locations[i].BatchID = id
			locations[i].CreatedAt = locRow.CreatedAt
			locations[i].UpdatedAt = locRow.UpdatedAt
		}
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}
	return r.GetBatch(ctx, id)
}

func (r *postgresPharmacyRepository) UpdateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*entities.BatchDetail, error) {
	tx, ok := TxFromCtx(ctx)
	own := false
	var err error
	if !ok {
		tx, err = r.Conn.Begin(ctx)
		if err != nil {
			return nil, err
		}
		own = true
		defer func() { _ = tx.Rollback(ctx) }()
	}
	q := r.queries.WithTx(tx)

	if err := q.UpdateBatch(ctx, db.UpdateBatchParams{
		ID:          b.ID,
		BatchNumber: b.BatchNumber,
		ExpiryDate:  b.ExpiryDate,
		Supplier:    b.Supplier,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("batch not found")
		}
		return nil, err
	}

	currentLocRows, err := q.ListBatchLocationsByBatch(ctx, b.ID)
	if err != nil {
		return nil, err
	}
	currentLocsMap := make(map[int64]db.BatchLocation)
	for _, row := range currentLocRows {
		loc := toBatchLocationRow(row)
		currentLocsMap[loc.ID] = loc
	}

	payloadLocsMap := make(map[int64]bool)
	for _, loc := range locations {
		if loc.ID > 0 {
			payloadLocsMap[loc.ID] = true
		}
	}

	for i := range locations {
		loc := &locations[i]
		if loc.Quantity < 0 {
			return nil, fmt.Errorf("location %q has negative quantity", loc.Location)
		}

		if loc.ID > 0 {
			currentLoc, exists := currentLocsMap[loc.ID]
			if !exists {
				return nil, fmt.Errorf("batch location with id %d not found", loc.ID)
			}
			if currentLoc.BatchID != b.ID {
				return nil, fmt.Errorf("batch location %d does not belong to batch %d", loc.ID, b.ID)
			}

			if currentLoc.Location != loc.Location || currentLoc.Quantity != loc.Quantity {
				if err := q.UpdateBatchLocation(ctx, db.UpdateBatchLocationParams{
					ID:       loc.ID,
					Location: loc.Location,
					Quantity: loc.Quantity,
				}); err != nil {
					return nil, fmt.Errorf("failed to update batch location %d: %w", loc.ID, err)
				}
			}
		} else {
			locRow, err := q.InsertBatchLocation(ctx, db.InsertBatchLocationParams{
				BatchID:  b.ID,
				Location: loc.Location,
				Quantity: loc.Quantity,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create batch location: %w", err)
			}
			loc.ID = locRow.ID
			loc.BatchID = b.ID
			loc.CreatedAt = locRow.CreatedAt
			loc.UpdatedAt = locRow.UpdatedAt
		}
	}

	for _, currentLoc := range currentLocsMap {
		if !payloadLocsMap[currentLoc.ID] {
			referenceCount, err := q.CountPrescriptionAllocationsForLocation(ctx, currentLoc.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to check prescription references: %w", err)
			}
			if referenceCount > 0 {
				return nil, fmt.Errorf("cannot delete batch location %d (location: %q) because it is allocated to %d prescription line(s)", currentLoc.ID, currentLoc.Location, referenceCount)
			}
			if err := q.DeleteBatchLocation(ctx, currentLoc.ID); err != nil {
				return nil, fmt.Errorf("failed to delete batch location %d: %w", currentLoc.ID, err)
			}
		}
	}

	if own {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
	}
	return r.GetBatch(ctx, b.ID)
}

func (r *postgresPharmacyRepository) DeleteBatch(ctx context.Context, batchID int64) error {
	q := r.q(ctx)

	if _, err := q.GetBatch(ctx, batchID); errors.Is(err, pgx.ErrNoRows) {
		return errors.New("batch not found")
	} else if err != nil {
		return err
	}

	// Check if any batch locations belonging to this batch are referenced by prescriptions
	referenceCount, err := q.CountPrescriptionAllocationsForBatch(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to check prescription references: %w", err)
	}

	if referenceCount > 0 {
		return fmt.Errorf("cannot delete batch because its locations are allocated to %d prescription line(s)", referenceCount)
	}

	return q.DeleteBatch(ctx, batchID)
}

// -----------------------------------------------------------------------------
//  LOCATIONS
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) ListBatchLocations(ctx context.Context, batchID int64) ([]db.BatchLocation, error) {
	rows, err := r.q(ctx).ListBatchLocationsByBatch(ctx, batchID)
	if err != nil {
		return nil, err
	}
	out := make([]db.BatchLocation, 0, len(rows))
	for _, row := range rows {
		out = append(out, toBatchLocationRow(row))
	}
	return out, nil
}

func (r *postgresPharmacyRepository) CreateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error) {
	locRow, err := r.q(ctx).InsertBatchLocation(ctx, db.InsertBatchLocationParams{
		BatchID:  loc.BatchID,
		Location: loc.Location,
		Quantity: loc.Quantity,
	})
	if err != nil {
		return nil, err
	}
	loc.ID = locRow.ID
	loc.CreatedAt = locRow.CreatedAt
	loc.UpdatedAt = locRow.UpdatedAt
	return r.GetBatchLocation(ctx, loc.ID)
}

func (r *postgresPharmacyRepository) GetBatchLocation(ctx context.Context, id int64) (*db.BatchLocation, error) {
	row, err := r.q(ctx).GetBatchLocation(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("batch location not found")
	}
	if err != nil {
		return nil, err
	}
	loc := toBatchLocationGetRow(row)
	return &loc, nil
}

func (r *postgresPharmacyRepository) UpdateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error) {
	if err := r.q(ctx).UpdateBatchLocation(ctx, db.UpdateBatchLocationParams{
		ID:       loc.ID,
		Location: loc.Location,
		Quantity: loc.Quantity,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("batch location not found")
		}
		return nil, err
	}
	return r.GetBatchLocation(ctx, loc.ID)
}

func (r *postgresPharmacyRepository) DeleteBatchLocation(ctx context.Context, id int64) error {
	q := r.q(ctx)
	referenceCount, err := q.CountPrescriptionAllocationsForLocation(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check prescription references: %w", err)
	}
	if referenceCount > 0 {
		return fmt.Errorf("cannot delete batch location because it is allocated to %d prescription line(s)", referenceCount)
	}
	return q.DeleteBatchLocation(ctx, id)
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
		// Use pointer-based expiry sort keys
		t1 := batches[i].ExpirySortKey
		t2 := batches[j].ExpirySortKey
		if t1 == nil && t2 != nil {
			return false
		}
		if t1 != nil && t2 == nil {
			return true
		}
		if t1 == nil && t2 == nil {
			return batches[i].DrugBatch.BatchNumber < batches[j].DrugBatch.BatchNumber
		}
		if t1.Equal(*t2) {
			return batches[i].DrugBatch.BatchNumber < batches[j].DrugBatch.BatchNumber
		}
		return t1.Before(*t2)
	})
	total := 0
	for _, b := range batches {
		total += int(b.DrugBatch.Quantity)
	}

	return &entities.DrugStock{
		Drug:     *dv,
		Batches:  batches,
		TotalQty: total,
	}, nil
}

// -----------------------------------------------------------------------------
// Helpers (conversion and null handling)
// -----------------------------------------------------------------------------

func toDrugFromGetRow(row db.GetDrugRow) db.Drug {
	return db.Drug{
		ID:                  row.ID,
		GenericName:         row.GenericName,
		BrandName:           row.BrandName,
		DrugCode:            row.DrugCode,
		DosageFormCode:      row.DosageFormCode,
		RouteCode:           row.RouteCode,
		StrengthNum:         row.StrengthNum,
		StrengthUnitNum:     row.StrengthUnitNum,
		StrengthDen:         row.StrengthDen,
		StrengthUnitDen:     row.StrengthUnitDen,
		DispenseUnit:        row.DispenseUnit,
		PieceContentAmount:  row.PieceContentAmount,
		PieceContentUnit:    row.PieceContentUnit,
		IsFractionalAllowed: row.IsFractionalAllowed,
		DisplayAsPercentage: row.DisplayAsPercentage,
		Barcode:             row.Barcode,
		Notes:               row.Notes,
		IsActive:            row.IsActive,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func toInsertDrugParams(d *db.Drug) db.InsertDrugParams {
	return db.InsertDrugParams{
		GenericName:        d.GenericName,
		BrandName:          d.BrandName,
		DrugCode:           d.DrugCode,
		DosageFormCode:     d.DosageFormCode,
		RouteCode:          d.RouteCode,
		StrengthNum:        d.StrengthNum,
		StrengthUnitNum:    d.StrengthUnitNum,
		StrengthDen:        d.StrengthDen,
		StrengthUnitDen:    d.StrengthUnitDen,
		DispenseUnit:       d.DispenseUnit,
		PieceContentAmount: d.PieceContentAmount,
		PieceContentUnit:   d.PieceContentUnit,
		Column13:           d.IsFractionalAllowed,
		Column14:           d.DisplayAsPercentage,
		Barcode:            d.Barcode,
		Notes:              d.Notes,
		Column17:           d.IsActive,
	}
}

func toUpdateDrugParams(d *db.Drug) db.UpdateDrugParams {
	return db.UpdateDrugParams{
		ID:                  d.ID,
		GenericName:         d.GenericName,
		BrandName:           d.BrandName,
		DrugCode:            d.DrugCode,
		DosageFormCode:      d.DosageFormCode,
		RouteCode:           d.RouteCode,
		StrengthNum:         d.StrengthNum,
		StrengthUnitNum:     d.StrengthUnitNum,
		StrengthDen:         d.StrengthDen,
		StrengthUnitDen:     d.StrengthUnitDen,
		DispenseUnit:        d.DispenseUnit,
		PieceContentAmount:  d.PieceContentAmount,
		PieceContentUnit:    d.PieceContentUnit,
		IsFractionalAllowed: d.IsFractionalAllowed,
		DisplayAsPercentage: d.DisplayAsPercentage,
		Barcode:             d.Barcode,
		Notes:               d.Notes,
		IsActive:            d.IsActive,
	}
}

func toBatchLocationRow(row db.ListBatchLocationsByBatchRow) db.BatchLocation {
	return db.BatchLocation{
		ID:        row.ID,
		BatchID:   row.BatchID,
		Location:  row.Location,
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toBatchLocationFromIDsRow(row db.ListBatchLocationsByBatchIDsRow) db.BatchLocation {
	return db.BatchLocation{
		ID:        row.ID,
		BatchID:   row.BatchID,
		Location:  row.Location,
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toBatchLocationGetRow(row db.GetBatchLocationRow) db.BatchLocation {
	return db.BatchLocation{
		ID:        row.ID,
		BatchID:   row.BatchID,
		Location:  row.Location,
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toBatchFromListRow(row db.ListBatchesByDrugRow) db.DrugBatch {
	return db.DrugBatch{
		ID:          row.ID,
		DrugID:      row.DrugID,
		BatchNumber: row.BatchNumber,
		ExpiryDate:  row.ExpiryDate,
		Supplier:    row.Supplier,
		Quantity:    row.Quantity,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toBatchFromGetRow(row db.GetBatchRow) db.DrugBatch {
	return db.DrugBatch{
		ID:          row.ID,
		DrugID:      row.DrugID,
		BatchNumber: row.BatchNumber,
		ExpiryDate:  row.ExpiryDate,
		Supplier:    row.Supplier,
		Quantity:    row.Quantity,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
