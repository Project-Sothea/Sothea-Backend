-- Admin ----------------------------------------------------------------------

-- name: GetAdmin :one
SELECT id,
       vid,
       family_group,
       reg_date,
       queue_no,
       name,
       khmer_name,
       dob,
       gender,
       village,
       contact_no,
       pregnant,
       last_menstrual_period,
       drug_allergies,
       sent_to_id
FROM admin
WHERE id = $1
  AND vid = $2;

-- name: InsertPatient :one
INSERT INTO admin (
  family_group,
  reg_date,
  queue_no,
  name,
  khmer_name,
  dob,
  gender,
  village,
  contact_no,
  pregnant,
  last_menstrual_period,
  drug_allergies,
  sent_to_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING id, vid;

-- name: InsertPatientVisit :one
INSERT INTO admin (
  id,
  family_group,
  reg_date,
  queue_no,
  name,
  khmer_name,
  dob,
  gender,
  village,
  contact_no,
  pregnant,
  last_menstrual_period,
  drug_allergies,
  sent_to_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING id, vid;

-- name: UpdateAdmin :exec
UPDATE admin
SET family_group         = $1,
    reg_date             = $2,
    queue_no             = $3,
    name                 = $4,
    khmer_name           = $5,
    dob                  = $6,
    gender               = $7,
    village              = $8,
    contact_no           = $9,
    pregnant             = $10,
    last_menstrual_period = $11,
    drug_allergies       = $12,
    sent_to_id           = $13
WHERE id = $14
  AND vid = $15;

-- name: CheckPatientExists :one
SELECT id
FROM admin
WHERE id = $1
LIMIT 1;

-- name: CheckPatientVisitExists :one
SELECT id, vid
FROM admin
WHERE id = $1
  AND vid = $2
LIMIT 1;

-- name: DeleteAdmin :exec
DELETE FROM admin
WHERE id = $1
  AND vid = $2;

-- name: GetLatestAdmin :one
SELECT id,
       vid,
       family_group,
       reg_date,
       queue_no,
       name,
       khmer_name
FROM admin
WHERE id = $1
ORDER BY reg_date DESC, vid DESC
LIMIT 1;

-- name: ListAdminVisits :many
SELECT vid, reg_date
FROM admin
WHERE id = $1
ORDER BY vid;

-- Past Medical History -------------------------------------------------------

-- name: GetPastMedicalHistory :one
SELECT id,
       vid,
       cough,
       fever,
       blocked_nose,
       sore_throat,
       night_sweats,
       unintentional_weight_loss,
       tuberculosis,
       tuberculosis_has_been_treated,
       diabetes,
       hypertension,
       hyperlipidemia,
       chronic_joint_pains,
       chronic_muscle_aches,
       sexually_transmitted_disease,
       specified_stds,
       others
FROM past_medical_history
WHERE id = $1
  AND vid = $2;

-- name: UpsertPastMedicalHistory :exec
INSERT INTO past_medical_history (
  id,
  vid,
  cough,
  fever,
  blocked_nose,
  sore_throat,
  night_sweats,
  unintentional_weight_loss,
  tuberculosis,
  tuberculosis_has_been_treated,
  diabetes,
  hypertension,
  hyperlipidemia,
  chronic_joint_pains,
  chronic_muscle_aches,
  sexually_transmitted_disease,
  specified_stds,
  others
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
)
ON CONFLICT (id, vid) DO UPDATE SET
  cough                       = EXCLUDED.cough,
  fever                       = EXCLUDED.fever,
  blocked_nose                = EXCLUDED.blocked_nose,
  sore_throat                 = EXCLUDED.sore_throat,
  night_sweats                = EXCLUDED.night_sweats,
  unintentional_weight_loss   = EXCLUDED.unintentional_weight_loss,
  tuberculosis                = EXCLUDED.tuberculosis,
  tuberculosis_has_been_treated = EXCLUDED.tuberculosis_has_been_treated,
  diabetes                    = EXCLUDED.diabetes,
  hypertension                = EXCLUDED.hypertension,
  hyperlipidemia              = EXCLUDED.hyperlipidemia,
  chronic_joint_pains         = EXCLUDED.chronic_joint_pains,
  chronic_muscle_aches        = EXCLUDED.chronic_muscle_aches,
  sexually_transmitted_disease = EXCLUDED.sexually_transmitted_disease,
  specified_stds              = EXCLUDED.specified_stds,
  others                      = EXCLUDED.others;

-- name: DeletePastMedicalHistory :exec
DELETE FROM past_medical_history
WHERE id = $1
  AND vid = $2;

-- Social History -------------------------------------------------------------

-- name: GetSocialHistory :one
SELECT id,
       vid,
       past_smoking_history,
       no_of_years,
       current_smoking_history,
       cigarettes_per_day,
       alcohol_history,
       how_regular
FROM social_history
WHERE id = $1
  AND vid = $2;

-- name: UpsertSocialHistory :exec
INSERT INTO social_history (
  id,
  vid,
  past_smoking_history,
  no_of_years,
  current_smoking_history,
  cigarettes_per_day,
  alcohol_history,
  how_regular
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (id, vid) DO UPDATE SET
  past_smoking_history = EXCLUDED.past_smoking_history,
  no_of_years          = EXCLUDED.no_of_years,
  current_smoking_history = EXCLUDED.current_smoking_history,
  cigarettes_per_day   = EXCLUDED.cigarettes_per_day,
  alcohol_history      = EXCLUDED.alcohol_history,
  how_regular          = EXCLUDED.how_regular;

-- name: DeleteSocialHistory :exec
DELETE FROM social_history
WHERE id = $1
  AND vid = $2;

-- Vital Statistics -----------------------------------------------------------

-- name: GetVitalStatistics :one
SELECT id,
       vid,
       temperature,
       spo2,
       systolic_bp1,
       diastolic_bp1,
       systolic_bp2,
       diastolic_bp2,
       avg_systolic_bp,
       avg_diastolic_bp,
       hr1,
       hr2,
       avg_hr,
       rand_blood_glucose_mmol_l,
       icope_high_bp
FROM vital_statistics
WHERE id = $1
  AND vid = $2;

-- name: UpsertVitalStatistics :exec
INSERT INTO vital_statistics (
  id,
  vid,
  temperature,
  spo2,
  systolic_bp1,
  diastolic_bp1,
  systolic_bp2,
  diastolic_bp2,
  avg_systolic_bp,
  avg_diastolic_bp,
  hr1,
  hr2,
  avg_hr,
  rand_blood_glucose_mmol_l,
  icope_high_bp
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
ON CONFLICT (id, vid) DO UPDATE SET
  temperature              = EXCLUDED.temperature,
  spo2                     = EXCLUDED.spo2,
  systolic_bp1             = EXCLUDED.systolic_bp1,
  diastolic_bp1            = EXCLUDED.diastolic_bp1,
  systolic_bp2             = EXCLUDED.systolic_bp2,
  diastolic_bp2            = EXCLUDED.diastolic_bp2,
  avg_systolic_bp          = EXCLUDED.avg_systolic_bp,
  avg_diastolic_bp         = EXCLUDED.avg_diastolic_bp,
  hr1                      = EXCLUDED.hr1,
  hr2                      = EXCLUDED.hr2,
  avg_hr                   = EXCLUDED.avg_hr,
  rand_blood_glucose_mmol_l = EXCLUDED.rand_blood_glucose_mmol_l,
  icope_high_bp            = EXCLUDED.icope_high_bp;

-- name: DeleteVitalStatistics :exec
DELETE FROM vital_statistics
WHERE id = $1
  AND vid = $2;

-- Height and Weight ----------------------------------------------------------

-- name: GetHeightAndWeight :one
SELECT id,
       vid,
       height,
       weight,
       bmi,
       bmi_analysis,
       paeds_height,
       paeds_weight,
       icope_lost_weight_past_months,
       icope_no_desire_to_eat
FROM height_and_weight
WHERE id = $1
  AND vid = $2;

-- name: UpsertHeightAndWeight :exec
INSERT INTO height_and_weight (
  id,
  vid,
  height,
  weight,
  bmi,
  bmi_analysis,
  paeds_height,
  paeds_weight,
  icope_lost_weight_past_months,
  icope_no_desire_to_eat
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id, vid) DO UPDATE SET
  height                     = EXCLUDED.height,
  weight                     = EXCLUDED.weight,
  bmi                        = EXCLUDED.bmi,
  bmi_analysis               = EXCLUDED.bmi_analysis,
  paeds_height               = EXCLUDED.paeds_height,
  paeds_weight               = EXCLUDED.paeds_weight,
  icope_lost_weight_past_months = EXCLUDED.icope_lost_weight_past_months,
  icope_no_desire_to_eat     = EXCLUDED.icope_no_desire_to_eat;

-- name: DeleteHeightAndWeight :exec
DELETE FROM height_and_weight
WHERE id = $1
  AND vid = $2;

-- Visual Acuity --------------------------------------------------------------

-- name: GetVisualAcuity :one
SELECT id,
       vid,
       l_eye_vision,
       r_eye_vision,
       additional_intervention,
       sent_to_opto,
       referred_for_glasses,
       icope_eye_problem,
       icope_treated_for_diabetes_or_bp
FROM visual_acuity
WHERE id = $1
  AND vid = $2;

-- name: UpsertVisualAcuity :exec
INSERT INTO visual_acuity (
  id,
  vid,
  l_eye_vision,
  r_eye_vision,
  additional_intervention,
  sent_to_opto,
  referred_for_glasses,
  icope_eye_problem,
  icope_treated_for_diabetes_or_bp
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id, vid) DO UPDATE SET
  l_eye_vision                = EXCLUDED.l_eye_vision,
  r_eye_vision                = EXCLUDED.r_eye_vision,
  additional_intervention     = EXCLUDED.additional_intervention,
  sent_to_opto                = EXCLUDED.sent_to_opto,
  referred_for_glasses        = EXCLUDED.referred_for_glasses,
  icope_eye_problem           = EXCLUDED.icope_eye_problem,
  icope_treated_for_diabetes_or_bp = EXCLUDED.icope_treated_for_diabetes_or_bp;

-- name: DeleteVisualAcuity :exec
DELETE FROM visual_acuity
WHERE id = $1
  AND vid = $2;

-- Dental ---------------------------------------------------------------------

-- name: GetDental :one
SELECT id,
       vid,
       fluoride_exposure,
       diet,
       bacterial_exposure,
       oral_symptoms,
       drink_other_water,
       risk_for_dental_carries,
       icope_difficulty_chewing,
       icope_pain_in_mouth,
       dental_notes
FROM dental
WHERE id = $1
  AND vid = $2;

-- name: UpsertDental :exec
INSERT INTO dental (
  id,
  vid,
  fluoride_exposure,
  diet,
  bacterial_exposure,
  oral_symptoms,
  drink_other_water,
  risk_for_dental_carries,
  icope_difficulty_chewing,
  icope_pain_in_mouth,
  dental_notes
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (id, vid) DO UPDATE SET
  fluoride_exposure      = EXCLUDED.fluoride_exposure,
  diet                   = EXCLUDED.diet,
  bacterial_exposure     = EXCLUDED.bacterial_exposure,
  oral_symptoms          = EXCLUDED.oral_symptoms,
  drink_other_water      = EXCLUDED.drink_other_water,
  risk_for_dental_carries = EXCLUDED.risk_for_dental_carries,
  icope_difficulty_chewing = EXCLUDED.icope_difficulty_chewing,
  icope_pain_in_mouth    = EXCLUDED.icope_pain_in_mouth,
  dental_notes           = EXCLUDED.dental_notes;

-- name: DeleteDental :exec
DELETE FROM dental
WHERE id = $1
  AND vid = $2;

-- Fall Risk ------------------------------------------------------------------

-- name: GetFallRisk :one
SELECT id,
       vid,
       side_to_side_balance,
       semi_tandem_balance,
       tandem_balance,
       gait_speed_test,
       chair_stand_test,
       fall_risk_score,
       icope_complete_chair_stands,
       icope_chair_stands_time
FROM fall_risk
WHERE id = $1
  AND vid = $2;

-- name: UpsertFallRisk :exec
INSERT INTO fall_risk (
  id,
  vid,
  side_to_side_balance,
  semi_tandem_balance,
  tandem_balance,
  gait_speed_test,
  chair_stand_test,
  fall_risk_score,
  icope_complete_chair_stands,
  icope_chair_stands_time
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id, vid) DO UPDATE SET
  side_to_side_balance   = EXCLUDED.side_to_side_balance,
  semi_tandem_balance    = EXCLUDED.semi_tandem_balance,
  tandem_balance         = EXCLUDED.tandem_balance,
  gait_speed_test        = EXCLUDED.gait_speed_test,
  chair_stand_test       = EXCLUDED.chair_stand_test,
  fall_risk_score        = EXCLUDED.fall_risk_score,
  icope_complete_chair_stands = EXCLUDED.icope_complete_chair_stands,
  icope_chair_stands_time = EXCLUDED.icope_chair_stands_time;

-- name: DeleteFallRisk :exec
DELETE FROM fall_risk
WHERE id = $1
  AND vid = $2;

-- Physiotherapy --------------------------------------------------------------

-- name: GetPhysiotherapy :one
SELECT id,
       vid,
       subjective_assessment,
       pain_scale,
       objective_assessment,
       intervention,
       evaluation
FROM physiotherapy
WHERE id = $1
  AND vid = $2;

-- name: UpsertPhysiotherapy :exec
INSERT INTO physiotherapy (
  id,
  vid,
  subjective_assessment,
  pain_scale,
  objective_assessment,
  intervention,
  evaluation
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (id, vid) DO UPDATE SET
  subjective_assessment = EXCLUDED.subjective_assessment,
  pain_scale            = EXCLUDED.pain_scale,
  objective_assessment  = EXCLUDED.objective_assessment,
  intervention          = EXCLUDED.intervention,
  evaluation            = EXCLUDED.evaluation;

-- name: DeletePhysiotherapy :exec
DELETE FROM physiotherapy
WHERE id = $1
  AND vid = $2;

-- Doctors Consultation -------------------------------------------------------

-- name: GetDoctorsConsultation :one
SELECT id,
       vid,
       well,
       msk,
       cvs,
       respi,
       gu,
       git,
       eye,
       derm,
       others,
       consultation_notes,
       diagnosis,
       treatment,
       referral_needed,
       referral_loc,
       remarks
FROM doctors_consultation
WHERE id = $1
  AND vid = $2;

-- name: UpsertDoctorsConsultation :exec
INSERT INTO doctors_consultation (
  id,
  vid,
  well,
  msk,
  cvs,
  respi,
  gu,
  git,
  eye,
  derm,
  others,
  consultation_notes,
  diagnosis,
  treatment,
  referral_needed,
  referral_loc,
  remarks
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
ON CONFLICT (id, vid) DO UPDATE SET
  well               = EXCLUDED.well,
  msk                = EXCLUDED.msk,
  cvs                = EXCLUDED.cvs,
  respi              = EXCLUDED.respi,
  gu                 = EXCLUDED.gu,
  git                = EXCLUDED.git,
  eye                = EXCLUDED.eye,
  derm               = EXCLUDED.derm,
  others             = EXCLUDED.others,
  consultation_notes = EXCLUDED.consultation_notes,
  diagnosis          = EXCLUDED.diagnosis,
  treatment          = EXCLUDED.treatment,
  referral_needed    = EXCLUDED.referral_needed,
  referral_loc       = EXCLUDED.referral_loc,
  remarks            = EXCLUDED.remarks;

-- name: DeleteDoctorsConsultation :exec
DELETE FROM doctors_consultation
WHERE id = $1
  AND vid = $2;

-- Patient Metadata -----------------------------------------------------------

-- name: GetPatientVisitMetaLatest :many
WITH LatestDates AS (
  SELECT id, MAX(reg_date) AS latest_reg_date
  FROM admin
  GROUP BY id
)
SELECT DISTINCT ON (a.id)
  a.id,
  a.vid,
  a.family_group,
  a.reg_date,
  a.queue_no,
  a.name,
  a.khmer_name,
  a.gender,
  a.village,
  a.contact_no,
  a.drug_allergies,
  a.sent_to_id,
  dc.referral_needed,
  EXISTS (
    SELECT 1
    FROM prescriptions pr
    JOIN prescription_lines pl ON pl.prescription_id = pr.id
    WHERE pr.patient_id = a.id
      AND pr.vid = a.vid
  ) AS has_prescription_with_drug,
  CASE
    WHEN EXISTS (
      SELECT 1
      FROM prescriptions pr
      JOIN prescription_lines pl ON pl.prescription_id = pr.id
      WHERE pr.patient_id = a.id
        AND pr.vid = a.vid
    ) THEN NOT EXISTS (
      SELECT 1
      FROM prescriptions pr
      JOIN prescription_lines pl ON pl.prescription_id = pr.id
      WHERE pr.patient_id = a.id
        AND pr.vid = a.vid
        AND (pl.is_packed IS NOT TRUE)
    )
    ELSE FALSE
  END AS all_prescription_drugs_packed,
  EXISTS (
    SELECT 1
    FROM prescriptions pr
    WHERE pr.patient_id = a.id
      AND pr.vid = a.vid
      AND pr.is_dispensed = TRUE
  ) AS prescription_dispensed
FROM admin a
LEFT JOIN doctors_consultation dc
  ON a.id = dc.id AND a.vid = dc.vid
INNER JOIN LatestDates ld
  ON a.id = ld.id AND a.reg_date = ld.latest_reg_date
ORDER BY a.id, a.vid DESC;

-- name: GetPatientVisitMetaByDate :many
SELECT DISTINCT ON (a.id)
  a.id,
  a.vid,
  a.family_group,
  a.reg_date,
  a.queue_no,
  a.name,
  a.khmer_name,
  a.gender,
  a.village,
  a.contact_no,
  a.drug_allergies,
  a.sent_to_id,
  dc.referral_needed,
  EXISTS (
    SELECT 1
    FROM prescriptions pr
    JOIN prescription_lines pl ON pl.prescription_id = pr.id
    WHERE pr.patient_id = a.id
      AND pr.vid = a.vid
  ) AS has_prescription_with_drug,
  CASE
    WHEN EXISTS (
      SELECT 1
      FROM prescriptions pr
      JOIN prescription_lines pl ON pl.prescription_id = pr.id
      WHERE pr.patient_id = a.id
        AND pr.vid = a.vid
    ) THEN NOT EXISTS (
      SELECT 1
      FROM prescriptions pr
      JOIN prescription_lines pl ON pl.prescription_id = pr.id
      WHERE pr.patient_id = a.id
        AND pr.vid = a.vid
        AND (pl.is_packed IS NOT TRUE)
    )
    ELSE FALSE
  END AS all_prescription_drugs_packed,
  EXISTS (
    SELECT 1
    FROM prescriptions pr
    WHERE pr.patient_id = a.id
      AND pr.vid = a.vid
      AND pr.is_dispensed = TRUE
  ) AS prescription_dispensed
FROM admin a
LEFT JOIN doctors_consultation dc
  ON a.id = dc.id AND a.vid = dc.vid
WHERE a.reg_date = $1
ORDER BY a.id, a.vid DESC;
