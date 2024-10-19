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
    systolic_bp1              NUMERIC(5, 1) NOT NULL,
    diastolic_bp1             NUMERIC(5, 1) NOT NULL,
    systolic_bp2              NUMERIC(5, 1) NOT NULL,
    diastolic_bp2             NUMERIC(5, 1) NOT NULL,
    avg_systolic_bp           NUMERIC(5, 1) NOT NULL,
    avg_diastolic_bp          NUMERIC(5, 1) NOT NULL,
    hr1                       NUMERIC(5, 1) NOT NULL,
    hr2                       NUMERIC(5, 1) NOT NULL,
    avg_hr                    NUMERIC(5, 1) NOT NULL,
    rand_blood_glucose_mmolL  NUMERIC(5, 1) NOT NULL,
    rand_blood_glucose_mmolLp NUMERIC(5, 1) NOT NULL,
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
    paeds_height NUMERIC(5, 1) NOT NULL,
    paeds_weight NUMERIC(5, 1) NOT NULL,
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

CREATE TABLE IF NOT EXISTS fallrisk
(
    id                          INTEGER NOT NULL,                            -- Use INTEGER to match the id type from admin
    vid                         INTEGER NOT NULL,                            -- Add vid to match the vid type from admin
    past_year_fall              BOOLEAN NOT NULL,
    unsteady_standing_walking   BOOLEAN NOT NULL,
    fall_worries                BOOLEAN NOT NULL,
    others                      TEXT,
    further_referral            BOOLEAN NOT NULL,
    PRIMARY KEY (id, vid),                                               -- Composite primary key
    CONSTRAINT fk_admin FOREIGN KEY (id, vid) REFERENCES admin (id, vid) -- Foreign key referencing the composite key in admin
);

CREATE TABLE IF NOT EXISTS doctorsconsultation
(
    id                 INTEGER NOT NULL,                                 -- Use INTEGER to match the id type from admin
    vid                INTEGER NOT NULL,                                 -- Add vid to match the vid type from admin
    healthy            BOOLEAN NOT NULL,
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
                             avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmolL,
                             rand_blood_glucose_mmolLp)
VALUES (1, 1, 36.5, 98, 120, 80, 122, 78, 121, 79, 72, 71, 71.5, 5.4, 5.3),
       (2, 1, 37.0, 97, 130, 85, 128, 82, 129, 83, 68, 70, 69, 5.7, 5.6),
       (3, 1, 36.8, 99, 118, 78, 120, 76, 119, 77, 75, 76, 75.5, 5.6, 5.5),
       (4, 1, 36.7, 98, 125, 82, 124, 80, 124.5, 81, 70, 72, 71, 5.3, 5.2);

INSERT INTO heightandweight (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight)
VALUES (1, 1, 170, 70, 24.2, 'normal weight', 90, 80),
       (2, 1, 165, 55, 20.2, 'normal weight', 95, 90),
       (3, 1, 180, 85, 26.2, 'overweight', 80, 95);

INSERT INTO visualacuity (id, vid, l_eye_vision, r_eye_vision, additional_intervention)
VALUES (1, 1, 20, 20, 'VISUAL FIELD TEST REQUIRED'),
       (2, 1, 15, 20, 'REFERRED TO BOC');

INSERT INTO fallrisk (id, vid, past_year_fall, unsteady_standing_walking, fall_worries, others, further_referral)
VALUES
    (1, 1, TRUE, TRUE, FALSE, 'Patient uses walking aid occasionally', TRUE),
    (2, 1, FALSE, FALSE, FALSE, NULL, FALSE);

INSERT INTO doctorsconsultation (id, vid, healthy, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (1, 1, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION');

/*******************
    Add additional admin entries for patient 1 and 2
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
                             avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmolL,
                             rand_blood_glucose_mmolLp)
VALUES (1, 2, 36.5, 98, 120, 80, 122, 78, 121, 79, 72, 71, 71.5, 5.4, 5.3),
       (2, 2, 37.0, 97, 130, 85, 128, 82, 129, 83, 68, 70, 69, 5.7, 5.6);

INSERT INTO heightandweight (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight)
VALUES (1, 2, 170, 70, 24.2, 'normal weight', 90, 80),
       (2, 2, 165, 55, 20.2, 'normal weight', 95, 90);

INSERT INTO visualacuity (id, vid, l_eye_vision, r_eye_vision, additional_intervention)
VALUES (1, 2, 20, 20, 'VISUAL FIELD TEST REQUIRED'),
       (2, 2, 15, 20, 'REFERRED TO BOC');

INSERT INTO fallrisk (id, vid, past_year_fall, unsteady_standing_walking, fall_worries, others, further_referral)
VALUES (1, 2, FALSE, TRUE, TRUE, 'Feels unsteady sometimes, but no falls in the past year', TRUE),
        (2, 2, TRUE, TRUE, FALSE, 'Uses a walking aid, had a fall last year', TRUE);

INSERT INTO doctorsconsultation (id, vid, healthy, msk, cvs, respi, gu, git, eye, derm, others,
                                 consultation_notes, diagnosis, treatment, referral_needed,
                                 referral_loc, remarks)
VALUES (1, 2, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION'),
        (2, 2, TRUE, FALSE, FALSE, TRUE, TRUE, FALSE, TRUE, FALSE, 'LEUKAEMIA',
        'CHEST PAIN, SHORTNESS OF BREATH, COUGH', 'ACUTE BRONCHITIS',
        'REST, HYDRATION, COUGH SYRUP', FALSE, NULL, 'MONITOR FOR RESOLUTION');

/*******************
    Add usernames and passwords
 */
CREATE TABLE users
(
    id SERIAL PRIMARY KEY,              -- Auto-incrementing ID for each user (optional but recommended)
    username VARCHAR(255) NOT NULL,     -- Username field, with a max length of 255 characters
    password_hash TEXT NOT NULL       -- Field to store the hashed password
);

-- Insert the users
INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$10$Y/FMxVZsuiPcVutd0O3kfuWEyWyqZUb4HC5yXF7.xVxNCt9GUfOPe');
INSERT INTO users (username, password_hash) VALUES ('dr1', '$2a$10$QJVQuKDvgYI7h1bdZ7FeVuo6W/osv1D0QKS80Z49wPYCwl7JM7XYm');
INSERT INTO users (username, password_hash) VALUES ('dr2', '$2a$10$YBdJa9dV2GjlbW6To7oWfunZzKm1gUWwa7rRL9oNqai5XPurvXfGu');
INSERT INTO users (username, password_hash) VALUES ('dr3', '$2a$10$7AUA6j4Gssu1NeedfH/.nu8qi3mkR/bo3kQ1PQIZgh0CL3J7YqQAW');
INSERT INTO users (username, password_hash) VALUES ('dr4', '$2a$10$yVvSf9HcrOCKU20iI3KFMOraUeMeZFz.YAaxy7bC22BiPEUK604C2');
INSERT INTO users (username, password_hash) VALUES ('dr5', '$2a$10$KY442JFYtKJQST31XYwJquoD61qhWgpL2ee743oQnLV0IRLY8.sUi');
INSERT INTO users (username, password_hash) VALUES ('dr6', '$2a$10$fW.NoRoM6Nxgx/lBLR/s5.Wp4dM2AtecwT.W.QqOj2fu04z2ED4BS');
INSERT INTO users (username, password_hash) VALUES ('dr7', '$2a$10$oPJxh3X2nGXqrbuy98tapeBrZ2tIvhvdxSrwo3eZ7BD88Q.gav8MS');
INSERT INTO users (username, password_hash) VALUES ('dr8', '$2a$10$eivqzqXzgEdu42kl3w9mveqW03UhEtCdJewRUDncFEfBPDxdfNeGi');
INSERT INTO users (username, password_hash) VALUES ('dr9', '$2a$10$iSD4RiAOHsNM90DeA2mkcejotEoQtIfFx5tDfIIwIobgCzncsE0ka');
INSERT INTO users (username, password_hash) VALUES ('dr10', '$2a$10$810uwCZxg3PMLmfF2hLSeeYIYTKM5ND3pqd/ui1ZPYO2SfgMTFSdK');
INSERT INTO users (username, password_hash) VALUES ('dr11', '$2a$10$WDe8w9bbz7RyBftb.ybamOvRUmpyON79mb6tXM/g7sN77QZY8VLMC');
INSERT INTO users (username, password_hash) VALUES ('dr12', '$2a$10$SJroXorL7Bj98/qRE6/FYORtdChWveQIyMjUhOO29xi/z8MYOnGba');
INSERT INTO users (username, password_hash) VALUES ('dr13', '$2a$10$UkbEtCJdeGuX/FHLTTlCE.OsrQ7xjVl0rY3Bo8NWY3SOkJK/KT0G.');
INSERT INTO users (username, password_hash) VALUES ('dr14', '$2a$10$I0f4b7QtrS4eH9SNETGUsOcwoqQqlQKRbUD2DITs6y5lB/cdlqOzC');
INSERT INTO users (username, password_hash) VALUES ('dr15', '$2a$10$UhtnHUmTU4DJTUXVaBR.5OFaS2fiimN6YWhOYTzaofZwMtKF/PKDC');
INSERT INTO users (username, password_hash) VALUES ('dr16', '$2a$10$gYTT5hStXUKrpK0Tv9sTluvsK3dwIa8vEJMzPeIuaZ6.MV1YCuJ9O');
INSERT INTO users (username, password_hash) VALUES ('dr17', '$2a$10$aIXgj2VmQuyih/KOMxvxeudRI6DRgiYUy1IMvZ3e.xFh2xZcyWpzq');
INSERT INTO users (username, password_hash) VALUES ('dr18', '$2a$10$xJQIZNVos4y5LkYekDRHV.KeSgKnEzV71uYN3tYRZ3vaJoDPNGSBK');
INSERT INTO users (username, password_hash) VALUES ('dr19', '$2a$10$pvo3eKjP4ZFNx.bE1l9mwuFT87Wnm.FqDHvlJSFbmokP7DkYwy9Oq');
INSERT INTO users (username, password_hash) VALUES ('dr20', '$2a$10$7ZKQ56ckZ2Z9Jye1em62iOrKeMLggM5oI2V1Jpjw52Cmg4nKTmNOG');
