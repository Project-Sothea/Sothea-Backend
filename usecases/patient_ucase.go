package usecases

import (
	"context"
	"time"

	"sothea-backend/entities"
	"sothea-backend/repository/postgres"
	db "sothea-backend/repository/sqlc"
)

type PatientUsecase struct {
	patientRepo    *postgres.PostgresPatientRepository
	contextTimeout time.Duration
}

// NewPatientUsecase constructs the patient usecase with timeout.
func NewPatientUsecase(p *postgres.PostgresPatientRepository, timeout time.Duration) *PatientUsecase {
	return &PatientUsecase{
		patientRepo:    p,
		contextTimeout: timeout,
	}
}

func (p *PatientUsecase) GetPatientVisit(ctx context.Context, id int32, vid int32) (*entities.Patient, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.GetPatientVisit(ctx, id, vid)
}

func (p *PatientUsecase) CreatePatient(ctx context.Context, patient *db.PatientDetail) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.CreatePatient(ctx, patient)
}

func (p *PatientUsecase) CreatePatientWithVisit(ctx context.Context, patient *db.PatientDetail, admin *db.Admin) (int32, int32, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.CreatePatientWithVisit(ctx, patient, admin)
}

func (p *PatientUsecase) UpdatePatient(ctx context.Context, id int32, patient *db.PatientDetail) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.UpdatePatient(ctx, id, patient)
}

func (p *PatientUsecase) DeletePatient(ctx context.Context, id int32) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.DeletePatient(ctx, id)
}

func (p *PatientUsecase) CreatePatientVisit(ctx context.Context, id int32, admin *db.Admin) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.CreatePatientVisit(ctx, id, admin)
}

func (p *PatientUsecase) DeletePatientVisit(ctx context.Context, id int32, vid int32) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.DeletePatientVisit(ctx, id, vid)
}

func (p *PatientUsecase) UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *entities.Patient) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.UpdatePatientVisit(ctx, id, vid, patient)
}

func (p *PatientUsecase) GetPatientMeta(ctx context.Context, id int32) (*entities.PatientMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.GetPatientMeta(ctx, id)
}

func (p *PatientUsecase) GetAllPatientVisitMeta(ctx context.Context, date time.Time) ([]entities.PatientVisitMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.GetAllPatientVisitMeta(ctx, date)
}
