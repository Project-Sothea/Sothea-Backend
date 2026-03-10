package usecases

import (
	"context"
	"time"

	"sothea-backend/entities"
	"sothea-backend/repository/postgres"
	db "sothea-backend/repository/sqlc"
)

type PharmacyUsecase struct {
	repo           *postgres.PostgresPharmacyRepository
	contextTimeout time.Duration
}

func NewPharmacyUsecase(r *postgres.PostgresPharmacyRepository, timeout time.Duration) *PharmacyUsecase {
	return &PharmacyUsecase{
		repo:           r,
		contextTimeout: timeout,
	}
}

func (u *PharmacyUsecase) ListDrugs(ctx context.Context, q *string) ([]entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListDrugs(ctx, q)
}

func (u *PharmacyUsecase) GetDrug(ctx context.Context, id int64) (*entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetDrug(ctx, id)
}

func (u *PharmacyUsecase) GetDrugStock(ctx context.Context, drugID int64) (*entities.DrugStock, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetDrugStock(ctx, drugID)
}

func (u *PharmacyUsecase) CreateDrug(ctx context.Context, d *db.Drug) (*entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateDrug(ctx, d)
}

func (u *PharmacyUsecase) UpdateDrug(ctx context.Context, d *db.Drug) (*entities.DrugView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateDrug(ctx, d)
}

func (u *PharmacyUsecase) DeleteDrug(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteDrug(ctx, id)
}

func (u *PharmacyUsecase) ListBatches(ctx context.Context, drugID int64) ([]entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListBatches(ctx, drugID)
}

func (u *PharmacyUsecase) GetBatch(ctx context.Context, batchID int64) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetBatch(ctx, batchID)
}

func (u *PharmacyUsecase) CreateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateBatch(ctx, b, locations)
}

func (u *PharmacyUsecase) UpdateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatch(ctx, b, locations)
}

func (u *PharmacyUsecase) DeleteBatch(ctx context.Context, batchID int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatch(ctx, batchID)
}

func (u *PharmacyUsecase) ListBatchLocations(ctx context.Context, batchID int64) ([]db.BatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListBatchLocations(ctx, batchID)
}

func (u *PharmacyUsecase) CreateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateBatchLocation(ctx, loc)
}

func (u *PharmacyUsecase) UpdateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatchLocation(ctx, loc)
}

func (u *PharmacyUsecase) DeleteBatchLocation(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatchLocation(ctx, id)
}
