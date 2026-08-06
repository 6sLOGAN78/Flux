// Package errs defines standard application error types and HTTP status mappings.
package errs

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrUnauthorized      = errors.New("unauthorized access")
	ErrForbidden         = errors.New("forbidden request")
	ErrExpired           = errors.New("resource has expired")
	ErrConflict          = errors.New("resource conflict")
	ErrRateLimitExceeded = errors.New("rate limit quota exceeded")
	ErrInternal          = errors.New("internal server error")
)

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
