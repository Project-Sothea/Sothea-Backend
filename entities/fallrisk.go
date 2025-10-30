package entities

import (
	"fmt"
)

type FallRisk struct {
	ID  int32 `json:"id" binding:"-"`
	VID int32 `json:"vid" binding:"-"`

	SideToSideBalance *int32 `json:"sideToSideBalance" binding:"required"`
	SemiTandemBalance *int32 `json:"semiTandemBalance" binding:"required"`
	TandemBalance     *int32 `json:"tandemBalance" binding:"required"`
	GaitSpeedTest     *int32 `json:"gaitSpeedTest" binding:"required"`
	ChairStandTest    *int32 `json:"chairStandTest" binding:"required"`
	FallRiskScore     *int32 `json:"fallRiskScore" binding:"required"`

	IcopeCompleteChairStands *bool `json:"icopeCompleteChairStands" binding:"required"`
	IcopeChairStandsTime     *bool `json:"icopeChairStandsTime" binding:"required"`
}

// TableName specifies the table name for the FallRisk model.
func (FallRisk) TableName() string {
	return "fallrisk"
}

// ToString generates a simple string representation of the FallRisk struct.
func (fr FallRisk) String() string {
	result := fmt.Sprintf("\nFALL RISK\n")
	result += fmt.Sprintf("ID: %d\n", fr.ID)
	result += fmt.Sprintf("VID: %d\n", fr.VID)
	result += fmt.Sprintf("SideToSideBalance: %d\n", *fr.SideToSideBalance)
	result += fmt.Sprintf("SemiTandemBalance: %d\n", *fr.SemiTandemBalance)
	result += fmt.Sprintf("TandemBalance: %d\n", *fr.TandemBalance)
	result += fmt.Sprintf("GaitSpeedTest: %d\n", *fr.GaitSpeedTest)
	result += fmt.Sprintf("ChairStandTest: %d\n", *fr.ChairStandTest)
	result += fmt.Sprintf("FallRiskScore: %d\n", *fr.FallRiskScore)
	result += fmt.Sprintf("IcopeCompleteChairStands: %t\n", *fr.IcopeCompleteChairStands)
	result += fmt.Sprintf("IcopeChairStandsTime: %t\n", *fr.IcopeChairStandsTime)
	return result
}
