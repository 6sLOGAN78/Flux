package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/model/analytics"

	"github.com/labstack/echo/v4"
)

type mockAnalyticsProvider struct{}

func (m *mockAnalyticsProvider) GetSummary(ctx context.Context, workspaceID string, from, to time.Time) (*analytics.AnalyticsSummary, error) {
	if workspaceID == "mock_ws" {
		return &analytics.AnalyticsSummary{TotalClicks: 42, UniqueVisitors: 10}, nil
	}
	return &analytics.AnalyticsSummary{}, nil
}

func (m *mockAnalyticsProvider) GetTimeseries(ctx context.Context, workspaceID string, from, to time.Time, interval string) (*analytics.TimeseriesResponse, error) {
	return &analytics.TimeseriesResponse{Data: []analytics.TimeseriesDataPoint{}}, nil
}

func (m *mockAnalyticsProvider) GetTopLinks(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.TopLinksResponse, error) {
	return &analytics.TopLinksResponse{Data: []analytics.TopLink{}}, nil
}

func (m *mockAnalyticsProvider) GetReferrers(ctx context.Context, workspaceID string, from, to time.Time, limit int) (*analytics.ReferrersResponse, error) {
	return &analytics.ReferrersResponse{Data: []analytics.ReferrerStat{}}, nil
}

func setupAnalyticsHandler() (*echo.Echo, *handler.AnalyticsHandler) {
	e := echo.New()
	provider := &mockAnalyticsProvider{}
	h := handler.NewAnalyticsHandler(provider)
	return e, h
}

func TestAnalyticsHandler_GetSummary_Unauthenticated(t *testing.T) {
	e, h := setupAnalyticsHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetSummary(c)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %v", err)
	}
}

func TestAnalyticsHandler_GetSummary_Authenticated(t *testing.T) {
	e, h := setupAnalyticsHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("tenant_id", "mock_ws")

	err := h.GetSummary(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var res analytics.AnalyticsSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.TotalClicks != 42 {
		t.Errorf("expected 42 clicks, got %d", res.TotalClicks)
	}
}

func TestAnalyticsHandler_DateRangeParsing(t *testing.T) {
	e, h := setupAnalyticsHandler()
	
	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary?from=2026-01-02T00:00:00Z&to=2026-01-01T00:00:00Z", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("tenant_id", "mock_ws")

	err := h.GetSummary(c)
	if err == nil {
		t.Fatalf("expected error for reversed dates")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %v", err)
	}
	
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary?from=2020-01-01T00:00:00Z&to=2026-01-01T00:00:00Z", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.Set("tenant_id", "mock_ws")

	err2 := h.GetSummary(c2)
	if err2 == nil {
		t.Fatalf("expected error for >1 year range")
	}
	he2, ok2 := err2.(*echo.HTTPError)
	if !ok2 || he2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %v", err2)
	}
}
