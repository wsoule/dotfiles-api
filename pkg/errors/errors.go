package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrCodeValidation     ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeConflict       ErrorCode = "CONFLICT"
	ErrCodeInternal       ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeRateLimit      ErrorCode = "RATE_LIMIT"
	ErrCodeInvalidToken   ErrorCode = "INVALID_TOKEN"
	ErrCodeExpiredToken   ErrorCode = "EXPIRED_TOKEN"
)

type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"status_code"`
	Internal   error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeUnauthorized,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeForbidden,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeConflict,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

func NewInternalError(message string, internal error) *AppError {
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Internal:   internal,
	}
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

func NewRateLimitError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeRateLimit,
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

func NewInvalidTokenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeInvalidToken,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewExpiredTokenError(message string) *AppError {
	return &AppError{
		Code:       ErrCodeExpiredToken,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}