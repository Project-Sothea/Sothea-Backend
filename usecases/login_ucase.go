package usecases

import (
	"context"
	"time"

	"sothea-backend/controllers/middleware"
	"sothea-backend/entities"
	"sothea-backend/repository/postgres"
	db "sothea-backend/repository/sqlc"
)

type LoginUsecase struct {
	userRepo       *postgres.PostgresUserRepository
	contextTimeout time.Duration
	secretKey      []byte
}

// NewLoginUseCase
func NewLoginUseCase(userRepo *postgres.PostgresUserRepository, timeout time.Duration, secretKey []byte) *LoginUsecase {
	return &LoginUsecase{
		userRepo:       userRepo,
		contextTimeout: timeout,
		secretKey:      secretKey,
	}
}

func (l *LoginUsecase) Login(ctx context.Context, user entities.LoginPayload) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, l.contextTimeout)
	defer cancel()

	dbUser, err := l.userRepo.GetUserByUsername(ctx, user.Username)
	if err != nil {
		return "", err
	}

	token, err := middleware.CreateToken(dbUser.ID, user.Username, l.secretKey)
	if err != nil {
		return "", err
	}
	return token, err
}

func (l *LoginUsecase) ListUsers(ctx context.Context) ([]db.User, error) {
	ctx, cancel := context.WithTimeout(ctx, l.contextTimeout)
	defer cancel()

	return l.userRepo.ListUsers(ctx)
}
