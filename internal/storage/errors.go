package storage

import "errors"

// Repository layer errors - specific to storage operations
var (
	// User-related errors
	ErrUserExists       = errors.New("user already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Ad-related errors
	ErrAdExists         = errors.New("ad already exists")
	ErrAdNotFound       = errors.New("ad not found")

	// Relationship errors
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	ErrInvalidUserReference = errors.New("invalid user reference")
)
