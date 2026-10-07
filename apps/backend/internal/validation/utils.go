// Package validation translates external request validation to safe public field errors.
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

const validationFailureMessage = "Validation failed"

// Validatable requires request validation before handler execution.
type Validatable interface {
	Validate() error
}

// CustomValidationError contains a safe field validation failure.
type CustomValidationError struct {
	Field   string
	Message string
}

// CustomValidationErrors collects field validation failures.
type CustomValidationErrors []CustomValidationError

func (c CustomValidationErrors) Error() string {
	return validationFailureMessage
}

// BindAndValidate binds a request and translates validation failures to the public error envelope.
func BindAndValidate(c echo.Context, payload Validatable) error {
	if isNil(payload) {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}
	if err := c.Bind(payload); err != nil {
		return errs.NewBadRequestError("Invalid request", false, nil, nil, nil)
	}

	if msg, fieldErrors := validateStruct(payload); msg != "" {
		return errs.NewBadRequestError(msg, true, nil, fieldErrors, nil)
	}

	return nil
}

func isNil(value interface{}) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.String, reflect.Struct, reflect.UnsafePointer:
		return false
	default:
		return false
	}
}

func validateStruct(v Validatable) (string, []errs.FieldError) {
	if err := v.Validate(); err != nil {
		return extractValidationErrors(err)
	}
	return "", nil
}

func extractValidationErrors(err error) (string, []errs.FieldError) {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return validationFailureMessage, customFieldErrors(err)
	}
	var fieldErrors []errs.FieldError

	for _, err := range validationErrors {
		if isNil(err) {
			continue
		}
		msg := fieldValidationMessage(err)

		fieldErrors = append(fieldErrors, errs.FieldError{
			Field: strings.ToLower(err.Field()),
			Error: msg,
		})
	}

	return validationFailureMessage, fieldErrors
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidUUID reports whether a string uses canonical UUID syntax.
func IsValidUUID(uuid string) bool {
	return uuidRegex.MatchString(uuid)
}

func fieldValidationMessage(err validator.FieldError) string {
	field := strings.ToLower(err.Field())
	var msg string

	switch err.Tag() {
	case "required":
		msg = "is required"
	case "min":
		if err.Type().Kind() == reflect.String {
			msg = fmt.Sprintf("must be at least %s characters", err.Param())
		} else {
			msg = fmt.Sprintf("must be at least %s", err.Param())
		}
	case "max":
		if err.Type().Kind() == reflect.String {
			msg = fmt.Sprintf("must not exceed %s characters", err.Param())
		} else {
			msg = fmt.Sprintf("must not exceed %s", err.Param())
		}
	case "oneof":
		msg = fmt.Sprintf("must be one of: %s", err.Param())
	case "email":
		msg = "must be a valid email address"
	case "e164":
		msg = "must be a valid phone number with country code"
	case "uuid":
		msg = "must be a valid UUID"
	case "uuidList":
		msg = "must be a comma-separated list of valid UUIDs"
	case "dive":
		msg = "some items are invalid"
	default:
		if err.Param() != "" {
			msg = fmt.Sprintf("%s: %s:%s", field, err.Tag(), err.Param())
		} else {
			msg = fmt.Sprintf("%s: %s", field, err.Tag())
		}
	}

	return msg
}

func customFieldErrors(err error) []errs.FieldError {
	var values CustomValidationErrors
	if !errors.As(err, &values) {
		var pointer *CustomValidationErrors
		if errors.As(err, &pointer) && pointer != nil {
			values = *pointer
		}
	}
	var fields []errs.FieldError
	for _, value := range values {
		fields = append(fields, errs.FieldError{Field: value.Field, Error: value.Message})
	}
	return fields
}
