package entities

import "fmt"

type HeightAndWeight struct {
	ID          int32    `json:"id" binding:"-"`
	VID         int32    `json:"vid" binding:"-"`
	Height      *float64 `json:"height" binding:"required"`
	Weight      *float64 `json:"weight" binding:"required"`
	BMI         *float64 `json:"bmi" binding:"required"`
	BMIAnalysis *string  `json:"bmiAnalysis" binding:"required"`
	PaedsHeight *float64 `json:"paedsHeight"`
	PaedsWeight *float64 `json:"paedsWeight"`

	IcopeLostWeightPastMonths *bool `json:"icopeLostWeightPastMonths"`
	IcopeNoDesireToEat        *bool `json:"icopeNoDesireToEat"`
	//AdminID      uint    `gorm:"uniqueIndex;not null"` // Foreign key referencing Admin's ID
	//Admin        Admin
}

// TableName specifies the table name for the HeightAndWeight model.
func (HeightAndWeight) TableName() string {
	return "heightandweight"
}

// ToString generates a simple string representation of the HeightAndWeight struct.
func (haw HeightAndWeight) String() string {
	result := fmt.Sprintf("\nHEIGHT AND WEIGHT\n")
	result += fmt.Sprintf("ID: %d\n", haw.ID)
	result += fmt.Sprintf("VID: %d\n", haw.VID)
	result += fmt.Sprintf("Height: %.2f\n", *haw.Height)
	result += fmt.Sprintf("Weight: %.2f\n", *haw.Weight)
	result += fmt.Sprintf("BMI: %.2f\n", *haw.BMI)
	result += fmt.Sprintf("BMI Analysis: %s\n", *haw.BMIAnalysis)
	result += fmt.Sprintf("Paeds Height: %.2f\n", SafeDeref(haw.PaedsHeight))
	result += fmt.Sprintf("Paeds Weight: %.2f\n", SafeDeref(haw.PaedsWeight))
	result += fmt.Sprintf("Rapid Weight Lost: %t\n", SafeDeref(haw.IcopeLostWeightPastMonths))
	result += fmt.Sprintf("No Desire to eat: %t\n", SafeDeref(haw.IcopeNoDesireToEat))
	return result
}
