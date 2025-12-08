package usecases

import (
	"context"
	"errors"
	"time"

	"sothea-backend/controllers/middleware"
	"sothea-backend/entities"
	"sothea-backend/repository/postgres"

	"github.com/jackc/pgx/v5"
)

type LoginUsecase struct {
	patientRepo    *postgres.PostgresPatientRepository
	contextTimeout time.Duration
	secretKey      []byte
}

// NewLoginUseCase builds a login usecase backed by the patient repository.
func NewLoginUseCase(p *postgres.PostgresPatientRepository, timeout time.Duration, secretKey []byte) *LoginUsecase {
	return &LoginUsecase{
		patientRepo:    p,
		contextTimeout: timeout,
		secretKey:      secretKey,
	}
}

func (l *LoginUsecase) Login(ctx context.Context, user entities.LoginPayload) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, l.contextTimeout)
	defer cancel()

	dbUser, err := l.patientRepo.GetDBUser(ctx, user.Username)
	if err != nil {
		return "", err
	}

	token, err := middleware.CreateToken(dbUser.ID, user.Username, l.secretKey)
	if err != nil {
		return "", err
	}
	return token, err
}
