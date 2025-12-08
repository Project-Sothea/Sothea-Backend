package usecases

import (
	"context"
	"time"

	"sothea-backend/entities"
	db "sothea-backend/repository/sqlc"
)

type patientUsecase struct {
	patientRepo    entities.PatientRepository
	contextTimeout time.Duration
}

// NewPatientUseCase
func NewPatientUsecase(p entities.PatientRepository, timeout time.Duration) entities.PatientUseCase {
	return &patientUsecase{
		patientRepo:    p,
		contextTimeout: timeout,
	}
}

func (p *patientUsecase) GetPatientVisit(ctx context.Context, id int32, vid int32) (*entities.Patient, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.GetPatientVisit(ctx, id, vid)
}

func (p *patientUsecase) CreatePatient(ctx context.Context, patient *db.PatientDetail) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.CreatePatient(ctx, patient)
}

func (p *patientUsecase) UpdatePatient(ctx context.Context, id int32, patient *db.PatientDetail) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.UpdatePatient(ctx, id, patient)
}

func (p *patientUsecase) DeletePatient(ctx context.Context, id int32) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.DeletePatient(ctx, id)
}

func (p *patientUsecase) CreatePatientVisit(ctx context.Context, id int32, admin *db.Admin) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.CreatePatientVisit(ctx, id, admin)
}

func (p *patientUsecase) DeletePatientVisit(ctx context.Context, id int32, vid int32) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.DeletePatientVisit(ctx, id, vid)
}

func (p *patientUsecase) UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *entities.Patient) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.UpdatePatientVisit(ctx, id, vid, patient)
}

func (p *patientUsecase) GetPatientMeta(ctx context.Context, id int32) (*entities.PatientMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.GetPatientMeta(ctx, id)
}

func (p *patientUsecase) GetAllPatientVisitMeta(ctx context.Context, date time.Time) ([]entities.PatientVisitMeta, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.GetAllPatientVisitMeta(ctx, date)
}
