// Package errs defines the public HTTP error envelope and its constructors.
package errs

import (
	"strings"
)

// FieldError identifies a rejected request field and its safe validation message.
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// ActionType identifies a client recovery action.
type ActionType string

// ActionTypeRedirect directs clients to a recovery URL.
const (
	ActionTypeRedirect ActionType = "redirect"
)

// Action describes an optional client recovery step.
type Action struct {
	Type    ActionType `json:"type"`
	Message string     `json:"message"`
	Value   string     `json:"value"`
}

// HTTPError contains the public error status, code, and safe details.
type HTTPError struct {
	Action   *Action      `json:"action"`
	Code     string       `json:"code"`
	Message  string       `json:"message"`
	Errors   []FieldError `json:"errors"`
	Status   int          `json:"status"`
	Override bool         `json:"override"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

// Is matches HTTP errors by their stable error code.
func (e *HTTPError) Is(target error) bool {
	other, ok := target.(*HTTPError)
	return ok && e != nil && other != nil && e.Code == other.Code
}

// WithMessage returns an error copy with an overridden public message.
func (e *HTTPError) WithMessage(message string) *HTTPError {
	return &HTTPError{
		Code:     e.Code,
		Message:  message,
		Status:   e.Status,
		Override: e.Override,
		Errors:   e.Errors,
		Action:   e.Action,
	}
}

// MakeUpperCaseWithUnderscores converts camel case messages to stable error codes.
func MakeUpperCaseWithUnderscores(str string) string {
	return strings.ToUpper(strings.ReplaceAll(str, " ", "_"))
}
