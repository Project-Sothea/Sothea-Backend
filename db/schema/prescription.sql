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

  CONSTRAINT fk_admin FOREIGN KEY (patient_id, vid) REFERENCES admin(id, vid)
);

/*******************
 Frequency
********************/
CREATE TABLE schedule_kinds (
  code TEXT PRIMARY KEY
);

-- Frequency codes lookup table
CREATE TABLE frequency_codes (
  code TEXT PRIMARY KEY,              -- OM, ON, BD, TDS, q8h, ...
  label TEXT NOT NULL,                -- human-readable label
  schedule_kind TEXT NOT NULL REFERENCES schedule_kinds(code),
  every_n INTEGER NOT NULL CHECK (every_n > 0),
  frequency_per_schedule NUMERIC(6,2) NOT NULL CHECK (frequency_per_schedule > 0)
);

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
