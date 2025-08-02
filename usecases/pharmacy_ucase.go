package usecases

import (
	"context"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
//  Struct + Constructor
// -----------------------------------------------------------------------------

type pharmacyUsecase struct {
	repo           entities.PharmacyRepository
	contextTimeout time.Duration
}

func NewPharmacyUsecase(r entities.PharmacyRepository, timeout time.Duration) entities.PharmacyUseCase {
	return &pharmacyUsecase{
		repo:           r,
		contextTimeout: timeout,
	}
}

// -----------------------------------------------------------------------------
//  Drug-level methods
// -----------------------------------------------------------------------------

func (u *pharmacyUsecase) ListDrugs(ctx context.Context) ([]entities.Drug, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.ListDrugs(ctx)
}

func (u *pharmacyUsecase) CreateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.CreateDrug(ctx, d)
}

func (u *pharmacyUsecase) GetDrug(ctx context.Context, id int64) (*entities.DrugDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	drug, err := u.repo.GetDrug(ctx, id)
	if err != nil {
		return nil, err
	}

	batches, err := u.repo.ListBatches(ctx, &id)
	if err != nil {
		return nil, err
	}

	return &entities.DrugDetail{Drug: *drug, Batches: batches}, nil
}

func (u *pharmacyUsecase) UpdateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateDrug(ctx, d)
}

func (u *pharmacyUsecase) DeleteDrug(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteDrug(ctx, id)
}

// -----------------------------------------------------------------------------
//  Batch-level methods
// -----------------------------------------------------------------------------

func (u *pharmacyUsecase) ListBatches(ctx context.Context, drugID *int64) ([]entities.DrugBatch, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.ListBatches(ctx, drugID)
}

func (u *pharmacyUsecase) CreateBatch(ctx context.Context, b *entities.DrugBatch) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.CreateBatch(ctx, b)
}

func (u *pharmacyUsecase) UpdateBatch(ctx context.Context, b *entities.DrugBatch) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatch(ctx, b)
}

func (u *pharmacyUsecase) DeleteBatch(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatch(ctx, id)
}
