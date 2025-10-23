package postgres

import (
	"context"
	"database/sql"

	"github.com/jieqiboh/sothea_backend/controllers/middleware"
	"github.com/jieqiboh/sothea_backend/entities"
)

// Small interface both *sql.DB and *sql.Tx satisfy:
type DBRunner interface {
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
}

// Get a DB-like thing bound to the tx if present; else fall back to the pool.
func DBFromCtx(ctx context.Context, db *sql.DB) DBRunner {
	if tx, ok := middleware.GetTx(ctx); ok && tx != nil {
		return tx
	}
	return db
}

// Only when you truly need the *sql.Tx (multi-statement)
func TxFromCtx(ctx context.Context) (*sql.Tx, bool) {
	return middleware.GetTx(ctx)
}

func CtxWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return middleware.CtxWithTx(ctx, tx)
}

func getUserByID(dbx DBRunner, id int64) (*entities.DBUser, error) {
	user := &entities.DBUser{}
	err := dbx.QueryRowContext(context.Background(), `
		SELECT id, username, password_hash FROM users
		WHERE id = $1
	`, id).Scan(&user.Id, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}
