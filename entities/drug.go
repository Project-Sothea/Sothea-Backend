// entities/drug.go
package entities

import (
	"context"
	"time"
)

// High-level identity (no strength/units here)
type Drug struct {
	ID          int64     `json:"id"`
	GenericName string    `json:"genericName"`
	BrandName   *string   `json:"brandName,omitempty"`
	ATCCode     *string   `json:"atcCode,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// How it’s supplied/dispensed. One drug can have many presentations.
type DrugPresentation struct {
	ID             int64  `json:"id"`
	DrugID         int64  `json:"drugId"`
	DosageFormCode string `json:"dosageFormCode"` // e.g. "TAB","SYR","CREAM"
	RouteCode      string `json:"routeCode"`      // e.g. "PO","TOP"

	// Strength / concentration
	StrengthNum     *int    `json:"strengthNum,omitempty"`     // e.g. 500
	StrengthUnitNum *string `json:"strengthUnitNum,omitempty"` // e.g. "mg"
	StrengthDen     *int    `json:"strengthDen,omitempty"`     // e.g. 5 (mL)
	StrengthUnitDen *string `json:"strengthUnitDen,omitempty"` // e.g. "mL"

	// Inventory base unit (what we actually deduct)
	DispenseUnit string `json:"dispenseUnit"` // "tab","mL","g","bottle"

	// Only for piece-dispensed liquids/creams (e.g., bottles/tubes)
	PieceContentAmount *int    `json:"pieceContentAmount,omitempty"` // e.g. 100 (mL per bottle)
	PieceContentUnit   *string `json:"pieceContentUnit,omitempty"`   // e.g. "mL"

	IsFractionalAllowed bool `json:"isFractionalAllowed"`

	Barcode   *string   `json:"barcode,omitempty"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Convenience for dropdowns / labels (computed server-side)
type DrugPresentationView struct {
	DrugPresentation
	DrugName        string `json:"drugName"`        // e.g. "Paracetamol"
	DisplayStrength string `json:"displayStrength"` // e.g. "500 mg tab" or "250 mg/5 mL syrup"
	DisplayRoute    string `json:"displayRoute"`    // e.g. "PO"
	DisplayLabel    string `json:"displayLabel"`    // e.g. "Paracetamol 500 mg tablet (PO)"
}

type DrugWithPresentations struct {
	Drug          Drug                   `json:"drug"`
	Presentations []DrugPresentationView `json:"presentations"`
}

// Batch = one lot for a presentation. Quantity is in the presentation’s DispenseUnit.
type DrugBatch struct {
	ID             int64      `json:"id"`
	PresentationID int64      `json:"presentationId"`
	BatchNumber    string     `json:"batchNumber"`
	ExpiryDate     *time.Time `json:"expiryDate,omitempty"`
	Supplier       *string    `json:"supplier,omitempty"`
	Quantity       int        `json:"quantity"` // current on-hand in DispenseUnit
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
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
	DispenseUnit   string              `json:"dispenseUnit"`  // from presentation
	ExpirySortKey  *time.Time          `json:"expirySortKey"` // for FEFO
	BatchLocations []DrugBatchLocation `json:"batchLocations"`
}

type PresentationStock struct {
	Presentation DrugPresentationView `json:"presentation"`
	Batches      []BatchDetail        `json:"batches"`
	TotalQty     int                  `json:"totalQty"` // sum of batch quantities
}

type PharmacyRepository interface {
	// DRUGS
	ListDrugs(ctx context.Context, q *string) ([]Drug, error) // optional search
	CreateDrug(ctx context.Context, d *Drug) (*Drug, error)
	GetDrug(ctx context.Context, id int64) (*Drug, error)
	UpdateDrug(ctx context.Context, d *Drug) (*Drug, error)
	DeleteDrug(ctx context.Context, id int64) error

	// PRESENTATIONS
	ListPresentations(ctx context.Context, drugID int64) ([]DrugPresentationView, error)
	GetPresentation(ctx context.Context, id int64) (*DrugPresentationView, error)
	CreatePresentation(ctx context.Context, p *DrugPresentation) (*DrugPresentationView, error)
	UpdatePresentation(ctx context.Context, p *DrugPresentation) (*DrugPresentationView, error)
	DeletePresentation(ctx context.Context, id int64) error

	// STOCK (batches & locations) — quantities are in the presentation’s DispenseUnit
	ListBatches(ctx context.Context, presentationID int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, batchID int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *DrugBatch, locations []DrugBatchLocation) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *DrugBatch) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, batchID int64) error

	ListBatchLocations(ctx context.Context, batchID int64) ([]DrugBatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error

	// Convenience for FEFO view (presentation + batches + totals)
	GetPresentationStock(ctx context.Context, presentationID int64) (*PresentationStock, error)
}

type PharmacyUseCase interface {
	// wrappers that also build DisplayLabel/DisplayStrength, totals, etc.
	ListDrugs(ctx context.Context, q *string) ([]Drug, error)
	GetDrugWithPresentations(ctx context.Context, drugID int64) (*DrugWithPresentations, error)

	GetPresentationStock(ctx context.Context, presentationID int64) (*PresentationStock, error)

	// Admin flows
	CreateDrug(ctx context.Context, d *Drug) (*Drug, error)
	UpdateDrug(ctx context.Context, d *Drug) (*Drug, error)
	DeleteDrug(ctx context.Context, id int64) error

	CreatePresentation(ctx context.Context, p *DrugPresentation) (*DrugPresentationView, error)
	UpdatePresentation(ctx context.Context, p *DrugPresentation) (*DrugPresentationView, error)
	DeletePresentation(ctx context.Context, id int64) error

	ListBatches(ctx context.Context, presentationID int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, batchID int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *DrugBatch, locations []DrugBatchLocation) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *DrugBatch) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, batchID int64) error

	ListBatchLocations(ctx context.Context, batchID int64) ([]DrugBatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *DrugBatchLocation) (*DrugBatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error
}
