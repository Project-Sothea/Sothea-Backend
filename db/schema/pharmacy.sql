/*******************
 Create Foundational Vocab
********************/

-- === Units (use numeric with 1 decimal place for amounts) =========================
CREATE TABLE units (
  code TEXT PRIMARY KEY,                  -- 'mg','g','mcg','mL','L','IU','tab','cap','drop','g'
  is_mass   BOOLEAN NOT NULL DEFAULT FALSE,
  is_volume BOOLEAN NOT NULL DEFAULT FALSE,
  is_piece  BOOLEAN NOT NULL DEFAULT FALSE
);

-- === Dosage forms ============================================================
CREATE TABLE dosage_forms (
  code TEXT PRIMARY KEY,                 -- 'TAB','CAP','SYR','SUSP','CREAM','DROP','INJ'
  label TEXT NOT NULL
);

-- === Routes ==================================================================
CREATE TABLE routes (
  code TEXT PRIMARY KEY,                 -- 'PO','IV','IM','TOP','OTIC','OPH'
  label TEXT NOT NULL
);

/*******************
 Drugs
********************/

CREATE TABLE drugs (
  id                BIGSERIAL PRIMARY KEY,
  generic_name      TEXT NOT NULL,           -- e.g., Amoxicillin
  brand_name        TEXT,                    -- optional
  drug_code         INTEGER,

  
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

  CONSTRAINT uq_drug_code UNIQUE (drug_code) DEFERRABLE INITIALLY DEFERRED,

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


/*******************
 Drug Batches and Batch Location
********************/
CREATE TABLE drug_batches (
  id             BIGSERIAL PRIMARY KEY,
  drug_id        BIGINT NOT NULL REFERENCES drugs(id) ON DELETE CASCADE,
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
