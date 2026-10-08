package errs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/6sLOGAN78/flux/internal/errs"
)

func TestHTTPErrorMatchesStableCode(t *testing.T) {
	var nilHTTP *errs.HTTPError
	for _, tc := range []struct {
		source error
		target error
		name   string
		want   bool
	}{
		{source: &errs.HTTPError{Code: "NOT_FOUND"},
			target: &errs.HTTPError{Code: "NOT_FOUND", Message: "different message"}, name: "same code", want: true},
		{source: &errs.HTTPError{Code: "NOT_FOUND"},
			target: &errs.HTTPError{Code: "UNAUTHORIZED"}, name: "different code"},
		{source: fmt.Errorf("operation: %w", &errs.HTTPError{Code: "NOT_FOUND"}),
			target: &errs.HTTPError{Code: "NOT_FOUND"}, name: "wrapped code", want: true},
		{source: fmt.Errorf("operation: %w", &errs.HTTPError{Code: "NOT_FOUND"}),
			target: &errs.HTTPError{Code: "UNAUTHORIZED"}, name: "wrapped different"},
		{source: &errs.HTTPError{Code: "NOT_FOUND"}, target: nilHTTP, name: "typed nil target"},
		{source: nilHTTP, target: &errs.HTTPError{Code: "NOT_FOUND"}, name: "typed nil receiver"},
		{source: &errs.HTTPError{Code: "NOT_FOUND"}, target: errors.New("NOT_FOUND"), name: "other type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := errors.Is(tc.source, tc.target); got != tc.want {
				t.Fatalf("stable code match = %v, want %v", got, tc.want)
			}
		})
	}
}
