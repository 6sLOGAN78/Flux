package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/router"

	"github.com/labstack/echo/v4"
)

func TestInitRouter_HealthEndpoint(t *testing.T) {
	e := echo.New()
	router.InitRouter(e, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 on /health endpoint, got %d", rec.Code)
	}
}
