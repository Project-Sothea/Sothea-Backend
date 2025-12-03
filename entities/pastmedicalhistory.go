package entities

import (
	"fmt"
)

type PastMedicalHistory struct {
	ID  int32 `json:"id" binding:"-"`
	VID int32 `json:"vid" binding:"-"`

	Cough                   *bool `json:"cough"`                   // Allow null for 'Nil' option
	Fever                   *bool `json:"fever"`                   // Allow null for 'Nil' option
	BlockedNose             *bool `json:"blockedNose"`             // Allow null for 'Nil' option
	SoreThroat              *bool `json:"soreThroat"`              // Allow null for 'Nil' option
	NightSweats             *bool `json:"nightSweats"`             // Allow null for 'Nil' option
	UnintentionalWeightLoss *bool `json:"unintentionalWeightLoss"` // Allow null for 'Nil' option

	Tuberculosis               *bool `json:"tuberculosis"`               // Allow null for 'Nil' option
	TuberculosisHasBeenTreated *bool `json:"tuberculosisHasBeenTreated"` // Allow null for 'Nil' option

	Diabetes                   *bool   `json:"diabetes"`                   // Allow null for 'Nil' option
	Hypertension               *bool   `json:"hypertension"`               // Allow null for 'Nil' option
	Hyperlipidemia             *bool   `json:"hyperlipidemia"`             // Allow null for 'Nil' option
	ChronicJointPains          *bool   `json:"chronicJointPains"`          // Allow null for 'Nil' option
	ChronicMuscleAches         *bool   `json:"chronicMuscleAches"`         // Allow null for 'Nil' option
	SexuallyTransmittedDisease *bool   `json:"sexuallyTransmittedDisease"` // Allow null for 'Nil' option
	SpecifiedSTDs              *string `json:"specifiedSTDs"`
	Others                     *string `json:"others"`
	//AdminID                    uint `gorm:"uniqueIndex;not null"` // Foreign key referencing Admin's ID
	//Admin                      Admin
}

// TableName specifies the table name for the PastMedicalHistory model.
func (PastMedicalHistory) TableName() string {
	return "pastmedicalhistory"
}

// ToString generates a simple string representation of the PastMedicalHistory struct.
func (pmh PastMedicalHistory) String() string {
	result := fmt.Sprintf("\nPAST MEDICAL HISTORY\n")
	result += fmt.Sprintf("ID: %d\n", pmh.ID)
	result += fmt.Sprintf("VID: %d\n", pmh.VID)

	result += fmt.Sprintf("Cough: %s\n", SafeDerefBool(pmh.Cough))
	result += fmt.Sprintf("Fever: %s\n", SafeDerefBool(pmh.Fever))
	result += fmt.Sprintf("Blocked Nose: %s\n", SafeDerefBool(pmh.BlockedNose))
	result += fmt.Sprintf("Sore Throat: %s\n", SafeDerefBool(pmh.SoreThroat))
	result += fmt.Sprintf("Night Sweats: %s\n", SafeDerefBool(pmh.NightSweats))
	result += fmt.Sprintf("Unintentional Weight Loss: %s\n", SafeDerefBool(pmh.UnintentionalWeightLoss))

	result += fmt.Sprintf("Tuberculosis: %s\n", SafeDerefBool(pmh.Tuberculosis))
	result += fmt.Sprintf("Has Tuberculosis been treated?: %s\n", SafeDerefBool(pmh.TuberculosisHasBeenTreated))

	result += fmt.Sprintf("Diabetes: %s\n", SafeDerefBool(pmh.Diabetes))
	result += fmt.Sprintf("Hypertension: %s\n", SafeDerefBool(pmh.Hypertension))
	result += fmt.Sprintf("Hyperlipidemia: %s\n", SafeDerefBool(pmh.Hyperlipidemia))
	result += fmt.Sprintf("ChronicJointPains: %s\n", SafeDerefBool(pmh.ChronicJointPains))
	result += fmt.Sprintf("ChronicMuscleAches: %s\n", SafeDerefBool(pmh.ChronicMuscleAches))
	result += fmt.Sprintf("SexuallyTransmittedDisease: %s\n", SafeDerefBool(pmh.SexuallyTransmittedDisease))
	result += fmt.Sprintf("SpecifiedSTDs: %s\n", SafeDeref(pmh.SpecifiedSTDs))
	result += fmt.Sprintf("Others: %s\n", SafeDeref(pmh.Others))
	return result
}
