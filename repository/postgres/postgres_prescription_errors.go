package postgres

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type InsufficientStockError struct {
	DrugID         int64 `json:"drug_id"`
	TotalRequired  int64 `json:"total_required"`
	TotalAvailable int64 `json:"total_available"`
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("Insufficient stock: %d requested, only %d available", e.TotalRequired, e.TotalAvailable)
}
func (e *InsufficientStockError) Code() string { return "INSUFFICIENT_STOCK" }

func mapPrescriptionSQLError(err error) error {
	var pe *pgconn.PgError
	if !errors.As(err, &pe) {
		return err
	}
	// 23514 = check_violation; you raised with CONSTRAINT 'ck_insufficient_stock'
	if pe.Code == "23514" && pe.ConstraintName == "ck_insufficient_stock" {
		var d InsufficientStockError
		if err := json.Unmarshal([]byte(pe.Detail), &d); err != nil {
			// If unmarshaling fails, return the original error with detail
			return fmt.Errorf("insufficient stock: %s", pe.Detail)
		}
		return &d
	}
	return err
}
