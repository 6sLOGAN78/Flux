package testing

import (
	"context"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo/v4"
)

// NewTestContext initializes a mock Echo Context for HTTP handler testing.
func NewTestContext(e *echo.Echo, method, path string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// TimeoutContext returns a context with a standard startup/teardown timeout limit.
func TimeoutContext(d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		d = 10 * time.Second
	}
	return context.WithTimeout(context.Background(), d)
}
