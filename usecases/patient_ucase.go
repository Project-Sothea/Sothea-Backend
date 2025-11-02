package usecases

import (
	"context"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
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

func (p *patientUsecase) CreatePatient(ctx context.Context, admin *entities.Admin) (int32, error) {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.CreatePatient(ctx, admin)
}

func (p *patientUsecase) CreatePatientVisit(ctx context.Context, id int32, admin *entities.Admin) (int32, error) {
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

func (p *patientUsecase) ExportDatabaseToCSV(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.contextTimeout)
	defer cancel()

	return p.patientRepo.ExportDatabaseToCSV(ctx)
}
