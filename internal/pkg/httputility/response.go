package httputility

import (
	"encoding/json"
	"net/http"

	"go-usersvc-demo/internal/domain"

	"github.com/gin-gonic/gin"
)

// ErrorResponse is the standard error response structure.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse is the standard success response structure.
type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// RespondError sends a structured error response.
func RespondError(c *gin.Context, err error) {
	if err == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "unknown error",
		})
		return
	}

	// Check if it's an AppError
	appErr, ok := domain.IsAppError(err)
	if ok {
		statusCode := mapErrorCodeToStatus(appErr.Code)
		c.JSON(statusCode, ErrorResponse{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		})
		return
	}

	// Default to internal error for non-AppError
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "an unexpected error occurred",
	})
}

// RespondSuccess sends a success response.
func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	if statusCode == http.StatusNoContent {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(statusCode, data)
}

// mapErrorCodeToStatus maps error codes to HTTP status codes.
func mapErrorCodeToStatus(code domain.ErrorCode) int {
	switch code {
	case domain.ErrCodeNotFound:
		return http.StatusNotFound
	case domain.ErrCodeConflict:
		return http.StatusConflict
	case domain.ErrCodeValidation:
		return http.StatusUnprocessableEntity
	case domain.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case domain.ErrCodeForbidden:
		return http.StatusForbidden
	case domain.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	case domain.ErrCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// ValidateAndRespond validates a struct and responds with error if invalid.
func ValidateAndRespond(c *gin.Context, v interface{}, validate func(interface{}) error) bool {
	if err := validate(v); err != nil {
		RespondError(c, domain.NewValidationError("validation failed", err.Error()))
		return false
	}
	return true
}

// WriteJSON writes JSON to response.
func WriteJSON(c *gin.Context, statusCode int, v interface{}) error {
	c.Header("Content-Type", "application/json")
	c.Writer.WriteHeader(statusCode)
	return json.NewEncoder(c.Writer).Encode(v)
}
