package usecases

import (
	"context"
	"time"

	"sothea-backend/controllers/middleware"
	"sothea-backend/entities"
)

type loginUsecase struct {
	patientRepo    entities.PatientRepository
	contextTimeout time.Duration
	secretKey      []byte
}

// NewLoginUseCase
func NewLoginUseCase(p entities.PatientRepository, timeout time.Duration, secretKey []byte) entities.LoginUseCase {
	return &loginUsecase{
		patientRepo:    p,
		contextTimeout: timeout,
		secretKey:      secretKey,
	}
}

func (l *loginUsecase) Login(ctx context.Context, user entities.LoginPayload) (string, error) {
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
