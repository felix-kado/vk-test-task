package services

import "errors"

// Service layer errors - business logic focused
var (
	// Authentication and authorization errors
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	
	// Resource errors
	ErrAdNotFound = errors.New("ad not found")
	
	// Input validation errors
	ErrInvalidInput = errors.New("invalid input")
	
	// Conflict errors
	ErrUserExists = errors.New("user already exists")
	ErrConflict   = errors.New("resource conflict")
)
