package validation_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/validation"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type validationRequest struct{ err error }

func (r *validationRequest) Validate() error { return r.err }

type failingBinder struct{ err error }

func (b failingBinder) Bind(interface{}, echo.Context) error { return b.err }

func bindSafely(t *testing.T, e *echo.Echo, payload validation.Validatable, body string) error {
	t.Helper()
	defer func() {
		if value := recover(); value != nil {
			t.Errorf("BindAndValidate panicked: %v", value)
		}
	}()
	r := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
	r.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return validation.BindAndValidate(e.NewContext(r, httptest.NewRecorder()), payload)
}

func assertBadRequest(t *testing.T, err error, message string, fields []errs.FieldError) {
	t.Helper()
	var httpErr *errs.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected typed bad request, received %v", err)
	}
	if httpErr.Status != http.StatusBadRequest || httpErr.Code != "BAD_REQUEST" || httpErr.Message != message {
		t.Errorf("unexpected error envelope: %+v", httpErr)
	}
	if !reflect.DeepEqual(httpErr.Errors, fields) {
		t.Errorf("field errors = %+v, want %+v", httpErr.Errors, fields)
	}
	body, marshalErr := json.Marshal(httpErr)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if strings.Contains(string(body), "secret-marker") {
		t.Errorf("diagnostic leaked: %s", body)
	}
}

func TestBindErrorsAreSanitized(t *testing.T) {
	t.Parallel()
	for _, err := range []error{errors.New("secret-marker"),
		errors.New("code=400, message=secret-marker"),
		echo.NewHTTPError(400,
			"secret-marker")} {
		t.Run(err.Error(), func(t *testing.T) {
			t.Parallel()
			e := echo.New()
			e.Binder = failingBinder{err: err}
			assertBadRequest(t, bindSafely(t, e, &validationRequest{}, "{}"), "Invalid request", nil)
		})
	}
}

func TestMalformedJSONIsSanitized(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`{"value":`, `{"secret-marker"`} {
		assertBadRequest(t, bindSafely(t, echo.New(), &validationRequest{}, body), "Invalid request", nil)
	}
}

func TestValidationErrorShapes(t *testing.T) {
	t.Parallel()
	v := validator.New()
	validationErr := v.Struct(struct {
		Email string `validate:"required,email"`
	}{})
	custom := validation.CustomValidationErrors{{Field: "destination", Message: "is invalid"}}
	cases := []struct {
		name   string
		err    error
		fields []errs.FieldError
	}{
		{name: "ordinary", err: errors.New("secret-marker"), fields: nil},
		{name: "invalid_validator_input", err: v.Struct(42), fields: nil},
		{name: "validator",
			err: validationErr,
			fields: []errs.FieldError{{Field: "email",
				Error: "is required"}}},
		{name: "wrapped_validator",
			err: fmt.Errorf("secret-marker: %w",
				validationErr),
			fields: []errs.FieldError{{Field: "email",
				Error: "is required"}}},
		{name: "custom", err: custom, fields: []errs.FieldError{{Field: "destination", Error: "is invalid"}}},
		{name: "wrapped_custom",
			err: fmt.Errorf("secret-marker: %w",
				custom),
			fields: []errs.FieldError{{Field: "destination",
				Error: "is invalid"}}},
		{name: "custom_pointer",
			err: &custom,
			fields: []errs.FieldError{{Field: "destination",
				Error: "is invalid"}}},
		{name: "nil_custom_slice", err: validation.CustomValidationErrors(nil), fields: nil},
		{name: "empty_custom_slice", err: validation.CustomValidationErrors{}, fields: nil},
		{name: "nil_custom_pointer", err: (*validation.CustomValidationErrors)(nil), fields: nil},
		{name: "nil_validator_slice", err: validator.ValidationErrors(nil), fields: nil},
		{name: "nil_validator_field", err: validator.ValidationErrors{nil}, fields: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertBadRequest(t,
				bindSafely(t,
					echo.New(),
					&validationRequest{err: tc.err},
					"{}"),
				"Validation failed",
				tc.fields)
		})
	}
}

func TestValidPayloadSucceeds(t *testing.T) {
	t.Parallel()
	if err := bindSafely(t, echo.New(), &validationRequest{}, "{}"); err != nil {
		t.Fatal(err)
	}
}

func TestNilPayloadIsRejected(t *testing.T) {
	t.Parallel()
	var typedNil *validationRequest
	for _, payload := range []validation.Validatable{nil, typedNil} {
		assertBadRequest(t, bindSafely(t, echo.New(), payload, "{}"), "Invalid request", nil)
	}
}
