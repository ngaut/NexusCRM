package errors

import "errors"

// Standard API Errors
var (
	ErrInvalidRequest   = errors.New("invalid request")
	ErrNotFound         = errors.New("resource not found")
	ErrInternalServer   = errors.New("internal server error")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("access forbidden")
	ErrValidationFailed = errors.New("validation failed")
)

// Helper functions to wrap errors with context
func Invalid(msg string) error {
	return errors.New(msg) // Simple wrapper for now, could be enhanced
}
