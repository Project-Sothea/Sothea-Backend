package entities

import (
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
	DispenserName *string            `json:"dispenser_name"`
	Lines         []PrescriptionLine `json:"lines"`
}

type PrescriptionLine struct {
	db.PrescriptionLine
	PackerName   *string                    `json:"packer_name"`
	UpdaterName  *string                    `json:"updater_name"`
	CreatorName  *string                    `json:"creator_name"`
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
}
