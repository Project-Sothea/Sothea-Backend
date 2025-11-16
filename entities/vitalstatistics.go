package entities

import "fmt"

type VitalStatistics struct {
	ID                      int32    `json:"id"`
	VID                     int32    `json:"vid" binding:"-"`
	Temperature             *float64 `json:"temperature" binding:"required"`
	SpO2                    *float64 `json:"spO2" binding:"required"`
	SystolicBP1             *float64 `json:"systolicBP1"`
	DiastolicBP1            *float64 `json:"diastolicBP1"`
	SystolicBP2             *float64 `json:"systolicBP2"`
	DiastolicBP2            *float64 `json:"diastolicBP2"`
	AverageSystolicBP       *float64 `json:"averageSystolicBP"`
	AverageDiastolicBP      *float64 `json:"averageDiastolicBP"`
	HR1                     *float64 `json:"hr1" binding:"required"`
	HR2                     *float64 `json:"hr2" binding:"required"`
	AverageHR               *float64 `json:"averageHR" binding:"required"`
	RandomBloodGlucoseMmolL *float64 `json:"randomBloodGlucoseMmolL"`
	IcopeHighBp             *bool    `json:"icopeHighBp"`
	//AdminID                  uint    `gorm:"uniqueIndex;not null"` // Foreign key referencing Admin's ID
	//Admin                    Admin
}

// TableName specifies the table name for the VitalStatistics model.
func (VitalStatistics) TableName() string {
	return "vitalstatistics"
}

// ToString generates a simple string representation of the VitalStatistics struct.
func (vs VitalStatistics) String() string {
	result := fmt.Sprintf("\nVITAL STATISTICS\n")
	result += fmt.Sprintf("ID: %d\n", vs.ID)
	result += fmt.Sprintf("VID: %d\n", vs.VID)
	result += fmt.Sprintf("Temperature: %.1f\n", *vs.Temperature)
	result += fmt.Sprintf("SpO2: %.1f\n", *vs.SpO2)
	result += fmt.Sprintf("SystolicBP1: %.1f\n", SafeDeref(vs.SystolicBP1))
	result += fmt.Sprintf("DiastolicBP1: %.1f\n", SafeDeref(vs.DiastolicBP1))
	result += fmt.Sprintf("SystolicBP2: %.1f\n", SafeDeref(vs.SystolicBP2))
	result += fmt.Sprintf("DiastolicBP2: %.1f\n", SafeDeref(vs.DiastolicBP2))
	result += fmt.Sprintf("AverageSystolicBP: %.1f\n", SafeDeref(vs.AverageSystolicBP))
	result += fmt.Sprintf("AverageDiastolicBP: %.1f\n", SafeDeref(vs.AverageDiastolicBP))
	result += fmt.Sprintf("HR1: %.1f\n", *vs.HR1)
	result += fmt.Sprintf("HR2: %.1f\n", *vs.HR2)
	result += fmt.Sprintf("AverageHR: %.1f\n", *vs.AverageHR)
	result += fmt.Sprintf("RandomBloodGlucoseMmolL: %.1f\n", *vs.RandomBloodGlucoseMmolL)
	result += fmt.Sprintf("IcopeHighBp: %t\n", SafeDeref(vs.IcopeHighBp))
	return result
}
