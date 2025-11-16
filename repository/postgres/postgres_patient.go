package postgres

import (
	"context"
	"database/sql"
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
	)
	if err != nil { // no admin found
		return nil, entities.ErrPatientVisitNotFound
	}

	rows = tx.QueryRowContext(ctx, "SELECT * FROM pastmedicalhistory WHERE pastmedicalhistory.id = $1 AND pastmedicalhistory.vid = $2;", id, vid)
	pastmedicalhistory := &entities.PastMedicalHistory{}
	err = rows.Scan(
		&pastmedicalhistory.ID,
		&pastmedicalhistory.VID,

		&pastmedicalhistory.Cough,
		&pastmedicalhistory.Fever,
		&pastmedicalhistory.BlockedNose,
		&pastmedicalhistory.SoreThroat,
		&pastmedicalhistory.NightSweats,
		&pastmedicalhistory.UnintentionalWeightLoss,

		&pastmedicalhistory.Tuberculosis,
		&pastmedicalhistory.TuberculosisHasBeenTreated,

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
		&vitalstatistics.IcopeHighBp,
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
		&heightandweight.IcopeLostWeightPastMonths,
		&heightandweight.IcopeNoDesireToEat,
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
		&visualacuity.SentToOpto,
		&visualacuity.ReferredForGlasses,
		&visualacuity.IcopeEyeProblem,
		&visualacuity.IcopeTreatedForDiabetesOrBp,
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
		&fallrisk.SideToSideBalance,
		&fallrisk.SemiTandemBalance,
		&fallrisk.TandemBalance,
		&fallrisk.GaitSpeedTest,
		&fallrisk.ChairStandTest,
		&fallrisk.FallRiskScore,
		&fallrisk.IcopeCompleteChairStands,
		&fallrisk.IcopeChairStandsTime,
	)
	if errors.Is(err, sql.ErrNoRows) { // no fallrisk found
		fallrisk = nil
	} else if err != nil { // unknown error
		return nil, err
	}

	rows = tx.QueryRowContext(ctx, `
		SELECT id, vid,
		       fluoride_exposure, diet, bacterial_exposure,
		       oral_symptoms, drink_other_water,
		       risk_for_dental_carries, icope_difficulty_chewing, icope_pain_in_mouth,
		       dental_notes
		FROM dental
		WHERE dental.id = $1 AND dental.vid = $2;`, id, vid)
	dental := &entities.Dental{}
	err = rows.Scan(
		&dental.ID,
		&dental.VID,
		&dental.FluorideExposure,
		&dental.Diet,
		&dental.BacterialExposure,
		&dental.OralSymptoms,
		&dental.DrinkOtherWater,
		&dental.RiskForDentalCarries,
		&dental.IcopeDifficultyChewing,
		&dental.IcopePainInMouth,
		&dental.DentalNotes,
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
		contact_no, pregnant, last_menstrual_period, drug_allergies, sent_to_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id`,
		admin.FamilyGroup, admin.RegDate, admin.QueueNo, admin.Name, admin.KhmerName, admin.Dob, admin.Age, admin.Gender, admin.Village, admin.ContactNo,
		admin.Pregnant, admin.LastMenstrualPeriod, admin.DrugAllergies, admin.SentToID)
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
		contact_no, pregnant, last_menstrual_period, drug_allergies, sent_to_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING vid`,
		id, admin.FamilyGroup, admin.RegDate, admin.QueueNo, admin.Name, admin.KhmerName, admin.Dob, admin.Age, admin.Gender, admin.Village, admin.ContactNo,
		admin.Pregnant, admin.LastMenstrualPeriod, admin.DrugAllergies, admin.SentToID)
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
		sent_to_id = $14 WHERE id = $15 AND vid = $16`, a.FamilyGroup, a.RegDate, a.QueueNo, a.Name, a.KhmerName, a.Dob, a.Age, a.Gender, a.Village, a.ContactNo,
			a.Pregnant, a.LastMenstrualPeriod, a.DrugAllergies, a.SentToID, id, vid)
		if err != nil {
			return err
		}
	}
	if pmh != nil { // Update pastmedicalhistory, use insert into on conflict update because not it isn't guaranteed to exist
		_, err = tx.ExecContext(ctx, `
		INSERT INTO pastmedicalhistory (
			id, vid,
			cough, fever, blocked_nose, sore_throat, night_sweats, unintentional_weight_loss,
			tuberculosis, tuberculosis_has_been_treated,
			diabetes, hypertension, hyperlipidemia, chronic_joint_pains,
			chronic_muscle_aches, sexually_transmitted_disease, specified_stds, others) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) 
		ON CONFLICT (id, vid) DO UPDATE SET
			cough = $3, 
			fever = $4,
			blocked_nose = $5,
			sore_throat = $6,
			night_sweats = $7,
			unintentional_weight_loss = $8,
			tuberculosis = $9,
			tuberculosis_has_been_treated = $10,
			diabetes = $11,
			hypertension = $12,
			hyperlipidemia = $13,
			chronic_joint_pains = $14,
			chronic_muscle_aches = $15,
			sexually_transmitted_disease = $16,
			specified_stds = $17,
			others = $18
		`,
			id, vid,
			pmh.Cough, pmh.Fever, pmh.BlockedNose, pmh.SoreThroat, pmh.NightSweats, pmh.UnintentionalWeightLoss,
			pmh.Tuberculosis, pmh.TuberculosisHasBeenTreated,
			pmh.Diabetes, pmh.Hypertension, pmh.Hyperlipidemia, pmh.ChronicJointPains,
			pmh.ChronicMuscleAches, pmh.SexuallyTransmittedDisease, pmh.SpecifiedSTDs, pmh.Others,
		)
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
		INSERT INTO vitalstatistics (
			id, vid,
			temperature, spO2, systolic_bp1, diastolic_bp1, systolic_bp2, diastolic_bp2, 
			avg_systolic_bp, avg_diastolic_bp, hr1, hr2, avg_hr, rand_blood_glucose_mmoll,
			icope_high_bp
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) 
		ON CONFLICT (id, vid) DO UPDATE SET
			temperature = $3,
			spO2 = $4,
			systolic_bp1 = $5,
			diastolic_bp1 = $6,
			systolic_bp2 = $7,
			diastolic_bp2 = $8,
			avg_systolic_bp = $9,
			avg_diastolic_bp = $10,
			hr1 = $11,
			hr2 = $12,
			avg_hr = $13,
			rand_blood_glucose_mmoll = $14,
			icope_high_bp = $15
		`,
			id, vid,
			vs.Temperature, vs.SpO2, vs.SystolicBP1, vs.DiastolicBP1, vs.SystolicBP2, vs.DiastolicBP2,
			vs.AverageSystolicBP, vs.AverageDiastolicBP, vs.HR1, vs.HR2, vs.AverageHR, vs.RandomBloodGlucoseMmolL,
			vs.IcopeHighBp,
		)

		if err != nil {
			return err
		}
	}
	if haw != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO heightandweight (
			id, vid,
			height, weight, bmi, bmi_analysis, paeds_height, paeds_weight,
			icope_lost_weight_past_months, icope_no_desire_to_eat
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		ON CONFLICT (id, vid) DO UPDATE SET
			height = $3,
			weight = $4,
			bmi = $5,
			bmi_analysis = $6,
			paeds_height = $7,
			paeds_weight = $8,
			icope_lost_weight_past_months = $9,
			icope_no_desire_to_eat = $10
		`, id, vid, haw.Height, haw.Weight, haw.BMI, haw.BMIAnalysis, haw.PaedsHeight, haw.PaedsWeight, haw.IcopeLostWeightPastMonths, haw.IcopeNoDesireToEat)

		if err != nil {
			return err
		}
	}
	if va != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO visualacuity (
			id, vid,
			l_eye_vision, r_eye_vision, additional_intervention,
			sent_to_opto, referred_for_glasses,
			icope_eye_problem, icope_treated_for_diabetes_or_bp
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT (id, vid) DO UPDATE SET
			l_eye_vision = $3,
			r_eye_vision = $4,
			additional_intervention = $5,
			sent_to_opto = $6,
			referred_for_glasses = $7,
			icope_eye_problem = $8,
			icope_treated_for_diabetes_or_bp = $9
		`, id, vid, va.LEyeVision, va.REyeVision, va.AdditionalIntervention,
			va.SentToOpto, va.ReferredForGlasses, va.IcopeEyeProblem, va.IcopeTreatedForDiabetesOrBp)

		if err != nil {
			return err
		}
	}
	if fr != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO fallrisk (
			id, vid,
			side_to_side_balance, semi_tandem_balance, tandem_balance,
			gait_speed_test, chair_stand_test,
			fall_risk_score,
			icope_complete_chair_stands, icope_chair_stands_time
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		ON CONFLICT (id, vid) DO UPDATE SET
		    side_to_side_balance = $3,
		    semi_tandem_balance = $4,
			tandem_balance = $5,
			gait_speed_test = $6,
			chair_stand_test = $7,
			fall_risk_score= $8,
			icope_complete_chair_stands = $9,
			icope_chair_stands_time = $10
		`,
			id, vid,
			fr.SideToSideBalance, fr.SemiTandemBalance, fr.TandemBalance,
			fr.GaitSpeedTest, fr.ChairStandTest,
			fr.FallRiskScore,
			fr.IcopeCompleteChairStands,
			fr.IcopeChairStandsTime,
		)

		if err != nil {
			return err
		}
	}
	if d != nil {
		_, err = tx.ExecContext(ctx, `
		INSERT INTO dental (
			id, vid,
			fluoride_exposure, diet, bacterial_exposure,
			oral_symptoms, drink_other_water,
			risk_for_dental_carries,
			icope_difficulty_chewing, icope_pain_in_mouth,
			dental_notes
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		ON CONFLICT (id, vid) DO UPDATE SET
			fluoride_exposure = $3,
			diet = $4,
			bacterial_exposure = $5,
			oral_symptoms =  $6,
			drink_other_water = $7,
			risk_for_dental_carries = $8,
			icope_difficulty_chewing = $9,
			icope_pain_in_mouth = $10,
			dental_notes = $11
		`,
			id, vid,
			d.FluorideExposure, d.Diet, d.BacterialExposure,
			d.OralSymptoms, d.DrinkOtherWater,
			d.RiskForDentalCarries,
			d.IcopeDifficultyChewing, d.IcopePainInMouth,
			d.DentalNotes,
		)

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
		rows, err = p.Conn.QueryContext(ctx, `WITH LatestDates AS (
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
													-- Has at least one prescription line (i.e. at least one drug) for this visit
													EXISTS (
														SELECT 1
														FROM prescriptions pr
														JOIN prescription_lines pl ON pl.prescription_id = pr.id
														WHERE pr.patient_id = a.id
														  AND pr.vid = a.vid
													) AS has_prescription_with_drug,
													-- If there are lines, TRUE only when none are unpacked
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
													-- Any prescription for this visit has been dispensed
													EXISTS (
														SELECT 1
														FROM prescriptions pr
														WHERE pr.patient_id = a.id
														  AND pr.vid = a.vid
														  AND pr.is_dispensed = TRUE
													) AS prescription_dispensed
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
			&patientVisitMeta.ReferralNeeded,
			&patientVisitMeta.HasPrescriptionWithDrug,
			&patientVisitMeta.AllPrescriptionDrugsPacked,
			&patientVisitMeta.PrescriptionDispensed)
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

func (p *postgresPatientRepository) ExportDatabaseToCSV(ctx context.Context) error {
	// Base query
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
        d.fluoride_exposure AS d_fluoride_exposure,
        d.diet AS d_diet,
        d.bacterial_exposure AS d_bacterial_exposure,
        d.oral_symptoms AS d_oral_symptoms,
        d.drink_other_water AS d_drink_other_water,
        d.risk_for_dental_carries AS d_risk_for_dental_carries,
        d.icope_difficulty_chewing AS d_icope_difficulty_chewing,
        d.icope_pain_in_mouth AS d_icope_pain_in_mouth,
        d.dental_notes AS d_dental_notes,
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
		p.medication_manage_symptoms AS p_medication_manage_symptoms
		FROM
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
	latestRow := p.Conn.QueryRowContext(ctx, `SELECT id, username, password_hash FROM users WHERE username = $1`, username)
	err := latestRow.Scan(&user.Id, &user.Username, &user.PasswordHash)
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
