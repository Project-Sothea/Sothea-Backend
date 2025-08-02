package entities

import "errors"

var (
	ErrInternalServerError  = errors.New("Internal Server Error")
	ErrPatientNotFound      = errors.New("Patient Not Found")
	ErrPatientVisitNotFound = errors.New("Patient Visit Not Found")
	ErrMissingAdminCategory = errors.New("Missing Admin field")
	ErrAuthenticationFailed = errors.New("Not Authenticated")
	ErrLoginFailed          = errors.New("Login Failed")

	ErrDrugNameTaken = errors.New("A drug with that name already exists")
)
