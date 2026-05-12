package domain

import "fmt"

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

type AppError struct {
	Code    ErrorCode
	Message string
	Details string
	Err     error 
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(message, details string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: message,
		Details: details,
	}
}

func NewConflictError(message, details string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
		Details: details,
	}
}

func NewValidationError(message, details string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
	}
}

func NewRateLimitedError(retryAfter string) *AppError {
	return &AppError{
		Code:    ErrCodeRateLimited,
		Message: "rate limited - too many requests",
		Details: retryAfter,
	}
}

func IsAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}
