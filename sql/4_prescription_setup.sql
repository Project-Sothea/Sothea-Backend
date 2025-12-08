CREATE TABLE prescriptions (
  id           BIGSERIAL PRIMARY KEY,
  patient_id   INTEGER NOT NULL,
  vid          INTEGER NOT NULL,
  notes        TEXT,
  is_dispensed BOOLEAN NOT NULL DEFAULT FALSE,
  dispensed_by BIGINT REFERENCES users(id),
  dispensed_at TIMESTAMPTZ,

  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by   BIGINT REFERENCES users(id),
  updated_at   TIMESTAMPTZ,
  updated_by   BIGINT REFERENCES users(id),

  CONSTRAINT fk_admin FOREIGN KEY (patient_id, vid) REFERENCES admin(id, vid) ON DELETE CASCADE
);

CREATE INDEX idx_rx_patient_visit ON prescriptions (patient_id, vid);

CREATE TRIGGER trg_rx_audit
BEFORE INSERT OR UPDATE ON prescriptions
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_rx_log
AFTER INSERT OR UPDATE OR DELETE ON prescriptions
FOR EACH ROW EXECUTE FUNCTION audit_row();

/*******************
 Frequency
********************/
CREATE TABLE schedule_kinds (
  code TEXT PRIMARY KEY
);

INSERT INTO schedule_kinds(code) VALUES
  ('hour'), ('day'), ('week'), ('month');

-- Frequency codes lookup table
CREATE TABLE frequency_codes (
  code TEXT PRIMARY KEY,              -- OM, ON, BD, TDS, q8h, ...
  label TEXT NOT NULL,                -- human-readable label
  schedule_kind TEXT NOT NULL REFERENCES schedule_kinds(code),
  every_n INTEGER NOT NULL CHECK (every_n > 0),
  frequency_per_schedule NUMERIC(6,2) NOT NULL CHECK (frequency_per_schedule > 0)
);

INSERT INTO frequency_codes (code, label, schedule_kind, every_n, frequency_per_schedule) VALUES
  ('OM',  'Once every morning',        'day',  1, 1),
  ('ON',  'Once every night',          'day',  1, 1),
  ('OA',  'Once every afternoon',      'day',  1, 1),
  ('BD',  '2 times a day',             'day',  1, 2),
  ('TDS', '3 times a day',             'day',  1, 3),
  ('QDS', '4 times a day',             'day',  1, 4),
  ('EOD', 'Every other day',           'day',  2, 1),
  ('q4h', 'Every 4 hours',             'hour', 4, 1),
  ('q6h', 'Every 6 hours',             'hour', 6, 1),
  ('q8h', 'Every 8 hours',             'hour', 8, 1),
  ('q12h','Every 12 hours',            'hour', 12, 1);

-- converts any duration unit to hours
CREATE OR REPLACE FUNCTION duration_to_hours(
  duration NUMERIC,
  duration_unit TEXT
) RETURNS NUMERIC AS $$
BEGIN
  IF duration_unit = 'hour' THEN
    RETURN duration;

  ELSIF duration_unit = 'day' THEN
    RETURN duration * 24;

  ELSIF duration_unit = 'week' THEN
    RETURN duration * 24 * 7;          -- 168

  ELSIF duration_unit = 'month' THEN
    RETURN duration * 730;             -- your chosen avg

  ELSE
    RAISE EXCEPTION 'Unsupported duration_unit: %', duration_unit;
  END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- How many administrations occur, given the periodic schedule
-- periods = ceil(duration / every_n)
-- doses   = periods * frequency_per_schedule
CREATE OR REPLACE FUNCTION dose_count_periodic_pure(
  schedule_kind TEXT,
  every_n INT,
  frequency_per_schedule NUMERIC,
  duration NUMERIC,
  duration_unit TEXT
) RETURNS INT AS $$
DECLARE
  duration_hours NUMERIC;
  every_n_hours NUMERIC;
  periods NUMERIC;
BEGIN
  IF schedule_kind NOT IN ('hour','day','week','month') THEN
    RAISE EXCEPTION 'Unknown schedule_kind: %', schedule_kind;
  END IF;

  IF duration_unit NOT IN ('hour','day','week','month') THEN
    RAISE EXCEPTION 'Unknown duration_unit: %', duration_unit;
  END IF;
  IF every_n <= 0 OR frequency_per_schedule <= 0 OR duration <= 0 THEN
    RAISE EXCEPTION 'every_n, frequency_per_schedule, duration must be > 0';
  END IF;

  -- standardise durtion in same unit as schedule_kind
  duration_hours := duration_to_hours(duration, duration_unit);
  every_n_hours := duration_to_hours(every_n, schedule_kind);
  periods := CEIL(duration_hours / every_n_hours);
  RETURN CEIL(periods * frequency_per_schedule);
END;
$$ LANGUAGE plpgsql IMMUTABLE;


/*******************
 Prescription Lines
********************/

CREATE TABLE prescription_lines (
  id                 BIGSERIAL PRIMARY KEY,
  prescription_id    BIGINT NOT NULL REFERENCES prescriptions(id) ON DELETE CASCADE,
  drug_id            BIGINT NOT NULL REFERENCES drugs(id),
  remarks            TEXT,
  prn                BOOLEAN NOT NULL DEFAULT FALSE, -- pro re nata (as needed)

  -- Clinical dose:
  dose_amount        NUMERIC(10,2) NOT NULL,            -- smallest clinical unit, up to 2 decimal places
  dose_unit          TEXT NOT NULL REFERENCES units(code),

  -- Frequency: must use frequency_code; schedule_* fields are derived
  frequency_code     TEXT NOT NULL REFERENCES frequency_codes(code), 

  duration      NUMERIC(6,2) NOT NULL,
  duration_unit TEXT NOT NULL REFERENCES schedule_kinds(code), 

  -- Computed target to pick (in dispense_unit of the presentation):
  total_to_dispense  INTEGER NOT NULL,            -- set by trigger

  is_packed          BOOLEAN NOT NULL DEFAULT FALSE, -- convenience flags (optional)
  packed_by          BIGINT REFERENCES users(id),
  packed_at          TIMESTAMPTZ,

  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by         BIGINT REFERENCES users(id),
  updated_at         TIMESTAMPTZ,
  updated_by         BIGINT REFERENCES users(id)
);

CREATE INDEX idx_lines_rx ON prescription_lines (prescription_id);
CREATE INDEX idx_lines_drug ON prescription_lines (drug_id);

CREATE TRIGGER trg_lines_audit
BEFORE INSERT OR UPDATE ON prescription_lines
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_lines_log
AFTER INSERT OR UPDATE OR DELETE ON prescription_lines
FOR EACH ROW EXECUTE FUNCTION audit_row();

CREATE OR REPLACE FUNCTION ck_line_stock_possible(
  p_drug_id         BIGINT,
  p_required        INTEGER,
  p_line_id         BIGINT  -- may be NULL on INSERT
)
RETURNS VOID
LANGUAGE plpgsql AS $$
DECLARE
  allocated   INTEGER := 0;
  need_extra  INTEGER;
  available   INTEGER;
BEGIN
  -- Sum current allocations for this line (if any)
  IF p_line_id IS NOT NULL THEN
    SELECT COALESCE(SUM(pbi.quantity), 0)
      INTO allocated
      FROM prescription_batch_items pbi
     WHERE pbi.line_id = p_line_id;
  END IF;

  -- Only the extra beyond what is already allocated must be available
  need_extra := GREATEST(p_required - allocated, 0);

  IF need_extra = 0 THEN
    RETURN; -- no new stock required
  END IF;

  -- Total available stock for the presentation
  SELECT COALESCE(SUM(bl.quantity), 0)
    INTO available
    FROM batch_locations bl
    JOIN drug_batches b ON b.id = bl.batch_id
   WHERE b.drug_id = p_drug_id;

  IF available < need_extra THEN
    RAISE EXCEPTION 'insufficient stock: total needed %, available %', p_required, available
      USING ERRCODE   = '23514',                 -- check_violation
            CONSTRAINT = 'ck_insufficient_stock',
            DETAIL     = json_build_object(
                            'drug_id', p_drug_id,
                            'total_required', p_required,
                            'total_available', available
                          )::text;
  END IF;
END;
$$;

-- PURE calculator: returns how many to pick in the *dispense unit*
-- Uses your new schedule fields.
CREATE OR REPLACE FUNCTION compute_total_to_dispense_pure(
  dose_amount NUMERIC(10,2),
  dose_unit TEXT,
  schedule_kind TEXT,
  every_n INT,
  frequency_per_schedule NUMERIC,
  duration NUMERIC,
  duration_unit TEXT,
  dispense_unit TEXT,
  strength_num NUMERIC(10,1),
  strength_unit_num TEXT,
  strength_den NUMERIC(10,1),                  -- NULL for solids
  strength_unit_den TEXT,            -- NULL for solids
  piece_content_amount NUMERIC(10,1),          -- only for bottle/tube/inhaler; NULL otherwise
  piece_content_unit TEXT            -- as above (e.g., 'mL', 'g', or 'puff')
) RETURNS INT AS $$
DECLARE
  per_dose_dispense NUMERIC;  -- in dispense_unit
  per_dose_liquid   NUMERIC;  -- in piece_content_unit (for bottles)
  total_liquid      NUMERIC;
  total_doses       INT;
  total             NUMERIC;
BEGIN
  -- 1) Compute how many administrations
  total_doses := dose_count_periodic_pure(schedule_kind, every_n, frequency_per_schedule, duration, duration_unit);

  -- 2) Per-administration conversion to "dispense unit"
  
  -- CASE 1: No strength data provided (all strength fields NULL)
  IF strength_num IS NULL AND strength_unit_num IS NULL AND strength_den IS NULL AND strength_unit_den IS NULL THEN
    -- dispense_unit is always valid
    IF dose_unit = dispense_unit THEN
      per_dose_dispense := dose_amount;
    -- If piece-based unit (bottle/tube/inhaler) with piece_content_unit, piece_content_unit is also valid
    ELSIF dispense_unit IN ('bottle', 'tube', 'inhaler') AND piece_content_unit IS NOT NULL AND dose_unit = piece_content_unit THEN
      per_dose_liquid := dose_amount;
      total_liquid := per_dose_liquid * total_doses;
      RETURN CEIL(total_liquid / NULLIF(piece_content_amount, 0));
    ELSE
      RAISE EXCEPTION 'No strength data: dose_unit % must be dispense_unit (%)%', 
        dose_unit, dispense_unit,
        CASE WHEN dispense_unit IN ('bottle', 'tube', 'inhaler') AND piece_content_unit IS NOT NULL 
          THEN ' or piece_content_unit (' || piece_content_unit || ')' 
          ELSE '' END;
    END IF;

  -- CASE 2: Only strength numerator provided (solid)
  ELSIF strength_den IS NULL AND strength_unit_den IS NULL THEN
    -- dispense_unit and strength_unit_num are both valid
    IF dose_unit = dispense_unit THEN
      per_dose_dispense := dose_amount;
    ELSIF dose_unit = strength_unit_num THEN
      -- Convert strength_unit_num to dispense_unit (pieces)
      per_dose_dispense := CEIL((dose_amount::NUMERIC / NULLIF(strength_num, 0)) * 100) / 100;
    ELSE
      RAISE EXCEPTION 'Solid: dose_unit % must be dispense_unit (%) or strength_unit_num (%)', 
        dose_unit, dispense_unit, strength_unit_num;
    END IF;

  -- CASE 3: Both numerator and denominator provided
  ELSE
    -- dispense_unit always allowed
    IF dose_unit = dispense_unit THEN
      per_dose_dispense := dose_amount;
    
    -- CASE 3.1 & 3.2: Not piece-based with content, OR piece-based without piece_content
    ELSIF dispense_unit NOT IN ('bottle', 'tube', 'inhaler') OR (dispense_unit IN ('bottle', 'tube', 'inhaler') AND piece_content_unit IS NULL) THEN
      -- If dispense_unit equals either strength_unit_num OR strength_unit_den,
      -- then BOTH strength_unit_num AND strength_unit_den are valid
      IF dispense_unit = strength_unit_num AND dose_unit = strength_unit_den THEN
        per_dose_dispense := CEIL((dose_amount::NUMERIC * strength_num / NULLIF(strength_den, 0)) * 100) / 100; -- dose_amt * strength_num / strength_den

      ELSIF dispense_unit = strength_unit_den AND dose_unit = strength_unit_num THEN
        per_dose_dispense := CEIL((dose_amount::NUMERIC * strength_den / NULLIF(strength_num, 0)) * 100) / 100; -- dose_amt * strength_den / strength_num
        
      ELSE
        -- dispense_unit doesn't match either strength unit, so only dispense_unit is valid
        -- This case should have been caught at line 275, so this is an error
        RAISE EXCEPTION 'dose_unit % must be dispense_unit (%). Strength unit conversions only allowed when dispense_unit equals strength_unit_num (%) or strength_unit_den (%)', 
          dose_unit, dispense_unit, strength_unit_num, strength_unit_den;
      END IF;

    -- CASE 3.3: Piece-based unit (bottle/tube/inhaler) with piece_content_unit
    ELSIF dispense_unit IN ('bottle', 'tube', 'inhaler') AND piece_content_unit IS NOT NULL THEN
      -- piece_content_unit is always valid
      IF dose_unit = piece_content_unit THEN
        per_dose_liquid := dose_amount;
      -- If piece_content_unit equals either strength_unit_num OR strength_unit_den,
      -- then BOTH strength_unit_num AND strength_unit_den are valid
      ELSIF piece_content_unit = strength_unit_num AND dose_unit = strength_unit_den THEN
        per_dose_liquid := CEIL((dose_amount::NUMERIC * strength_num / NULLIF(strength_den, 0)) * 100) / 100; -- dose_amt * strength_num / strength_den

      ELSIF piece_content_unit = strength_unit_den AND dose_unit = strength_unit_num THEN
        per_dose_liquid := CEIL((dose_amount::NUMERIC * strength_den / NULLIF(strength_num, 0)) * 100) / 100; -- dose_amt * strength_den / strength_num

      ELSE
        -- piece_content_unit doesn't match either strength unit, so only piece_content_unit is valid
        -- This case should have been caught at line 296, so this is an error
        RAISE EXCEPTION 'dose_unit % must be piece_content_unit (%). Strength unit conversions only allowed when piece_content_unit equals strength_unit_num (%) or strength_unit_den (%)', 
          dose_unit, piece_content_unit, strength_unit_num, strength_unit_den;
      END IF;
      
      total_liquid := per_dose_liquid * total_doses;
      RETURN CEIL(total_liquid / NULLIF(piece_content_amount, 0));
    END IF;
  END IF;

  -- 3) Multiply per-dose pick by total doses
  total := per_dose_dispense * total_doses;
  RETURN CEIL(total);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Before insert/update: fill total_to_dispense based on current fields
CREATE OR REPLACE FUNCTION trg_set_total_to_dispense()
RETURNS TRIGGER AS $$
DECLARE
  dp RECORD;
  fc RECORD;
  required_qty INTEGER;
  available_qty INTEGER;
  must_check   BOOLEAN := (TG_OP = 'INSERT');
BEGIN
  -- frequency_code must always be provided
  IF NEW.frequency_code IS NULL THEN
    RAISE EXCEPTION 'frequency_code must be provided';
  END IF;

  -- Lookup the schedule fields from frequency_codes
  SELECT schedule_kind, every_n, frequency_per_schedule
    INTO fc
    FROM frequency_codes
   WHERE code = NEW.frequency_code;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'frequency_code % not found', NEW.frequency_code;
  END IF;

  SELECT dispense_unit, strength_num, strength_unit_num,
         strength_den, strength_unit_den,
         piece_content_amount, piece_content_unit
    INTO dp
    FROM drugs
   WHERE id = NEW.drug_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'drug % not found for line', NEW.drug_id;
  END IF;

  NEW.total_to_dispense := compute_total_to_dispense_pure(
    NEW.dose_amount, NEW.dose_unit,
    fc.schedule_kind, fc.every_n, fc.frequency_per_schedule,
    NEW.duration, NEW.duration_unit,
    dp.dispense_unit, dp.strength_num, dp.strength_unit_num,
    dp.strength_den, dp.strength_unit_den,
    dp.piece_content_amount, dp.piece_content_unit
  );
  required_qty := NEW.total_to_dispense;

  -- Recheck for update when qty-driving fields change
  IF TG_OP = 'UPDATE' THEN
    must_check := must_check OR (
      NEW.drug_id                IS DISTINCT FROM OLD.drug_id OR
      NEW.dose_amount            IS DISTINCT FROM OLD.dose_amount OR
      NEW.dose_unit              IS DISTINCT FROM OLD.dose_unit OR
      NEW.frequency_code         IS DISTINCT FROM OLD.frequency_code OR
      NEW.duration               IS DISTINCT FROM OLD.duration OR
      NEW.duration_unit          IS DISTINCT FROM OLD.duration_unit
    );
  END IF;
  
  IF must_check THEN
    PERFORM ck_line_stock_possible(NEW.drug_id, NEW.total_to_dispense, NEW.id);
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER biu_lines_compute_total
BEFORE INSERT OR UPDATE ON prescription_lines
FOR EACH ROW EXECUTE FUNCTION trg_set_total_to_dispense();

CREATE TABLE prescription_batch_items (
  id                    BIGSERIAL PRIMARY KEY,
  line_id               BIGINT NOT NULL REFERENCES prescription_lines(id) ON DELETE CASCADE,
  batch_location_id     BIGINT NOT NULL REFERENCES batch_locations(id),
  quantity              INTEGER NOT NULL CHECK (quantity > 0), -- in dispense_unit

  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_by            BIGINT REFERENCES users(id),
  updated_at            TIMESTAMPTZ,
  updated_by            BIGINT REFERENCES users(id)
);

CREATE INDEX idx_pbi_line ON prescription_batch_items (line_id);
CREATE INDEX idx_pbi_bl   ON prescription_batch_items (batch_location_id);

CREATE TRIGGER trg_pbi_audit
BEFORE INSERT OR UPDATE ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION set_audit_fields();

CREATE TRIGGER trg_pbi_log
AFTER INSERT OR UPDATE OR DELETE ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION audit_row();

/*
CREATE OR REPLACE FUNCTION assert_packed_quantity()
RETURNS TRIGGER AS $$
DECLARE
  need INT; got INT;
BEGIN
  IF NEW.is_packed AND (OLD.is_packed IS DISTINCT FROM TRUE) THEN
    SELECT total_to_dispense INTO need FROM prescription_lines WHERE id = NEW.id;
    SELECT COALESCE(SUM(quantity),0) INTO got FROM prescription_batch_items WHERE line_id = NEW.id;
    IF need <> got THEN
      RAISE EXCEPTION 'Packed line must allocate exactly total_to_dispense (need %, have %)', need, got;
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS ck_line_packed_alloc ON prescription_lines;
CREATE CONSTRAINT TRIGGER ck_line_packed_alloc
AFTER UPDATE OF is_packed ON prescription_lines
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION assert_packed_quantity();
*/

-- Reserve/return stock when prescription_batch_items change
CREATE OR REPLACE FUNCTION trg_pbi_adjust_stock()
RETURNS TRIGGER AS $$
DECLARE
  old_loc  BIGINT;
  new_loc  BIGINT;
  delta    INTEGER;
  res      INTEGER;
BEGIN
  IF TG_OP = 'INSERT' THEN
    new_loc := NEW.batch_location_id;
    delta   := NEW.quantity;

    -- Decrement with guard + lock row
    UPDATE batch_locations
      SET quantity = quantity - delta
    WHERE id = new_loc AND quantity >= delta
    RETURNING 1 INTO res;

    IF res IS NULL THEN
      RAISE EXCEPTION 'Insufficient stock at location % for reserve of %', new_loc, delta;
    END IF;

    RETURN NEW;

  ELSIF TG_OP = 'DELETE' THEN
    old_loc := OLD.batch_location_id;
    delta   := OLD.quantity;

    -- Return quantity; no guard required
    UPDATE batch_locations
      SET quantity = quantity + delta
    WHERE id = old_loc;

    RETURN OLD;

  ELSIF TG_OP = 'UPDATE' THEN
    old_loc := OLD.batch_location_id;
    new_loc := NEW.batch_location_id;

    IF new_loc = old_loc THEN
      -- Same location: apply quantity delta
      delta := COALESCE(NEW.quantity,0) - COALESCE(OLD.quantity,0);
      IF delta > 0 THEN
        -- take more
        UPDATE batch_locations
          SET quantity = quantity - delta
        WHERE id = new_loc AND quantity >= delta
        RETURNING 1 INTO res;

        IF res IS NULL THEN
          RAISE EXCEPTION 'Insufficient stock at location % for additional reserve of %', new_loc, delta;
        END IF;

      ELSIF delta < 0 THEN
        -- return surplus
        UPDATE batch_locations
          SET quantity = quantity + (-delta)
        WHERE id = new_loc;
      END IF;

    ELSE
      -- Location changed: return OLD, then take NEW (guarded)
      IF OLD.quantity > 0 THEN
        UPDATE batch_locations
          SET quantity = quantity + OLD.quantity
        WHERE id = old_loc;
      END IF;

      IF NEW.quantity > 0 THEN
        UPDATE batch_locations
          SET quantity = quantity - NEW.quantity
        WHERE id = new_loc AND quantity >= NEW.quantity
        RETURNING 1 INTO res;

        IF res IS NULL THEN
          RAISE EXCEPTION 'Insufficient stock at location % to move allocation of %', new_loc, NEW.quantity;
        END IF;
      END IF;

    END IF;

    RETURN NEW;

  END IF;

  RETURN NULL; -- not reached
END;
$$ LANGUAGE plpgsql;

-- Attach as BEFORE triggers so a failing reserve blocks the write
DROP TRIGGER IF EXISTS pbi_adjust_stock_ins ON prescription_batch_items;
DROP TRIGGER IF EXISTS pbi_adjust_stock_upd ON prescription_batch_items;
DROP TRIGGER IF EXISTS pbi_adjust_stock_del ON prescription_batch_items;

CREATE TRIGGER pbi_adjust_stock_ins
BEFORE INSERT ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION trg_pbi_adjust_stock();

CREATE TRIGGER pbi_adjust_stock_upd
BEFORE UPDATE ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION trg_pbi_adjust_stock();

CREATE TRIGGER pbi_adjust_stock_del
BEFORE DELETE ON prescription_batch_items
FOR EACH ROW EXECUTE FUNCTION trg_pbi_adjust_stock();

-- DEFERRABLE constraint trigger to allow multi-row transactions
CREATE OR REPLACE FUNCTION ck_pbi_drug_match()
RETURNS TRIGGER AS $$
DECLARE
  bl_batch_id BIGINT;
  bl_drug_id  BIGINT;
  line_drug_id BIGINT;
BEGIN
  SELECT b.id, b.drug_id INTO bl_batch_id, bl_drug_id
  FROM batch_locations bl
  JOIN drug_batches b ON b.id = bl.batch_id
  WHERE bl.id = NEW.batch_location_id;

  SELECT drug_id INTO line_drug_id FROM prescription_lines WHERE id = NEW.line_id;

  IF bl_drug_id IS DISTINCT FROM line_drug_id THEN
    RAISE EXCEPTION 'Batch location % belongs to drug %, but line % is for drug %',
      NEW.batch_location_id, bl_drug_id, NEW.line_id, line_drug_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS pbi_ck_drug ON prescription_batch_items;
CREATE CONSTRAINT TRIGGER pbi_ck_drug
AFTER INSERT OR UPDATE ON prescription_batch_items
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION ck_pbi_drug_match();
