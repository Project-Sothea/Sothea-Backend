package postgres

import (
	"context"

	"sothea-backend/controllers/middleware"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Small interface both *pgxpool.Pool and pgx.Tx satisfy:
type DBRunner interface {
	Exec(ctx context.Context, q string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, q string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, q string, args ...any) pgx.Row
}

// Only when you truly need the pgx.Tx (multi-statement)
func TxFromCtx(ctx context.Context) (pgx.Tx, bool) {
	if txAny, ok := middleware.GetTx(ctx); ok {
		if tx, ok2 := txAny.(pgx.Tx); ok2 {
			return tx, true
		}
	}
	return nil, false
}
