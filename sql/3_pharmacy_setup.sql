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
  ('bottle',FALSE,FALSE,TRUE),
  ('sachet',FALSE,FALSE,TRUE),   -- discrete pieces like tablets
  ('inhaler',FALSE,FALSE,TRUE),  -- discrete pieces like tablets
  ('puff',FALSE,FALSE,TRUE),    -- content unit for inhalers (e.g., 200 puffs per inhaler)
  ('tube',FALSE,FALSE,TRUE);     -- discrete pieces like bottles, can have piece_content (e.g., 30g tube)

-- === Dosage forms ============================================================
CREATE TABLE dosage_forms (
  code TEXT PRIMARY KEY,                 -- 'TAB','CAP','SYR','SUSP','CREAM','DROP','INJ'
  label TEXT NOT NULL
);

INSERT INTO dosage_forms(code,label) VALUES
  ('TAB','Tablet'),('CAP','Capsule'),('SYR','Syrup'),('SUSP','Suspension'),
  ('CREAM','Cream/Ointment'),('DROP','Drops'),('INJ','Injection'),('INH','Inhaler'), ('SAT','Sachet');

-- === Routes ==================================================================
CREATE TABLE routes (
  code TEXT PRIMARY KEY,                 -- 'PO','IV','IM','TOP','OTIC','OPH'
  label TEXT NOT NULL
);

INSERT INTO routes(code,label) VALUES
  ('PO','Oral'),('IV','Intravenous'),('IM','Intramuscular'),
  ('TOP','Topical'),('OTIC','Ear'),('OPH','Eye'),('INH','Inhalation'), ('NAS','Nasal');


/*******************
 Drugs
********************/

CREATE TABLE drugs (
  id                BIGSERIAL PRIMARY KEY,
  generic_name      TEXT NOT NULL,           -- e.g., Amoxicillin
  brand_name        TEXT,                    -- optional
  atc_code          TEXT,                    -- optional coding
  
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
  display_as_percentage BOOLEAN DEFAULT FALSE,   -- If true, show concentration as % (e.g., 1% instead of 1 g/100 g)

  is_active           BOOLEAN NOT NULL DEFAULT TRUE,

  barcode             TEXT,                      -- optional scan support
  notes               TEXT,

  created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by          BIGINT REFERENCES users(id),
  updated_at          TIMESTAMPTZ,
  updated_by          BIGINT REFERENCES users(id),

  CONSTRAINT uq_drug_identity UNIQUE (
    generic_name, brand_name, dosage_form_code, route_code,
    strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit
  ),

  -- Either solid OR liquid/cream style is valid, OR unknown strength (all strength fields NULL):
  CONSTRAINT ck_drug_style CHECK (
    -- Solid with known strength
    (strength_den IS NULL AND strength_unit_den IS NULL AND strength_num IS NOT NULL AND strength_unit_num IS NOT NULL)
      OR
    -- Liquid/cream with known concentration
    (strength_den IS NOT NULL AND strength_unit_den IS NOT NULL AND strength_num IS NOT NULL AND strength_unit_num IS NOT NULL)
      OR
    -- Unknown strength (all strength fields NULL) - allowed for piece-based dispensing
    (strength_den IS NULL AND strength_unit_den IS NULL AND strength_num IS NULL AND strength_unit_num IS NULL)
  )
);

CREATE TRIGGER trg_drugs_audit
BEFORE INSERT OR UPDATE ON drugs
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_drugs_log
AFTER INSERT OR UPDATE OR DELETE ON drugs
FOR EACH ROW EXECUTE FUNCTION audit_row();




/*******************
 Drug Batches and Batch Location
********************/
CREATE TABLE drug_batches (
  id             BIGSERIAL PRIMARY KEY,
  drug_id        BIGINT NOT NULL REFERENCES drugs(id) ON DELETE RESTRICT,
  batch_number   TEXT NOT NULL,
  expiry_date    DATE,                   -- nullable allowed
  supplier       TEXT,
  quantity       INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0), -- in dispense_unit

  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by     BIGINT REFERENCES users(id),
  updated_at     TIMESTAMPTZ,
  updated_by     BIGINT REFERENCES users(id),

  UNIQUE (drug_id, batch_number)
);

CREATE INDEX idx_batches_drug ON drug_batches (drug_id);
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

-- Helper function to get dispense_unit from drug (used in triggers)
CREATE OR REPLACE FUNCTION get_drug_dispense_unit(p_drug_id BIGINT)
RETURNS TEXT AS $$
DECLARE
  du TEXT;
BEGIN
  SELECT dispense_unit INTO du FROM drugs WHERE id = p_drug_id;
  RETURN du;
END;
$$ LANGUAGE plpgsql STABLE;


DROP TRIGGER IF EXISTS trg_bl_sync_qty ON batch_locations;
CREATE TRIGGER trg_bl_sync_qty
AFTER INSERT OR UPDATE OR DELETE ON batch_locations
FOR EACH ROW EXECUTE FUNCTION sync_batch_quantity();

BEGIN;

-- ---------------------------------------------------------------------------
-- DRUGS (combined with presentations)
--    All numeric strengths support up to 1 decimal place; strength_den NULL for solids.
-- ---------------------------------------------------------------------------

-- Paracetamol 500 mg TAB PO (solid → tab piece, no denominator)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Paracetamol', 'Panadol', 'N02BE01', 'TAB', 'PO', 500, 'mg', NULL, NULL, 'tab', NULL, NULL, FALSE, NULL, '500 mg tablet')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Amoxicillin 250 mg/5 mL SUSP PO (liquid → bottle piece w/ 100 mL per bottle)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Amoxicillin', 'Amoxil', 'J01CA04', 'SUSP', 'PO', 250, 'mg', 5, 'mL', 'bottle', 100, 'mL', FALSE, NULL, '250 mg/5 mL; 100 mL bottle')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Ibuprofen 100 mg/5 mL SYR PO (liquid → continuous mL, fractional allowed)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Ibuprofen', 'Nurofen', 'M01AE01', 'SYR', 'PO', 100, 'mg', 5, 'mL', 'mL', NULL, NULL, TRUE, NULL, '100 mg/5 mL syrup')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Hydrocortisone 1 g/100 g CREAM TOP (cream → continuous g)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Hydrocortisone', 'Hytone', 'D07AA02', 'CREAM', 'TOP', 1, 'g', 100, 'g', 'g', NULL, NULL, FALSE, NULL, '1% (1 g/100 g) cream')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Gentamicin eye drops 3 mg/mL DROP OPH (liquid → bottle 10 mL)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Gentamicin', 'Garamycin', 'S01AA11', 'DROP', 'OPH', 3, 'mg', 1, 'mL', 'bottle', 10, 'mL', FALSE, NULL, '0.3% (3 mg/mL) ophthalmic drops; 10 mL')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Ciprofloxacin 200 mg/100 mL INJ IV (infusion → continuous mL)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Ciprofloxacin', 'Cipro IV', 'J01MA02', 'INJ', 'IV', 200, 'mg', 100, 'mL', 'mL', NULL, NULL, FALSE, NULL, '200 mg/100 mL IV bag')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- Vitamin D3 1000 IU/tab TAB PO (solid → tab)
INSERT INTO drugs (
  generic_name, brand_name, atc_code, dosage_form_code, route_code,
  strength_num, strength_unit_num, strength_den, strength_unit_den,
  dispense_unit, piece_content_amount, piece_content_unit, is_fractional_allowed, barcode, notes
)
VALUES
  ('Cholecalciferol', 'Vit D3 1000IU', 'A11CC05', 'TAB', 'PO', 1000, 'IU', NULL, NULL, 'tab', NULL, NULL, FALSE, NULL, 'Vitamin D3 1000 IU tablet')
ON CONFLICT (generic_name, brand_name, dosage_form_code, route_code, strength_num, strength_unit_num, strength_den, strength_unit_den, dispense_unit)
DO NOTHING;

-- ---------------------------------------------------------------------------
-- 3) BATCHES + LOCATIONS
--    We insert batches (quantity=0), then split stock by locations.
--    The sync trigger will recompute batch quantity from locations.
-- ---------------------------------------------------------------------------

-- Helper: upsert a batch and split stock to locations in one shot
-- (repeatable pattern via CTE)
-- Paracetamol 500 mg TAB: batches PAN500-A, PAN500-B (piece = tabs)
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Paracetamol' AND brand_name='Panadol'
    AND dosage_form_code='TAB' AND route_code='PO'
    AND strength_num=500 AND strength_unit_num='mg'
    AND strength_den IS NULL AND dispense_unit='tab'
),
b1 AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'PAN500-A', DATE '2027-01-31', 'Acme Pharma', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
  RETURNING id, drug_id
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
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT (SELECT drug_id FROM d), 'PAN500-B', DATE '2028-03-31', 'Acme Pharma', 0
  ON CONFLICT (drug_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b2), 'Main Pharmacy', 60
ON CONFLICT (batch_id, location) DO NOTHING;

-- Amoxicillin 250 mg/5 mL SUSP (bottle pieces): AMX250-100ML-01
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Amoxicillin' AND brand_name='Amoxil'
    AND dosage_form_code='SUSP' AND route_code='PO'
    AND strength_num=250 AND strength_unit_num='mg'
    AND strength_den=5 AND strength_unit_den='mL'
    AND dispense_unit='bottle' AND piece_content_amount=100 AND piece_content_unit='mL'
),
b AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'AMX250-100ML-01', DATE '2026-11-30', 'MediSupply Co', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Main Pharmacy', 30
ON CONFLICT (batch_id, location) DO NOTHING;

-- Ibuprofen 100 mg/5 mL SYR (continuous mL): IBU100-5ML-LOT1 total 500 mL
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Ibuprofen' AND brand_name='Nurofen'
    AND dosage_form_code='SYR' AND route_code='PO'
    AND strength_num=100 AND strength_unit_num='mg'
    AND strength_den=5 AND strength_unit_den='mL'
    AND dispense_unit='mL'
),
b AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'IBU100-5ML-LOT1', DATE '2026-08-31', 'WellPharma', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
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
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Hydrocortisone' AND brand_name='Hytone'
    AND dosage_form_code='CREAM' AND route_code='TOP'
    AND strength_num=1 AND strength_unit_num='g'
    AND strength_den=100 AND strength_unit_den='g'
    AND dispense_unit='g'
),
b AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'HC1-2027A', DATE '2027-05-31', 'DermaPro', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
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
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Gentamicin' AND brand_name='Garamycin'
    AND dosage_form_code='DROP' AND route_code='OPH'
    AND strength_num=3 AND strength_unit_num='mg'
    AND strength_den=1 AND strength_unit_den='mL'
    AND dispense_unit='bottle' AND piece_content_amount=10 AND piece_content_unit='mL'
),
b AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'GEN-OPH-10ML', DATE '2026-04-30', 'EyeCare Dist', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Main Pharmacy', 25
ON CONFLICT (batch_id, location) DO NOTHING;

-- Ciprofloxacin IV 200 mg/100 mL (continuous mL): CIPRO-IV-200-LOTX total 300 mL
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Ciprofloxacin' AND brand_name='Cipro IV'
    AND dosage_form_code='INJ' AND route_code='IV'
    AND strength_num=200 AND strength_unit_num='mg'
    AND strength_den=100 AND strength_unit_den='mL'
    AND dispense_unit='mL'
),
b AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'CIPRO-IV-200-LOTX', DATE '2026-09-30', 'Hospisupply', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
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
WITH d AS (
  SELECT id AS drug_id
  FROM drugs
  WHERE generic_name='Cholecalciferol' AND brand_name='Vit D3 1000IU'
    AND dosage_form_code='TAB' AND route_code='PO'
    AND strength_num=1000 AND strength_unit_num='IU'
    AND strength_den IS NULL AND dispense_unit='tab'
),
b AS (
  INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier, quantity)
  SELECT d.drug_id, 'VITD3-1K-01', DATE '2028-12-31', 'NutraHealth', 0 FROM d
  ON CONFLICT (drug_id, batch_number) DO NOTHING
  RETURNING id
)
INSERT INTO batch_locations (batch_id, location, quantity)
SELECT (SELECT id FROM b), 'Main Pharmacy', 200
ON CONFLICT (batch_id, location) DO NOTHING;

COMMIT;
