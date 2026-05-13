package postgres

import (
	"context"
	"errors"
	"time"

	"sothea-backend/entities"
	db "sothea-backend/repository/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPatientRepository struct {
	Conn     *pgxpool.Pool
	queries  *db.Queries
	timezone *time.Location
}

// NewPostgresPatientRepository will create an object that represent the patient.Repository interface
func NewPostgresPatientRepository(conn *pgxpool.Pool, timezone *time.Location) *PostgresPatientRepository {
	return &PostgresPatientRepository{
		Conn:     conn,
		queries:  db.New(conn),
		timezone: timezone,
	}
}

// GetPatientVisit returns a Patient struct representing a single visit based on ID and visit ID. Patient and Admin are guaranteed when found.
func (p *PostgresPatientRepository) GetPatientVisit(ctx context.Context, id int32, vid int32) (*entities.Patient, error) {
	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)

	patientRow, err := q.GetPatient(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entities.ErrPatientNotFound
	}
	if err != nil {
		return nil, err
	}

	adminRow, err := q.GetAdmin(ctx, db.GetAdminParams{ID: id, Vid: vid})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entities.ErrPatientVisitNotFound
	}
	if err != nil {
		return nil, err
	}
	admin := &adminRow

	var pastmedicalhistory *db.PastMedicalHistory
	if row, err := q.GetPastMedicalHistory(ctx, db.GetPastMedicalHistoryParams{ID: id, Vid: vid}); err == nil {
		pastmedicalhistory = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var socialhistory *db.SocialHistory
	if row, err := q.GetSocialHistory(ctx, db.GetSocialHistoryParams{ID: id, Vid: vid}); err == nil {
		socialhistory = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var vitalstatistics *db.VitalStatistic
	if row, err := q.GetVitalStatistics(ctx, db.GetVitalStatisticsParams{ID: id, Vid: vid}); err == nil {
		vitalstatistics = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var heightandweight *db.HeightAndWeight
	if row, err := q.GetHeightAndWeight(ctx, db.GetHeightAndWeightParams{ID: id, Vid: vid}); err == nil {
		heightandweight = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var visualacuity *db.VisualAcuity
	if row, err := q.GetVisualAcuity(ctx, db.GetVisualAcuityParams{ID: id, Vid: vid}); err == nil {
		visualacuity = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var fallrisk *db.FallRisk
	if row, err := q.GetFallRisk(ctx, db.GetFallRiskParams{ID: id, Vid: vid}); err == nil {
		fallrisk = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var dental *db.Dental
	if row, err := q.GetDental(ctx, db.GetDentalParams{ID: id, Vid: vid}); err == nil {
		dental = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var physiotherapy *db.Physiotherapy
	if row, err := q.GetPhysiotherapy(ctx, db.GetPhysiotherapyParams{ID: id, Vid: vid}); err == nil {
		physiotherapy = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var doctorsconsultation *db.DoctorsConsultation
	if row, err := q.GetDoctorsConsultation(ctx, db.GetDoctorsConsultationParams{ID: id, Vid: vid}); err == nil {
		doctorsconsultation = &row
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	patient := entities.Patient{
		PatientDetails:      &patientRow,
		Admin:               admin,
		PastMedicalHistory:  pastmedicalhistory,
		SocialHistory:       socialhistory,
		VitalStatistics:     vitalstatistics,
		HeightAndWeight:     heightandweight,
		VisualAcuity:        visualacuity,
		Dental:              dental,
		FallRisk:            fallrisk,
		Physiotherapy:       physiotherapy,
		DoctorsConsultation: doctorsconsultation,
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &patient, nil
}

// CreatePatient inserts a new patient record and returns the patient id.
func (p *PostgresPatientRepository) CreatePatient(ctx context.Context, patient *db.PatientDetail) (int32, error) {
	if patient == nil {
		return -1, entities.ErrMissingPatientData
	}

	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)
	patientParams, err := toInsertPatientParams(patient)
	if err != nil {
		return -1, err
	}
	patientParams.Dob = patientParams.Dob.In(p.timezone)

	patientID, err := q.InsertPatient(ctx, patientParams)
	if err != nil {
		return -1, err
	}

	if err = tx.Commit(ctx); err != nil {
		return -1, err
	}
	return patientID, nil
}

// CreatePatientWithVisit inserts a patient and first visit in a single transaction.
func (p *PostgresPatientRepository) CreatePatientWithVisit(ctx context.Context, patient *db.PatientDetail, admin *db.Admin) (int32, int32, error) {
	if patient == nil {
		return -1, -1, entities.ErrMissingPatientData
	}
	if admin == nil {
		return -1, -1, entities.ErrMissingAdminCategory
	}

	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return -1, -1, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)

	patientParams, err := toInsertPatientParams(patient)
	if err != nil {
		return -1, -1, err
	}
	patientParams.Dob = patientParams.Dob.In(p.timezone)

	patientID, err := q.InsertPatient(ctx, patientParams)
	if err != nil {
		return -1, -1, err
	}

	adminParams, err := toInsertPatientVisitParams(patientID, admin)
	if err != nil {
		return -1, -1, err
	}
	adminParams.RegDate = adminParams.RegDate.In(p.timezone)

	row, err := q.InsertPatientVisit(ctx, adminParams)
	if err != nil {
		return -1, -1, err
	}

	if err = tx.Commit(ctx); err != nil {
		return -1, -1, err
	}
	return patientID, row.Vid, nil
}

// UpdatePatient updates demographic data for a patient.
func (p *PostgresPatientRepository) UpdatePatient(ctx context.Context, id int32, patient *db.PatientDetail) error {
	if patient == nil {
		return entities.ErrMissingPatientData
	}

	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)
	exists, err := checkPatientExists(ctx, q, id)
	if err != nil {
		return err
	}
	if !exists {
		return entities.ErrPatientNotFound
	}

	params := toUpdatePatientParams(id, patient)
	if err := q.UpdatePatient(ctx, params); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeletePatient deletes a patient and all associated visits/data.
func (p *PostgresPatientRepository) DeletePatient(ctx context.Context, id int32) error {
	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)

	exists, err := checkPatientExists(ctx, q, id)
	if err != nil {
		return err
	}
	if !exists {
		return entities.ErrPatientNotFound
	}

	if _, err := tx.Exec(ctx, "DELETE FROM patient_details WHERE id = $1", id); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// CreatePatientVisit inserts a new Admin category for an existing patient and returns the new vid if successful.
func (p *PostgresPatientRepository) CreatePatientVisit(ctx context.Context, id int32, admin *db.Admin) (int32, error) {
	if admin == nil {
		return -1, entities.ErrMissingAdminCategory
	}

	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)
	exists, err := checkPatientExists(ctx, q, id)
	if err != nil {
		return -1, err
	}
	if !exists {
		return -1, entities.ErrPatientNotFound
	}

	params, err := toInsertPatientVisitParams(id, admin)
	if err != nil {
		return -1, err
	}

	params.RegDate = params.RegDate.In(p.timezone)

	row, err := q.InsertPatientVisit(ctx, params)
	if err != nil {
		return -1, err
	}

	if err = tx.Commit(ctx); err != nil {
		return -1, err
	}
	return row.Vid, nil
}

// DeletePatientVisit removes a visit and cascades through related tables.
func (p *PostgresPatientRepository) DeletePatientVisit(ctx context.Context, id int32, vid int32) error {
	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)

	// Delete admin row; ON DELETE CASCADE handles dependent tables.
	if err = runDelete(ctx, q.DeleteAdmin, db.DeleteAdminParams{ID: id, Vid: vid}); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

// UpdatePatientVisit updates a visit for an existing patient, filling out or overriding any of its fields.
func (p *PostgresPatientRepository) UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *entities.Patient) error {
	tx, err := p.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	q := p.queries.WithTx(tx)

	exists, err := checkPatientVisitExists(ctx, q, id, vid)
	if err != nil {
		return err
	}
	if !exists {
		return entities.ErrPatientVisitNotFound
	}

	if patient.Admin != nil {
		params, err := toUpdateAdminParams(id, vid, patient.Admin)
		if err != nil {
			return err
		}
		if err = q.UpdateAdmin(ctx, params); err != nil {
			return err
		}
	}

	if patient.PastMedicalHistory != nil {
		params := toPastMedicalHistoryParams(id, vid, patient.PastMedicalHistory)
		if err = q.UpsertPastMedicalHistory(ctx, params); err != nil {
			return err
		}
	}

	if patient.SocialHistory != nil {
		params, err := toSocialHistoryParams(id, vid, patient.SocialHistory)
		if err != nil {
			return err
		}
		if err = q.UpsertSocialHistory(ctx, params); err != nil {
			return err
		}
	}

	if patient.VitalStatistics != nil {
		params, err := toVitalStatisticsParams(id, vid, patient.VitalStatistics)
		if err != nil {
			return err
		}
		if err = q.UpsertVitalStatistics(ctx, params); err != nil {
			return err
		}
	}

	if patient.HeightAndWeight != nil {
		params, err := toHeightAndWeightParams(id, vid, patient.HeightAndWeight)
		if err != nil {
			return err
		}
		if err = q.UpsertHeightAndWeight(ctx, params); err != nil {
			return err
		}
	}

	if patient.VisualAcuity != nil {
		params, err := toVisualAcuityParams(id, vid, patient.VisualAcuity)
		if err != nil {
			return err
		}
		if err = q.UpsertVisualAcuity(ctx, params); err != nil {
			return err
		}
	}

	if patient.FallRisk != nil {
		params, err := toFallRiskParams(id, vid, patient.FallRisk)
		if err != nil {
			return err
		}
		if err = q.UpsertFallRisk(ctx, params); err != nil {
			return err
		}
	}

	if patient.Dental != nil {
		params, err := toDentalParams(id, vid, patient.Dental)
		if err != nil {
			return err
		}
		if err = q.UpsertDental(ctx, params); err != nil {
			return err
		}
	}

	if patient.Physiotherapy != nil {
		params := toPhysiotherapyParams(id, vid, patient.Physiotherapy)
		if err = q.UpsertPhysiotherapy(ctx, params); err != nil {
			return err
		}
	}

	if patient.DoctorsConsultation != nil {
		params, err := toDoctorsConsultationParams(id, vid, patient.DoctorsConsultation)
		if err != nil {
			return err
		}
		if err = q.UpsertDoctorsConsultation(ctx, params); err != nil {
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (p *PostgresPatientRepository) GetPatientMeta(ctx context.Context, id int32) (*entities.PatientMeta, error) {
	q := p.queries

	exists, err := checkPatientExists(ctx, q, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, entities.ErrPatientNotFound
	}

	patientRow, err := q.GetPatient(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entities.ErrPatientNotFound
	}
	if err != nil {
		return nil, err
	}

	latestRow, err := q.GetLatestAdmin(ctx, id)
	if err != nil {
		return nil, err
	}

	patientMeta := entities.PatientMeta{
		ID:          patientRow.ID,
		Vid:         latestRow.Vid,
		FamilyGroup: patientRow.FamilyGroup,
		RegDate:     latestRow.RegDate,
		QueueNo:     latestRow.QueueNo,
		Name:        patientRow.Name,
		KhmerName:   patientRow.KhmerName,
		Visits:      make(map[int32]time.Time),
	}

	visitRows, err := q.ListAdminVisits(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, row := range visitRows {
		patientMeta.Visits[row.Vid] = row.RegDate
	}

	return &patientMeta, nil
}

func (p *PostgresPatientRepository) GetAllPatientVisitMeta(ctx context.Context, date time.Time) ([]entities.PatientVisitMeta, error) {
	q := p.queries
	if date.IsZero() {
		rows, err := q.GetPatientVisitMetaLatest(ctx)
		if err != nil {
			return nil, err
		}
		result := make([]entities.PatientVisitMeta, 0, len(rows))
		for _, row := range rows {
			result = append(result, toPatientVisitMetaLatest(row))
		}
		return result, nil
	}

	rows, err := q.GetPatientVisitMetaByDate(ctx, date)
	if err != nil {
		return nil, err
	}
	result := make([]entities.PatientVisitMeta, 0, len(rows))
	for _, row := range rows {
		result = append(result, toPatientVisitMetaByDate(row))
	}
	return result, nil
}

// Helpers --------------------------------------------------------------------

func checkPatientExists(ctx context.Context, q *db.Queries, id int32) (bool, error) {
	_, err := q.CheckPatientExists(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func checkPatientVisitExists(ctx context.Context, q *db.Queries, id int32, vid int32) (bool, error) {
	_, err := q.CheckPatientVisitExists(ctx, db.CheckPatientVisitExistsParams{ID: id, Vid: vid})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func runDelete[T any](ctx context.Context, fn func(context.Context, T) error, params T) error {
	return fn(ctx, params)
}

// Parameter builders ---------------------------------------------------------

func toInsertPatientParams(patient *db.PatientDetail) (db.InsertPatientParams, error) {
	return db.InsertPatientParams{
		Name:          patient.Name,
		FamilyGroup:   patient.FamilyGroup,
		KhmerName:     patient.KhmerName,
		Dob:           patient.Dob,
		Gender:        patient.Gender,
		Village:       patient.Village,
		ContactNo:     patient.ContactNo,
		DrugAllergies: patient.DrugAllergies,
	}, nil
}

func toInsertPatientVisitParams(id int32, admin *db.Admin) (db.InsertPatientVisitParams, error) {
	return db.InsertPatientVisitParams{
		ID:                  id,
		RegDate:             admin.RegDate,
		QueueNo:             admin.QueueNo,
		Pregnant:            admin.Pregnant,
		LastMenstrualPeriod: admin.LastMenstrualPeriod,
		SentToID:            admin.SentToID,
	}, nil
}

func toUpdatePatientParams(id int32, patient *db.PatientDetail) db.UpdatePatientParams {
	return db.UpdatePatientParams{
		Name:          patient.Name,
		FamilyGroup:   patient.FamilyGroup,
		KhmerName:     patient.KhmerName,
		Dob:           patient.Dob,
		Gender:        patient.Gender,
		Village:       patient.Village,
		ContactNo:     patient.ContactNo,
		DrugAllergies: patient.DrugAllergies,
		ID:            id,
	}
}

func toUpdateAdminParams(id int32, vid int32, admin *db.Admin) (db.UpdateAdminParams, error) {
	return db.UpdateAdminParams{
		RegDate:             admin.RegDate,
		QueueNo:             admin.QueueNo,
		Pregnant:            admin.Pregnant,
		LastMenstrualPeriod: admin.LastMenstrualPeriod,
		SentToID:            admin.SentToID,
		ID:                  id,
		Vid:                 vid,
	}, nil
}

func toPastMedicalHistoryParams(id int32, vid int32, pmh *db.PastMedicalHistory) db.UpsertPastMedicalHistoryParams {
	return db.UpsertPastMedicalHistoryParams{
		ID:                         id,
		Vid:                        vid,
		Cough:                      pmh.Cough,
		Fever:                      pmh.Fever,
		BlockedNose:                pmh.BlockedNose,
		SoreThroat:                 pmh.SoreThroat,
		NightSweats:                pmh.NightSweats,
		UnintentionalWeightLoss:    pmh.UnintentionalWeightLoss,
		Tuberculosis:               pmh.Tuberculosis,
		TuberculosisHasBeenTreated: pmh.TuberculosisHasBeenTreated,
		Diabetes:                   pmh.Diabetes,
		Hypertension:               pmh.Hypertension,
		Hyperlipidemia:             pmh.Hyperlipidemia,
		ChronicJointPains:          pmh.ChronicJointPains,
		ChronicMuscleAches:         pmh.ChronicMuscleAches,
		SexuallyTransmittedDisease: pmh.SexuallyTransmittedDisease,
		SpecifiedStds:              pmh.SpecifiedStds,
		Others:                     pmh.Others,
	}
}

func toSocialHistoryParams(id int32, vid int32, sh *db.SocialHistory) (db.UpsertSocialHistoryParams, error) {
	return db.UpsertSocialHistoryParams{
		ID:                    id,
		Vid:                   vid,
		PastSmokingHistory:    sh.PastSmokingHistory,
		NoOfYears:             sh.NoOfYears,
		CurrentSmokingHistory: sh.CurrentSmokingHistory,
		CigarettesPerDay:      sh.CigarettesPerDay,
		AlcoholHistory:        sh.AlcoholHistory,
		HowRegular:            sh.HowRegular,
	}, nil
}

func toVitalStatisticsParams(id int32, vid int32, vs *db.VitalStatistic) (db.UpsertVitalStatisticsParams, error) {
	return db.UpsertVitalStatisticsParams{
		ID:                    id,
		Vid:                   vid,
		Temperature:           vs.Temperature,
		Spo2:                  vs.Spo2,
		SystolicBp1:           vs.SystolicBp1,
		DiastolicBp1:          vs.DiastolicBp1,
		SystolicBp2:           vs.SystolicBp2,
		DiastolicBp2:          vs.DiastolicBp2,
		AvgSystolicBp:         vs.AvgSystolicBp,
		AvgDiastolicBp:        vs.AvgDiastolicBp,
		Hr1:                   vs.Hr1,
		Hr2:                   vs.Hr2,
		AvgHr:                 vs.AvgHr,
		RandBloodGlucoseMmolL: vs.RandBloodGlucoseMmolL,
		IcopeHighBp:           vs.IcopeHighBp,
	}, nil
}

func toHeightAndWeightParams(id int32, vid int32, haw *db.HeightAndWeight) (db.UpsertHeightAndWeightParams, error) {
	return db.UpsertHeightAndWeightParams{
		ID:                        id,
		Vid:                       vid,
		Height:                    haw.Height,
		Weight:                    haw.Weight,
		Bmi:                       haw.Bmi,
		BmiAnalysis:               haw.BmiAnalysis,
		PaedsHeight:               haw.PaedsHeight,
		PaedsWeight:               haw.PaedsWeight,
		IcopeLostWeightPastMonths: haw.IcopeLostWeightPastMonths,
		IcopeNoDesireToEat:        haw.IcopeNoDesireToEat,
	}, nil
}

func toVisualAcuityParams(id int32, vid int32, va *db.VisualAcuity) (db.UpsertVisualAcuityParams, error) {
	return db.UpsertVisualAcuityParams{
		ID:                          id,
		Vid:                         vid,
		LEyeVision:                  va.LEyeVision,
		REyeVision:                  va.REyeVision,
		AdditionalIntervention:      va.AdditionalIntervention,
		SentToOpto:                  va.SentToOpto,
		ReferredForGlasses:          va.ReferredForGlasses,
		IcopeEyeProblem:             va.IcopeEyeProblem,
		IcopeTreatedForDiabetesOrBp: va.IcopeTreatedForDiabetesOrBp,
	}, nil
}

func toFallRiskParams(id int32, vid int32, fr *db.FallRisk) (db.UpsertFallRiskParams, error) {
	return db.UpsertFallRiskParams{
		ID:                       id,
		Vid:                      vid,
		SideToSideBalance:        fr.SideToSideBalance,
		SemiTandemBalance:        fr.SemiTandemBalance,
		TandemBalance:            fr.TandemBalance,
		GaitSpeedTest:            fr.GaitSpeedTest,
		ChairStandTest:           fr.ChairStandTest,
		FallRiskScore:            fr.FallRiskScore,
		IcopeCompleteChairStands: fr.IcopeCompleteChairStands,
		IcopeChairStandsTime:     fr.IcopeChairStandsTime,
	}, nil
}

func toDentalParams(id int32, vid int32, d *db.Dental) (db.UpsertDentalParams, error) {
	return db.UpsertDentalParams{
		ID:                     id,
		Vid:                    vid,
		FluorideExposure:       d.FluorideExposure,
		Diet:                   d.Diet,
		BacterialExposure:      d.BacterialExposure,
		OralSymptoms:           d.OralSymptoms,
		DrinkOtherWater:        d.DrinkOtherWater,
		RiskForDentalCarries:   d.RiskForDentalCarries,
		IcopeDifficultyChewing: d.IcopeDifficultyChewing,
		IcopePainInMouth:       d.IcopePainInMouth,
		DentalNotes:            d.DentalNotes,
	}, nil
}

func toPhysiotherapyParams(id int32, vid int32, phy *db.Physiotherapy) db.UpsertPhysiotherapyParams {
	return db.UpsertPhysiotherapyParams{
		ID:                   id,
		Vid:                  vid,
		SubjectiveAssessment: phy.SubjectiveAssessment,
		PainScale:            phy.PainScale,
		ObjectiveAssessment:  phy.ObjectiveAssessment,
		Intervention:         phy.Intervention,
		Evaluation:           phy.Evaluation,
	}
}

func toDoctorsConsultationParams(id int32, vid int32, dc *db.DoctorsConsultation) (db.UpsertDoctorsConsultationParams, error) {
	return db.UpsertDoctorsConsultationParams{
		ID:                id,
		Vid:               vid,
		Well:              dc.Well,
		Msk:               dc.Msk,
		Cvs:               dc.Cvs,
		Respi:             dc.Respi,
		Gu:                dc.Gu,
		Git:               dc.Git,
		Eye:               dc.Eye,
		Derm:              dc.Derm,
		Others:            dc.Others,
		ConsultationNotes: dc.ConsultationNotes,
		Diagnosis:         dc.Diagnosis,
		Treatment:         dc.Treatment,
		ReferralNeeded:    dc.ReferralNeeded,
		ReferralLoc:       dc.ReferralLoc,
		Remarks:           dc.Remarks,
	}, nil
}

// Entity builders ------------------------------------------------------------

func toPatientVisitMetaLatest(row db.GetPatientVisitMetaLatestRow) entities.PatientVisitMeta {
	return buildPatientVisitMeta(
		row.ID,
		row.Vid,
		row.FamilyGroup,
		row.RegDate,
		row.QueueNo,
		row.Name,
		row.KhmerName,
		row.Gender,
		row.Village,
		row.ContactNo,
		row.DrugAllergies,
		row.SentToID,
		row.ReferralNeeded,
		row.HasPrescriptionWithDrug,
		row.AllPrescriptionDrugsPacked,
		row.PrescriptionDispensed,
	)
}

func toPatientVisitMetaByDate(row db.GetPatientVisitMetaByDateRow) entities.PatientVisitMeta {
	return buildPatientVisitMeta(
		row.ID,
		row.Vid,
		row.FamilyGroup,
		row.RegDate,
		row.QueueNo,
		row.Name,
		row.KhmerName,
		row.Gender,
		row.Village,
		row.ContactNo,
		row.DrugAllergies,
		row.SentToID,
		row.ReferralNeeded,
		row.HasPrescriptionWithDrug,
		row.AllPrescriptionDrugsPacked,
		row.PrescriptionDispensed,
	)
}

func buildPatientVisitMeta(
	id int32,
	vid int32,
	familyGroup string,
	regDate time.Time,
	queueNo string,
	name string,
	khmerName string,
	gender string,
	village string,
	contactNo string,
	drugAllergies *string,
	sentToID bool,
	referralNeeded *bool,
	hasPrescription bool,
	allPrescriptionPacked bool,
	prescriptionDispensed bool,
) entities.PatientVisitMeta {
	return entities.PatientVisitMeta{
		ID:                         id,
		Vid:                        vid,
		FamilyGroup:                familyGroup,
		RegDate:                    regDate,
		QueueNo:                    queueNo,
		Name:                       name,
		KhmerName:                  khmerName,
		Gender:                     gender,
		Village:                    village,
		ContactNo:                  contactNo,
		DrugAllergies:              drugAllergies,
		SentToID:                   sentToID,
		ReferralNeeded:             referralNeeded,
		HasPrescriptionWithDrug:    hasPrescription,
		AllPrescriptionDrugsPacked: allPrescriptionPacked,
		PrescriptionDispensed:      prescriptionDispensed,
	}
}
