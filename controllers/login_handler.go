package controllers

import (
	"net/http"

	"sothea-backend/controllers/middleware"
	"sothea-backend/entities"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// LoginHandler represent the httphandler for patient
type LoginHandler struct {
	Usecase   entities.LoginUseCase
	secretKey []byte
}

// NewLoginHandler will initialize the resources endpoint
func NewLoginHandler(r gin.IRouter, us entities.LoginUseCase, secretKey []byte) {
	handler := &LoginHandler{
		Usecase: us,
	}
	r.POST("/login", handler.Login)
	r.GET("/login/is-valid-token", middleware.AuthRequired(secretKey), handler.IsValidToken)
}

func (l *LoginHandler) Login(c *gin.Context) {
	// username and password are in the json body
	var u entities.LoginPayload
	if err := c.ShouldBindJSON(&u); err != nil {
		// Use type assertion to check if err is of type validator.ValidationErrors
		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username or password must be a non-empty string!"})
			return // exit on first error
		} else {
			// Handle other types of errors (e.g., JSON binding errors)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	ctx := c.Request.Context()

	tokenString, err := l.Usecase.Login(ctx, u)
	if err != nil {
		if err == entities.ErrLoginFailed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": entities.ErrLoginFailed.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (l *LoginHandler) IsValidToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Valid Token"})
}
