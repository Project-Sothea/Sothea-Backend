package entities

import "fmt"

type FallRisk struct {
	ID                      int32   `json:"id" binding:"-"`
	VID                     int32   `json:"vid" binding:"-"`
	PastYearFall            *bool   `json:"pastYearFall" binding:"required"`
	UnsteadyStandingFalling *bool   `json:"unsteadyStandingFalling" binding:"required"`
	FallWorries             *bool   `json:"fallWorries" binding:"required"`
	Others                  *string `json:"others"`
	FurtherReferral         *bool   `json:"furtherReferral" binding:"required"`
	//AdminID      uint    `gorm:"uniqueIndex;not null"` // Foreign key referencing Admin's ID
	//Admin        Admin
}

// TableName specifies the table name for the FallRisk model.
func (FallRisk) TableName() string {
	return "fallrisk"
}

// ToString generates a simple string representation of the HeightAndWeight struct.
func (fr FallRisk) String() string {
	result := fmt.Sprintf("\nHEIGHT AND WEIGHT\n")
	result += fmt.Sprintf("ID: %d\n", fr.ID)
	result += fmt.Sprintf("VID: %d\n", fr.VID)
	result += fmt.Sprintf("Past Year Fall: %t\n", *fr.PastYearFall)
	result += fmt.Sprintf("Unsteady Standing Falling: %t\n", *fr.UnsteadyStandingFalling)
	result += fmt.Sprintf("Fall Worries: %t\n", *fr.FallWorries)
	result += fmt.Sprintf("Others: %s\n", *fr.Others)
	result += fmt.Sprintf("Further Referral: %t\n", *fr.FurtherReferral)
	return result
}
