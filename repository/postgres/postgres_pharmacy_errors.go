package postgres

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type DuplicateBatchNumberError struct {
	DrugID      int64  `json:"drug_id"`
	BatchNumber string `json:"batch_number"`
}

func (e *DuplicateBatchNumberError) Error() string {
	return fmt.Sprintf("Batch number '%s' already exists for this drug", e.BatchNumber)
}

func (e *DuplicateBatchNumberError) Code() string { return "DUPLICATE_BATCH_NUMBER" }

func mapPharmacySQLError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	// Unique violation on (drug_id, batch_number)
	if pgErr.Code == "23505" && pgErr.ConstraintName == "drug_batches_drug_id_batch_number_key" {
		batchNum := extractBatchNumber(pgErr.Detail)
		return &DuplicateBatchNumberError{
			DrugID:      0, // not provided by pgErr; caller knows the drug_id in context
			BatchNumber: batchNum,
		}
	}

	return err
}

// detail example: "Key (drug_id, batch_number)=(1, PAN500-A) already exists."
func extractBatchNumber(detail string) string {
	parts := strings.Split(detail, ")=(")
	if len(parts) < 2 {
		return ""
	}
	rest := parts[1]
	rest = strings.TrimSuffix(rest, ") already exists.")
	fields := strings.Split(rest, ",")
	if len(fields) < 2 {
		return ""
	}
	return strings.TrimSpace(fields[1])
}
