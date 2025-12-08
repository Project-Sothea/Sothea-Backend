CREATE TABLE patient_details
(
  id             SERIAL PRIMARY KEY,
  name           TEXT       NOT NULL,
  family_group   TEXT       NOT NULL,
  khmer_name     TEXT       NOT NULL,
  dob            DATE       NOT NULL,
  gender         VARCHAR(1) NOT NULL,
  village        TEXT       NOT NULL,
  contact_no     TEXT       NOT NULL,
  drug_allergies TEXT

);

CREATE TABLE admin
(
  id                    INTEGER NOT NULL REFERENCES patient_details (id) ON DELETE CASCADE,
  vid                   INTEGER NOT NULL,
  reg_date              DATE    NOT NULL,
  queue_no              TEXT    NOT NULL,
  pregnant              BOOLEAN NOT NULL,
  last_menstrual_period DATE,
  sent_to_id            BOOLEAN NOT NULL,

  PRIMARY KEY (id, vid) -- Composite primary key
);

CREATE TABLE past_medical_history
(
    id                           INTEGER NOT NULL,                       -- Use INTEGER to match the id type from admin
    vid                          INTEGER NOT NULL,                       -- Add vid to match the vid type from admin
    cough                        BOOLEAN,                               -- Allow NULL for 'Nil' option
    fever                        BOOLEAN,                               -- Allow NULL for 'Nil' option
    blocked_nose                 BOOLEAN,                               -- Allow NULL for 'Nil' option
    sore_throat                  BOOLEAN,                               -- Allow NULL for 'Nil' option
    night_sweats                 BOOLEAN,                               -- Allow NULL for 'Nil' option
    unintentional_weight_loss    BOOLEAN,                               -- Allow NULL for 'Nil' option
    tuberculosis                 BOOLEAN,                               -- Allow NULL for 'Nil' option
    tuberculosis_has_been_treated BOOLEAN,                              -- Allow NULL for 'Nil' option
    diabetes                     BOOLEAN,                               -- Allow NULL for 'Nil' option
    hypertension                 BOOLEAN,                               -- Allow NULL for 'Nil' option
    hyperlipidemia               BOOLEAN,                               -- Allow NULL for 'Nil' option
    chronic_joint_pains          BOOLEAN,                               -- Allow NULL for 'Nil' option
    chronic_muscle_aches         BOOLEAN,                               -- Allow NULL for 'Nil' option
    sexually_transmitted_disease BOOLEAN,                               -- Allow NULL for 'Nil' option
    specified_stds               TEXT,
    others                       TEXT,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE social_history
(
    id                      INTEGER NOT NULL,                            -- Use INTEGER to match the id type from admin
    vid                     INTEGER NOT NULL,                            -- Add vid to match the vid type from admin
    past_smoking_history    BOOLEAN NOT NULL,
    no_of_years             INTEGER,
    current_smoking_history BOOLEAN NOT NULL,
    cigarettes_per_day      INTEGER,
    alcohol_history         BOOLEAN NOT NULL,
    how_regular             VARCHAR(1),

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE vital_statistics
(
    id                        INTEGER       NOT NULL,                    -- Use INTEGER to match the id type from admin
    vid                       INTEGER       NOT NULL,                    -- Add vid to match the vid type from admin
    temperature               NUMERIC(5, 1) NOT NULL,
    spo2                      NUMERIC(5, 1) NOT NULL,
    systolic_bp1              NUMERIC(5, 1),
    diastolic_bp1             NUMERIC(5, 1),
    systolic_bp2              NUMERIC(5, 1),
    diastolic_bp2             NUMERIC(5, 1),
    avg_systolic_bp           NUMERIC(5, 1),
    avg_diastolic_bp          NUMERIC(5, 1),
    hr1                       NUMERIC(5, 1) NOT NULL,
    hr2                       NUMERIC(5, 1) NOT NULL,
    avg_hr                    NUMERIC(5, 1) NOT NULL,
    rand_blood_glucose_mmol_l  NUMERIC(5, 1),
    icope_high_bp             BOOLEAN,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE height_and_weight
(
    id           INTEGER       NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid          INTEGER       NOT NULL,                                 -- Add vid to match the vid type from admin
    height       NUMERIC(5, 1) NOT NULL,
    weight       NUMERIC(5, 1) NOT NULL,
    bmi          NUMERIC(5, 1) NOT NULL,
    bmi_analysis TEXT          NOT NULL,
    paeds_height NUMERIC(5, 1),
    paeds_weight NUMERIC(5, 1),

    icope_lost_weight_past_months BOOLEAN,
    icope_no_desire_to_eat BOOLEAN,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE visual_acuity
(
    id                      INTEGER NOT NULL,                            -- Use INTEGER to match the id type from admin
    vid                     INTEGER NOT NULL,                            -- Add vid to match the vid type from admin
    l_eye_vision            INTEGER NOT NULL,
    r_eye_vision            INTEGER NOT NULL,
    additional_intervention TEXT,

    sent_to_opto BOOLEAN NOT NULL,
    referred_for_glasses BOOLEAN NOT NULL,
    icope_eye_problem BOOLEAN,
    icope_treated_for_diabetes_or_bp BOOLEAN,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE dental
(
    id                   INTEGER NOT NULL, -- Use INTEGER to match the id type from admin
    vid                  INTEGER NOT NULL,
    fluoride_exposure    TEXT    NOT NULL,
    diet                 TEXT    NOT NULL,
    bacterial_exposure   TEXT    NOT NULL,
    oral_symptoms        BOOLEAN NOT NULL,
    drink_other_water    BOOLEAN NOT NULL,
    risk_for_dental_carries TEXT NOT NULL,
    icope_difficulty_chewing BOOLEAN,
    icope_pain_in_mouth  BOOLEAN,
    dental_notes         TEXT,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE fall_risk
(
    id           INTEGER       NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid          INTEGER       NOT NULL,                                 -- Add vid to match the vid type from admin
    side_to_side_balance INTEGER NOT NULL,
    semi_tandem_balance INTEGER NOT NULL,
    tandem_balance INTEGER NOT NULL,
    gait_speed_test INTEGER NOT NULL,
    chair_stand_test INTEGER NOT NULL,
    fall_risk_score INTEGER NOT NULL,
    icope_complete_chair_stands BOOLEAN,
    icope_chair_stands_time BOOLEAN,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE doctors_consultation
(
    id                 INTEGER NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid                INTEGER NOT NULL,                                 -- Add vid to match the vid type from admin
    well               BOOLEAN,
    msk                BOOLEAN,
    cvs                BOOLEAN,
    respi              BOOLEAN,
    gu                 BOOLEAN,
    git                BOOLEAN,
    eye                BOOLEAN,
    derm               BOOLEAN,
    others             TEXT,
    consultation_notes TEXT,
    diagnosis          TEXT,
    treatment          TEXT,
    referral_needed    BOOLEAN NOT NULL,
    referral_loc       TEXT,
    remarks            TEXT,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);

CREATE TABLE physiotherapy
(
    id                         INTEGER NOT NULL,                         -- Use INTEGER to match the id type from patients
    vid                        INTEGER NOT NULL,                         -- Add vid to match the vid type from visits
    subjective_assessment      TEXT,                                     -- Subjective Assessment (Open Ended)
    pain_scale                 INTEGER,                                  -- Pain Scale (1-10)
    objective_assessment       TEXT,                                     -- Objective Assessment (Open Ended)
    intervention               TEXT,                                     -- Intervention (Open Ended)
    evaluation                 TEXT,                                     -- Evaluation (Open Ended)

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) ON DELETE CASCADE -- Foreign key referencing the composite key in admin
);
