package entities

import (
	"fmt"
)

type PastMedicalHistory struct {
	ID                         int32   `json:"id" binding:"-"`
	VID                        int32   `json:"vid" binding:"-"`
	Tuberculosis               *bool   `json:"tuberculosis" binding:"required"`
	Diabetes                   *bool   `json:"diabetes" binding:"required"`
	Hypertension               *bool   `json:"hypertension" binding:"required"`
	Hyperlipidemia             *bool   `json:"hyperlipidemia" binding:"required"`
	ChronicJointPains          *bool   `json:"chronicJointPains" binding:"required"`
	ChronicMuscleAches         *bool   `json:"chronicMuscleAches" binding:"required"`
	SexuallyTransmittedDisease *bool   `json:"sexuallyTransmittedDisease" binding:"required"`
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
	result += fmt.Sprintf("Tuberculosis: %t\n", *pmh.Tuberculosis)
	result += fmt.Sprintf("Diabetes: %t\n", *pmh.Diabetes)
	result += fmt.Sprintf("Hypertension: %t\n", *pmh.Hypertension)
	result += fmt.Sprintf("Hyperlipidemia: %t\n", *pmh.Hyperlipidemia)
	result += fmt.Sprintf("ChronicJointPains: %t\n", *pmh.ChronicJointPains)
	result += fmt.Sprintf("ChronicMuscleAches: %t\n", *pmh.ChronicMuscleAches)
	result += fmt.Sprintf("SexuallyTransmittedDisease: %t\n", *pmh.SexuallyTransmittedDisease)
	result += fmt.Sprintf("SpecifiedSTDs: %s\n", SafeDeref(pmh.SpecifiedSTDs))
	result += fmt.Sprintf("Others: %s\n", SafeDeref(pmh.Others))
	return result
}
