package storage

import "errors"

// Repository layer errors - specific to storage operations
var (
	// User-related errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	
	// Ad-related errors
	ErrAdNotFound = errors.New("ad not found")
	
	// Relationship errors
	ErrInvalidUserReference = errors.New("invalid user reference")
	
	// Generic storage errors
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrConstraintViolation = errors.New("constraint violation")
)
