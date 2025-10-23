package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	return &postgresPharmacyRepository{conn}
}

// -----------------------------------------------------------------------------
//  DRUGS  (catalog)
// -----------------------------------------------------------------------------

const qListDrugs = `
SELECT id, name, unit, default_size, notes
FROM   drugs
ORDER  BY name;`

func (r *postgresPharmacyRepository) ListDrugs(ctx context.Context) ([]entities.Drug, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	rows, err := dbx.QueryContext(ctx, qListDrugs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Drug{}
	for rows.Next() {
		var d entities.Drug
		if err := rows.Scan(&d.ID, &d.Name, &d.Unit, &d.DefaultSize, &d.Notes); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

const qCreateDrug = `
INSERT INTO drugs (name, unit, default_size, notes)
VALUES ($1,$2,$3,$4)
ON CONFLICT (name) DO NOTHING
RETURNING id;`

func (r *postgresPharmacyRepository) CreateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	var id int64
	dbx := DBFromCtx(ctx, r.Conn)
	err := dbx.QueryRowContext(ctx, qCreateDrug,
		d.Name, d.Unit, d.DefaultSize, d.Notes,
	).Scan(&id)
	switch {
	case err == sql.ErrNoRows: // duplicate → no row returned
		return nil, entities.ErrDrugNameTaken

	case err != nil: // other DB error
		return nil, err

	default: // success
		return r.GetDrug(ctx, id)
	}
}

const qGetDrug = `
SELECT id, name, unit, default_size, notes
FROM   drugs
WHERE  id=$1;`

func (r *postgresPharmacyRepository) GetDrug(ctx context.Context, id int64) (*entities.Drug, error) {
	var d entities.Drug
	dbx := DBFromCtx(ctx, r.Conn)
	err := dbx.QueryRowContext(ctx, qGetDrug, id).
		Scan(&d.ID, &d.Name, &d.Unit, &d.DefaultSize, &d.Notes)
	if err == sql.ErrNoRows {
		return nil, errors.New("drug not found")
	}
	return &d, err
}

const qUpdateDrug = `
UPDATE drugs
SET    name=$2,
       unit=$3,
       default_size=$4,
       notes=$5,
       updated_at=NOW()
WHERE  id=$1;`

func (r *postgresPharmacyRepository) UpdateDrug(
	ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qUpdateDrug,
		d.ID, d.Name, d.Unit, d.DefaultSize, d.Notes)
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return nil, errors.New("drug not found")
	}
	return r.GetDrug(ctx, d.ID)
}

const qDeleteDrug = `DELETE FROM drugs WHERE id=$1;`

func (r *postgresPharmacyRepository) DeleteDrug(
	ctx context.Context, id int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qDeleteDrug, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return errors.New("drug not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
//  DRUG BATCHES
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) ListBatchDetails(
	ctx context.Context,
	drugID *int64, // nil = all
) ([]entities.BatchDetail, error) {

	// 1) Fetch batches (optionally filtered by drug)
	var (
		args []any
		q    = `
			SELECT id, drug_id, batch_number, expiry_date, supplier
			FROM drug_batches
		`
	)
	if drugID != nil {
		q += " WHERE drug_id = $1"
		args = append(args, *drugID)
	}
	q += " ORDER BY drug_id, expiry_date, id"

	dbx := DBFromCtx(ctx, r.Conn)
	rows, err := dbx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	batches := make([]entities.DrugBatch, 0, 64)
	batchIDs := make([]int64, 0, 64)

	for rows.Next() {
		var b entities.DrugBatch
		if err := rows.Scan(&b.ID, &b.DrugID, &b.BatchNumber, &b.ExpiryDate, &b.Supplier); err != nil {
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

	// 2) Fetch all locations for those batches in one query
	locRows, err := dbx.QueryContext(ctx, `
		SELECT id, batch_id, location, quantity
		FROM batch_locations
		WHERE batch_id = ANY($1)
		ORDER BY batch_id, location, id
	`, pq.Array(batchIDs))
	if err != nil {
		return nil, err
	}
	defer locRows.Close()

	// batchID -> []locations
	locsByBatch := make(map[int64][]entities.DrugBatchLocation, len(batchIDs))
	for locRows.Next() {
		var l entities.DrugBatchLocation
		if err := locRows.Scan(&l.ID, &l.BatchID, &l.Location, &l.Quantity); err != nil {
			return nil, err
		}
		locsByBatch[l.BatchID] = append(locsByBatch[l.BatchID], l)
	}
	if err := locRows.Err(); err != nil {
		return nil, err
	}

	// 3) Combine
	out := make([]entities.BatchDetail, 0, len(batches))
	for _, b := range batches {
		out = append(out, entities.BatchDetail{
			DrugBatch:      b,
			BatchLocations: locsByBatch[b.ID],
		})
	}
	return out, nil
}

const qGetBatch = `
SELECT id, drug_id, batch_number, expiry_date, supplier
FROM   drug_batches
WHERE  id=$1;`

func (r *postgresPharmacyRepository) GetBatch(ctx context.Context, id int64) (*entities.BatchDetail, error) {
	var batch entities.DrugBatch
	dbx := DBFromCtx(ctx, r.Conn)
	err := dbx.QueryRowContext(ctx, qGetBatch, id).
		Scan(&batch.ID, &batch.DrugID, &batch.BatchNumber, &batch.ExpiryDate, &batch.Supplier)
	if err == sql.ErrNoRows {
		return nil, errors.New("batch not found")
	}

	batchLocations, err := r.ListBatchLocations(ctx, batch.ID)
	if err != nil {
		return nil, err
	}

	var batchDetails = &entities.BatchDetail{
		DrugBatch:      batch,
		BatchLocations: batchLocations,
	}
	return batchDetails, err
}

const qCreateBatch = `
INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier)
VALUES ($1, $2, $3, $4)
RETURNING id;
`

func (r *postgresPharmacyRepository) CreateBatch(ctx context.Context, b *entities.BatchDetail) (*entities.BatchDetail, error) {
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

	// 1) Insert batch
	var batchID int64
	if err := tx.QueryRowContext(ctx, qCreateBatch,
		b.DrugID, b.BatchNumber, b.ExpiryDate, b.Supplier).Scan(&batchID); err != nil {
		return nil, err
	}

	// 2) Insert nested locations if provided
	if len(b.BatchLocations) > 0 {
		stmt, err := tx.PrepareContext(ctx, qCreateBatchLocation)
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		for i := range b.BatchLocations {
			loc := &b.BatchLocations[i]
			if loc.Quantity < 0 {
				return nil, fmt.Errorf("location %q has negative quantity", loc.Location)
			}
			// ensure FK is set
			loc.BatchID = batchID

			var locID int64
			if err := stmt.QueryRowContext(ctx, batchID, loc.Location, loc.Quantity).Scan(&locID); err != nil {
				return nil, err
			}
			loc.ID = locID
		}
	}

	// 3) Commit the whole thing atomically
	if ownTx {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}

	// 4) Return the hydrated detail (locations included)
	return r.GetBatch(ctx, batchID)
}

const qUpdateBatch = `
UPDATE drug_batches
SET drug_id=$2,
    batch_number=$3,
    expiry_date=$4,
    supplier=$5
WHERE id=$1;
`

func (r *postgresPharmacyRepository) UpdateBatch(ctx context.Context, b *entities.DrugBatch) (*entities.BatchDetail, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qUpdateBatch,
		b.ID, b.DrugID, b.BatchNumber, b.ExpiryDate, b.Supplier)
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return nil, errors.New("batch not found")
	}
	return r.GetBatch(ctx, b.ID)
}

const qDeleteBatch = `DELETE FROM drug_batches WHERE id=$1;`

func (r *postgresPharmacyRepository) DeleteBatch(ctx context.Context, id int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qDeleteBatch, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return errors.New("batch not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
//  DRUG BATCH LOCATIONS  (stock entries)
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) ListBatchLocations(ctx context.Context, batchID int64) ([]entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	rows, err := dbx.QueryContext(ctx, `
		SELECT id, batch_id, location, quantity
		FROM batch_locations
		WHERE batch_id = $1
		ORDER BY location, id
	`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entities.DrugBatchLocation
	for rows.Next() {
		var l entities.DrugBatchLocation
		if err := rows.Scan(&l.ID, &l.BatchID, &l.Location, &l.Quantity); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

const qGetBatchLocation = `
SELECT id, batch_id, location, quantity
FROM   batch_locations
WHERE  id=$1;`

func (r *postgresPharmacyRepository) GetBatchLocation(ctx context.Context, id int64) (*entities.DrugBatchLocation, error) {
	var batchLocation entities.DrugBatchLocation
	dbx := DBFromCtx(ctx, r.Conn)
	err := dbx.QueryRowContext(ctx, qGetBatchLocation, id).
		Scan(&batchLocation.ID, &batchLocation.BatchID, &batchLocation.Location, &batchLocation.Quantity)
	if err == sql.ErrNoRows {
		return nil, errors.New("batch location not found")
	}
	return &batchLocation, err
}

const qCreateBatchLocation = `
	INSERT INTO batch_locations (batch_id, location, quantity)
	VALUES ($1, $2, $3)
	RETURNING id
`

func (r *postgresPharmacyRepository) CreateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	var id int64
	dbx := DBFromCtx(ctx, r.Conn)
	err := dbx.QueryRowContext(ctx, qCreateBatchLocation,
		loc.BatchID, loc.Location, loc.Quantity).Scan(&id)
	if err != nil {
		return nil, err
	}
	return r.GetBatchLocation(ctx, id)
}

const qUpdateBatchLocation = `
UPDATE batch_locations
SET batch_id=$2,
    location=$3,
    quantity=$4
WHERE id=$1;
`

func (r *postgresPharmacyRepository) UpdateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qUpdateBatchLocation,
		loc.ID, loc.BatchID, loc.Location, loc.Quantity)
	if err != nil {
		return nil, err
	}
	if aff, _ := res.RowsAffected(); aff == 0 {
		return nil, errors.New("batch location not found")
	}
	return r.GetBatchLocation(ctx, loc.ID)
}

const qDeleteBatchLocation = `
	DELETE FROM batch_locations WHERE id=$1
`

func (r *postgresPharmacyRepository) DeleteBatchLocation(ctx context.Context, id int64) error {
	dbx := DBFromCtx(ctx, r.Conn)
	res, err := dbx.ExecContext(ctx, qDeleteBatchLocation, id)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return errors.New("location not found")
	}
	return nil
}
