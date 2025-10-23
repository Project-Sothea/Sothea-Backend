package entities

import (
	"context"
	"time"
)

type Prescription struct {
	ID              int64              `json:"id"`
	VID             int32              `json:"vid"`
	PatientID       int64              `json:"patientId"`
	Notes           *string            `json:"notes"`
	CreatedBy       *int64             `json:"createdBy"`
	CreatorName     *string            `json:"creatorName,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	PrescribedDrugs []DrugPrescription `json:"prescribedDrugs"`
	IsDispensed     bool               `json:"isDispensed"`
	DispensedBy     *int64             `json:"dispensedBy,omitempty"`
	DispenserName   *string            `json:"dispenserName,omitempty"`
	DispensedAt     *time.Time         `json:"dispensedAt,omitempty"`
}

type DrugPrescription struct {
	ID             int64                   `json:"id"`
	PrescriptionID int64                   `json:"prescriptionId"`
	DrugID         int64                   `json:"drugId"`
	Remarks        *string                 `json:"remarks"` // aka instructions
	RequestedQty   int64                   `json:"requestedQty"`
	IsPacked       bool                    `json:"isPacked"`
	PackedBy       *int64                  `json:"packedBy,omitempty"`
	PackerName     *string                 `json:"packerName,omitempty"`
	PackedAt       *time.Time              `json:"packedAt,omitempty"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
	Batches        []PrescriptionBatchItem `json:"batches"`
}

type PrescriptionBatchItem struct {
	ID                 int64     `json:"id"`
	DrugPrescriptionID int64     `json:"drugPrescriptionId"`
	BatchLocationId    int64     `json:"batchLocationId"`
	Quantity           int       `json:"quantity"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type PrescriptionRepository interface {
	CreatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	GetPrescriptionByID(ctx context.Context, id int64) (*Prescription, error)
	ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*Prescription, error)
	UpdatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	DeletePrescription(ctx context.Context, id int64) error
}

type PrescriptionUseCase interface {
	CreatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	GetPrescriptionByID(ctx context.Context, id int64) (*Prescription, error)
	ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*Prescription, error)
	UpdatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	DeletePrescription(ctx context.Context, id int64) error
}
