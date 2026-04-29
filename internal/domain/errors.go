package domain

import "fmt"

// Custom error types for the application
type ErrorCode string

const (
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeRateLimited  ErrorCode = "RATE_LIMITED"
)

// AppError represents a structured application error.
type AppError struct {
	Code    ErrorCode
	Message string
	Details string
	Err     error // underlying error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Constructor functions for common errors

// NewNotFoundError creates a NOT_FOUND error.
func NewNotFoundError(message, details string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: message,
		Details: details,
	}
}

// NewConflictError creates a CONFLICT error.
func NewConflictError(message, details string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
		Details: details,
	}
}

// NewValidationError creates a VALIDATION_ERROR.
func NewValidationError(message, details string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

// NewUnauthorizedError creates an UNAUTHORIZED error.
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates a FORBIDDEN error.
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

// NewInternalError creates an INTERNAL_ERROR with underlying cause.
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
	}
}

// NewRateLimitedError creates a RATE_LIMITED error.
func NewRateLimitedError(retryAfter string) *AppError {
	return &AppError{
		Code:    ErrCodeRateLimited,
		Message: "rate limited - too many requests",
		Details: retryAfter,
	}
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
