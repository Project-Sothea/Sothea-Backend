/*******************
    Drop the tables
********************/
DROP TABLE IF EXISTS past_medical_history;
DROP TABLE IF EXISTS social_history;
DROP TABLE IF EXISTS vital_statistics;
DROP TABLE IF EXISTS height_and_weight;
DROP TABLE IF EXISTS visual_acuity;
DROP TABLE IF EXISTS fall_risk;
DROP TABLE IF EXISTS doctors_consultation;
DROP TABLE IF EXISTS admin;

/*******************
Create the schema and Load Extensions
********************/

CREATE TABLE IF NOT EXISTS admin
(
  id                    SERIAL, -- Use SERIAL to auto-increment the ID
  vid                   INTEGER    NOT NULL,
  family_group          TEXT       NOT NULL,
  reg_date              DATE       NOT NULL,
  queue_no              TEXT       NOT NULL,
  name                  TEXT       NOT NULL,
  khmer_name            TEXT       NOT NULL,
  dob                   DATE       NOT NULL,
  gender                VARCHAR(1) NOT NULL,
  village               TEXT       NOT NULL,
  contact_no            TEXT       NOT NULL,
  pregnant              BOOLEAN    NOT NULL,
  last_menstrual_period DATE,
  drug_allergies        TEXT,
  sent_to_id            BOOLEAN    NOT NULL,
  PRIMARY KEY (id, vid)         -- Composite primary key
);

CREATE TABLE IF NOT EXISTS past_medical_history
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS social_history
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS vital_statistics
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS height_and_weight
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS visual_acuity
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS dental
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS fall_risk
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS doctors_consultation
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
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS physiotherapy
(
    id                         INTEGER NOT NULL,                         -- Use INTEGER to match the id type from patients
    vid                        INTEGER NOT NULL,                         -- Add vid to match the vid type from visits
    subjective_assessment      TEXT,                                     -- Subjective Assessment (Open Ended)
    pain_scale                 INTEGER,                                  -- Pain Scale (1-10)
    objective_assessment       TEXT,                                     -- Objective Assessment (Open Ended)
    intervention               TEXT,                                     -- Intervention (Open Ended)
    evaluation                 TEXT,                                     -- Evaluation (Open Ended)
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

/*******************
    Create the trigger function
*******************/

CREATE OR REPLACE FUNCTION set_entry_id() RETURNS TRIGGER AS
$$
DECLARE
    max_entry_id INTEGER;
BEGIN
    -- Check if the ID already exists in the table
    SELECT COALESCE(MAX(VID), 0)
    INTO max_entry_id
    FROM admin
    WHERE ID = NEW.ID;

    -- Increment Entry_ID based on the max_entry_id
    NEW.VID := max_entry_id + 1;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER before_insert_admin
    BEFORE INSERT
    ON admin
    FOR EACH ROW
EXECUTE FUNCTION set_entry_id();

/*******************
    Create new patients
 */

INSERT INTO admin (family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no, pregnant,
                   last_menstrual_period, drug_allergies, sent_to_id)
VALUES ('S001', '2024-01-10', '1A', 'John Doe', '១២៣៤ ៥៦៧៨៩០ឥឲ', '1994-01-10', 30, 'M', 'SO', '12345678', FALSE, NULL,
        'panadol', FALSE),
       ('S002A', '2024-01-10', '2A', 'Jane Smith', '១២៣៤ ៥៦៧៨៩០ឥឲ', '1999-01-10', 25, 'F', 'SO', '12345679', FALSE,
        NULL, NULL, FALSE),
       ('S002B', '2024-01-10', '2B', 'Bob Smith', '១២៣៤ ៥៦៧៨៩០ឥឲ', '1999-01-10', 25, 'M', 'R1', '99999999', FALSE, NULL,
        'aspirin', FALSE),
       ('S003', '2024-01-10', '3A', 'Bob Johnson', '១២៣៤ ៥៦៧៨៩០ឥឲ', '1989-01-10', 35, 'M', 'R1', '11111111', FALSE,
        NULL, NULL, FALSE),
       ('S004', '2024-01-10', '4B', 'Alice Brown', '១២៣៤ ៥៦៧៨៩០ឥឲ', '1996-01-10', 28, 'F', 'R1', '17283948', FALSE,
        NULL, NULL, FALSE),
       ('S005A', '2024-01-10', '5C', 'Charlie Davis', '១២៣៤ ៥៦៧៨៩០ឥឲ', '1982-01-10', 40, 'M', 'R1', '09876543', FALSE,
        NULL, NULL, FALSE);

INSERT INTO past_medical_history
  (id, vid,
   cough, fever, blocked_nose, sore_throat, night_sweats, unintentional_weight_loss,
   tuberculosis, tuberculosis_has_been_treated,
   diabetes, hypertension, hyperlipidemia,
   chronic_joint_pains, chronic_muscle_aches,
   sexually_transmitted_disease, specified_stds, others)
VALUES
  (1, 1, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,  TRUE,  FALSE, FALSE, TRUE,  FALSE, FALSE, TRUE,  TRUE, 'TRICHOMONAS', 'None'),
  (2, 1, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,  FALSE,  TRUE, TRUE,  TRUE,  FALSE, FALSE, FALSE, '', 'CHILDHOOD LEUKAEMIA'),
  (3, 1, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,  TRUE,  FALSE, FALSE, FALSE, FALSE, TRUE,  TRUE,  FALSE, '', ''),
  (4, 1, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,  FALSE, FALSE, TRUE,  FALSE, TRUE,  FALSE,  TRUE,  'Syphilis', NULL),
  (5, 1, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,  FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, '','');

INSERT INTO social_history (id, vid, past_smoking_history, no_of_years, current_smoking_history,
                           cigarettes_per_day, alcohol_history, how_regular)
VALUES (1, 1, TRUE, 15, FALSE, NULL, TRUE, 'A'),
       (2, 1, FALSE, NULL, TRUE, 10, TRUE, 'D'),
       (3, 1, TRUE, 20, TRUE, 5, FALSE, NULL),
       (4, 1, TRUE, 10, FALSE, NULL, TRUE, 'B'),
       (5, 1, FALSE, NULL, FALSE, NULL, FALSE, NULL);

INSERT INTO vital_statistics (id, vid, temperature, spo2, systolic_bp1, diastolic_bp1, systolic_bp2, diastolic_bp2,
                             avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmol_l, icope_high_bp)
VALUES (1, 1, 36.5, 98, 120, 80, 122, 78, 121, 79, 72, 71, 71.5, 5.4, FALSE),
       (2, 1, 37.0, 97, 130, 85, 128, 82, 129, 83, 68, 70, 69, 5.7, TRUE),
       (3, 1, 36.8, 99, 118, 78, 120, 76, 119, 77, 75, 76, 75.5, 5.6, TRUE),
       (4, 1, 36.7, 98, 125, 82, 124, 80, 124.5, 81, 70, 72, 71, 5.3, FALSE);

INSERT INTO height_and_weight
  (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight,
   icope_lost_weight_past_months, icope_no_desire_to_eat)
VALUES
  (1, 1, 170, 70, 24.2, 'normal weight', 90, 80, FALSE, FALSE),
  (2, 1, 165, 55, 20.2, 'normal weight', 95, 90, FALSE, FALSE),
  (3, 1, 180, 85, 26.2, 'overweight',     80, 95, FALSE, FALSE);

INSERT INTO visual_acuity
  (id, vid, l_eye_vision, r_eye_vision, additional_intervention,
   sent_to_opto, referred_for_glasses, icope_eye_problem, icope_treated_for_diabetes_or_bp)
VALUES
  (1, 1, 20, 20, 'VISUAL FIELD TEST REQUIRED', FALSE, FALSE, FALSE, FALSE),
  (2, 1, 15, 20, 'REFERRED TO BOC',            FALSE, FALSE, FALSE, FALSE);

INSERT INTO fall_risk
  (id, vid,
   side_to_side_balance, semi_tandem_balance, tandem_balance,
   gait_speed_test, chair_stand_test, fall_risk_score,
   icope_complete_chair_stands, icope_chair_stands_time)
VALUES
  (1, 1, 1, 1, 2, 3, 4, 11, TRUE,  TRUE),
  (2, 1, 0, 1, 1, 2, 0,  4, FALSE, FALSE);

-- First visit rows rewritten
INSERT INTO dental
  (id, vid,
   fluoride_exposure, diet, bacterial_exposure,
   oral_symptoms, drink_other_water,
   risk_for_dental_carries, icope_difficulty_chewing, icope_pain_in_mouth,
   dental_notes)
VALUES
  (1, 1,
   '6, 7', '2-3', 'None in last 2 years',
   TRUE,  FALSE,
   'Low Risk', FALSE, TRUE,
   'None'),

  (2, 1,
   '5, 4, 3', '≥4', 'Yes in last 7 - 23 months',
   FALSE, TRUE,
   'Middle Risk',  FALSE, FALSE,
   'None');


INSERT INTO doctors_consultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (1, 1, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION');

INSERT INTO physiotherapy (id, vid, subjective_assessment, pain_scale, objective_assessment, intervention, evaluation)
VALUES (1, 1, 'Patient reports chronic lower back pain, worse in the morning. Difficulty with daily activities.', 7, 'Limited range of motion in lumbar spine. Muscle tension noted.', 'Prescribed stretching exercises and heat therapy.', 'Patient shows improvement with exercises. Continue current plan.'),
       (2, 1, 'Complains of shoulder stiffness after work. Pain increases with overhead movements.', 5, 'Reduced shoulder mobility. Tenderness in rotator cuff area.', 'Manual therapy and strengthening exercises recommended.', 'Moderate improvement. Patient to continue home exercises.');

/*******************
    Add additional entries for patient 1 and 2
 */
INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
       pregnant, last_menstrual_period, drug_allergies, sent_to_id)
VALUES (1, 'Family 1', '2025-07-01', 'Q123', 'John Doe', 'ខេមរ', '1990-01-01', 34, 'M', 'Village 1', '123456789', false, '2023-06-01', 'None', false);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
       pregnant, last_menstrual_period, drug_allergies, sent_to_id)
VALUES (1, 'Family 2', '2024-12-02', 'Q124', 'Jane Doe', 'ចន ឌូ', '1990-01-11', 34, 'F', 'Village 2', '987654321',
  true, '2023-06-15', 'Penicillin', true);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
       pregnant, last_menstrual_period, drug_allergies, sent_to_id)
VALUES (1, 'Family 1', '2023-07-03', 'Q125', 'Alice Doe', 'អាលីស ស្ម៊ីត', '1990-01-01', 35, 'F', 'Village 1',
  '555666777', false, '2023-05-01', 'None', false);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
       pregnant, last_menstrual_period, drug_allergies, sent_to_id)
VALUES (2, 'B009', '2024-12-03', 'Q125', 'Walter White', 'អាលីស ស្ម៊ីត', '1990-01-01', 52, 'M', 'ABQ',
  '555666777', false, '2023-05-01', 'None', false);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
       pregnant, last_menstrual_period, drug_allergies, sent_to_id)
VALUES (2, 'B009', '2023-10-03', 'Q125', 'Walter White', 'អាលីស ស្ម៊ីត', '1990-01-01', 52, 'M', 'ABQ',
  '555666777', false, '2023-05-01', 'None', false);

/*******************
    Add remaining categories for second visit for patient 1 and 2
 */

INSERT INTO past_medical_history
  (id, vid,
   cough, fever, blocked_nose, sore_throat, night_sweats, unintentional_weight_loss,
   tuberculosis, tuberculosis_has_been_treated,
   diabetes, hypertension, hyperlipidemia,
   chronic_joint_pains, chronic_muscle_aches,
   sexually_transmitted_disease, specified_stds, others)
VALUES
  (1, 2, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE,  FALSE, FALSE, TRUE,  FALSE, FALSE, TRUE,  TRUE, 'TRICHOMONAS', 'None'),
  (2, 2, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE,  TRUE,  TRUE,  FALSE, FALSE, FALSE, '', 'CHILDHOOD LEUKAEMIA');

INSERT INTO social_history (id, vid, past_smoking_history, no_of_years, current_smoking_history,
                           cigarettes_per_day, alcohol_history, how_regular)
VALUES (1, 2, TRUE, 15, FALSE, NULL, TRUE, 'A'),
       (2, 2, FALSE, NULL, TRUE, 10, TRUE, 'D');

INSERT INTO vital_statistics (id, vid, temperature, spo2, systolic_bp1, diastolic_bp1, systolic_bp2, diastolic_bp2,
                             avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmol_l, icope_high_bp)
VALUES (1, 2, 36.5, 98, 120, 80, 122, 78, 121, 79, 72, 71, 71.5, 5.4, TRUE),
       (2, 2, 37.0, 97, 130, 85, 128, 82, 129, 83, 68, 70, 69, 5.7, FALSE);

INSERT INTO height_and_weight
  (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight,
   icope_lost_weight_past_months, icope_no_desire_to_eat)
VALUES
  (1, 2, 170, 70, 24.2, 'normal weight', 90, 80, FALSE, FALSE),
  (2, 2, 165, 55, 20.2, 'normal weight', 95, 90, FALSE, FALSE);

INSERT INTO visual_acuity
  (id, vid, l_eye_vision, r_eye_vision, additional_intervention,
   sent_to_opto, referred_for_glasses, icope_eye_problem, icope_treated_for_diabetes_or_bp)
VALUES
  (1, 2, 20, 20, 'VISUAL FIELD TEST REQUIRED', FALSE, FALSE, FALSE, FALSE),
  (2, 2, 15, 20, 'REFERRED TO BOC',            FALSE, FALSE, FALSE, FALSE);

INSERT INTO doctors_consultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (1, 2, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'others',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', TRUE, NULL, 'MONITOR FOR RESOLUTION');
INSERT INTO doctors_consultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (2, 2, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION');
