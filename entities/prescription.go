package entities

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// High-level prescription aggregates
// -----------------------------------------------------------------------------

type Prescription struct {
	ID            int64      `json:"id"`
	VID           int32      `json:"vid"`
	PatientID     int64      `json:"patientId"`
	Notes         *string    `json:"notes,omitempty"`
	IsDispensed   bool       `json:"isDispensed"`
	DispensedBy   *int64     `json:"dispensedBy,omitempty"`
	DispenserName *string    `json:"dispenserName,omitempty"`
	DispensedAt   *time.Time `json:"dispensedAt,omitempty"`
	CreatedBy     *int64     `json:"createdBy,omitempty"`
	CreatorName   *string    `json:"creatorName,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

	// One line per presentation (form/strength/dispense unit)
	Lines []PrescriptionLine `json:"lines"`
}

// -----------------------------------------------------------------------------
// Scheduling constants (optional but helpful)
// -----------------------------------------------------------------------------

type ScheduleKind string

const (
	ScheduleHour  ScheduleKind = "hour"
	ScheduleDay   ScheduleKind = "day"
	ScheduleWeek  ScheduleKind = "week"
	ScheduleMonth ScheduleKind = "month"
)

// -----------------------------------------------------------------------------
// One line = one presentation + dose + schedule
// -----------------------------------------------------------------------------

type PrescriptionLine struct {
	ID             int64   `json:"id"`
	PrescriptionID int64   `json:"prescriptionId"`
	PresentationID int64   `json:"presentationId"`
	Remarks        *string `json:"remarks,omitempty"` // SIG / instructions

	// Clinical dose input (per administration)
	DoseAmount float64 `json:"doseAmount"`
	DoseUnit   string  `json:"doseUnit"` // "mg","mL","tab","bottle", etc.

	// Periodic schedule model:
	// periods = ceil(duration / everyN); doses = periods * frequencyPerSchedule
	// Examples:
	//  - TID x 7 days        → kind=day,  everyN=1, freq=3,  duration=7
	//  - q8h x 5 days        → kind=hour, everyN=8, freq=1,  duration=120
	//  - every other day x14 → kind=day,  everyN=2, freq=1,  duration=14
	//  - weekly x 4 weeks    → kind=week, everyN=1, freq=1,  duration=4
	ScheduleKind         string  `json:"scheduleKind"`         // "hour" | "day" | "week" | "month"
	EveryN               int     `json:"everyN"`               // > 0
	FrequencyPerSchedule float64 `json:"frequencyPerSchedule"` // administrations per schedule period

	Duration     float64 `json:"duration"`     // in units of scheduleKind
	DurationUnit string  `json:"durationUnit"` // "hour" | "day" | "week" | "month"

	// Computed by backend (in presentation’s dispense unit)
	TotalToDispense int    `json:"totalToDispense"`
	DispenseUnit    string `json:"dispenseUnit"` // copied from presentation for quick rendering

	// Packing workflow
	IsPacked   bool       `json:"isPacked"`
	PackedBy   *int64     `json:"packedBy,omitempty"`
	PackerName *string    `json:"packerName,omitempty"`
	PackedAt   *time.Time `json:"packedAt,omitempty"`

	// Name of person who created / last
	UpdaterName *string `json:"updaterName,omitempty"`

	// Helpful denormalized display (optional)
	DrugName        string `json:"drugName,omitempty"`
	DisplayStrength string `json:"displayStrength,omitempty"` // e.g., "500 mg/tab" or "250 mg/5 mL"
	DisplayRoute    string `json:"displayRoute,omitempty"`    // e.g., "PO"
	DisplayLabel    string `json:"displayLabel,omitempty"`    // full text

	// Allocations (sum must equal TotalToDispense when IsPacked = true)
	Allocations []LineAllocation `json:"allocations"`
}

type AddLineReq struct {
	PresentationID       int64   `json:"presentationId" binding:"required"`
	Remarks              *string `json:"remarks"`
	DoseAmount           float64 `json:"doseAmount" binding:"required,gt=0"`
	DoseUnit             string  `json:"doseUnit" binding:"required"`
	ScheduleKind         string  `json:"scheduleKind" binding:"required,oneof=hour day week month"`
	EveryN               int     `json:"everyN" binding:"required,gt=0"`               // e.g. every 2 days
	FrequencyPerSchedule float64 `json:"frequencyPerSchedule" binding:"required,gt=0"` // e.g. 3 doses per 'day' window
	Duration             float64 `json:"duration" binding:"required,gt=0"`             // in units of ScheduleKind * EveryN
	DurationUnit         string  `json:"durationUnit" binding:"required,oneof=hour day week month"`
}

type LineAllocation struct {
	ID              int64     `json:"id"`
	LineID          int64     `json:"lineId"`
	BatchLocationID int64     `json:"batchLocationId"`
	Quantity        int       `json:"quantity"` // in DispenseUnit
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	// (Optional) denormalized for UI lists
	BatchNumber string     `json:"batchNumber,omitempty"`
	Location    string     `json:"location,omitempty"`
	ExpiryDate  *time.Time `json:"expiryDate,omitempty"`
}

type SetAllocReq struct {
	Allocations []struct {
		BatchLocationID int64 `json:"batchLocationId" validate:"required"`
		Quantity        int   `json:"quantity" validate:"gt=0"`
	} `json:"allocations" validate:"dive"`
}

// -----------------------------------------------------------------------------
// Repository & UseCase interfaces (unchanged signatures)
// -----------------------------------------------------------------------------

type PrescriptionRepository interface {
	// PRESCRIPTIONS
	CreatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	GetPrescriptionByID(ctx context.Context, id int64) (*Prescription, error)
	ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*Prescription, error)
	UpdatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	DeletePrescription(ctx context.Context, id int64) error

	// LINES
	GetLine(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	AddLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	UpdateLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	RemoveLine(ctx context.Context, lineID int64) error

	// ALLOCATIONS (packing plan per line; stock reservations handled by DB triggers)
	ListLineAllocations(ctx context.Context, lineID int64) ([]LineAllocation, error)
	SetLineAllocations(ctx context.Context, lineID int64, allocs []LineAllocation) ([]LineAllocation, error)

	// STATE CHANGES
	MarkLinePacked(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	UnpackLine(ctx context.Context, lineID int64) (*PrescriptionLine, error)

	// DISPENSE: finalize & stamp (no stock mutation here; triggers already reserved)
	DispensePrescription(ctx context.Context, prescriptionID int64) (*Prescription, error)
}

type PrescriptionUseCase interface {
	CreatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	GetPrescriptionByID(ctx context.Context, id int64) (*Prescription, error)
	ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*Prescription, error)
	UpdatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	DeletePrescription(ctx context.Context, id int64) error

	// Line CRUD
	AddLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	UpdateLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	RemoveLine(ctx context.Context, lineID int64) error

	// Packing helpers
	SuggestFEFOAllocations(ctx context.Context, lineID int64) ([]LineAllocation, error)
	ListLineAllocations(ctx context.Context, lineID int64) ([]LineAllocation, error)
	SetLineAllocations(ctx context.Context, lineID int64, allocs []LineAllocation) ([]LineAllocation, error)
	MarkLinePacked(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	UnpackLine(ctx context.Context, lineID int64) (*PrescriptionLine, error)

	// Finalize
	DispensePrescription(ctx context.Context, prescriptionID int64) (*Prescription, error)
}
