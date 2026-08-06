package analytics_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/modules/analytics"

	"github.com/labstack/echo/v4"
)

// MockAnalyticsProvider implements analytics summary & metrics querying for unit tests
type MockAnalyticsProvider struct{}

func (m *MockAnalyticsProvider) GetSummary(ctx context.Context, userID string, page, limit int) (*analytics.AnalyticsSummaryResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	return &analytics.AnalyticsSummaryResponse{
		TotalClicks: 1500,
		UniqueUsers: 1200,
		TopCountries: map[string]int64{
			"US": 800,
			"IN": 400,
			"DE": 300,
		},
		TopBrowsers: map[string]int64{
			"Chrome": 1000,
			"Safari": 500,
		},
		TopDevices: map[string]int64{
			"Desktop": 900,
			"Mobile":  600,
		},
		Page:  page,
		Limit: limit,
	}, nil
}

func (m *MockAnalyticsProvider) GetLinkMetrics(ctx context.Context, linkID string) (*analytics.LinkMetricsResponse, error) {
	if linkID == "unknown" {
		return nil, analytics.ErrLinkMetricsNotFound
	}
	return &analytics.LinkMetricsResponse{
		LinkID:      linkID,
		TotalClicks: 450,
		DailyStats: map[string]int64{
			"2026-08-05": 200,
			"2026-08-06": 250,
		},
	}, nil
}

func TestAnalyticsAPI_GetSummary(t *testing.T) {
	provider := &MockAnalyticsProvider{}
	handler := analytics.NewAnalyticsHandler(provider)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/analytics/summary?page=1&limit=20", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", "usr_test123")

	err := handler.GetSummary(c)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 OK, got %d", rec.Code)
	}
}

func TestAnalyticsAPI_GetLinkMetrics(t *testing.T) {
	provider := &MockAnalyticsProvider{}
	handler := analytics.NewAnalyticsHandler(provider)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/analytics/links/link_999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("link_999")

	err := handler.GetLinkMetrics(c)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 OK, got %d", rec.Code)
	}
}

func TestAnalyticsAPI_GetLinkMetrics_NotFound(t *testing.T) {
	provider := &MockAnalyticsProvider{}
	handler := analytics.NewAnalyticsHandler(provider)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/analytics/links/unknown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("unknown")

	err := handler.GetLinkMetrics(c)
	if err == nil {
		t.Fatalf("expected error for non-existent link, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusNotFound {
		t.Errorf("expected HTTP 404 Not Found, got: %v", err)
	}
}
