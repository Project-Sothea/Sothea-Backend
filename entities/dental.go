package entities

import (
	"fmt"
)

type Dental struct {
	ID  int32 `json:"id" binding:"-"`
	VID int32 `json:"vid" binding:"-"`

	CleanTeethFreq        *int    `json:"cleanTeethFreq" binding:"required"`
	SugarConsumeFreq      *string `json:"sugarConsumeFreq" binding:"required"`
	BacterialExposure     *bool   `json:"bacterialExposure" binding:"required"`
	NumLossFromToothDecay *int    `json:"numLossFromToothDecay" binding:"required"`
	OralSymptoms          *bool   `json:"oralSymptoms" binding:"required"`
	DrinkOtherWater       *bool   `json:"drinkOtherWater" binding:"required"`

	RiskForDentalCarries   *string `json:"riskForDentalCarries" binding:"required"`
	IcopeDifficultyChewing *bool   `json:"icopeDifficultyChewing" binding:"required"`
	IcopePainInMouth       *bool   `json:"icopePainInMouth" binding:"required"`

	DentalNotes    *string `json:"dentalNotes"`
	ReferralNeeded *bool   `json:"referralNeeded" binding:"required"`
	ReferralLoc    *string `json:"referralLoc"`

	Tooth11 *bool `json:"tooth11"`
	Tooth12 *bool `json:"tooth12"`
	Tooth13 *bool `json:"tooth13"`
	Tooth14 *bool `json:"tooth14"`
	Tooth15 *bool `json:"tooth15"`
	Tooth16 *bool `json:"tooth16"`
	Tooth17 *bool `json:"tooth17"`
	Tooth18 *bool `json:"tooth18"`
	Tooth21 *bool `json:"tooth21"`
	Tooth22 *bool `json:"tooth22"`
	Tooth23 *bool `json:"tooth23"`
	Tooth24 *bool `json:"tooth24"`
	Tooth25 *bool `json:"tooth25"`
	Tooth26 *bool `json:"tooth26"`
	Tooth27 *bool `json:"tooth27"`
	Tooth28 *bool `json:"tooth28"`
	Tooth31 *bool `json:"tooth31"`
	Tooth32 *bool `json:"tooth32"`
	Tooth33 *bool `json:"tooth33"`
	Tooth34 *bool `json:"tooth34"`
	Tooth35 *bool `json:"tooth35"`
	Tooth36 *bool `json:"tooth36"`
	Tooth37 *bool `json:"tooth37"`
	Tooth38 *bool `json:"tooth38"`
	Tooth41 *bool `json:"tooth41"`
	Tooth42 *bool `json:"tooth42"`
	Tooth43 *bool `json:"tooth43"`
	Tooth44 *bool `json:"tooth44"`
	Tooth45 *bool `json:"tooth45"`
	Tooth46 *bool `json:"tooth46"`
	Tooth47 *bool `json:"tooth47"`
	Tooth48 *bool `json:"tooth48"`
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
	result += fmt.Sprintf("CleanTeethFreq: %d\n", *fr.CleanTeethFreq)
	result += fmt.Sprintf("SugarConsumeFreq: %s\n", *fr.SugarConsumeFreq)
	result += fmt.Sprintf("BacterialExposure: %t\n", *fr.BacterialExposure)
	result += fmt.Sprintf("NumLossFromToothDecay: %d\n", *fr.NumLossFromToothDecay)
	result += fmt.Sprintf("OralSymptoms: %t\n", *fr.OralSymptoms)
	result += fmt.Sprintf("DrinkOtherThanWater: %t\n", *fr.DrinkOtherWater)

	result += fmt.Sprintf("RiskForDentalCarry: %s\n", *fr.RiskForDentalCarries)

	result += fmt.Sprintf("IcopeDifficultyChewing: %t\n", *fr.IcopeDifficultyChewing)
	result += fmt.Sprintf("IcopePainInMouth: %t\n", *fr.IcopePainInMouth)

	result += fmt.Sprintf("DentalNotes: %s\n", SafeDeref(fr.DentalNotes))
	result += fmt.Sprintf("ReferralNeeded: %t\n", *fr.ReferralNeeded)
	result += fmt.Sprintf("ReferralLoc: %s\n", SafeDeref(fr.ReferralLoc))
	result += fmt.Sprintf("Tooth11: %t\n", SafeDeref(fr.Tooth11))
	result += fmt.Sprintf("Tooth12: %t\n", SafeDeref(fr.Tooth12))
	result += fmt.Sprintf("Tooth13: %t\n", SafeDeref(fr.Tooth13))
	result += fmt.Sprintf("Tooth14: %t\n", SafeDeref(fr.Tooth14))
	result += fmt.Sprintf("Tooth15: %t\n", SafeDeref(fr.Tooth15))
	result += fmt.Sprintf("Tooth16: %t\n", SafeDeref(fr.Tooth16))
	result += fmt.Sprintf("Tooth17: %t\n", SafeDeref(fr.Tooth17))
	result += fmt.Sprintf("Tooth18: %t\n", SafeDeref(fr.Tooth18))
	result += fmt.Sprintf("Tooth21: %t\n", SafeDeref(fr.Tooth21))
	result += fmt.Sprintf("Tooth22: %t\n", SafeDeref(fr.Tooth22))
	result += fmt.Sprintf("Tooth23: %t\n", SafeDeref(fr.Tooth23))
	result += fmt.Sprintf("Tooth24: %t\n", SafeDeref(fr.Tooth24))
	result += fmt.Sprintf("Tooth25: %t\n", SafeDeref(fr.Tooth25))
	result += fmt.Sprintf("Tooth26: %t\n", SafeDeref(fr.Tooth26))
	result += fmt.Sprintf("Tooth27: %t\n", SafeDeref(fr.Tooth27))
	result += fmt.Sprintf("Tooth28: %t\n", SafeDeref(fr.Tooth28))
	result += fmt.Sprintf("Tooth31: %t\n", SafeDeref(fr.Tooth31))
	result += fmt.Sprintf("Tooth32: %t\n", SafeDeref(fr.Tooth32))
	result += fmt.Sprintf("Tooth33: %t\n", SafeDeref(fr.Tooth33))
	result += fmt.Sprintf("Tooth34: %t\n", SafeDeref(fr.Tooth34))
	result += fmt.Sprintf("Tooth35: %t\n", SafeDeref(fr.Tooth35))
	result += fmt.Sprintf("Tooth36: %t\n", SafeDeref(fr.Tooth36))
	result += fmt.Sprintf("Tooth37: %t\n", SafeDeref(fr.Tooth37))
	result += fmt.Sprintf("Tooth38: %t\n", SafeDeref(fr.Tooth38))
	result += fmt.Sprintf("Tooth41: %t\n", SafeDeref(fr.Tooth41))
	result += fmt.Sprintf("Tooth42: %t\n", SafeDeref(fr.Tooth42))
	result += fmt.Sprintf("Tooth43: %t\n", SafeDeref(fr.Tooth43))
	result += fmt.Sprintf("Tooth44: %t\n", SafeDeref(fr.Tooth44))
	result += fmt.Sprintf("Tooth45: %t\n", SafeDeref(fr.Tooth45))
	result += fmt.Sprintf("Tooth46: %t\n", SafeDeref(fr.Tooth46))
	result += fmt.Sprintf("Tooth47: %t\n", SafeDeref(fr.Tooth47))
	result += fmt.Sprintf("Tooth48: %t\n", SafeDeref(fr.Tooth48))
	return result
}
