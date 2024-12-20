/*******************
    Drop the tables
********************/
DROP TABLE IF EXISTS pastmedicalhistory;
DROP TABLE IF EXISTS socialhistory;
DROP TABLE IF EXISTS vitalstatistics;
DROP TABLE IF EXISTS heightandweight;
DROP TABLE IF EXISTS visualacuity;
DROP TABLE IF EXISTS fallrisk;
DROP TABLE IF EXISTS doctorsconsultation;
DROP TABLE IF EXISTS admin;
DROP TABLE IF EXISTS users;

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
    dob                   DATE,
    age                   INTEGER,
    gender                VARCHAR(1) NOT NULL,
    village               TEXT       NOT NULL,
    contact_no            TEXT       NOT NULL,
    pregnant              BOOLEAN    NOT NULL,
    last_menstrual_period Date,
    drug_allergies        TEXT,
    sent_to_id            BOOLEAN    NOT NULL,
    photo                 BYTEA,
    PRIMARY KEY (id, vid)         -- Composite primary key
);

CREATE TABLE IF NOT EXISTS pastmedicalhistory
(
    id                           INTEGER NOT NULL,                       -- Use INTEGER to match the id type from admin
    vid                          INTEGER NOT NULL,                       -- Add vid to match the vid type from admin
    tuberculosis                 BOOLEAN NOT NULL,
    diabetes                     BOOLEAN NOT NULL,
    hypertension                 BOOLEAN NOT NULL,
    hyperlipidemia               BOOLEAN NOT NULL,
    chronic_joint_pains          BOOLEAN NOT NULL,
    chronic_muscle_aches         BOOLEAN NOT NULL,
    sexually_transmitted_disease BOOLEAN NOT NULL,
    specified_stds               TEXT,
    others                       TEXT,
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS socialhistory
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

CREATE TABLE IF NOT EXISTS vitalstatistics
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
    rand_blood_glucose_mmoll  NUMERIC(5, 1) NOT NULL,
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS heightandweight
(
    id           INTEGER       NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid          INTEGER       NOT NULL,                                 -- Add vid to match the vid type from admin
    height       NUMERIC(5, 1) NOT NULL,
    weight       NUMERIC(5, 1) NOT NULL,
    bmi          NUMERIC(5, 1) NOT NULL,
    bmi_analysis TEXT          NOT NULL,
    paeds_height NUMERIC(5, 1),
    paeds_weight NUMERIC(5, 1),
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS visualacuity
(
    id                      INTEGER NOT NULL,                            -- Use INTEGER to match the id type from admin
    vid                     INTEGER NOT NULL,                            -- Add vid to match the vid type from admin
    l_eye_vision            INTEGER NOT NULL,
    r_eye_vision            INTEGER NOT NULL,
    additional_intervention TEXT,
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS dental
(
    id                 INTEGER NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid                INTEGER NOT NULL,                                 -- Add vid to match the vid type from admin
    clean_teeth_freq   INTEGER NOT NULL CHECK (clean_teeth_freq BETWEEN 0 AND 7),
    sugar_consume_freq INTEGER NOT NULL CHECK (sugar_consume_freq BETWEEN 0 AND 6),
    past_year_decay    BOOLEAN NOT NULL,
    brush_teeth_pain   BOOLEAN NOT NULL,
    drink_other_water  BOOLEAN NOT NULL,
    dental_notes       TEXT,
    referral_needed    BOOLEAN NOT NULL,
    referral_loc       TEXT,

    -- Teeth Chart (FDI numbering), True if cavity exists
    tooth_11           BOOLEAN,                                          -- Right Upper
    tooth_12           BOOLEAN,
    tooth_13           BOOLEAN,
    tooth_14           BOOLEAN,
    tooth_15           BOOLEAN,
    tooth_16           BOOLEAN,
    tooth_17           BOOLEAN,
    tooth_18           BOOLEAN,

    tooth_21           BOOLEAN,                                          -- Left Upper
    tooth_22           BOOLEAN,
    tooth_23           BOOLEAN,
    tooth_24           BOOLEAN,
    tooth_25           BOOLEAN,
    tooth_26           BOOLEAN,
    tooth_27           BOOLEAN,
    tooth_28           BOOLEAN,

    tooth_31           BOOLEAN,                                          -- Left Lower
    tooth_32           BOOLEAN,
    tooth_33           BOOLEAN,
    tooth_34           BOOLEAN,
    tooth_35           BOOLEAN,
    tooth_36           BOOLEAN,
    tooth_37           BOOLEAN,
    tooth_38           BOOLEAN,

    tooth_41           BOOLEAN,                                          -- Right Lower
    tooth_42           BOOLEAN,
    tooth_43           BOOLEAN,
    tooth_44           BOOLEAN,
    tooth_45           BOOLEAN,
    tooth_46           BOOLEAN,
    tooth_47           BOOLEAN,
    tooth_48           BOOLEAN,

    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS fallrisk
(
    id                  INTEGER    NOT NULL,                             -- Use INTEGER to match the id type from admin
    vid                 INTEGER    NOT NULL,                             -- Add vid to match the vid type from admin
    fall_worries        VARCHAR(1) NOT NULL,                             -- How often do you worry about falling? (a, b, c, d)
    fall_history        VARCHAR(1) NOT NULL,                             -- History of fall within past 12 months (a, b, c, d)
    cognitive_status    VARCHAR(1) NOT NULL,                             -- Cognitive status (a, b, c, d)
    continence_problems VARCHAR(1) NOT NULL,                             -- Continence problems (a, b, c, d, e)
    safety_awareness    VARCHAR(1) NOT NULL,                             -- Safety awareness (a, b, c, d)
    unsteadiness        VARCHAR(1) NOT NULL,                             -- Unsteadiness when standing, transferring and/or walking (a, b, c, d)
    fall_risk_score     INTEGER    NOT NULL,                             -- Fall risk score
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS doctorsconsultation
(
    id                 INTEGER NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid                INTEGER NOT NULL,                                 -- Add vid to match the vid type from admin
    well               BOOLEAN NOT NULL,
    msk                BOOLEAN NOT NULL,
    cvs                BOOLEAN NOT NULL,
    respi              BOOLEAN NOT NULL,
    gu                 BOOLEAN NOT NULL,
    git                BOOLEAN NOT NULL,
    eye                BOOLEAN NOT NULL,
    derm               BOOLEAN NOT NULL,
    others             TEXT    NOT NULL,
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
    id                         INTEGER NOT NULL,                         -- Use INTEGER to match the id type from admin
    vid                        INTEGER NOT NULL,                         -- Add vid to match the vid type from admin
    pain_stiffness_day         INTEGER NOT NULL,                         -- Pain/stiffness during the day: How severe was your usual joint or muscle pain and/or stiffness overall during the day in the last 2 weeks? 0, 1, 2, 3, 4, 5
    pain_stiffness_night       INTEGER NOT NULL,                         -- Pain/stiffness during the night: How severe was your usual joint or muscle pain and/or stiffness overall during the night in the last 2 weeks? 0, 1, 2, 3, 4, 5
    symptoms_interfere_tasks   TEXT    NOT NULL,                         -- How much has your symptoms interfered with your ability to walk or do everyday tasks like cooking, cleaning or dressing in the last 2 weeks? Never, Rarely, Sometimes, Often, Always
    symptoms_change            TEXT    NOT NULL,                         -- Have your symptoms improved, worsened, or stayed the same over the last 2 weeks? Never, Rarely, Sometimes, Often, Always
    symptoms_need_help         TEXT    NOT NULL,                         -- How often have you needed help from others (including family, friends or carers) because of your joint or muscle symptoms in the last 2 weeks? Never, Rarely, Sometimes, Often, Always
    trouble_sleep_symptoms     TEXT    NOT NULL,                         -- How often have you had trouble with either falling asleep or staying asleep because of your joint or muscle symptoms in the last 2 weeks? Never, Rarely, Sometimes, Often, Always
    how_much_fatigue           INTEGER NOT NULL,                         -- How much fatigue or low energy have you felt in the last 2 weeks? 0, 1, 2, 3, 4, 5
    anxious_low_mood           INTEGER NOT NULL,                         -- How much have you felt anxious or low in your mood because of your joint or muscle symptoms in the last 2 weeks? 0, 1, 2, 3, 4, 5
    medication_manage_symptoms TEXT    NOT NULL,                         -- Have you used any medication to manage your symptoms in the last 2 weeks? If yes, how often? Never, Rarely, Sometimes, Often, Always
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

INSERT INTO pastmedicalhistory(id, vid, tuberculosis, diabetes, hypertension, hyperlipidemia, chronic_joint_pains,
                               chronic_muscle_aches, sexually_transmitted_disease, specified_stds, others)
VALUES (1, 1, TRUE, FALSE, TRUE, FALSE, FALSE, TRUE, TRUE, 'TRICHOMONAS', 'None'),
       (2, 1, FALSE, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE, '', 'CHILDHOOD LEUKAEMIA'),
       (3, 1, TRUE, FALSE, FALSE, FALSE, TRUE, TRUE, FALSE, '', ''),
       (4, 1, FALSE, FALSE, TRUE, FALSE, TRUE, FALSE, TRUE, 'Syphilis', NULL),
       (5, 1, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, '', '');

INSERT INTO socialhistory (id, vid, past_smoking_history, no_of_years, current_smoking_history,
                           cigarettes_per_day, alcohol_history, how_regular)
VALUES (1, 1, TRUE, 15, FALSE, NULL, TRUE, 'A'),
       (2, 1, FALSE, NULL, TRUE, 10, TRUE, 'D'),
       (3, 1, TRUE, 20, TRUE, 5, FALSE, NULL),
       (4, 1, TRUE, 10, FALSE, NULL, TRUE, 'B'),
       (5, 1, FALSE, NULL, FALSE, NULL, FALSE, NULL);

INSERT INTO vitalstatistics (id, vid, temperature, spo2, systolic_bp1, diastolic_bp1, systolic_bp2, diastolic_bp2,
                             avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmoll)
VALUES (1, 1, 36.5, 98, 120, 80, 122, 78, 121, 79, 72, 71, 71.5, 5.4),
       (2, 1, 37.0, 97, 130, 85, 128, 82, 129, 83, 68, 70, 69, 5.7),
       (3, 1, 36.8, 99, 118, 78, 120, 76, 119, 77, 75, 76, 75.5, 5.6),
       (4, 1, 36.7, 98, 125, 82, 124, 80, 124.5, 81, 70, 72, 71, 5.3);

INSERT INTO heightandweight (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight)
VALUES (1, 1, 170, 70, 24.2, 'normal weight', 90, 80),
       (2, 1, 165, 55, 20.2, 'normal weight', 95, 90),
       (3, 1, 180, 85, 26.2, 'overweight', 80, 95);

INSERT INTO visualacuity (id, vid, l_eye_vision, r_eye_vision, additional_intervention)
VALUES (1, 1, 20, 20, 'VISUAL FIELD TEST REQUIRED'),
       (2, 1, 15, 20, 'REFERRED TO BOC');

INSERT INTO fallrisk (id, vid, fall_worries, fall_history, cognitive_status, continence_problems, safety_awareness,
                      unsteadiness, fall_risk_score)
VALUES (1, 1, 'a', 'a', 'b', 'e', 'd', 'c', 6),
       (2, 1, 'b', 'd', 'd', 'c', 'b', 'a', 10);

INSERT INTO dental (id, vid, clean_teeth_freq, sugar_consume_freq, past_year_decay, brush_teeth_pain, drink_other_water,
                    dental_notes, referral_needed, referral_loc, tooth_11, tooth_21, tooth_22, tooth_35, tooth_47,
                    tooth_48)
VALUES (1, 1, 2, 3, TRUE, TRUE, FALSE, 'None', TRUE, 'Dentist',
        TRUE, FALSE, TRUE, FALSE, TRUE, FALSE);

INSERT INTO dental (id, vid, clean_teeth_freq, sugar_consume_freq, past_year_decay, brush_teeth_pain, drink_other_water,
                    dental_notes, referral_needed, referral_loc, tooth_15, tooth_28, tooth_33, tooth_41, tooth_48)
VALUES (2, 1, 3, 4, FALSE, FALSE, TRUE, 'None', FALSE, NULL,
        FALSE, FALSE, FALSE, FALSE, FALSE);

INSERT INTO doctorsconsultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (1, 1, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION');

INSERT INTO physiotherapy (id, vid, pain_stiffness_day, pain_stiffness_night, symptoms_interfere_tasks, symptoms_change,
                           symptoms_need_help, trouble_sleep_symptoms, how_much_fatigue, anxious_low_mood,
                           medication_manage_symptoms)
VALUES (1, 1, 3, 2, 'Sometimes', 'Often', 'Sometimes', 'Sometimes', 3, 2, 'Sometimes'),
         (2, 1, 4, 3, 'Often', 'Always', 'Often', 'Often', 4, 3, 'Often');

/*******************
    Add additional entries for patient 1 and 2
 */
INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
                   pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo)
VALUES (1, 'Family 1', '2025-07-01', 'Q123', 'John Doe', 'ខេមរ', '1990-01-01', 34, 'M', 'Village 1', '123456789', false,
        '2023-06-01', 'None', false, NULL);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
                   pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo)
VALUES (1, 'Family 2', '2024-12-02', 'Q124', 'Jane Doe', 'ចន ឌូ', '1990-01-011', 34, 'F', 'Village 2', '987654321',
        true, '2023-06-15', 'Penicillin', true, NULL);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
                   pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo)
VALUES (1, 'Family 1', '2023-07-03', 'Q125', 'Alice Doe', 'អាលីស ស្ម៊ីត', '1990-01-010', 35, 'F', 'Village 1',
        '555666777', false, '2023-05-01', 'None', false, NULL);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
                   pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo)
VALUES (2, 'B009', '2024-12-03', 'Q125', 'Walter White', 'អាលីស ស្ម៊ីត', '1990-01-010', 52, 'M', 'ABQ',
        '555666777', false, '2023-05-01', 'None', false, NULL);

INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, contact_no,
                   pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo)
VALUES (2, 'B009', '2023-10-03', 'Q125', 'Walter White', 'អាលីស ស្ម៊ីត', '1990-01-010', 52, 'M', 'ABQ',
        '555666777', false, '2023-05-01', 'None', false, NULL);

/*******************
    Add remaining categories for second visit for patient 1 and 2
 */

INSERT INTO pastmedicalhistory(id, vid, tuberculosis, diabetes, hypertension, hyperlipidemia, chronic_joint_pains,
                               chronic_muscle_aches, sexually_transmitted_disease, specified_stds, others)
VALUES (1, 2, TRUE, FALSE, TRUE, FALSE, FALSE, TRUE, TRUE, 'TRICHOMONAS', 'None'),
       (2, 2, FALSE, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE, '', 'CHILDHOOD LEUKAEMIA');

INSERT INTO socialhistory (id, vid, past_smoking_history, no_of_years, current_smoking_history,
                           cigarettes_per_day, alcohol_history, how_regular)
VALUES (1, 2, TRUE, 15, FALSE, NULL, TRUE, 'A'),
       (2, 2, FALSE, NULL, TRUE, 10, TRUE, 'D');

INSERT INTO vitalstatistics (id, vid, temperature, spo2, systolic_bp1, diastolic_bp1, systolic_bp2, diastolic_bp2,
                             avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmoll)
VALUES (1, 2, 36.5, 98, 120, 80, 122, 78, 121, 79, 72, 71, 71.5, 5.4),
       (2, 2, 37.0, 97, 130, 85, 128, 82, 129, 83, 68, 70, 69, 5.7);

INSERT INTO heightandweight (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight)
VALUES (1, 2, 170, 70, 24.2, 'normal weight', 90, 80),
       (2, 2, 165, 55, 20.2, 'normal weight', 95, 90);

INSERT INTO visualacuity (id, vid, l_eye_vision, r_eye_vision, additional_intervention)
VALUES (1, 2, 20, 20, 'VISUAL FIELD TEST REQUIRED'),
       (2, 2, 15, 20, 'REFERRED TO BOC');

INSERT INTO doctorsconsultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (1, 2, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'others',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', TRUE, NULL, 'MONITOR FOR RESOLUTION');
INSERT INTO doctorsconsultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (2, 2, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION');

/*******************
    Add usernames and passwords
 */
CREATE TABLE users
(
    id            SERIAL PRIMARY KEY,    -- Auto-incrementing ID for each user (optional but recommended)
    username      VARCHAR(255) NOT NULL, -- Username field, with a max length of 255 characters
    password_hash TEXT         NOT NULL  -- Field to store the hashed password
);

-- Insert the users
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$Y/FMxVZsuiPcVutd0O3kfuWEyWyqZUb4HC5yXF7.xVxNCt9GUfOPe');
INSERT INTO users (username, password_hash)
VALUES ('user', '$2a$10$cqFdRzsZVwGF6fI4N.YEYOOEZL7B/93RmEZRkVPVHHqyBMfNwy48i');
