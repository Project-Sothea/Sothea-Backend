-- Prescriptions --------------------------------------------------------------

-- name: InsertPrescription :one
INSERT INTO prescriptions (patient_id, vid, notes)
VALUES ($1,$2,$3)
RETURNING id, created_at, updated_at;

-- name: GetPrescriptionHeader :one
SELECT p.id, p.patient_id, p.vid, p.notes,
       p.created_by, p.created_at, p.updated_at, p.updated_by,
       p.is_dispensed, p.dispensed_by, p.dispensed_at,
       uc.name AS creator_name,
       uu.name AS updater_name,
       ud.name AS dispenser_name
FROM prescriptions p
LEFT JOIN users uc ON uc.id = p.created_by
LEFT JOIN users uu ON uu.id = p.updated_by
LEFT JOIN users ud ON ud.id = p.dispensed_by
WHERE p.id = $1;

-- name: ListPrescriptionLines :many
SELECT
  pl.id, pl.prescription_id, pl.drug_id, pl.remarks, pl.prn,
  pl.dose_amount, pl.dose_unit,
  pl.frequency_code,
  pl.duration, pl.duration_unit,
  pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
  u_packer.name AS packer_name,
  u_updater.name AS updater_name,
  u_creator.name AS creator_name,
  d.generic_name AS drug_name,
  d.route_code AS route_code,
  d.dispense_unit AS dispense_unit,
  CASE
    WHEN d.strength_num IS NULL AND d.strength_unit_num IS NULL THEN d.dosage_form_code
    WHEN d.strength_den IS NULL THEN d.strength_num::text || ' ' || d.strength_unit_num || '/' || d.dispense_unit
    ELSE d.strength_num::text || ' ' || d.strength_unit_num || '/' ||
         d.strength_den::text || ' ' || d.strength_unit_den
  END AS display_strength
FROM prescription_lines pl
LEFT JOIN drugs d ON d.id = pl.drug_id
LEFT JOIN users u_packer ON u_packer.id = pl.packed_by
LEFT JOIN users u_updater ON u_updater.id = pl.updated_by
LEFT JOIN users u_creator ON u_creator.id = pl.created_by
WHERE pl.prescription_id = $1
ORDER BY pl.id;

-- name: ListAllocationsByLineIDs :many
SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
FROM prescription_batch_items
WHERE line_id = ANY($1::bigint[])
ORDER BY line_id, id;

-- name: ListPrescriptionsAll :many
SELECT id, patient_id, vid, notes,
       created_by, created_at, updated_at,
       is_dispensed, dispensed_by, dispensed_at
FROM prescriptions
ORDER BY created_at DESC;

-- name: ListPrescriptionsByPatient :many
SELECT id, patient_id, vid, notes,
       created_by, created_at, updated_at,
       is_dispensed, dispensed_by, dispensed_at
FROM prescriptions
WHERE patient_id = $1
ORDER BY created_at DESC;

-- name: ListPrescriptionsByPatientVisit :many
SELECT id, patient_id, vid, notes,
       created_by, created_at, updated_at,
       is_dispensed, dispensed_by, dispensed_at
FROM prescriptions
WHERE patient_id = $1 AND vid = $2
ORDER BY created_at DESC;

-- name: GetPrescriptionDispensed :one
SELECT is_dispensed FROM prescriptions WHERE id=$1;

-- name: UpdatePrescriptionNotes :exec
UPDATE prescriptions
SET notes=$2, updated_at=now()
WHERE id=$1;

-- name: UpdatePrescriptionFull :exec
UPDATE prescriptions
SET patient_id=$2, vid=$3, notes=$4, updated_at=now()
WHERE id=$1;

-- name: DeleteAllocationsByPrescription :exec
DELETE FROM prescription_batch_items
WHERE line_id IN (SELECT id FROM prescription_lines WHERE prescription_id=$1);

-- name: DeleteLinesByPrescription :exec
DELETE FROM prescription_lines WHERE prescription_id=$1;

-- name: DeletePrescription :exec
DELETE FROM prescriptions WHERE id=$1;

-- Lines ----------------------------------------------------------------------

-- name: InsertLine :one
INSERT INTO prescription_lines (
  prescription_id, drug_id, remarks, prn,
  dose_amount, dose_unit,
  frequency_code,
  duration, duration_unit
) VALUES (
  $1,$2,$3,$4,
  $5,$6,
  $7,
  $8,$9
) RETURNING id, total_to_dispense, is_packed;

-- name: GetPrescriptionIDForLine :one
SELECT prescription_id FROM prescription_lines WHERE id=$1;

-- name: GetLineGuardForUpdate :one
SELECT drug_id,
       dose_amount, dose_unit,
       frequency_code,
       duration, duration_unit,
       remarks, prn
FROM prescription_lines
WHERE id=$1
FOR UPDATE;

-- name: UpdateLine :one
UPDATE prescription_lines SET
  drug_id=$2, remarks=$3, prn=$4,
  dose_amount=$5, dose_unit=$6,
  frequency_code=$7,
  duration=$8, duration_unit=$9,
  is_packed=FALSE, packed_by=NULL, packed_at=NULL,
  updated_at=NOW()
WHERE id=$1
RETURNING total_to_dispense;

-- name: DeleteAllocationsByLine :exec
DELETE FROM prescription_batch_items WHERE line_id=$1;

-- name: DeleteLine :exec
DELETE FROM prescription_lines WHERE id=$1;

-- name: ListAllocationsByLine :many
SELECT id, line_id, batch_location_id, quantity, created_at, updated_at
FROM prescription_batch_items
WHERE line_id=$1
ORDER BY id;

-- name: InsertAllocation :one
INSERT INTO prescription_batch_items (line_id, batch_location_id, quantity)
VALUES ($1,$2,$3)
RETURNING id, created_at, updated_at;

-- name: MarkLinePacked :exec
UPDATE prescription_lines
SET is_packed=TRUE,
    packed_by = current_setting('sothea.user_id')::bigint,
    packed_at=NOW()
WHERE id=$1;

-- name: UnpackLine :exec
UPDATE prescription_lines
SET is_packed=FALSE, packed_by=NULL, packed_at=NULL, updated_at=NOW()
WHERE id=$1;

-- name: CountLinesForPrescription :one
SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1;

-- name: CountPackedLinesForPrescription :one
SELECT COUNT(*) FROM prescription_lines WHERE prescription_id=$1 AND is_packed=TRUE;

-- name: DispensePrescription :exec
UPDATE prescriptions
SET is_dispensed=TRUE,
    dispensed_by = current_setting('sothea.user_id')::bigint,
    dispensed_at=NOW(),
    updated_at=NOW()
WHERE id=$1;

-- name: GetLineWithDispenseUnit :one
SELECT
  pl.id, pl.prescription_id, pl.drug_id, pl.remarks, pl.prn,
  pl.dose_amount, pl.dose_unit,
  pl.frequency_code,
  pl.duration, pl.duration_unit,
  pl.total_to_dispense, pl.is_packed, pl.packed_by, pl.packed_at,
  (SELECT dispense_unit FROM drugs WHERE id=pl.drug_id) AS du
FROM prescription_lines pl
WHERE pl.id=$1;
