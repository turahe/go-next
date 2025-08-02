package responses

import (
	"errors"
	"net/http"
)

// Custom error types for better error handling
var (
	ErrNotFound     = errors.New("resource not found")
	ErrValidation   = errors.New("validation error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
)

// CustomError represents a custom error with status code
type CustomError struct {
	Message    string
	StatusCode int
}

func (e *CustomError) Error() string {
	return e.Message
}

// NewError creates a new custom error
func NewError(message string, statusCode int) error {
	return &CustomError{
		Message:    message,
		StatusCode: statusCode,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message string) error {
	return &CustomError{
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) error {
	return &CustomError{
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) error {
	return &CustomError{
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) error {
	return &CustomError{
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) error {
	return &CustomError{
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// Error checking functions
func IsNotFoundError(err error) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.StatusCode == http.StatusNotFound
	}
	return errors.Is(err, ErrNotFound)
}

func IsValidationError(err error) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.StatusCode == http.StatusBadRequest
	}
	return errors.Is(err, ErrValidation)
}

func IsUnauthorizedError(err error) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.StatusCode == http.StatusUnauthorized
	}
	return errors.Is(err, ErrUnauthorized)
}

func IsForbiddenError(err error) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.StatusCode == http.StatusForbidden
	}
	return errors.Is(err, ErrForbidden)
}

func IsConflictError(err error) bool {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.StatusCode == http.StatusConflict
	}
	return errors.Is(err, ErrConflict)
}

// GetErrorStatusCode returns the status code for an error
func GetErrorStatusCode(err error) int {
	var customErr *CustomError
	if errors.As(err, &customErr) {
		return customErr.StatusCode
	}
	return http.StatusInternalServerError
}
