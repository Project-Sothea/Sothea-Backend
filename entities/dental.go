package entities

import (
	"fmt"
)

type Dental struct {
	ID  int32 `json:"id" binding:"-"`
	VID int32 `json:"vid" binding:"-"`

	FluorideExposure       *string `json:"fluorideExposure" binding:"required"`
	Diet                   *string `json:"diet" binding:"required"`
	BacterialExposure      *string `json:"bacterialExposure" binding:"required"`
	OralSymptoms           *bool   `json:"oralSymptoms" binding:"required"`
	DrinkOtherWater        *bool   `json:"drinkOtherWater" binding:"required"`
	RiskForDentalCarries   *string `json:"riskForDentalCarries" binding:"required"`
	IcopeDifficultyChewing *bool   `json:"icopeDifficultyChewing"`
	IcopePainInMouth       *bool   `json:"icopePainInMouth"`
	DentalNotes            *string `json:"dentalNotes"`
}

// TableName specifies the table name for the Dental model.
func (Dental) TableName() string {
	return "dental"
}

// ToString generates a simple string representation of the Dental struct.
func (fr Dental) String() string {
	result := fmt.Sprintf("\nDENTAL\n")
	result += fmt.Sprintf("ID: %d\n", fr.ID)
	result += fmt.Sprintf("VID: %d\n", fr.VID)
	result += fmt.Sprintf("FluorideExposure: %s\n", SafeDeref(fr.FluorideExposure))
	result += fmt.Sprintf("Diet: %s\n", SafeDeref(fr.Diet))
	result += fmt.Sprintf("BacterialExposure: %s\n", SafeDeref(fr.BacterialExposure))
	result += fmt.Sprintf("OralSymptoms: %t\n", *fr.OralSymptoms)
	result += fmt.Sprintf("DrinkOtherThanWater: %t\n", *fr.DrinkOtherWater)

	result += fmt.Sprintf("RiskForDentalCarry: %s\n", *fr.RiskForDentalCarries)

	result += fmt.Sprintf("IcopeDifficultyChewing: %t\n", SafeDeref(fr.IcopeDifficultyChewing))
	result += fmt.Sprintf("IcopePainInMouth: %t\n", SafeDeref(fr.IcopePainInMouth))

	result += fmt.Sprintf("DentalNotes: %s\n", SafeDeref(fr.DentalNotes))
	return result
}
