package postgres

import (
	"context"

	db "sothea-backend/repository/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	Conn    *pgxpool.Pool
	queries *db.Queries
}

func NewPostgresUserRepository(conn *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{
		Conn:    conn,
		queries: db.New(conn),
	}
}

func (r *PostgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*db.User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &db.User{
		ID:       row.ID,
		Username: row.Username,
		Name:     row.Name,
	}, nil
}

func (r *PostgresUserRepository) ListUsers(ctx context.Context) ([]db.User, error) {
	rows, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]db.User, len(rows))
	for i, row := range rows {
		users[i] = db.User{
			ID:       row.ID,
			Username: row.Username,
			Name:     row.Name,
		}
	}
	return users, nil
}
