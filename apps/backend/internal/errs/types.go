package errs

import (
	"net/http"
)

// NewUnauthorizedError constructs an authentication failure.
func NewUnauthorizedError(message string, override bool) *HTTPError {
	return &HTTPError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusUnauthorized)),
		Message:  message,
		Status:   http.StatusUnauthorized,
		Override: override,
	}
}

// NewForbiddenError constructs an authorization failure.
func NewForbiddenError(message string, override bool) *HTTPError {
	return &HTTPError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusForbidden)),
		Message:  message,
		Status:   http.StatusForbidden,
		Override: override,
	}
}

// NewBadRequestError constructs an invalid request failure.
func NewBadRequestError(message string, override bool, code *string, errors []FieldError, action *Action) *HTTPError {
	formattedCode := MakeUpperCaseWithUnderscores(http.StatusText(http.StatusBadRequest))

	if code != nil {
		formattedCode = *code
	}

	return &HTTPError{
		Code:     formattedCode,
		Message:  message,
		Status:   http.StatusBadRequest,
		Override: override,
		Errors:   errors,
		Action:   action,
	}
}

// NewNotFoundError constructs a missing resource failure.
func NewNotFoundError(message string, override bool, code *string) *HTTPError {
	formattedCode := MakeUpperCaseWithUnderscores(http.StatusText(http.StatusNotFound))

	if code != nil {
		formattedCode = *code
	}

	return &HTTPError{
		Code:     formattedCode,
		Message:  message,
		Status:   http.StatusNotFound,
		Override: override,
	}
}

// NewInternalServerError constructs a safe internal failure.
func NewInternalServerError() *HTTPError {
	return &HTTPError{
		Code:     MakeUpperCaseWithUnderscores(http.StatusText(http.StatusInternalServerError)),
		Message:  http.StatusText(http.StatusInternalServerError),
		Status:   http.StatusInternalServerError,
		Override: false,
	}
}

// ValidationError constructs a request validation failure with field details.
func ValidationError(err error) *HTTPError {
	return NewBadRequestError("Validation failed: "+err.Error(), false, nil, nil, nil)
}
