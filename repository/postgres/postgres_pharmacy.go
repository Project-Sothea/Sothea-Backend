package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jieqiboh/sothea_backend/entities"
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
	rows, err := r.Conn.QueryContext(ctx, qListDrugs)
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
	err := r.Conn.QueryRowContext(ctx, qCreateDrug,
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
	err := r.Conn.QueryRowContext(ctx, qGetDrug, id).
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

	res, err := r.Conn.ExecContext(ctx, qUpdateDrug,
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

	res, err := r.Conn.ExecContext(ctx, qDeleteDrug, id)
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
//  DRUG BATCHES  (stock entries)
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) ListBatches(
	ctx context.Context,
	drugID *int64, // nil = show all
) ([]entities.DrugBatch, error) {

	base := `
SELECT id, drug_id, batch_no, location,
       quantity, expiry_date, supplier, depleted_at
FROM   drug_batches`
	var rows *sql.Rows
	var err error

	if drugID != nil {
		rows, err = r.Conn.QueryContext(ctx,
			base+" WHERE drug_id=$1 ORDER BY expiry_date", *drugID)
	} else {
		rows, err = r.Conn.QueryContext(ctx,
			base+" ORDER BY drug_id, expiry_date")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.DrugBatch{}
	for rows.Next() {
		var b entities.DrugBatch
		if err := rows.Scan(&b.ID, &b.DrugID, &b.BatchNumber, &b.Location,
			&b.Quantity, &b.ExpiryDate, &b.Supplier, &b.DepletedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

const qCreateBatch = `
INSERT INTO drug_batches
    (drug_id, batch_no, location, quantity, expiry_date, supplier)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id;`

func (r *postgresPharmacyRepository) CreateBatch(ctx context.Context, b *entities.DrugBatch) (int64, error) {
	var id int64
	err := r.Conn.QueryRowContext(ctx, qCreateBatch,
		b.DrugID, b.BatchNumber, b.Location,
		b.Quantity, b.ExpiryDate, b.Supplier,
	).Scan(&id)

	return id, err
}

const qUpdateBatch = `
UPDATE drug_batches
SET    drug_id=$2,
       batch_no=$3,
       location=$4,
       quantity=$5,
       expiry_date=$6,
       supplier=$7,
       depleted_at=$8,
       updated_at=NOW()
WHERE  id=$1;`

func (r *postgresPharmacyRepository) UpdateBatch(
	ctx context.Context, b *entities.DrugBatch) error {

	res, err := r.Conn.ExecContext(ctx, qUpdateBatch,
		b.ID, b.DrugID, b.BatchNumber, b.Location,
		b.Quantity, b.ExpiryDate, b.Supplier, b.DepletedAt)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return errors.New("batch not found")
	}
	return nil
}

const qDeleteBatch = `DELETE FROM drug_batches WHERE id=$1;`

func (r *postgresPharmacyRepository) DeleteBatch(
	ctx context.Context, id int64) error {

	res, err := r.Conn.ExecContext(ctx, qDeleteBatch, id)
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
//  (Optional) FEFO helper – not wired yet
// -----------------------------------------------------------------------------

func (r *postgresPharmacyRepository) earliestBatches(
	ctx context.Context, drugID int64,
) ([]entities.DrugBatch, error) {

	const q = `
SELECT id, drug_id, batch_no, location,
       quantity, expiry_date, supplier, depleted_at
FROM   drug_batches
WHERE  drug_id = $1 AND quantity > 0
ORDER  BY expiry_date;`

	rows, err := r.Conn.QueryContext(ctx, q, drugID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entities.DrugBatch
	for rows.Next() {
		var b entities.DrugBatch
		if err := rows.Scan(&b.ID, &b.DrugID, &b.BatchNumber, &b.Location,
			&b.Quantity, &b.ExpiryDate, &b.Supplier, &b.DepletedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
