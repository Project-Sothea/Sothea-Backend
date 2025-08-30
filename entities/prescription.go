package entities

import (
	"context"
	"time"
)

type Prescription struct {
	ID              int64              `json:"id"`
	VID             int32              `json:"vid"`
	PatientID       int64              `json:"patientId"`
	StaffID         *int64             `json:"staffId"` // Optional
	Notes           *string            `json:"notes"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	PrescribedDrugs []DrugPrescription `json:"prescribedDrugs"`
	IsPacked        bool               `json:"isPacked"`
}

type DrugPrescription struct {
	ID             int64                   `json:"id"`
	PrescriptionID int64                   `json:"prescriptionId"`
	DrugID         int64                   `json:"drugId"`
	Quantity       int                     `json:"quantity"`
	Remarks        *string                 `json:"remarks"` // aka instructions
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
	Batches        []PrescriptionBatchItem `json:"batches"`
}

type PrescriptionBatchItem struct {
	ID                 int64     `json:"id"`
	DrugPrescriptionID int64     `json:"drugPrescriptionId"`
	BatchId            int64     `json:"batchId"`
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
