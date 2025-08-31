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

	batcheDetails, err := u.repo.ListBatchDetails(ctx, &id)
	if err != nil {
		return nil, err
	}

	return &entities.DrugDetail{Drug: *drug, Batches: batcheDetails}, nil
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

func (u *pharmacyUsecase) ListBatches(ctx context.Context, drugID *int64) ([]entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.ListBatchDetails(ctx, drugID)
}

func (u *pharmacyUsecase) CreateBatch(ctx context.Context, batchDetail *entities.BatchDetail) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.CreateBatch(ctx, batchDetail)
}

func (u *pharmacyUsecase) UpdateBatch(ctx context.Context, b *entities.DrugBatch) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatch(ctx, b)
}

func (u *pharmacyUsecase) DeleteBatch(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatch(ctx, id)
}

// -----------------------------------------------------------------------------
//  BatchLocation-level methods
// -----------------------------------------------------------------------------

func (u *pharmacyUsecase) CreateBatchLocation(ctx context.Context, batchLocation *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.CreateBatchLocation(ctx, batchLocation)
}

func (u *pharmacyUsecase) UpdateBatchLocation(ctx context.Context, batchLocation *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatchLocation(ctx, batchLocation)
}

func (u *pharmacyUsecase) DeleteBatchLocation(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatchLocation(ctx, id)
}
