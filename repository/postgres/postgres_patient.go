package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/jieqiboh/sothea_backend/util"
	"github.com/joho/sqltocsv"
	_ "github.com/lib/pq"
)

type postgresPatientRepository struct {
	Conn *sql.DB
}

// NewPostgresPatientRepository will create an object that represent the patient.Repository interface
func NewPostgresPatientRepository(conn *sql.DB) entities.PatientRepository {
	return &postgresPatientRepository{conn}
}

// GetPatientVisit returns a Patient struct representing a single visit based on ID, and Visit ID. Only guaranteed field is Admin
func (p *postgresPatientRepository) GetPatientVisit(ctx context.Context, id int32, vid int32) (res *entities.Patient, err error) {
	// Start a new transaction
	tx, err := p.Conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	rows := tx.QueryRowContext(ctx, "SELECT * FROM admin WHERE id = $1 AND vid = $2;", id, vid)
	admin := entities.Admin{}
	var photoBytes []byte
	err = rows.Scan(
		&admin.ID,
		&admin.VID,
		&admin.FamilyGroup,
		&admin.RegDate,
		&admin.QueueNo,
		&admin.Name,
		&admin.KhmerName,
		&admin.Dob,
		&admin.Age,
		&admin.Gender,
		&admin.Village,
		&admin.ContactNo,
		&admin.Pregnant,
		&admin.LastMenstrualPeriod,
		&admin.DrugAllergies,
		&admin.SentToID,
		&photoBytes,
	)
	photoBase64 := base64.StdEncoding.EncodeToString(photoBytes)
	admin.Photo = &photoBase64
	if err != nil { // no admin found
		return nil, entities.ErrPatientVisitNotFound
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM pastmedicalhistory WHERE pastmedicalhistory.id = $1 AND pastmedicalhistory.vid = $2;", id, vid)
	pastmedicalhistory := &entities.PastMedicalHistory{}
	err = rows.Scan(
		&pastmedicalhistory.ID,
		&pastmedicalhistory.VID,
		&pastmedicalhistory.Tuberculosis,
		&pastmedicalhistory.Diabetes,
		&pastmedicalhistory.Hypertension,
		&pastmedicalhistory.Hyperlipidemia,
		&pastmedicalhistory.ChronicJointPains,
		&pastmedicalhistory.ChronicMuscleAches,
		&pastmedicalhistory.SexuallyTransmittedDisease,
		&pastmedicalhistory.SpecifiedSTDs,
		&pastmedicalhistory.Others,
	)
	if errors.Is(err, sql.ErrNoRows) { // no pastmedicalhistory found
		pastmedicalhistory = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM socialhistory WHERE socialhistory.id = $1 AND socialhistory.vid = $2;", id, vid)
	socialhistory := &entities.SocialHistory{}
	err = rows.Scan(
		&socialhistory.ID,
		&socialhistory.VID,
		&socialhistory.PastSmokingHistory,
		&socialhistory.NumberOfYears,
		&socialhistory.CurrentSmokingHistory,
		&socialhistory.CigarettesPerDay,
		&socialhistory.AlcoholHistory,
		&socialhistory.HowRegular,
	)
	if errors.Is(err, sql.ErrNoRows) { // no socialhistory found
		socialhistory = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM vitalstatistics WHERE vitalstatistics.id = $1 AND vitalstatistics.vid = $2;", id, vid)
	vitalstatistics := &entities.VitalStatistics{}
	err = rows.Scan(
		&vitalstatistics.ID,
		&vitalstatistics.VID,
		&vitalstatistics.Temperature,
		&vitalstatistics.SpO2,
		&vitalstatistics.SystolicBP1,
		&vitalstatistics.DiastolicBP1,
		&vitalstatistics.SystolicBP2,
		&vitalstatistics.DiastolicBP2,
		&vitalstatistics.AverageSystolicBP,
		&vitalstatistics.AverageDiastolicBP,
		&vitalstatistics.HR1,
		&vitalstatistics.HR2,
		&vitalstatistics.AverageHR,
		&vitalstatistics.RandomBloodGlucoseMmolL,
	)
	if errors.Is(err, sql.ErrNoRows) { // no vitalstatistics found
		vitalstatistics = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM heightandweight WHERE heightandweight.id = $1 AND heightandweight.vid = $2;", id, vid)
	heightandweight := &entities.HeightAndWeight{}
	err = rows.Scan(
		&heightandweight.ID,
		&heightandweight.VID,
		&heightandweight.Height,
		&heightandweight.Weight,
		&heightandweight.BMI,
		&heightandweight.BMIAnalysis,
		&heightandweight.PaedsHeight,
		&heightandweight.PaedsWeight,
	)
	if errors.Is(err, sql.ErrNoRows) { // no heightandweight found
		heightandweight = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM visualacuity WHERE visualacuity.id = $1 AND visualacuity.vid = $2;", id, vid)
	visualacuity := &entities.VisualAcuity{}
	err = rows.Scan(
		&visualacuity.ID,
		&visualacuity.VID,
		&visualacuity.LEyeVision,
		&visualacuity.REyeVision,
		&visualacuity.AdditionalIntervention,
	)
	if errors.Is(err, sql.ErrNoRows) { // no visualacuity found
		visualacuity = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM fallrisk WHERE fallrisk.id = $1 AND fallrisk.vid = $2;", id, vid)
	fallrisk := &entities.FallRisk{}
	err = rows.Scan(
		&fallrisk.ID,
		&fallrisk.VID,
		&fallrisk.FallWorries,
		&fallrisk.FallHistory,
		&fallrisk.CognitiveStatus,
		&fallrisk.ContinenceProblems,
		&fallrisk.SafetyAwareness,
		&fallrisk.Unsteadiness,
		&fallrisk.FallRiskScore,
	)
	if errors.Is(err, sql.ErrNoRows) { // no fallrisk found
		fallrisk = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM dental WHERE dental.id = $1 AND dental.vid = $2;", id, vid)
	dental := &entities.Dental{}
	err = rows.Scan(
		&dental.ID,
		&dental.VID,
		&dental.CleanTeethFreq,
		&dental.SugarConsumeFreq,
		&dental.PastYearDecay,
		&dental.BrushTeethPain,
		&dental.DrinkOtherWater,
		&dental.DentalNotes,
		&dental.ReferralNeeded,
		&dental.ReferralLoc,
		&dental.Tooth11,
		&dental.Tooth12,
		&dental.Tooth13,
		&dental.Tooth14,
		&dental.Tooth15,
		&dental.Tooth16,
		&dental.Tooth17,
		&dental.Tooth18,
		&dental.Tooth21,
		&dental.Tooth22,
		&dental.Tooth23,
		&dental.Tooth24,
		&dental.Tooth25,
		&dental.Tooth26,
		&dental.Tooth27,
		&dental.Tooth28,
		&dental.Tooth31,
		&dental.Tooth32,
		&dental.Tooth33,
		&dental.Tooth34,
		&dental.Tooth35,
		&dental.Tooth36,
		&dental.Tooth37,
		&dental.Tooth38,
		&dental.Tooth41,
		&dental.Tooth42,
		&dental.Tooth43,
		&dental.Tooth44,
		&dental.Tooth45,
		&dental.Tooth46,
		&dental.Tooth47,
		&dental.Tooth48,
	)
	if errors.Is(err, sql.ErrNoRows) { // no dental found
		dental = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM physiotherapy WHERE physiotherapy.id = $1 AND physiotherapy.vid = $2;", id, vid)
	physiotherapy := &entities.Physiotherapy{}
	err = rows.Scan(
		&physiotherapy.ID,
		&physiotherapy.VID,
		&physiotherapy.PainStiffnessDay,
		&physiotherapy.PainStiffnessNight,
		&physiotherapy.SymptomsInterfereTasks,
		&physiotherapy.SymptomsChange,
		&physiotherapy.SymptomsNeedHelp,
		&physiotherapy.TroubleSleepSymptoms,
		&physiotherapy.HowMuchFatigue,
		&physiotherapy.AnxiousLowMood,
		&physiotherapy.MedicationManageSymptoms,
	)
	if errors.Is(err, sql.ErrNoRows) { // no physiotherapy found
		physiotherapy = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM doctorsconsultation WHERE doctorsconsultation.id = $1 AND doctorsconsultation.vid = $2;", id, vid)
	doctorsconsultation := &entities.DoctorsConsultation{}
	err = rows.Scan(
		&doctorsconsultation.ID,
		&doctorsconsultation.VID,
		&doctorsconsultation.Well,
		&doctorsconsultation.Msk,
		&doctorsconsultation.Cvs,
		&doctorsconsultation.Respi,
		&doctorsconsultation.Gu,
		&doctorsconsultation.Git,
		&doctorsconsultation.Eye,
		&doctorsconsultation.Derm,
		&doctorsconsultation.Others,
		&doctorsconsultation.ConsultationNotes,
		&doctorsconsultation.Diagnosis,
		&doctorsconsultation.Treatment,
		&doctorsconsultation.ReferralNeeded,
		&doctorsconsultation.ReferralLoc,
		&doctorsconsultation.Remarks,
	)
	if errors.Is(err, sql.ErrNoRows) { // no doctorsconsultation found
		doctorsconsultation = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	patient := entities.Patient{
		Admin:               &admin,
		PastMedicalHistory:  pastmedicalhistory,
		SocialHistory:       socialhistory,
		VitalStatistics:     vitalstatistics,
		HeightAndWeight:     heightandweight,
		VisualAcuity:        visualacuity,
		FallRisk:            fallrisk,
		Dental:              dental,
		Physiotherapy:       physiotherapy,
		DoctorsConsultation: doctorsconsultation,
	}

	if err = tx.Commit(); err != nil { // commit transaction
		return nil, err
	}

	return &patient, nil
}

// CreatePatient inserts a new Admin category for a new patient and returns the new id if successful.
func (p *postgresPatientRepository) CreatePatient(ctx context.Context, admin *entities.Admin) (int32, error) {
	// Start a new transaction
	tx, err := p.Conn.BeginTx(ctx, nil)
	if err != nil { // error starting transaction
		return -1, err
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	var patientid int32
	if admin == nil { // no admin field
		return -1, entities.ErrMissingAdminCategory
	}
	rows := tx.QueryRowContext(ctx, `INSERT INTO admin (family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, 
	contact_no, pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`,
		admin.FamilyGroup, admin.RegDate, admin.QueueNo, admin.Name, admin.KhmerName, admin.Dob, admin.Age, admin.Gender, admin.Village, admin.ContactNo,
		admin.Pregnant, admin.LastMenstrualPeriod, admin.DrugAllergies, admin.SentToID, admin.Photo)
	err = rows.Scan(&patientid)
	if err != nil { // error inserting admin
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}
	return patientid, nil
}

// CreatePatientVisit inserts a new Admin category for an existing patient and returns the new vid if successful. Only required field is Admin
func (p *postgresPatientRepository) CreatePatientVisit(ctx context.Context, id int32, admin *entities.Admin) (int32, error) {
	// Start a new transaction
	tx, err := p.Conn.BeginTx(ctx, nil)
	if err != nil { // error starting transaction
		return -1, err
	}
	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	var patientid int32
	if admin == nil { // no admin field
		return -1, entities.ErrMissingAdminCategory
	}

	// Check that patient exists
	doesPatientExist, err := p.checkPatientExists(ctx, id)
	if err != nil { // query error
		return -1, err
	} else if !doesPatientExist { // no query error, and patient doesn't exist
		return -1, entities.ErrPatientNotFound
	}

	rows := tx.QueryRowContext(ctx, `INSERT INTO admin (id, family_group, reg_date, queue_no, name, khmer_name, dob, age, gender, village, 
	contact_no, pregnant, last_menstrual_period, drug_allergies, sent_to_id, photo) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) RETURNING vid`,
		id, admin.FamilyGroup, admin.RegDate, admin.QueueNo, admin.Name, admin.KhmerName, admin.Dob, admin.Age, admin.Gender, admin.Village, admin.ContactNo,
		admin.Pregnant, admin.LastMenstrualPeriod, admin.DrugAllergies, admin.SentToID, admin.Photo)
	err = rows.Scan(&patientid)
	if err != nil { // error inserting admin
		return -1, err
	}

	if err = tx.Commit(); err != nil {
		return -1, err
	}
	return patientid, nil
}

// Deletes all patient entries where id and vid match
func (p *postgresPatientRepository) DeletePatientVisit(ctx context.Context, id int32, vid int32) error {
	// Start a new transaction
	tx, err := p.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM pastmedicalhistory WHERE pastmedicalhistory.id = $1 AND pastmedicalhistory.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM socialhistory WHERE socialhistory.id = $1 AND socialhistory.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM vitalstatistics WHERE vitalstatistics.id = $1 AND vitalstatistics.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM heightandweight WHERE heightandweight.id = $1 AND heightandweight.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM visualacuity WHERE visualacuity.id = $1 AND visualacuity.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM fallrisk WHERE fallrisk.id = $1 AND fallrisk.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM dental WHERE dental.id = $1 AND dental.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM physiotherapy WHERE physiotherapy.id = $1 AND physiotherapy.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM doctorsconsultation WHERE doctorsconsultation.id = $1 AND doctorsconsultation.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM prescriptions WHERE prescriptions.id = $1 AND prescriptions.vid = $2;", id, vid)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM admin WHERE id = $1 AND vid = $2", id, vid)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// UpdatePatientVisit updates a visit for an existing patient, filling out or overriding any of its fields
func (p *postgresPatientRepository) UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *entities.Patient) error {
	// Checks that a patient exists by searching for admin field
	// Then for each non-nil field in patient, updates it
	// Start a new transaction
	tx, err := p.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Defer a rollback in case anything fails.
	defer tx.Rollback()

	// Check that patient visit exists
	doesPatientVisitExist, err := p.checkPatientVisitExists(ctx, id, vid)
	if err != nil {
		return err
	} else if !doesPatientVisitExist {
		return entities.ErrPatientVisitNotFound
	}

	a := patient.Admin
	pmh := patient.PastMedicalHistory
	socialhistory := patient.SocialHistory
	vs := patient.VitalStatistics
	haw := patient.HeightAndWeight
	va := patient.VisualAcuity
	fr := patient.FallRisk
	d := patient.Dental
	phy := patient.Physiotherapy
	dc := patient.DoctorsConsultation
	if a != nil { // Update admin
		_, err = tx.ExecContext(ctx, `UPDATE admin SET family_group = $1, reg_date = $2, queue_no = $3, name = $4, khmer_name = $5, dob = $6, age = $7, 
		gender = $8, village = $9, contact_no = $10, pregnant = $11, last_menstrual_period = $12, drug_allergies = $13,
		sent_to_id = $14, photo = $15 WHERE id = $16 AND vid = $17`, a.FamilyGroup, a.RegDate, a.QueueNo, a.Name, a.KhmerName, a.Dob, a.Age, a.Gender, a.Village, a.ContactNo,
			a.Pregnant, a.LastMenstrualPeriod, a.DrugAllergies, a.SentToID, a.Photo, id, vid)
		if err != nil {
			return err
		}
	}
	if pmh != nil { // Update pastmedicalhistory, use insert into on conflict update because not it isn't guaranteed to exist
		_, err = tx.ExecContext(ctx, `
		INSERT INTO pastmedicalhistory (id, vid, tuberculosis, diabetes, hypertension, hyperlipidemia, chronic_joint_pains,
										 chronic_muscle_aches, sexually_transmitted_disease, specified_stds, others) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		ON CONFLICT (id, vid) DO UPDATE SET
			tuberculosis = $3, 
			diabetes = $4,
			hypertension = $5,
			hyperlipidemia = $6,
			chronic_joint_pains = $7,
			chronic_muscle_aches = $8,
			sexually_transmitted_disease = $9,
			specified_stds = $10,
			others = $11
		`, id, vid, pmh.Tuberculosis, pmh.Diabetes, pmh.Hypertension, pmh.Hyperlipidemia,
			pmh.ChronicJointPains, pmh.ChronicMuscleAches, pmh.SexuallyTransmittedDisease,
			pmh.SpecifiedSTDs, pmh.Others)
		if err != nil {
			return err
		}
	}
	if socialhistory != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO socialhistory (id, vid, past_smoking_history, no_of_years, current_smoking_history, cigarettes_per_day, 
		alcohol_history, how_regular) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (id, vid) DO UPDATE SET
			past_smoking_history = $3,
			no_of_years = $4,
			current_smoking_history = $5,
			cigarettes_per_day = $6,
			alcohol_history = $7,
			how_regular = $8
		`, id, vid, socialhistory.PastSmokingHistory, socialhistory.NumberOfYears, socialhistory.CurrentSmokingHistory,
			socialhistory.CigarettesPerDay, socialhistory.AlcoholHistory, socialhistory.HowRegular)

		if err != nil {
			return err
		}
	}
	if vs != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO vitalstatistics (id, vid, temperature, spO2, systolic_bp1, diastolic_bp1, systolic_bp2, diastolic_bp2, 
		avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmoll) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) 
		ON CONFLICT (id, vid) DO UPDATE SET
			temperature = $3,
			spO2 = $4,
			systolic_bp1 = $4,
			diastolic_bp1 = $6,
			systolic_bp2 = $7,
			diastolic_bp2 = $8,
			avg_systolic_bp = $9,
			avg_diastolic_bp = $10,
			hr1 = $11,
			hr2 = $12,
			avg_hr = $13,
			rand_blood_glucose_mmoll = $14
		`, id, vid, vs.Temperature, vs.SpO2, vs.SystolicBP1, vs.DiastolicBP1, vs.SystolicBP2, vs.DiastolicBP2,
			vs.AverageSystolicBP, vs.AverageDiastolicBP, vs.HR1, vs.HR2, vs.AverageHR, vs.RandomBloodGlucoseMmolL)

		if err != nil {
			return err
		}
	}
	if haw != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO heightandweight (id, vid, height, weight, bmi, bmi_analysis, paeds_height, paeds_weight) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		ON CONFLICT (id, vid) DO UPDATE SET
			height = $3,
			weight = $4,
			bmi = $5,
			bmi_analysis = $6,
			paeds_height = $7,
			paeds_weight = $8
		`, id, vid, haw.Height, haw.Weight, haw.BMI, haw.BMIAnalysis, haw.PaedsHeight, haw.PaedsWeight)

		if err != nil {
			return err
		}
	}
	if va != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO visualacuity (id, vid, l_eye_vision, r_eye_vision, additional_intervention) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (id, vid) DO UPDATE SET
			l_eye_vision = $3,
			r_eye_vision = $4,
			additional_intervention = $5
		`, id, vid, va.LEyeVision, va.REyeVision, va.AdditionalIntervention)

		if err != nil {
			return err
		}
	}
	if fr != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO fallrisk (id, vid, fall_worries, fall_history, cognitive_status, continence_problems, safety_awareness, unsteadiness, fall_risk_score) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT (id, vid) DO UPDATE SET
		    fall_worries = $3,
		    fall_history = $4,
			cognitive_status = $5,
			continence_problems = $6,
			safety_awareness= $7,
			unsteadiness = $8,
			fall_risk_score = $9
		`, id, vid, fr.FallWorries, fr.FallHistory, fr.CognitiveStatus, fr.ContinenceProblems, fr.SafetyAwareness, fr.Unsteadiness, fr.FallRiskScore)

		if err != nil {
			return err
		}
	}
	if d != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO dental (id, vid, clean_teeth_freq, sugar_consume_freq, past_year_decay, brush_teeth_pain, drink_other_water, 
		dental_notes, referral_needed, referral_loc, tooth_11, tooth_12, tooth_13, tooth_14, tooth_15, tooth_16, tooth_17, tooth_18, 
		tooth_21, tooth_22, tooth_23, tooth_24, tooth_25, tooth_26, tooth_27, tooth_28, tooth_31, tooth_32, tooth_33, tooth_34, tooth_35, 
		tooth_36, tooth_37, tooth_38, tooth_41, tooth_42, tooth_43, tooth_44, tooth_45, tooth_46, tooth_47, tooth_48) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, 
		$26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42) 
		ON CONFLICT (id, vid) DO UPDATE SET
			clean_teeth_freq = $3,
			sugar_consume_freq = $4,
			past_year_decay = $5,
			brush_teeth_pain = $6,
			drink_other_water = $7,
			dental_notes = $8,
			referral_needed = $9,
			referral_loc = $10,
			tooth_11 = $11,
			tooth_12 = $12,
			tooth_13 = $13,
			tooth_14 = $14,
			tooth_15 = $15,
			tooth_16 = $16,
			tooth_17 = $17,
			tooth_18 = $18,
			tooth_21 = $19,
			tooth_22 = $20,
			tooth_23 = $21,
			tooth_24 = $22,
			tooth_25 = $23,
			tooth_26 = $24,
			tooth_27 = $25,
			tooth_28 = $26,
			tooth_31 = $27,
			tooth_32 = $28,
			tooth_33 = $29,
			tooth_34 = $30,
			tooth_35 = $31,
			tooth_36 = $32,
			tooth_37 = $33,
			tooth_38 = $34,
			tooth_41 = $35,
			tooth_42 = $36,
			tooth_43 = $37,
			tooth_44 = $38,
			tooth_45 = $39,
			tooth_46 = $40,
			tooth_47 = $41,
			tooth_48 = $42
		`, id, vid, d.CleanTeethFreq, d.SugarConsumeFreq, d.PastYearDecay, d.BrushTeethPain, d.DrinkOtherWater, d.DentalNotes, d.ReferralNeeded, d.ReferralLoc, d.Tooth11, d.Tooth12, d.Tooth13, d.Tooth14, d.Tooth15, d.Tooth16, d.Tooth17, d.Tooth18, d.Tooth21, d.Tooth22, d.Tooth23, d.Tooth24, d.Tooth25, d.Tooth26, d.Tooth27, d.Tooth28, d.Tooth31, d.Tooth32, d.Tooth33, d.Tooth34, d.Tooth35, d.Tooth36, d.Tooth37, d.Tooth38, d.Tooth41, d.Tooth42, d.Tooth43, d.Tooth44, d.Tooth45, d.Tooth46, d.Tooth47, d.Tooth48)

		if err != nil {
			return err
		}
	}
	if phy != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO physiotherapy (id, vid, pain_stiffness_day, pain_stiffness_night, symptoms_interfere_tasks, symptoms_change, 
		symptoms_need_help, trouble_sleep_symptoms, how_much_fatigue, anxious_low_mood, medication_manage_symptoms) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		ON CONFLICT (id, vid) DO UPDATE SET
			pain_stiffness_day = $3,
			pain_stiffness_night = $4,
			symptoms_interfere_tasks = $5,
			symptoms_change = $6,
			symptoms_need_help = $7,
			trouble_sleep_symptoms = $8,
			how_much_fatigue = $9,
			anxious_low_mood = $10,
			medication_manage_symptoms = $11
		`, id, vid, phy.PainStiffnessDay, phy.PainStiffnessNight, phy.SymptomsInterfereTasks, phy.SymptomsChange, phy.SymptomsNeedHelp, phy.TroubleSleepSymptoms, phy.HowMuchFatigue, phy.AnxiousLowMood, phy.MedicationManageSymptoms)

		if err != nil {
			return err
		}
	}
	if dc != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO doctorsconsultation (id, vid, well, msk, cvs, respi, gu, git, eye, derm, others, 
		consultation_notes, diagnosis, treatment, referral_needed, referral_loc, remarks) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) 
		ON CONFLICT(id, vid) 
		DO UPDATE SET
			well = $3,
			msk = $4,
			cvs = $5,
			respi = $6,
			gu = $7,
			git = $8,
			eye = $9,
			derm = $10,
			others = $11,
			consultation_notes = $12,
			diagnosis = $13,
			treatment = $14,
			referral_needed = $15,
			referral_loc = $16,
			remarks = $17
		`,
			id, vid, dc.Well, dc.Msk, dc.Cvs, dc.Respi, dc.Gu, dc.Git, dc.Eye, dc.Derm, dc.Others, dc.ConsultationNotes,
			dc.Diagnosis, dc.Treatment, dc.ReferralNeeded, dc.ReferralLoc, dc.Remarks)

		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p *postgresPatientRepository) GetPatientMeta(ctx context.Context, id int32) (*entities.PatientMeta, error) {
	// Check that patient exists
	doesPatientExist, err := p.checkPatientExists(ctx, id)
	if err != nil { // query error
		return nil, err
	} else if !doesPatientExist { // no query error, and patient doesn't exist
		return nil, entities.ErrPatientNotFound
	}

	// Gets metadata for a specific patient, invoked when navigating to other visits of a patient
	// For FamilyGroup, RegDate, QueueNo, Name and KhmerName, the values from the latest visit are used
	patientMeta := entities.PatientMeta{}
	patientMeta.Visits = make(map[int32]time.Time) // Initialize the Visits map

	// Get latest row
	latestRow := p.Conn.QueryRowContext(ctx, `SELECT id, vid, family_group, reg_date, queue_no, name, khmer_name FROM admin WHERE id = $1 ORDER BY reg_date DESC LIMIT 1`, id)
	err = latestRow.Scan(&patientMeta.ID, &patientMeta.VID, &patientMeta.FamilyGroup, &patientMeta.RegDate, &patientMeta.QueueNo, &patientMeta.Name, &patientMeta.KhmerName)
	if err != nil {
		return nil, err
	}

	// Get vid and reg_date
	rows, err := p.Conn.QueryContext(ctx, "SELECT vid, reg_date FROM ADMIN WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// Iterate through the result set and populate the Visits map
	for rows.Next() {
		var vid int32
		var visitDate time.Time
		if err := rows.Scan(&vid, &visitDate); err != nil {
			return nil, err
		}
		patientMeta.Visits[vid] = visitDate
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &patientMeta, nil
}

func (p *postgresPatientRepository) GetAllPatientVisitMeta(ctx context.Context, date time.Time) ([]entities.PatientVisitMeta, error) {
	// If date is non-empty, for every patient, return patientvisitmeta of their visit on that date if it exists
	// If date is empty aka default constructor, for every patient, return patientvisitmeta of their latest visit
	var rows *sql.Rows
	var err error
	result := make([]entities.PatientVisitMeta, 0)

	if date.IsZero() { // Date is empty
		rows, err = p.Conn.QueryContext(ctx, `WITH LatestDates AS (SELECT id, MAX(reg_date) AS latest_reg_date FROM admin GROUP BY id)
													SELECT DISTINCT ON (a.id) 
														a.id, a.vid, a.family_group, a.reg_date, a.queue_no, a.name, 
														a.khmer_name, a.gender, a.village, a.contact_no, a.drug_allergies, 
														a.sent_to_id, dc.referral_needed
													FROM 
														admin a
													LEFT JOIN 
														doctorsconsultation dc
													ON 
														a.id = dc.id AND a.vid = dc.vid -- assuming the foreign key relationship
													INNER JOIN 
														LatestDates ld
													ON 
														a.id = ld.id AND a.reg_date = ld.latest_reg_date
													ORDER BY 
														a.id, 
														a.vid DESC;`)
		if err != nil {
			return nil, err
		}
	} else { // Date is non-empty
		formattedDate := date.Format("2006-01-02")
		rows, err = p.Conn.QueryContext(ctx, `SELECT DISTINCT ON (a.id) 
													a.id, a.vid, a.family_group, a.reg_date, a.queue_no, a.name, 
													a.khmer_name, a.gender, a.village, a.contact_no, a.drug_allergies, 
													a.sent_to_id, dc.referral_needed
												FROM 
													admin a
												LEFT JOIN 
													doctorsconsultation dc
												ON 
													a.id = dc.id AND a.vid = dc.vid
												WHERE 
													a.reg_date = $1
												ORDER BY 
													a.id, 
													a.vid DESC;`, formattedDate)

		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		patientVisitMeta := entities.PatientVisitMeta{}
		err = rows.Scan(
			&patientVisitMeta.ID,
			&patientVisitMeta.VID,
			&patientVisitMeta.FamilyGroup,
			&patientVisitMeta.RegDate,
			&patientVisitMeta.QueueNo,
			&patientVisitMeta.Name,
			&patientVisitMeta.KhmerName,
			&patientVisitMeta.Gender,
			&patientVisitMeta.Village,
			&patientVisitMeta.ContactNo,
			&patientVisitMeta.DrugAllergies,
			&patientVisitMeta.SentToID,
			&patientVisitMeta.ReferralNeeded)
		if err != nil {
			return nil, err
		}
		result = append(result, patientVisitMeta)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *postgresPatientRepository) ExportDatabaseToCSV(ctx context.Context, includePhoto bool) error {
	// Base query without the photo field
	query := `SELECT
        a.id,
        a.vid,
        a.family_group AS a_family_group,
        a.reg_date AS a_reg_date,
        a.queue_no AS a_queue_no,
		a.name AS a_name,
		a.khmer_name AS a_khmer_name,
		a.dob AS a_dob,
		a.age AS a_age,
		a.gender AS a_gender,
		a.village AS a_village,
		a.contact_no AS a_contact_no,
		a.pregnant AS a_pregnant,
		a.last_menstrual_period AS a_last_menstrual_period,
		a.drug_allergies AS a_drug_allergies,
		a.sent_to_id AS a_sent_to_id,
        -- Past Medical History
		pmh.tuberculosis AS pmh_tuberculosis,
		pmh.diabetes AS pmh_diabetes,
		pmh.hypertension AS pmh_hypertension,
		pmh.hyperlipidemia AS pmh_hyperlipidemia,
		pmh.chronic_joint_pains AS pmh_chronic_joint_pains,
		pmh.chronic_muscle_aches AS pmh_chronic_muscle_aches,
		pmh.sexually_transmitted_disease AS pmh_sexually_transmitted_disease,
		pmh.specified_stds AS pmh_specified_stds,
		pmh.others AS pmh_others,
        -- Social History
		sh.past_smoking_history AS sh_past_smoking_history,
		sh.no_of_years AS sh_no_of_years,
		sh.current_smoking_history AS sh_current_smoking_history,
		sh.cigarettes_per_day AS sh_cigarettes_per_day,
		sh.alcohol_history AS sh_alcohol_history,
		sh.how_regular AS sh_how_regular,
        -- Vital Statistics
		vs.temperature AS vs_temperature,
		vs.spo2 AS vs_spo2,
		vs.systolic_bp1 AS vs_systolic_bp1,
		vs.diastolic_bp1 AS vs_diastolic_bp1,
		vs.systolic_bp2 AS vs_systolic_bp2,
		vs.diastolic_bp2 AS vs_diastolic_bp2,
		vs.avg_systolic_bp AS vs_avg_systolic_bp,
		vs.avg_diastolic_bp AS vs_avg_diastolic_bp,
		vs.hr1 AS vs_hr1,
		vs.hr2 AS vs_hr2,
		vs.avg_hr AS vs_avg_hr,
		vs.rand_blood_glucose_mmoll AS vs_rand_blood_glucose_mmoll,
        -- Height and Weight
        haw.height AS haw_height,
        haw.weight AS haw_weight,
        haw.bmi AS haw_bmi,
        haw.bmi_analysis AS haw_bmi_analysis,
        haw.paeds_height AS haw_paeds_height,
        haw.paeds_weight AS haw_paeds_weight,
        -- Visual Acuity
        va.l_eye_vision AS va_l_eye_vision,
        va.r_eye_vision AS va_r_eye_vision,
        va.additional_intervention AS va_additional_intervention,
		-- Dental
        d.clean_teeth_freq AS d_clean_teeth_freq,
        d.sugar_consume_freq AS d_sugar_consume_freq,
        d.past_year_decay AS d_past_year_decay,
        d.brush_teeth_pain AS d_brush_teeth_pain,
        d.drink_other_water AS d_drink_other_water,
        d.dental_notes AS d_dental_notes,
        d.referral_needed AS d_referral_needed,
        d.referral_loc AS d_referral_loc,
        d.tooth_11 AS d_tooth_11,
        d.tooth_12 AS d_tooth_12,
        d.tooth_13 AS d_tooth_13,
        d.tooth_14 AS d_tooth_14,
        d.tooth_15 AS d_tooth_15,
        d.tooth_16 AS d_tooth_16,
        d.tooth_17 AS d_tooth_17,
        d.tooth_18 AS d_tooth_18, -- Right Upper
        d.tooth_21 AS d_tooth_21,
        d.tooth_22 AS d_tooth_22,
        d.tooth_23 AS d_tooth_23,
        d.tooth_24 AS d_tooth_24,
        d.tooth_25 AS d_tooth_25,
        d.tooth_26 AS d_tooth_26,
        d.tooth_27 AS d_tooth_27,
        d.tooth_28 AS d_tooth_28, -- Left Upper
        d.tooth_31 AS d_tooth_31,
        d.tooth_32 AS d_tooth_32,
        d.tooth_33 AS d_tooth_33,
        d.tooth_34 AS d_tooth_34,
        d.tooth_35 AS d_tooth_35,
        d.tooth_36 AS d_tooth_36,
        d.tooth_37 AS d_tooth_37,
        d.tooth_38 AS d_tooth_38, -- Left Lower
        d.tooth_41 AS d_tooth_41,
        d.tooth_42 AS d_tooth_42,
        d.tooth_43 AS d_tooth_43,
        d.tooth_44 AS d_tooth_44,
        d.tooth_45 AS d_tooth_45,
        d.tooth_46 AS d_tooth_46,
        d.tooth_47 AS d_tooth_47,
        d.tooth_48 AS d_tooth_48,  -- Right Lower
        -- Fall Risk
        fr.fall_worries AS fr_fall_worries,
        fr.fall_history AS fr_fall_history,
        fr.cognitive_status AS fr_cognitive_status,
        fr.continence_problems AS fr_continence_problems,
        fr.safety_awareness AS fr_safety_awareness,
        fr.unsteadiness AS fr_unsteadiness,
        fr.fall_risk_score AS fr_fall_risk_score,
        -- Doctors Consultation
        dc.well AS dc_well,
        dc.msk AS dc_msk,
        dc.cvs AS dc_cvs,
        dc.respi AS dc_respi,
        dc.gu AS dc_gu,
        dc.git AS dc_git,
        dc.eye AS dc_eye,
        dc.derm AS dc_derm,
        dc.others AS dc_others,
        dc.consultation_notes AS dc_consultation_notes,
        dc.diagnosis AS dc_diagnosis,
        dc.treatment AS dc_treatment,
        dc.referral_needed AS dc_referral_needed,
        dc.referral_loc AS dc_referral_loc,
        dc.remarks AS dc_remarks,
		-- Physiotherapy
		p.pain_stiffness_day AS p_pain_stiffness_day,
		p.pain_stiffness_night AS p_pain_stiffness_night,
		p.symptoms_interfere_tasks AS p_symptoms_interfere_tasks,
		p.symptoms_change AS p_symptoms_change,
		p.symptoms_need_help AS p_symptoms_need_help,
		p.trouble_sleep_symptoms AS p_trouble_sleep_symptoms,
		p.how_much_fatigue AS p_how_much_fatigue,
		p.anxious_low_mood AS p_anxious_low_mood,
		p.medication_manage_symptoms AS p_medication_manage_symptoms`

	// Conditionally add the photo field at the end of the query
	if includePhoto {
		query += `, a.photo`
	} else {
		query += `, NULL AS a_photo`
	}

	// Now query will include a.photo at the end if includePhoto is true, or NULL if false
	query += ` FROM
        admin a
    LEFT JOIN
        pastmedicalhistory pmh ON a.id = pmh.id AND a.vid = pmh.vid
    LEFT JOIN
        socialhistory sh ON a.id = sh.id AND a.vid = sh.vid
    LEFT JOIN
        vitalstatistics vs ON a.id = vs.id AND a.vid = vs.vid
    LEFT JOIN
        heightandweight haw ON a.id = haw.id AND a.vid = haw.vid
    LEFT JOIN
        visualacuity va ON a.id = va.id AND a.vid = va.vid
    LEFT JOIN 
		dental d ON a.id = d.id AND a.vid = d.vid
    LEFT JOIN
        fallrisk fr ON a.id = fr.id AND a.vid = fr.vid
	LEFT JOIN
		physiotherapy p ON a.id = p.id AND a.vid = p.vid
    LEFT JOIN
        doctorsconsultation dc ON a.id = dc.id AND a.vid = dc.vid`

	// Execute the query
	rows, err := p.Conn.QueryContext(ctx, query)
	if err != nil {
		return err
	}

	filePath := util.MustGitPath("repository/tmp/output.csv")
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	conv := sqltocsv.New(rows)
	conv.TimeFormat = "2006-01-02"

	err = conv.WriteFile(filePath)
	if err != nil {
		panic(err)
	}

	return nil
}

func (p *postgresPatientRepository) GetDBUser(ctx context.Context, username string) (*entities.DBUser, error) {
	user := entities.DBUser{}

	// Get latest row
	latestRow := p.Conn.QueryRowContext(ctx, `SELECT username, password_hash FROM users WHERE username = $1`, username)
	err := latestRow.Scan(&user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *postgresPatientRepository) checkPatientExists(ctx context.Context, id int32) (bool, error) {
	// Helper method to check that a patient exists
	var resId int32
	err := p.Conn.QueryRowContext(ctx, "SELECT id FROM admin WHERE id = $1;", id).Scan(&resId)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Fatalf("query error: %v\n", err)
		return false, err
	}

	return true, nil
}

func (p *postgresPatientRepository) checkPatientVisitExists(ctx context.Context, id int32, vid int32) (bool, error) {
	var resId int32
	var resVid int32
	err := p.Conn.QueryRowContext(ctx, "SELECT id, vid FROM admin WHERE id = $1 AND vid = $2;", id, vid).Scan(&resId, &resVid)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		log.Fatalf("query error: %v\n", err)
		return false, err
	}

	return true, nil
}
