// Package errs defines standard application error types and HTTP status mappings.
package errs

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrForbidden         = errors.New("forbidden request")
	ErrExpired           = errors.New("resource has expired")
	ErrConflict          = errors.New("resource conflict")
	ErrRateLimitExceeded = errors.New("rate limit quota exceeded")
	ErrQuotaExceeded     = errors.New("resource quota exceeded")
	ErrInternal          = errors.New("internal server error")
)

// FieldError represents a specific input field validation failure.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// HTTPError represents a structured API HTTP error.
type HTTPError struct {
	StatusCode  int          `json:"status_code"`
	Message     string       `json:"message"`
	Code        *string      `json:"code,omitempty"`
	IsUserError bool         `json:"is_user_error"`
	FieldErrors []FieldError `json:"field_errors,omitempty"`
	Err         error        `json:"-"`
}

func (e *HTTPError) Error() string {
	if e.Code != nil {
		return fmt.Sprintf("[%s] %s (status %d)", *e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("%s (status %d)", e.Message, e.StatusCode)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

func NewBadRequestError(message string, isUserError bool, code *string, fieldErrors []FieldError, cause error) *HTTPError {
	return &HTTPError{
		StatusCode:  http.StatusBadRequest,
		Message:     message,
		Code:        code,
		IsUserError: isUserError,
		FieldErrors: fieldErrors,
		Err:         cause,
	}
}

func NewNotFoundError(message string, isUserError bool, code *string) *HTTPError {
	return &HTTPError{
		StatusCode:  http.StatusNotFound,
		Message:     message,
		Code:        code,
		IsUserError: isUserError,
		Err:         ErrNotFound,
	}
}

func NewInternalServerError() *HTTPError {
	code := "INTERNAL_SERVER_ERROR"
	return &HTTPError{
		StatusCode:  http.StatusInternalServerError,
		Message:     "An internal server error occurred",
		Code:        &code,
		IsUserError: false,
		Err:         ErrInternal,
	}
}

// AppError represents a domain error with code and message.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
