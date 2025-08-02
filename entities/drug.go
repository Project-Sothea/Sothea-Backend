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
	DefaultSize *int    `json:"default_size,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// Represents each batch of the drug
type DrugBatch struct {
	ID          int64      `json:"id"`
	DrugID      int64      `json:"drug_id"`
	BatchNumber string     `json:"batch_no"`
	Location    *string    `json:"location,omitempty"`
	Quantity    int        `json:"quantity"`
	ExpiryDate  time.Time  `json:"expiry_date"`
	Supplier    *string    `json:"supplier,omitempty"`
	DepletedAt  *time.Time `json:"depleted_at,omitempty"`
}

type DrugDetail struct {
	Drug
	Batches []DrugBatch `json:"batches"`
}

type PharmacyRepository interface {
	ListDrugs(ctx context.Context) ([]Drug, error)
	CreateDrug(ctx context.Context, d *Drug) (*Drug, error)
	GetDrug(ctx context.Context, id int64) (*Drug, error)
	UpdateDrug(ctx context.Context, d *Drug) (*Drug, error)
	DeleteDrug(ctx context.Context, id int64) error

	ListBatches(ctx context.Context, drugID *int64) ([]DrugBatch, error) // set drugID = nil for all batches
	CreateBatch(ctx context.Context, b *DrugBatch) (int64, error)
	UpdateBatch(ctx context.Context, b *DrugBatch) error
	DeleteBatch(ctx context.Context, id int64) error
}

type PharmacyUseCase interface {
	// Drug-level
	ListDrugs(ctx context.Context) ([]Drug, error)
	CreateDrug(ctx context.Context, d *Drug) (*Drug, error)
	GetDrug(ctx context.Context, id int64) (*DrugDetail, error)
	UpdateDrug(ctx context.Context, d *Drug) (*Drug, error)
	DeleteDrug(ctx context.Context, id int64) error

	// Batch-level
	ListBatches(ctx context.Context, drugID *int64) ([]DrugBatch, error) // set drugID = nil for all batches
	CreateBatch(ctx context.Context, b *DrugBatch) (int64, error)
	UpdateBatch(ctx context.Context, b *DrugBatch) error
	DeleteBatch(ctx context.Context, id int64) error
}
