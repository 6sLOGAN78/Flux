package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/handler"

	"github.com/labstack/echo/v4"
)

func TestHealthHandler_NilDBPool(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	healthH := handler.NewHealthHandler(nil)
	err := healthH.HealthCheck(c)
	if err != nil {
		t.Fatalf("unexpected error running HealthCheck: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200, got %d", rec.Code)
	}
}

func TestRedirectHandler_MissingSlug(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	redirectH := handler.NewRedirectHandler(nil, nil)
	err := redirectH.HandleRedirect(c)
	if err == nil {
		t.Fatal("expected HTTP error for missing slug parameter")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusNotFound {
		t.Errorf("expected HTTP 404 for missing slug, got %v", err)
	}
}

func TestAnalyticsHandler_MissingProvider(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("tenant_id", "mock_tenant")

	analyticsH := handler.NewAnalyticsHandler(nil)
	err := analyticsH.GetSummary(c)
	if err == nil {
		t.Fatal("expected HTTP error for uninitialized analytics provider")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusInternalServerError {
		t.Errorf("expected HTTP 500 when provider is nil, got %v", err)
	}
}
