-- Drugs ----------------------------------------------------------------------

-- name: ListDrugs :many
SELECT id, generic_name, brand_name, drug_code, dosage_form_code, route_code,
       strength_num, strength_unit_num, strength_den, strength_unit_den,
       dispense_unit, piece_content_amount, piece_content_unit,
       is_fractional_allowed, display_as_percentage, barcode, notes, is_active, created_at, updated_at
FROM drugs
ORDER BY generic_name, COALESCE(brand_name, ''), dosage_form_code, route_code;

-- name: SearchDrugs :many
SELECT id, generic_name, brand_name, drug_code, dosage_form_code, route_code,
       strength_num, strength_unit_num, strength_den, strength_unit_den,
       dispense_unit, piece_content_amount, piece_content_unit,
       is_fractional_allowed, display_as_percentage, barcode, notes, is_active, created_at, updated_at
FROM drugs
WHERE generic_name ILIKE $1 OR COALESCE(brand_name, '') ILIKE $1
ORDER BY generic_name, COALESCE(brand_name, ''), dosage_form_code, route_code;

-- name: InsertDrug :one
INSERT INTO drugs (
  generic_name, brand_name, drug_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit,
  is_fractional_allowed, display_as_percentage, barcode, notes, is_active
) VALUES (
  $1,$2,$3,$4,$5,
  $6,$7,$8,$9,
  $10,$11,$12,
  COALESCE($13,FALSE),COALESCE($14,FALSE),$15,$16,COALESCE($17,TRUE)
) RETURNING id;

-- name: GetDrug :one
SELECT id, generic_name, brand_name, drug_code, dosage_form_code, route_code,
       strength_num, strength_unit_num, strength_den, strength_unit_den,
       dispense_unit, piece_content_amount, piece_content_unit,
       is_fractional_allowed, display_as_percentage, barcode, notes, is_active, created_at, updated_at
FROM drugs
WHERE id = $1;

-- name: UpdateDrug :exec
UPDATE drugs SET
  generic_name=$2, brand_name=$3, drug_code=$4,
  dosage_form_code=$5, route_code=$6,
  strength_num=$7, strength_unit_num=$8, strength_den=$9, strength_unit_den=$10,
  dispense_unit=$11, piece_content_amount=$12, piece_content_unit=$13,
  is_fractional_allowed=$14, display_as_percentage=$15, barcode=$16, notes=$17, is_active=$18, updated_at=NOW()
WHERE id=$1;

-- name: DeleteDrug :exec
DELETE FROM drugs WHERE id=$1;

-- name: CountPrescriptionLinesForDrug :one
SELECT COUNT(*) AS cnt
FROM prescription_lines
WHERE drug_id = $1;

-- Batches --------------------------------------------------------------------

-- name: ListBatchesByDrug :many
SELECT id, drug_id, batch_number, expiry_date, supplier, quantity, created_at, updated_at
FROM drug_batches
WHERE drug_id = $1
ORDER BY expiry_date NULLS LAST, batch_number, id;

-- name: GetBatch :one
SELECT id, drug_id, batch_number, expiry_date, supplier, quantity, created_at, updated_at
FROM drug_batches
WHERE id = $1;

-- name: InsertBatch :one
INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
VALUES ($1,$2,$3,$4,COALESCE($5,0))
RETURNING id;

-- name: UpdateBatch :exec
UPDATE drug_batches
SET batch_number=$2, expiry_date=$3, supplier=$4
WHERE id=$1;

-- name: DeleteBatch :exec
DELETE FROM drug_batches WHERE id=$1;

-- name: ListBatchLocationsByBatch :many
SELECT id, batch_id, location, quantity, created_at, updated_at
FROM batch_locations
WHERE batch_id = $1
ORDER BY location, id;

-- name: ListBatchLocationsByBatchIDs :many
SELECT id, batch_id, location, quantity, created_at, updated_at
FROM batch_locations
WHERE batch_id = ANY($1::bigint[])
ORDER BY batch_id, location, id;

-- name: InsertBatchLocation :one
INSERT INTO batch_locations (batch_id, location, quantity)
VALUES ($1,$2,$3)
RETURNING id, created_at, updated_at;

-- name: GetBatchLocation :one
SELECT id, batch_id, location, quantity, created_at, updated_at
FROM batch_locations
WHERE id = $1;

-- name: UpdateBatchLocation :exec
UPDATE batch_locations
SET location=$2, quantity=$3, updated_at=NOW()
WHERE id=$1;

-- name: DeleteBatchLocation :exec
DELETE FROM batch_locations WHERE id=$1;

-- name: CountPrescriptionAllocationsForLocation :one
SELECT COUNT(*) AS cnt
FROM prescription_batch_items
WHERE batch_location_id = $1;

-- name: CountPrescriptionAllocationsForBatch :one
SELECT COUNT(*) AS cnt
FROM prescription_batch_items pbi
JOIN batch_locations bl ON bl.id = pbi.batch_location_id
WHERE bl.batch_id = $1;
