package usecases

import (
	"context"
	"time"

	"github.com/jieqiboh/sothea_backend/controllers/middleware"
	"github.com/jieqiboh/sothea_backend/entities"
	"golang.org/x/crypto/bcrypt"
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

func (l *loginUsecase) Login(ctx context.Context, user entities.User) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, l.contextTimeout)
	defer cancel()

	dbUser, err := l.patientRepo.GetDBUser(ctx, user.Username)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.PasswordHash), []byte(user.Password))
	if err != nil {
		return "", entities.ErrLoginFailed
	}

	token, err := middleware.CreateToken(dbUser.Id, user.Username, l.secretKey)
	if err != nil {
		return "", err
	}
	return token, err
}
