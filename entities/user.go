package entities

import (
	"context"
)

// Use pointers so that some structs are optional
// User is a struct that represents a user when a user is trying to log in
type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// DBUser is a struct that represents how a user is stored in the database
type DBUser struct {
	Id           int64 `json:"id"`
	Username     string
	Name         string
	PasswordHash string
}

type LoginUseCase interface {
	Login(ctx context.Context, user User) (string, error)
}
