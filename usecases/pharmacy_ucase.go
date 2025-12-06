package usecases

import (
	"context"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
// Struct + Constructor
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
// Helpers (formatting & small validations)
// -----------------------------------------------------------------------------

func valOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
func valOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// -----------------------------------------------------------------------------
// PharmacyUseCase implementation
// -----------------------------------------------------------------------------

// ----------------- DRUGS -----------------

func (u *pharmacyUsecase) ListDrugs(ctx context.Context, q *string) ([]entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListDrugs(ctx, q)
}

func (u *pharmacyUsecase) GetDrug(ctx context.Context, id int64) (*entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetDrug(ctx, id)
}

func (u *pharmacyUsecase) GetDrugStock(ctx context.Context, drugID int64) (*entities.DrugStock, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetDrugStock(ctx, drugID)
}

func (u *pharmacyUsecase) CreateDrug(ctx context.Context, d *entities.Drug) (*entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateDrug(ctx, d)
}

func (u *pharmacyUsecase) UpdateDrug(ctx context.Context, d *entities.Drug) (*entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateDrug(ctx, d)
}

func (u *pharmacyUsecase) DeleteDrug(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteDrug(ctx, id)
}

// ----------------- BATCHES -----------------

func (u *pharmacyUsecase) ListBatches(ctx context.Context, drugID int64) ([]entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListBatches(ctx, drugID)
}

func (u *pharmacyUsecase) GetBatch(ctx context.Context, batchID int64) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetBatch(ctx, batchID)
}

func (u *pharmacyUsecase) CreateBatch(ctx context.Context, b *entities.DrugBatch, locations []entities.DrugBatchLocation) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateBatch(ctx, b, locations)
}

func (u *pharmacyUsecase) UpdateBatch(ctx context.Context, b *entities.DrugBatch, locations []entities.DrugBatchLocation) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatch(ctx, b, locations)
}

func (u *pharmacyUsecase) DeleteBatch(ctx context.Context, batchID int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatch(ctx, batchID)
}

// ------------- BATCH LOCATIONS -------------

func (u *pharmacyUsecase) ListBatchLocations(ctx context.Context, batchID int64) ([]entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListBatchLocations(ctx, batchID)
}

func (u *pharmacyUsecase) CreateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateBatchLocation(ctx, loc)
}

func (u *pharmacyUsecase) UpdateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatchLocation(ctx, loc)
}

func (u *pharmacyUsecase) DeleteBatchLocation(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatchLocation(ctx, id)
}
