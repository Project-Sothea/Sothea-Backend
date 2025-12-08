package entities

import (
	"context"
	"time"

	db "sothea-backend/repository/sqlc"
)

// Aggregated patient view returned by APIs.
type Patient struct {
	PatientDetails      *db.PatientDetail       `json:"patient_details"`
	Admin               *db.Admin               `json:"admin"`
	PastMedicalHistory  *db.PastMedicalHistory  `json:"past_medical_history"`
	SocialHistory       *db.SocialHistory       `json:"social_history"`
	VitalStatistics     *db.VitalStatistic      `json:"vital_statistics"`
	HeightAndWeight     *db.HeightAndWeight     `json:"height_and_weight"`
	VisualAcuity        *db.VisualAcuity        `json:"visual_acuity"`
	Dental              *db.Dental              `json:"dental"`
	FallRisk            *db.FallRisk            `json:"fall_risk"`
	Physiotherapy       *db.Physiotherapy       `json:"physiotherapy"`
	DoctorsConsultation *db.DoctorsConsultation `json:"doctors_consultation"`
}

type PatientMeta struct {
	ID          int32               `json:"id"`
	Vid         int32               `json:"vid"`
	FamilyGroup string              `json:"family_group"`
	RegDate     time.Time           `json:"reg_date"`
	QueueNo     string              `json:"queue_no"`
	Name        string              `json:"name"`
	KhmerName   string              `json:"khmer_name"`
	Visits      map[int32]time.Time `json:"visits"`
}

type PatientVisitMeta struct {
	ID                         int32     `json:"id"`
	Vid                        int32     `json:"vid"`
	FamilyGroup                string    `json:"family_group"`
	RegDate                    time.Time `json:"reg_date"`
	QueueNo                    string    `json:"queue_no"`
	Name                       string    `json:"name"`
	KhmerName                  string    `json:"khmer_name"`
	Gender                     string    `json:"gender"`
	Village                    string    `json:"village"`
	ContactNo                  string    `json:"contact_no"`
	DrugAllergies              *string   `json:"drug_allergies"`
	SentToID                   bool      `json:"sent_to_id"`
	ReferralNeeded             *bool     `json:"referral_needed"`
	HasPrescriptionWithDrug    bool      `json:"has_prescription_with_drug"`
	AllPrescriptionDrugsPacked bool      `json:"all_prescription_drugs_packed"`
	PrescriptionDispensed      bool      `json:"prescription_dispensed"`
}

// Drug view helpers retained for UI convenience.
type DrugView struct {
	db.Drug

	DisplayStrength string `json:"display_strength"`
	DisplayRoute    string `json:"display_route"`
	DisplayLabel    string `json:"display_label"`
}

type BatchDetail struct {
	db.DrugBatch

	DispenseUnit   string             `json:"dispense_unit"`
	ExpirySortKey  *time.Time         `json:"expiry_sort_key"`
	BatchLocations []db.BatchLocation `json:"batch_locations"`
}

type DrugStock struct {
	Drug     DrugView      `json:"drug"`
	Batches  []BatchDetail `json:"batches"`
	TotalQty int           `json:"total_qty"`
}

type Prescription struct {
	db.Prescription
	Lines []PrescriptionLine `json:"lines"`
}

type PrescriptionLine struct {
	db.PrescriptionLine
	Allocations  []db.PrescriptionBatchItem `json:"allocations"`
	DispenseUnit string                     `json:"dispense_unit,omitempty"`
}

type AddLineReq struct {
	DrugID        int64   `json:"drug_id" binding:"required"`
	Remarks       *string `json:"remarks"`
	Prn           bool    `json:"prn"`
	DoseAmount    float64 `json:"dose_amount" binding:"required,gt=0"`
	DoseUnit      string  `json:"dose_unit" binding:"required"`
	FrequencyCode string  `json:"frequency_code" binding:"required"`
	Duration      float64 `json:"duration" binding:"required,gt=0"`
	DurationUnit  string  `json:"duration_unit" binding:"required,oneof=hour day week month"`
}

type SetAllocReq struct {
	Allocations []struct {
		BatchLocationID int64 `json:"batch_location_id" validate:"required"`
		Quantity        int   `json:"quantity" validate:"gt=0"`
	} `json:"allocations" validate:"dive"`
}

type LoginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ---------------- Interfaces ----------------

type PatientUseCase interface {
	GetPatientVisit(ctx context.Context, id int32, vid int32) (*Patient, error)
	CreatePatient(ctx context.Context, patient *db.PatientDetail) (int32, error)
	UpdatePatient(ctx context.Context, id int32, patient *db.PatientDetail) error
	DeletePatient(ctx context.Context, id int32) error
	CreatePatientVisit(ctx context.Context, id int32, admin *db.Admin) (int32, error)
	DeletePatientVisit(ctx context.Context, id int32, vid int32) error
	UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *Patient) error
	GetPatientMeta(ctx context.Context, id int32) (*PatientMeta, error)
	GetAllPatientVisitMeta(ctx context.Context, date time.Time) ([]PatientVisitMeta, error)
}

type PatientRepository interface {
	GetPatientVisit(ctx context.Context, id int32, vid int32) (*Patient, error)
	CreatePatient(ctx context.Context, patient *db.PatientDetail) (int32, error)
	UpdatePatient(ctx context.Context, id int32, patient *db.PatientDetail) error
	DeletePatient(ctx context.Context, id int32) error
	CreatePatientVisit(ctx context.Context, id int32, admin *db.Admin) (int32, error)
	DeletePatientVisit(ctx context.Context, id int32, vid int32) error
	UpdatePatientVisit(ctx context.Context, id int32, vid int32, patient *Patient) error
	GetPatientMeta(ctx context.Context, id int32) (*PatientMeta, error)
	GetAllPatientVisitMeta(ctx context.Context, date time.Time) ([]PatientVisitMeta, error)
	GetDBUser(ctx context.Context, username string) (*db.User, error)
}

type PharmacyRepository interface {
	ListDrugs(ctx context.Context, q *string) ([]DrugView, error)
	CreateDrug(ctx context.Context, d *db.Drug) (*DrugView, error)
	GetDrug(ctx context.Context, id int64) (*DrugView, error)
	UpdateDrug(ctx context.Context, d *db.Drug) (*DrugView, error)
	DeleteDrug(ctx context.Context, id int64) error
	ListBatches(ctx context.Context, drugID int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, batchID int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, batchID int64) error
	ListBatchLocations(ctx context.Context, batchID int64) ([]db.BatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error
	GetDrugStock(ctx context.Context, drugID int64) (*DrugStock, error)
}

type PharmacyUseCase interface {
	ListDrugs(ctx context.Context, q *string) ([]DrugView, error)
	GetDrug(ctx context.Context, id int64) (*DrugView, error)
	GetDrugStock(ctx context.Context, drugID int64) (*DrugStock, error)
	CreateDrug(ctx context.Context, d *db.Drug) (*DrugView, error)
	UpdateDrug(ctx context.Context, d *db.Drug) (*DrugView, error)
	DeleteDrug(ctx context.Context, id int64) error
	ListBatches(ctx context.Context, drugID int64) ([]BatchDetail, error)
	GetBatch(ctx context.Context, batchID int64) (*BatchDetail, error)
	CreateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*BatchDetail, error)
	UpdateBatch(ctx context.Context, b *db.DrugBatch, locations []db.BatchLocation) (*BatchDetail, error)
	DeleteBatch(ctx context.Context, batchID int64) error
	ListBatchLocations(ctx context.Context, batchID int64) ([]db.BatchLocation, error)
	CreateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error)
	UpdateBatchLocation(ctx context.Context, loc *db.BatchLocation) (*db.BatchLocation, error)
	DeleteBatchLocation(ctx context.Context, id int64) error
}

type PrescriptionRepository interface {
	CreatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	GetPrescriptionByID(ctx context.Context, id int64) (*Prescription, error)
	ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*Prescription, error)
	UpdatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	DeletePrescription(ctx context.Context, id int64) error
	GetLine(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	AddLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	UpdateLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	RemoveLine(ctx context.Context, lineID int64) error
	ListLineAllocations(ctx context.Context, lineID int64) ([]db.PrescriptionBatchItem, error)
	SetLineAllocations(ctx context.Context, lineID int64, allocs []db.PrescriptionBatchItem) ([]db.PrescriptionBatchItem, error)
	MarkLinePacked(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	UnpackLine(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	DispensePrescription(ctx context.Context, prescriptionID int64) (*Prescription, error)
}

type PrescriptionUseCase interface {
	CreatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	GetPrescriptionByID(ctx context.Context, id int64) (*Prescription, error)
	ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*Prescription, error)
	UpdatePrescription(ctx context.Context, p *Prescription) (*Prescription, error)
	DeletePrescription(ctx context.Context, id int64) error
	AddLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	UpdateLine(ctx context.Context, line *PrescriptionLine) (*PrescriptionLine, error)
	RemoveLine(ctx context.Context, lineID int64) error
	SuggestFEFOAllocations(ctx context.Context, lineID int64) ([]db.PrescriptionBatchItem, error)
	ListLineAllocations(ctx context.Context, lineID int64) ([]db.PrescriptionBatchItem, error)
	SetLineAllocations(ctx context.Context, lineID int64, allocs []db.PrescriptionBatchItem) ([]db.PrescriptionBatchItem, error)
	MarkLinePacked(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	UnpackLine(ctx context.Context, lineID int64) (*PrescriptionLine, error)
	DispensePrescription(ctx context.Context, prescriptionID int64) (*Prescription, error)
}

type LoginUseCase interface {
	Login(ctx context.Context, user LoginPayload) (string, error)
}
