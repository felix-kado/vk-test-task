package storage

import "errors"

var (
	// ErrExists is returned when a resource already exists.
	ErrExists = errors.New("already exists")
	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")
	// ErrInvalidCredentials is returned when provided credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
)
