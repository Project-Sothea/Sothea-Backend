/*******************
 Drop in dependency order
********************/
DROP TABLE IF EXISTS prescription_batch_items;
DROP TABLE IF EXISTS drug_prescriptions;
DROP TABLE IF EXISTS prescriptions;
DROP TABLE IF EXISTS drug_batches;
DROP TABLE IF EXISTS drugs;

/*******************
 Helper: updated_at trigger used by multiple tables
********************/
CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

/*******************
 Create Drug Catalog
********************/
CREATE TABLE IF NOT EXISTS drugs (
  id           BIGSERIAL PRIMARY KEY,
  name         TEXT NOT NULL UNIQUE,
  unit         TEXT NOT NULL,             -- e.g., "tablet", "ml", "sachet"
  default_size INTEGER,                   -- e.g., 500 (mg), 200 (mg), etc.
  notes        TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_drugs_touch_updated_at
BEFORE UPDATE ON drugs
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

/*******************
 Create Inventory Table
********************/
CREATE TABLE IF NOT EXISTS drug_batches (
  id           BIGSERIAL PRIMARY KEY,
  drug_id      BIGINT NOT NULL REFERENCES drugs(id) ON DELETE CASCADE,
  batch_number TEXT NOT NULL,
  notes        TEXT,
  expiry_date  DATE NOT NULL,
  supplier     TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (drug_id, batch_number)
);

CREATE TRIGGER trg_drug_batches_touch_updated_at
BEFORE UPDATE ON drug_batches
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

/*******************
 Create Batch Location Table
********************/
CREATE TABLE IF NOT EXISTS batch_locations (
  id             BIGSERIAL PRIMARY KEY,
  batch_id        BIGINT NOT NULL REFERENCES drug_batches(id) ON DELETE CASCADE,
  location       TEXT NOT NULL,
  quantity       INTEGER NOT NULL CHECK (quantity >= 0),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (batch_id, location)
);

CREATE TRIGGER trg_batch_locations_touch_updated_at
BEFORE UPDATE ON batch_locations
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Helpful indexes (FEFO joins & lookups by location text)
CREATE INDEX IF NOT EXISTS idx_bl_batch ON batch_locations (batch_id);
CREATE INDEX IF NOT EXISTS idx_bl_location ON batch_locations (location);

/*******************
 Prescriptions (FK to admin(id, vid))
********************/
CREATE TABLE IF NOT EXISTS prescriptions (
  id         BIGSERIAL PRIMARY KEY,
  patient_id INTEGER NOT NULL,
  vid        INTEGER NOT NULL,
  staff_id   INTEGER,
  notes      TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  is_packed BOOLEAN NOT NULL DEFAULT false,
  CONSTRAINT fk_admin FOREIGN KEY (patient_id, vid) REFERENCES admin(id, vid)
);

CREATE TRIGGER trg_prescriptions_touch_updated_at
BEFORE UPDATE ON prescriptions
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE INDEX IF NOT EXISTS idx_prescriptions_patient_visit ON prescriptions (patient_id, vid);

/*******************
 Drug prescriptions (line items)
********************/
CREATE TABLE IF NOT EXISTS drug_prescriptions (
  id               BIGSERIAL PRIMARY KEY,
  prescription_id  BIGINT NOT NULL REFERENCES prescriptions(id) ON DELETE CASCADE,
  drug_id          BIGINT NOT NULL REFERENCES drugs(id),
  remarks          TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_drug_prescriptions_touch_updated_at
BEFORE UPDATE ON drug_prescriptions
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE INDEX IF NOT EXISTS idx_drug_prescriptions_prescription ON drug_prescriptions (prescription_id);
CREATE INDEX IF NOT EXISTS idx_drug_prescriptions_drug ON drug_prescriptions (drug_id);

/*******************
 Batch splits for each drug prescription
********************/
CREATE TABLE IF NOT EXISTS prescription_batch_items (
  id                      BIGSERIAL PRIMARY KEY,
  drug_prescription_id    BIGINT NOT NULL REFERENCES drug_prescriptions(id) ON DELETE CASCADE,
  drug_batch_location_id  BIGINT NOT NULL REFERENCES batch_locations(id),
  quantity               INTEGER NOT NULL CHECK (quantity > 0),
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_prescription_batch_items_touch_updated_at
BEFORE UPDATE ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE INDEX IF NOT EXISTS idx_pbi_prescription ON prescription_batch_items (drug_prescription_id);
CREATE INDEX IF NOT EXISTS idx_pbi_batch ON prescription_batch_items (drug_batch_location_id);

/*******************
 Seed data
 (Assumes you already have admin rows like (patient_id, vid) = (1,1), (1,2), (2,1))
********************/

-- Drugs
INSERT INTO drugs (name, unit, default_size, notes) VALUES
  ('Paracetamol',   'tablet', 500, 'Analgesic/antipyretic'),
  ('Amoxicillin',   'capsule', 500, 'Antibiotic'),
  ('Ibuprofen',     'tablet', 200, 'NSAID'),
  ('Oral Rehydration Salts', 'sachet', NULL, 'ORS sachets'),
  ('Metformin',     'tablet', 500, 'For T2DM')
ON CONFLICT (name) DO NOTHING;

-- Drug batches (NO location/quantity columns anymore)
INSERT INTO drug_batches (drug_id, batch_number, notes, expiry_date, supplier)
SELECT id, 'PCM-2401', 'FEFO older batch', DATE '2025-12-31', 'Acme Pharma'
FROM drugs WHERE name = 'Paracetamol'
UNION ALL
SELECT id, 'PCM-2407', NULL, DATE '2026-06-30', 'Acme Pharma'
FROM drugs WHERE name = 'Paracetamol'
UNION ALL
SELECT id, 'AMX-2402', NULL, DATE '2025-11-30', 'GoodMeds Co.'
FROM drugs WHERE name = 'Amoxicillin'
UNION ALL
SELECT id, 'AMX-2410', NULL, DATE '2026-10-31', 'GoodMeds Co.'
FROM drugs WHERE name = 'Amoxicillin'
UNION ALL
SELECT id, 'IBU-2403', NULL, DATE '2026-03-31', 'Healthy Labs'
FROM drugs WHERE name = 'Ibuprofen'
UNION ALL
SELECT id, 'ORS-2401', NULL, DATE '2026-01-31', 'HydraHealth'
FROM drugs WHERE name = 'Oral Rehydration Salts'
UNION ALL
SELECT id, 'MET-2405', NULL, DATE '2026-05-31', 'BetaCare'
FROM drugs WHERE name = 'Metformin';

-- Per-location quantities (FREE-TEXT location)
-- Mirrors your previous single-location + quantity values
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Main Store', 300
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Paracetamol' AND b.batch_number = 'PCM-2401';

INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Mobile Clinic A', 500
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Paracetamol' AND b.batch_number = 'PCM-2407';

INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Main Store', 200
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Amoxicillin' AND b.batch_number = 'AMX-2402';

INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Main Store', 400
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Amoxicillin' AND b.batch_number = 'AMX-2410';

INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Mobile Clinic B', 600
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Ibuprofen' AND b.batch_number = 'IBU-2403';

INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Main Store', 250
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Oral Rehydration Salts' AND b.batch_number = 'ORS-2401';

INSERT INTO batch_locations (batch_id, location, quantity)
SELECT b.id, 'Main Store', 350
FROM drug_batches b
JOIN drugs d ON d.id = b.drug_id
WHERE d.name = 'Metformin' AND b.batch_number = 'MET-2405';

-- Prescriptions (link to existing patients/visits)
-- P1: patient 1, visit 1
INSERT INTO prescriptions (patient_id, vid, staff_id, notes)
VALUES (1, 1, 101, 'Fever and myalgia; start PCM + Ibuprofen');

-- P2: patient 1, visit 2
INSERT INTO prescriptions (patient_id, vid, staff_id, notes)
VALUES (1, 2, 102, 'Dehydration; give ORS; continue PCM as needed');

-- P3: patient 2, visit 1
INSERT INTO prescriptions (patient_id, vid, staff_id, notes)
VALUES (2, 1, 201, 'Pharyngitis; amoxicillin course; PCM PRN');

-- Drug prescriptions (NOTE: no quantity column anymore; totals live in batch splits)
-- For P1
INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks)
SELECT p.id, d.id, 'Paracetamol 500mg: 1 tab q6h PRN, 4 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 1 AND d.name = 'Paracetamol';

INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks)
SELECT p.id, d.id, 'Ibuprofen 200mg: 1 tab q8h PRN, 4 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 1 AND d.name = 'Ibuprofen';

-- For P2
INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks)
SELECT p.id, d.id, 'ORS: 1 sachet after each loose stool (max 6)'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 2 AND d.name = 'Oral Rehydration Salts';

INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks)
SELECT p.id, d.id, 'Paracetamol 500mg: 1 tab q8h PRN, 3 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 2 AND d.name = 'Paracetamol';

-- For P3
INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks)
SELECT p.id, d.id, 'Amoxicillin 500mg: 1 cap TID, 7 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 2 AND p.vid = 1 AND d.name = 'Amoxicillin';

INSERT INTO drug_prescriptions (prescription_id, drug_id, remarks)
SELECT p.id, d.id, 'Paracetamol 500mg: 1 tab q8h PRN, 4 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 2 AND p.vid = 1 AND d.name = 'Paracetamol';

-- Batch splits (FEFO-style) — now insert INTO prescription_batch_items using batch_location rows
-- P1: PCM 16 tabs -> 10 from PCM-2401 (Main Store), 6 from PCM-2407 (Mobile Clinic A)
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 10
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'PCM-2401'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Main Store'
WHERE p.patient_id = 1 AND p.vid = 1;

INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 6
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'PCM-2407'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Mobile Clinic A'
WHERE p.patient_id = 1 AND p.vid = 1;

-- P1: Ibuprofen 12 tabs -> from IBU-2403 at Mobile Clinic B
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 12
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Ibuprofen'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'IBU-2403'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Mobile Clinic B'
WHERE p.patient_id = 1 AND p.vid = 1;

-- P2: ORS 6 sachets -> ORS-2401 at Main Store
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 6
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Oral Rehydration Salts'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'ORS-2401'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Main Store'
WHERE p.patient_id = 1 AND p.vid = 2;

-- P2: PCM 8 tabs -> take from older PCM-2401 at Main Store
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 8
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'PCM-2401'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Main Store'
WHERE p.patient_id = 1 AND p.vid = 2;

-- P3: Amoxicillin 21 caps -> split 11/10 across older/newer batches (both Main Store)
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 11
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Amoxicillin'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'AMX-2402'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Main Store'
WHERE p.patient_id = 2 AND p.vid = 1;

INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 10
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Amoxicillin'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'AMX-2410'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Main Store'
WHERE p.patient_id = 2 AND p.vid = 1;

-- P3: PCM 12 tabs -> consume older PCM-2401 at Main Store
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_location_id, quantity)
SELECT dp.id, bl.id, 12
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches b ON b.drug_id = d.id AND b.batch_number = 'PCM-2401'
JOIN batch_locations bl ON bl.batch_id = b.id AND bl.location = 'Main Store'
WHERE p.patient_id = 2 AND p.vid = 1;
