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
  batch_no     TEXT NOT NULL,
  location     TEXT,                      -- e.g., "Main Store", "Mobile Clinic A"
  notes        TEXT,
  quantity     INTEGER NOT NULL CHECK (quantity >= 0),
  expiry_date  DATE NOT NULL,
  supplier     TEXT,
  depleted_at  TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (drug_id, batch_no)
);

CREATE TRIGGER trg_drug_batches_touch_updated_at
BEFORE UPDATE ON drug_batches
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- Helpful indexes for FEFO picking and lookups
CREATE INDEX IF NOT EXISTS idx_drug_batches_drug_expiry ON drug_batches (drug_id, expiry_date);
CREATE INDEX IF NOT EXISTS idx_drug_batches_not_depleted ON drug_batches (drug_id) WHERE depleted_at IS NULL;

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
  quantity         INTEGER NOT NULL CHECK (quantity > 0),  -- total quantity prescribed for this drug
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
  id                     BIGSERIAL PRIMARY KEY,
  drug_prescription_id   BIGINT NOT NULL REFERENCES drug_prescriptions(id) ON DELETE CASCADE,
  drug_batch_id          BIGINT NOT NULL REFERENCES drug_batches(id),
  quantity               INTEGER NOT NULL CHECK (quantity > 0),
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_prescription_batch_items_touch_updated_at
BEFORE UPDATE ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

CREATE INDEX IF NOT EXISTS idx_pbi_prescription ON prescription_batch_items (drug_prescription_id);
CREATE INDEX IF NOT EXISTS idx_pbi_batch ON prescription_batch_items (drug_batch_id);

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

-- Capture the IDs for consistent seeding below
WITH d AS (
  SELECT id, name FROM drugs
)
SELECT * FROM d;

-- Batches
-- Make sure each (drug_id, batch_no) is unique and quantities cover the splits below
INSERT INTO drug_batches (drug_id, batch_no, location, notes, quantity, expiry_date, supplier)
SELECT id, 'PCM-2401', 'Main Store',  'FEFO older batch', 300, DATE '2025-12-31', 'Acme Pharma'
FROM drugs WHERE name = 'Paracetamol'
UNION ALL
SELECT id, 'PCM-2407', 'Mobile Clinic A', NULL, 500, DATE '2026-06-30', 'Acme Pharma'
FROM drugs WHERE name = 'Paracetamol'
UNION ALL
SELECT id, 'AMX-2402', 'Main Store',  NULL, 200, DATE '2025-11-30', 'GoodMeds Co.'
FROM drugs WHERE name = 'Amoxicillin'
UNION ALL
SELECT id, 'AMX-2410', 'Main Store',  NULL, 400, DATE '2026-10-31', 'GoodMeds Co.'
FROM drugs WHERE name = 'Amoxicillin'
UNION ALL
SELECT id, 'IBU-2403', 'Mobile Clinic B', NULL, 600, DATE '2026-03-31', 'Healthy Labs'
FROM drugs WHERE name = 'Ibuprofen'
UNION ALL
SELECT id, 'ORS-2401', 'Main Store',  NULL, 250, DATE '2026-01-31', 'HydraHealth'
FROM drugs WHERE name = 'Oral Rehydration Salts'
UNION ALL
SELECT id, 'MET-2405', 'Main Store',  NULL, 350, DATE '2026-05-31', 'BetaCare'
FROM drugs WHERE name = 'Metformin';

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

-- Drug prescriptions
-- For P1
INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
SELECT p.id, d.id, 16, 'Paracetamol 500mg: 1 tab q6h PRN, 4 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 1 AND d.name = 'Paracetamol';

INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
SELECT p.id, d.id, 12, 'Ibuprofen 200mg: 1 tab q8h PRN, 4 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 1 AND d.name = 'Ibuprofen';

-- For P2
INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
SELECT p.id, d.id, 6, 'ORS: 1 sachet after each loose stool (max 6)'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 2 AND d.name = 'Oral Rehydration Salts';

INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
SELECT p.id, d.id, 8, 'Paracetamol 500mg: 1 tab q8h PRN, 3 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 1 AND p.vid = 2 AND d.name = 'Paracetamol';

-- For P3
INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
SELECT p.id, d.id, 21, 'Amoxicillin 500mg: 1 cap TID, 7 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 2 AND p.vid = 1 AND d.name = 'Amoxicillin';

INSERT INTO drug_prescriptions (prescription_id, drug_id, quantity, remarks)
SELECT p.id, d.id, 12, 'Paracetamol 500mg: 1 tab q8h PRN, 4 days'
FROM prescriptions p, drugs d
WHERE p.patient_id = 2 AND p.vid = 1 AND d.name = 'Paracetamol';

-- Batch splits (FEFO-style: earlier expiry first)
-- P1: PCM 16 tabs -> split across older then newer batch
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 10
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'PCM-2401'
WHERE p.patient_id = 1 AND p.vid = 1;

INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 6
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'PCM-2407'
WHERE p.patient_id = 1 AND p.vid = 1;

-- P1: Ibuprofen 12 tabs -> single batch
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 12
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Ibuprofen'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'IBU-2403'
WHERE p.patient_id = 1 AND p.vid = 1;

-- P2: ORS 6 sachets -> single batch
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 6
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Oral Rehydration Salts'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'ORS-2401'
WHERE p.patient_id = 1 AND p.vid = 2;

-- P2: PCM 8 tabs -> take from older batch
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 8
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'PCM-2401'
WHERE p.patient_id = 1 AND p.vid = 2;

-- P3: Amoxicillin 21 caps -> split 11/10 across older/newer batches
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 11
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Amoxicillin'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'AMX-2402'
WHERE p.patient_id = 2 AND p.vid = 1;

INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 10
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Amoxicillin'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'AMX-2410'
WHERE p.patient_id = 2 AND p.vid = 1;

-- P3: PCM 12 tabs -> consume older batch first
INSERT INTO prescription_batch_items (drug_prescription_id, drug_batch_id, quantity)
SELECT dp.id, db.id, 12
FROM drug_prescriptions dp
JOIN prescriptions p ON p.id = dp.prescription_id
JOIN drugs d ON d.id = dp.drug_id AND d.name = 'Paracetamol'
JOIN drug_batches db ON db.drug_id = d.id AND db.batch_no = 'PCM-2401'
WHERE p.patient_id = 2 AND p.vid = 1;
