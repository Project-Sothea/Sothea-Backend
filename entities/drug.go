// entities/drug.go
package entities

import (
	"context"
	"time"
)

// Represents the "type" of drug (e.g. Benadryl)
type Drug struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Unit        string  `json:"unit"`
	DefaultSize *int    `json:"defaultSize,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// Represents each batch of the drug
type DrugBatch struct {
	ID          int64     `json:"id"`
	DrugID      int64     `json:"drugId"`
	BatchNumber string    `json:"batchNumber"`
	ExpiryDate  time.Time `json:"expiryDate"`
	Notes       *string   `json:"notes,omitempty"`
	Supplier    *string   `json:"supplier,omitempty"`
}

// Represents each batch of the drug
type DrugBatchLocation struct {
	ID       int64  `json:"id"`
	BatchID  int64  `json:"batchId"`
	Location string `json:"location"`
	Quantity int64  `json:"quantity"`
}

type BatchDetail struct {
	DrugBatch
	BatchLocations []DrugBatchLocation `json:"batchLocations"`
}

type DrugDetail struct {
	Drug
	Batches []BatchDetail `json:"batches"`
}

type PharmacyRepository interface {
	ListDrugs(ctx context.Context) ([]Drug, error)
	CreateDrug(ctx context.Context, d *Drug) (*Drug, error)
	GetDrug(ctx context.Context, id int64) (*Drug, error)
	UpdateDrug(ctx context.Context, d *Drug) (*Drug, error)
	DeleteDrug(ctx context.Context, id int64) error

	ListBatchDetails(ctx context.Context, drugID *int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, id int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *BatchDetail) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *DrugBatch) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, id int64) error

	ListBatchLocations(ctx context.Context, batchID int64) ([]DrugBatchLocation, error)
	GetBatchLocation(ctx context.Context, id int64) (*DrugBatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error
}

type PharmacyUseCase interface {
	// Drug-level
	ListDrugs(ctx context.Context) ([]Drug, error)
	CreateDrug(ctx context.Context, d *Drug) (*Drug, error)
	GetDrug(ctx context.Context, id int64) (*DrugDetail, error)
	UpdateDrug(ctx context.Context, d *Drug) (*Drug, error)
	DeleteDrug(ctx context.Context, id int64) error

	// Batch-level
	ListBatches(ctx context.Context, drugID *int64) ([]BatchDetail, error) // set drugID = nil for all batches
	CreateBatch(ctx context.Context, b *BatchDetail) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *DrugBatch) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, id int64) error

	// BatchLocation CRUD (independent)
	CreateBatchLocation(ctx context.Context, batchLocation *DrugBatchLocation) (*DrugBatchLocation, error)
	UpdateBatchLocation(ctx context.Context, batchLocation *DrugBatchLocation) (*DrugBatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error
}
