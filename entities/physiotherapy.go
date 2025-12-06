package entities

import (
	"fmt"
)

type Physiotherapy struct {
	ID                   int32   `json:"id" binding:"-"`
	VID                  int32   `json:"vid" binding:"-"`
	SubjectiveAssessment *string `json:"subjectiveAssessment"`
	PainScale            *int32  `json:"painScale"` // 1-10
	ObjectiveAssessment  *string `json:"objectiveAssessment"`
	Intervention         *string `json:"intervention"`
	Evaluation           *string `json:"evaluation"`
}

// TableName specifies the table name for the Physiotherapy model.
func (Physiotherapy) TableName() string {
	return "physiotherapy"
}

// ToString generates a simple string representation of the Physiotherapy struct.
func (p Physiotherapy) String() string {
	result := fmt.Sprintf("\nPHYSIOTHERAPY\n")
	result += fmt.Sprintf("ID: %d\n", p.ID)
	result += fmt.Sprintf("VID: %d\n", p.VID)
	result += fmt.Sprintf("SubjectiveAssessment: %v\n", SafeDeref(p.SubjectiveAssessment))
	result += fmt.Sprintf("PainScale: %v\n", SafeDeref(p.PainScale))
	result += fmt.Sprintf("ObjectiveAssessment: %v\n", SafeDeref(p.ObjectiveAssessment))
	result += fmt.Sprintf("Intervention: %v\n", SafeDeref(p.Intervention))
	result += fmt.Sprintf("Evaluation: %v\n", SafeDeref(p.Evaluation))
	return result
}
