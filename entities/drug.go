// entities/drug.go
package entities

import (
	"context"
	"time"
)

// Drug
type Drug struct {
	ID          int64   `json:"id"`
	GenericName string  `json:"genericName"`
	BrandName   *string `json:"brandName,omitempty"`
	DrugCode    *int64  `json:"drugCode,omitempty"`

	DosageFormCode string `json:"dosageFormCode"` // e.g. "TAB","SYR","CREAM"
	RouteCode      string `json:"routeCode"`      // e.g. "PO","TOP"

	// Strength / concentration
	StrengthNum     *float64 `json:"strengthNum,omitempty"`     // e.g. 500.0 (supports 1 decimal place)
	StrengthUnitNum *string  `json:"strengthUnitNum,omitempty"` // e.g. "mg"
	StrengthDen     *float64 `json:"strengthDen,omitempty"`     // e.g. 5.0 (mL) (supports 1 decimal place)
	StrengthUnitDen *string  `json:"strengthUnitDen,omitempty"` // e.g. "mL"

	// Inventory base unit (what we actually deduct)
	DispenseUnit string `json:"dispenseUnit"` // "tab","mL","g","bottle"

	// Only for piece-dispensed items (e.g., bottles, tubes, inhalers)
	PieceContentAmount *float64 `json:"pieceContentAmount,omitempty"` // e.g. 100.0 (mL per bottle, g per tube, puffs per inhaler) (supports 1 decimal place)
	PieceContentUnit   *string  `json:"pieceContentUnit,omitempty"`   // e.g. "mL", "g", "puff"

	IsFractionalAllowed bool `json:"isFractionalAllowed"`

	DisplayAsPercentage bool `json:"displayAsPercentage"` // If true, show concentration as % (e.g., 1% instead of 1 g/100 g)

	Barcode   *string   `json:"barcode,omitempty"`
	Notes     *string   `json:"notes,omitempty"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Convenience for dropdowns / labels (computed server-side)
type DrugView struct {
	Drug
	DisplayStrength string `json:"displayStrength"` // e.g. "500 mg tab" or "250 mg/5 mL syrup"
	DisplayRoute    string `json:"displayRoute"`    // e.g. "PO"
	DisplayLabel    string `json:"displayLabel"`    // e.g. "Paracetamol 500 mg tablet (PO)"
}

// Batch = one lot for a drug. Quantity is in the drug's DispenseUnit.
type DrugBatch struct {
	ID          int64      `json:"id"`
	DrugID      int64      `json:"drugId"`
	BatchNumber string     `json:"batchNumber"`
	ExpiryDate  *time.Time `json:"expiryDate,omitempty"`
	Supplier    *string    `json:"supplier,omitempty"`
	Quantity    int        `json:"quantity"` // current on-hand in DispenseUnit
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type DrugBatchLocation struct {
	ID        int64     `json:"id"`
	BatchID   int64     `json:"batchId"`
	Location  string    `json:"location"`
	Quantity  int       `json:"quantity"` // in DispenseUnit
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Useful for FEFO picking UI
type BatchDetail struct {
	DrugBatch
	DispenseUnit   string              `json:"dispenseUnit"`  // from drug
	ExpirySortKey  *time.Time          `json:"expirySortKey"` // for FEFO
	BatchLocations []DrugBatchLocation `json:"batchLocations"`
}

type DrugStock struct {
	Drug     DrugView      `json:"drug"`
	Batches  []BatchDetail `json:"batches"`
	TotalQty int           `json:"totalQty"` // sum of batch quantities
}

type PharmacyRepository interface {
	// DRUGS (combined with presentations)
	ListDrugs(ctx context.Context, q *string) ([]DrugView, error) // optional search
	CreateDrug(ctx context.Context, d *Drug) (*DrugView, error)
	GetDrug(ctx context.Context, id int64) (*DrugView, error)
	UpdateDrug(ctx context.Context, d *Drug) (*DrugView, error)
	DeleteDrug(ctx context.Context, id int64) error

	// STOCK (batches & locations) — quantities are in the drug's DispenseUnit
	ListBatches(ctx context.Context, drugID int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, batchID int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *DrugBatch, locations []DrugBatchLocation) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *DrugBatch, locations []DrugBatchLocation) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, batchID int64) error

	ListBatchLocations(ctx context.Context, batchID int64) ([]DrugBatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error

	// Convenience for FEFO view (drug + batches + totals)
	GetDrugStock(ctx context.Context, drugID int64) (*DrugStock, error)
}

type PharmacyUseCase interface {
	// wrappers that also build DisplayLabel/DisplayStrength, totals, etc.
	ListDrugs(ctx context.Context, q *string) ([]DrugView, error)
	GetDrug(ctx context.Context, id int64) (*DrugView, error)
	GetDrugStock(ctx context.Context, drugID int64) (*DrugStock, error)

	// Admin flows
	CreateDrug(ctx context.Context, d *Drug) (*DrugView, error)
	UpdateDrug(ctx context.Context, d *Drug) (*DrugView, error)
	DeleteDrug(ctx context.Context, id int64) error

	ListBatches(ctx context.Context, drugID int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, batchID int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *DrugBatch, locations []DrugBatchLocation) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *DrugBatch, locations []DrugBatchLocation) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, batchID int64) error

	ListBatchLocations(ctx context.Context, batchID int64) ([]DrugBatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error
}
