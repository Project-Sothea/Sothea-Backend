/*******************
 Create Foundational Vocab
********************/

-- === Units (use numeric with 1 decimal place for amounts) =========================
CREATE TABLE IF NOT EXISTS units (
  code TEXT PRIMARY KEY,                  -- 'mg','g','mcg','mL','L','IU','tab','cap','drop','g'
  is_mass   BOOLEAN NOT NULL DEFAULT FALSE,
  is_volume BOOLEAN NOT NULL DEFAULT FALSE,
  is_piece  BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO units(code,is_mass,is_volume,is_piece) VALUES
  ('mcg', TRUE,FALSE,FALSE),
  ('mg',  TRUE,FALSE,FALSE),
  ('g',   TRUE,FALSE,FALSE),
  ('mL',  FALSE,TRUE,FALSE),
  ('L',   FALSE,TRUE,FALSE),
  ('IU',  TRUE,FALSE,FALSE),     -- treat IU like mass-potency
  ('tab', FALSE,FALSE,TRUE),
  ('cap', FALSE,FALSE,TRUE),
  ('drop',FALSE,FALSE,TRUE),
  ('bottle',FALSE,FALSE,TRUE);

-- === Dosage forms ============================================================
CREATE TABLE dosage_forms (
  code TEXT PRIMARY KEY,                 -- 'TAB','CAP','SYR','SUSP','CREAM','DROP','INJ'
  label TEXT NOT NULL
);

INSERT INTO dosage_forms(code,label) VALUES
  ('TAB','Tablet'),('CAP','Capsule'),('SYR','Syrup'),('SUSP','Suspension'),
  ('CREAM','Cream/Ointment'),('DROP','Drops'),('INJ','Injection'),('INH','Inhaler');

-- === Routes ==================================================================
CREATE TABLE routes (
  code TEXT PRIMARY KEY,                 -- 'PO','IV','IM','TOP','OTIC','OPH'
  label TEXT NOT NULL
);

INSERT INTO routes(code,label) VALUES
  ('PO','Oral'),('IV','Intravenous'),('IM','Intramuscular'),
  ('TOP','Topical'),('OTIC','Ear'),('OPH','Eye'),('INH','Inhalation');


/*******************
 Drugs
********************/

CREATE TABLE drugs (
  id          BIGSERIAL PRIMARY KEY,
  generic_name TEXT NOT NULL,           -- e.g., Amoxicillin
  brand_name   TEXT,                    -- optional
  atc_code     TEXT,                    -- optional coding
  notes        TEXT,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by   BIGINT REFERENCES users(id),
  updated_at   TIMESTAMPTZ,
  updated_by   BIGINT REFERENCES users(id),

  CONSTRAINT uq_drug_identity UNIQUE (generic_name, brand_name)
);

CREATE TRIGGER trg_drugs_audit
BEFORE INSERT OR UPDATE ON drugs
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_drugs_log
AFTER INSERT OR UPDATE OR DELETE ON drugs
FOR EACH ROW EXECUTE FUNCTION audit_row();

/*******************
 Presentation (drug metadata)
********************/

CREATE TABLE drug_presentations (
  id                BIGSERIAL PRIMARY KEY,
  drug_id           BIGINT NOT NULL REFERENCES drugs(id) ON DELETE CASCADE,
  dosage_form_code  TEXT NOT NULL REFERENCES dosage_forms(code),
  route_code        TEXT NOT NULL REFERENCES routes(code),

  -- Strength/concentration:
  -- Solids:  strength_num / piece (e.g., 500 mg per tab) => den* fields NULL
  -- Liquids/creams: strength_num / strength_den (e.g., 250 mg / 5 mL, or 1 g / 100 g)
  strength_num        NUMERIC(10,1),                   -- numerator amount (up to 1 decimal place)
  strength_unit_num   TEXT REFERENCES units(code),
  strength_den        NUMERIC(10,1),                   -- denominator amount (NULL for solids, up to 1 decimal place)
  strength_unit_den   TEXT REFERENCES units(code),

  -- What pharmacy counts down (inventory base unit):
  dispense_unit       TEXT NOT NULL REFERENCES units(code),

  piece_content_amount NUMERIC(10,1),
  piece_content_unit   TEXT REFERENCES units(code),

  is_fractional_allowed BOOLEAN DEFAULT FALSE,

  barcode             TEXT,                      -- optional scan support
  notes               TEXT,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by          BIGINT REFERENCES users(id),
  updated_at          TIMESTAMPTZ,
  updated_by          BIGINT REFERENCES users(id),

  CONSTRAINT uq_presentation UNIQUE (
    drug_id, dosage_form_code, route_code,
    strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit
  ),

  -- Either solid OR liquid/cream style is valid, OR unknown strength (all strength fields NULL):
  CONSTRAINT ck_presentation_style CHECK (
    -- Solid with known strength
    (strength_den IS NULL AND strength_unit_den IS NULL AND strength_num IS NOT NULL AND strength_unit_num IS NOT NULL)
      OR
    -- Liquid/cream with known concentration
    (strength_den IS NOT NULL AND strength_unit_den IS NOT NULL AND strength_num IS NOT NULL AND strength_unit_num IS NOT NULL)
      OR
    -- Unknown strength (all strength fields NULL) - allowed for piece-based dispensing
    (strength_den IS NULL AND strength_unit_den IS NULL AND strength_num IS NULL AND strength_unit_num IS NULL)
  ),


);

CREATE TRIGGER trg_presentations_audit
BEFORE INSERT OR UPDATE ON drug_presentations
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_presentations_log
AFTER INSERT OR UPDATE OR DELETE ON drug_presentations
FOR EACH ROW EXECUTE FUNCTION audit_row();




/*******************
 Drug Batches and Batch Location
********************/
CREATE TABLE drug_batches (
  id             BIGSERIAL PRIMARY KEY,
  presentation_id BIGINT NOT NULL REFERENCES drug_presentations(id) ON DELETE RESTRICT,
  batch_number   TEXT NOT NULL,
  expiry_date    DATE,                   -- nullable allowed
  supplier       TEXT,
  quantity       INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0), -- in dispense_unit

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by     BIGINT REFERENCES users(id),
  updated_at     TIMESTAMPTZ,
  updated_by     BIGINT REFERENCES users(id),

  UNIQUE (presentation_id, batch_number)
);

CREATE INDEX idx_batches_presentation ON drug_batches (presentation_id);
CREATE INDEX idx_batches_expiry ON drug_batches (expiry_date);

CREATE TRIGGER trg_batches_audit
BEFORE INSERT OR UPDATE ON drug_batches
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_batches_log
AFTER INSERT OR UPDATE OR DELETE ON drug_batches
FOR EACH ROW EXECUTE FUNCTION audit_row();

-- Per-location splits
CREATE TABLE batch_locations (
  id             BIGSERIAL PRIMARY KEY,
  batch_id       BIGINT NOT NULL REFERENCES drug_batches(id) ON DELETE CASCADE,
  location       TEXT NOT NULL,
  quantity       INTEGER NOT NULL CHECK (quantity >= 0), -- in dispense_unit

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by     BIGINT REFERENCES users(id),
  updated_at     TIMESTAMPTZ,
  updated_by     BIGINT REFERENCES users(id),

  UNIQUE (batch_id, location)
);

CREATE INDEX idx_bl_batch ON batch_locations (batch_id);
CREATE INDEX idx_bl_location ON batch_locations (location);

CREATE TRIGGER trg_bl_audit
BEFORE INSERT OR UPDATE ON batch_locations
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_bl_log
AFTER INSERT OR UPDATE OR DELETE ON batch_locations
FOR EACH ROW EXECUTE FUNCTION audit_row();

CREATE OR REPLACE FUNCTION sync_batch_quantity()
RETURNS TRIGGER AS $$
DECLARE
  bid BIGINT := COALESCE(NEW.batch_id, OLD.batch_id);
  s   INT;
BEGIN
  SELECT COALESCE(SUM(quantity),0) INTO s FROM batch_locations WHERE batch_id = bid;
  UPDATE drug_batches SET quantity = s WHERE id = bid;
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS trg_bl_sync_qty ON batch_locations;
CREATE TRIGGER trg_bl_sync_qty
AFTER INSERT OR UPDATE OR DELETE ON batch_locations
FOR EACH ROW EXECUTE FUNCTION sync_batch_quantity();

BEGIN;

-- ---------------------------------------------------------------------------
-- 1) DRUGS
-- ---------------------------------------------------------------------------
INSERT INTO drugs (generic_name, brand_name, atc_code, notes)
VALUES
  ('Paracetamol',      'Panadol',        'N02BE01', 'Analgesic/antipyretic'),
  ('Amoxicillin',      'Amoxil',         'J01CA04', 'Penicillin antibiotic'),
  ('Ibuprofen',        'Nurofen',        'M01AE01', 'NSAID'),
  ('Hydrocortisone',   'Hytone',         'D07AA02', 'Topical steroid'),
  ('Gentamicin',       'Garamycin',      'S01AA11', 'Aminoglycoside (ophthalmic)'),
  ('Ciprofloxacin',    'Cipro IV',       'J01MA02', 'Fluoroquinolone (IV infusion)'),
  ('Cholecalciferol',  'Vit D3 1000IU',  'A11CC05', 'Vitamin D3 tablets')
ON CONFLICT (generic_name, brand_name) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 2) PRESENTATIONS
--    Note: use SELECT to bind to existing drug IDs. All numeric strengths
--    support up to 1 decimal place; strength_den NULL for solids.
-- ---------------------------------------------------------------------------

-- Paracetamol 500 mg TAB PO (solid → tab piece, no denominator)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'TAB','PO', 500,'mg', NULL,NULL, 'tab', NULL,NULL, FALSE, NULL, '500 mg tablet'
FROM drugs d
WHERE d.generic_name='Paracetamol' AND d.brand_name='Panadol'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Amoxicillin 250 mg/5 mL SUSP PO (liquid → bottle piece w/ 100 mL per bottle)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'SUSP','PO', 250,'mg', 5,'mL', 'bottle', 100,'mL', FALSE, NULL, '250 mg/5 mL; 100 mL bottle'
FROM drugs d
WHERE d.generic_name='Amoxicillin' AND d.brand_name='Amoxil'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Ibuprofen 100 mg/5 mL SYR PO (liquid → continuous mL, fractional allowed)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'SYR','PO', 100,'mg', 5,'mL', 'mL', NULL,NULL, TRUE, NULL, '100 mg/5 mL syrup'
FROM drugs d
WHERE d.generic_name='Ibuprofen' AND d.brand_name='Nurofen'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Hydrocortisone 1 g/100 g CREAM TOP (cream → continuous g)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'CREAM','TOP', 1,'g', 100,'g', 'g', NULL,NULL, FALSE, NULL, '1% (1 g/100 g) cream'
FROM drugs d
WHERE d.generic_name='Hydrocortisone' AND d.brand_name='Hytone'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Gentamicin eye drops 3 mg/mL DROP OPH (liquid → bottle 10 mL)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'DROP','OPH', 3,'mg', 1,'mL', 'bottle', 10,'mL', FALSE, NULL, '0.3% (3 mg/mL) ophthalmic drops; 10 mL'
FROM drugs d
WHERE d.generic_name='Gentamicin' AND d.brand_name='Garamycin'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Ciprofloxacin 200 mg/100 mL INJ IV (infusion → continuous mL)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'INJ','IV', 200,'mg', 100,'mL', 'mL', NULL,NULL, FALSE, NULL, '200 mg/100 mL IV bag'
FROM drugs d
WHERE d.generic_name='Ciprofloxacin' AND d.brand_name='Cipro IV'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Vitamin D3 1000 IU/tab TAB PO (solid → tab)
INSERT INTO drug_presentations (
  drug_id, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
SELECT d.id, 'TAB','PO', 1000,'IU', NULL,NULL, 'tab', NULL,NULL, FALSE, NULL, 'Vitamin D3 1000 IU tablet'
FROM drugs d
WHERE d.generic_name='Cholecalciferol' AND d.brand_name='Vit D3 1000IU'
ON CONFLICT (drug_id, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3) BATCHES + LOCATIONS
--    We insert batches (quantity=0), then split stock by locations.
--    The sync trigger will recompute batch quantity from locations.
-- ---------------------------------------------------------------------------

-- Helper: upsert a batch and split stock to locations in one shot
-- (repeatable pattern via CTE)
-- Paracetamol 500 mg TAB: batches PAN500-A, PAN500-B (piece = tabs)
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Paracetamol' AND d.brand_name='Panadol'
    AND dp.dosage_form_code='TAB' AND dp.route_code='PO'
    AND dp.strength_num=500 AND dp.strength_unit_num='mg'
    AND dp.strength_den IS NULL AND dp.dispense_unit='tab'
),
b1 AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'PAN500-A', DATE '2027-01-31', 'Acme Pharma', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id, presentation_id
),
b1q AS (
  -- split: Main=80, Clinic A=40
  INSERT INTO batch_locations (batch_id, location, quantity)
  SELECT b1.id, 'Main Pharmacy', 80 FROM b1
  ON CONFLICT (batch_id, location) DO NOTHING
),
b1q2 AS (
  INSERT INTO batch_locations (batch_id, location, quantity)
  SELECT (SELECT id FROM b1), 'Clinic A', 40
  ON CONFLICT (batch_id, location) DO NOTHING
),
b2 AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'PAN500-B', DATE '2028-03-31', 'Acme Pharma', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b2), 'Main Pharmacy', 60
ON CONFLICT (batch_id, location) DO NOTHING;

-- Amoxicillin 250 mg/5 mL SUSP (bottle pieces): AMX250-100ML-01
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Amoxicillin' AND d.brand_name='Amoxil'
    AND dp.dosage_form_code='SUSP' AND dp.route_code='PO'
    AND dp.strength_num=250 AND dp.strength_unit_num='mg'
    AND dp.strength_den=5 AND dp.strength_unit_den='mL'
    AND dp.dispense_unit='bottle' AND dp.piece_content_amount=100 AND dp.piece_content_unit='mL'
),
b AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'AMX250-100ML-01', DATE '2026-11-30', 'MediSupply Co', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Main Pharmacy', 30
ON CONFLICT (batch_id, location) DO NOTHING;

-- Ibuprofen 100 mg/5 mL SYR (continuous mL): IBU100-5ML-LOT1 total 500 mL
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Ibuprofen' AND d.brand_name='Nurofen'
    AND dp.dosage_form_code='SYR' AND dp.route_code='PO'
    AND dp.strength_num=100 AND dp.strength_unit_num='mg'
    AND dp.strength_den=5 AND dp.strength_unit_den='mL'
    AND dp.dispense_unit='mL'
),
b AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'IBU100-5ML-LOT1', DATE '2026-08-31', 'WellPharma', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
),
l1 AS (
  INSERT INTO batch_locations (batch_id, location, quantity)
  SELECT (SELECT id FROM b), 'Main Pharmacy', 300
  ON CONFLICT (batch_id, location) DO NOTHING
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Clinic A', 200
ON CONFLICT (batch_id, location) DO NOTHING;

-- Hydrocortisone 1% cream (continuous g): HC1-2027A total 400 g
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Hydrocortisone' AND d.brand_name='Hytone'
    AND dp.dosage_form_code='CREAM' AND dp.route_code='TOP'
    AND dp.strength_num=1 AND dp.strength_unit_num='g'
    AND dp.strength_den=100 AND dp.strength_unit_den='g'
    AND dp.dispense_unit='g'
),
b AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'HC1-2027A', DATE '2027-05-31', 'DermaPro', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
),
l1 AS (
  INSERT INTO batch_locations (batch_id, location, quantity)
  SELECT (SELECT id FROM b), 'Main Pharmacy', 250
  ON CONFLICT (batch_id, location) DO NOTHING
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Ward A', 150
ON CONFLICT (batch_id, location) DO NOTHING;

-- Gentamicin eye drops 3 mg/mL (bottle 10 mL pieces): GEN-OPH-10ML
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Gentamicin' AND d.brand_name='Garamycin'
    AND dp.dosage_form_code='DROP' AND dp.route_code='OPH'
    AND dp.strength_num=3 AND dp.strength_unit_num='mg'
    AND dp.strength_den=1 AND dp.strength_unit_den='mL'
    AND dp.dispense_unit='bottle' AND dp.piece_content_amount=10 AND dp.piece_content_unit='mL'
),
b AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'GEN-OPH-10ML', DATE '2026-04-30', 'EyeCare Dist', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Main Pharmacy', 25
ON CONFLICT (batch_id, location) DO NOTHING;

-- Ciprofloxacin IV 200 mg/100 mL (continuous mL): CIPRO-IV-200-LOTX total 300 mL
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Ciprofloxacin' AND d.brand_name='Cipro IV'
    AND dp.dosage_form_code='INJ' AND dp.route_code='IV'
    AND dp.strength_num=200 AND dp.strength_unit_num='mg'
    AND dp.strength_den=100 AND dp.strength_unit_den='mL'
    AND dp.dispense_unit='mL'
),
b AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'CIPRO-IV-200-LOTX', DATE '2026-09-30', 'Hospisupply', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
),
l1 AS (
  INSERT INTO batch_locations (batch_id, location, quantity)
  SELECT (SELECT id FROM b), 'Main Pharmacy', 200
  ON CONFLICT (batch_id, location) DO NOTHING
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Ward B', 100
ON CONFLICT (batch_id, location) DO NOTHING;

-- Vitamin D3 1000 IU/tab (solid → tab pieces): VITD3-1K-01
WITH pr AS (
  SELECT dp.id AS presentation_id
  FROM drug_presentations dp
  JOIN drugs d ON d.id = dp.drug_id
  WHERE d.generic_name='Cholecalciferol' AND d.brand_name='Vit D3 1000IU'
    AND dp.dosage_form_code='TAB' AND dp.route_code='PO'
    AND dp.strength_num=1000 AND dp.strength_unit_num='IU'
    AND dp.strength_den IS NULL AND dp.dispense_unit='tab'
),
b AS (
  INSERT INTO drug_batches (presentation_id, batch_number, expiry_date, supplier, quantity)
  SELECT pr.presentation_id, 'VITD3-1K-01', DATE '2028-12-31', 'NutraHealth', 0 FROM pr
  ON CONFLICT (presentation_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Main Pharmacy', 200
ON CONFLICT (batch_id, location) DO NOTHING;

COMMIT;
