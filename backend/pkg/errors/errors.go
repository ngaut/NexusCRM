package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is the base interface for all application errors
type AppError interface {
	error
	HTTPStatus() int
	Code() string
}

// NotFoundError represents a resource that was not found
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) HTTPStatus() int {
	return http.StatusNotFound
}

func (e *NotFoundError) Code() string {
	return "NOT_FOUND"
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

// ValidationError represents invalid input
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) HTTPStatus() int {
	return http.StatusBadRequest
}

func (e *ValidationError) Code() string {
	return "VALIDATION_ERROR"
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// PermissionError represents insufficient permissions
type PermissionError struct {
	Action   string
	Resource string
	UserID   string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: cannot %s %s", e.Action, e.Resource)
}

func (e *PermissionError) HTTPStatus() int {
	return http.StatusForbidden
}

func (e *PermissionError) Code() string {
	return "PERMISSION_DENIED"
}

// NewPermissionError creates a new PermissionError
func NewPermissionError(action, resource string) *PermissionError {
	return &PermissionError{Action: action, Resource: resource}
}

// UnauthorizedError represents authentication failures
type UnauthorizedError struct {
	Reason string
}

func (e *UnauthorizedError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("unauthorized: %s", e.Reason)
	}
	return "unauthorized"
}

func (e *UnauthorizedError) HTTPStatus() int {
	return http.StatusUnauthorized
}

func (e *UnauthorizedError) Code() string {
	return "UNAUTHORIZED"
}

// NewUnauthorizedError creates a new UnauthorizedError
func NewUnauthorizedError(reason string) *UnauthorizedError {
	return &UnauthorizedError{Reason: reason}
}

// ConflictError represents a conflict with existing data
type ConflictError struct {
	Resource string
	Field    string
	Value    string
}

func (e *ConflictError) Error() string {
	if e.Field != "" && e.Value != "" {
		return fmt.Sprintf("%s already exists with %s='%s'", e.Resource, e.Field, e.Value)
	}
	return fmt.Sprintf("%s already exists", e.Resource)
}

func (e *ConflictError) HTTPStatus() int {
	return http.StatusConflict
}

func (e *ConflictError) Code() string {
	return "CONFLICT"
}

// NewConflictError creates a new ConflictError
func NewConflictError(resource, field, value string) *ConflictError {
	return &ConflictError{Resource: resource, Field: field, Value: value}
}

// InternalError represents unexpected server errors
type InternalError struct {
	Message string
	Cause   error
}

func (e *InternalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("internal error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("internal error: %s", e.Message)
}

func (e *InternalError) HTTPStatus() int {
	return http.StatusInternalServerError
}

func (e *InternalError) Code() string {
	return "INTERNAL_ERROR"
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}

// NewInternalError creates a new InternalError
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{Message: message, Cause: cause}
}

// Helper functions for error checking

// IsNotFound checks if an error is a NotFoundError
func IsNotFound(err error) bool {
	var notFound *NotFoundError
	return errors.As(err, &notFound)
}

// IsValidation checks if an error is a ValidationError
func IsValidation(err error) bool {
	var validation *ValidationError
	return errors.As(err, &validation)
}

// IsPermission checks if an error is a PermissionError
func IsPermission(err error) bool {
	var permission *PermissionError
	return errors.As(err, &permission)
}

// IsUnauthorized checks if an error is an UnauthorizedError
func IsUnauthorized(err error) bool {
	var unauthorized *UnauthorizedError
	return errors.As(err, &unauthorized)
}

// IsConflict checks if an error is a ConflictError
func IsConflict(err error) bool {
	var conflict *ConflictError
	return errors.As(err, &conflict)
}

// GetHTTPStatus returns the HTTP status code for an error
// Returns 500 if the error doesn't implement AppError
func GetHTTPStatus(err error) int {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code for an error
// Returns "UNKNOWN_ERROR" if the error doesn't implement AppError
func GetErrorCode(err error) string {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr.Code()
	}
	return "UNKNOWN_ERROR"
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ToResponse converts an error to an ErrorResponse
func ToResponse(err error) ErrorResponse {
	return ErrorResponse{
		Code:    GetErrorCode(err),
		Message: err.Error(),
	}
}
