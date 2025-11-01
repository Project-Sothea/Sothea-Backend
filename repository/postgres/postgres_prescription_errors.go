package postgres

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type InsufficientStockError struct {
	PresentationID int64 `json:"presentation+id"`
	TotalRequired  int64 `json:"total_required"`
	TotalAvailable int64 `json:"total_available"`
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("Insufficient stock: %d requested, only %d available", e.TotalRequired, e.TotalAvailable)
}
func (e *InsufficientStockError) Code() string { return "INSUFFICIENT_STOCK" }

func mapPrescriptionSQLError(err error) error {
	var pe *pq.Error
	if !errors.As(err, &pe) {
		return err
	}
	// 23514 = check_violation; you raised with CONSTRAINT 'ck_insufficient_stock'
	if string(pe.Code) == "23514" && pe.Constraint == "ck_insufficient_stock" {
		var d InsufficientStockError
		_ = json.Unmarshal([]byte(pe.Detail), &d)
		return &d
	}
	return err
}
