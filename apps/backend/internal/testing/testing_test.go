package testing_test

import (
	"errors"
	"testing"
	"time"

	customTesting "flux/apps/backend/internal/testing"

	"github.com/labstack/echo/v4"
)

func TestAssertions(t *testing.T) {
	customTesting.AssertNoError(t, nil, "should pass with nil error")
	customTesting.AssertError(t, errors.New("sample error"), "should pass with non-nil error")
	customTesting.AssertEqual(t, 200, 200, "status codes match")
	customTesting.AssertStatus(t, 200, 200)
}

func TestHelpers(t *testing.T) {
	e := echo.New()
	c, rec := customTesting.NewTestContext(e, "GET", "/test", nil)

	if c == nil || rec == nil {
		t.Fatalf("expected non-nil Echo test context and recorder")
	}

	ctx, cancel := customTesting.TimeoutContext(5 * time.Second)
	defer cancel()

	if ctx == nil {
		t.Fatalf("expected non-nil context")
	}
}
