package usecases

import (
	"context"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
// Struct + Constructor
// -----------------------------------------------------------------------------

type prescriptionUsecase struct {
	repo           entities.PrescriptionRepository
	contextTimeout time.Duration
}

func NewPrescriptionUsecase(r entities.PrescriptionRepository, timeout time.Duration) entities.PrescriptionUseCase {
	return &prescriptionUsecase{
		repo:           r,
		contextTimeout: timeout,
	}
}

// -----------------------------------------------------------------------------
// Core Methods
// -----------------------------------------------------------------------------

func (u *prescriptionUsecase) CreatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreatePrescription(ctx, p)
}

func (u *prescriptionUsecase) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetPrescriptionByID(ctx, id)
}

func (u *prescriptionUsecase) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListPrescriptions(ctx, patientID, vid)
}

func (u *prescriptionUsecase) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdatePrescription(ctx, p)
}

func (u *prescriptionUsecase) DeletePrescription(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeletePrescription(ctx, id)
}
