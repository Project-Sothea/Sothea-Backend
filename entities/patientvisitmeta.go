package entities

import (
	"fmt"
	"time"
)

type PatientVisitMeta struct {
	ID             int32      `json:"id" binding:"-"`
	VID            int32      `json:"vid" binding:"-"`
	FamilyGroup    *string    `json:"familyGroup" binding:"required"`
	RegDate        *time.Time `json:"regDate" binding:"required"`
	QueueNo        *string    `json:"queueNo" binding:"required"`
	Name           *string    `json:"name" binding:"required"`
	KhmerName      *string    `json:"khmerName" binding:"required"`
	Gender         *string    `json:"gender" binding:"required"`
	Village        *string    `json:"village" binding:"required"`
	ContactNo      *string    `json:"contactNo" binding:"required"`
	DrugAllergies  *string    `json:"drugAllergies"`
	SentToID       *bool      `json:"sentToId" binding:"required"`
	ReferralNeeded *bool      `json:"referralNeeded" binding:"required"`

	// Derived visit-level prescription / dispensing flags
	HasPrescriptionWithDrug    *bool `json:"hasPrescriptionWithDrug" binding:"required"`
	AllPrescriptionDrugsPacked *bool `json:"allPrescriptionDrugsPacked" binding:"required"`
	PrescriptionDispensed      *bool `json:"prescriptionDispensed" binding:"required"`
}

// TableName specifies the table name for the PatientVisitMeta model.
func (PatientVisitMeta) TableName() string {
	return "patientvisitmeta"
}

// ToString generates a simple string representation of the PatientVisitMeta struct.
func (pvm PatientVisitMeta) String() string {
	result := fmt.Sprintf("\nPATIENTVISITMETA\n")
	result += fmt.Sprintf("ID: %d\n", pvm.ID)
	result += fmt.Sprintf("VID: %d\n", pvm.VID)
	result += fmt.Sprintf("FamilyGroup: %s\n", *pvm.FamilyGroup)
	result += fmt.Sprintf("RegDate: %s\n", pvm.RegDate.Format("2006-01-02"))
	result += fmt.Sprintf("QueueNo: %s\n", *pvm.QueueNo)
	result += fmt.Sprintf("Name: %s\n", *pvm.Name)
	result += fmt.Sprintf("KhmerName: %s\n", *pvm.KhmerName)
	result += fmt.Sprintf("Gender: %s\n", *pvm.Gender)
	result += fmt.Sprintf("Village: %s\n", *pvm.Village)
	result += fmt.Sprintf("ContactNo: %s\n", *pvm.ContactNo)
	result += fmt.Sprintf("DrugAllergies: %v\n", SafeDeref(pvm.DrugAllergies))
	result += fmt.Sprintf("SentToID: %t\n", *pvm.SentToID)
	result += fmt.Sprintf("ReferralNeeded: %t\n", *pvm.ReferralNeeded)
	result += fmt.Sprintf("HasPrescriptionWithDrug: %t\n", SafeDeref(pvm.HasPrescriptionWithDrug))
	result += fmt.Sprintf("AllPrescriptionDrugsPacked: %t\n", SafeDeref(pvm.AllPrescriptionDrugsPacked))
	result += fmt.Sprintf("PrescriptionDispensed: %t\n", SafeDeref(pvm.PrescriptionDispensed))
	return result
}
