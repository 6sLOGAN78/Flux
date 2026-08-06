package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	customMiddleware "flux/apps/backend/internal/middleware"

	"github.com/labstack/echo/v4"
)

// MockLimiterStore implements LimiterStore in memory for testing
type MockLimiterStore struct {
	counts map[string]int
}

func NewMockLimiterStore() *MockLimiterStore {
	return &MockLimiterStore{counts: make(map[string]int)}
}

func (m *MockLimiterStore) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
	current := m.counts[key]
	if current >= limit {
		return false, 0, window, nil
	}
	m.counts[key] = current + 1
	remaining := limit - (current + 1)
	return true, remaining, window, nil
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	store := NewMockLimiterStore()
	middleware := customMiddleware.RateLimitMiddleware(store, 5, time.Minute)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("expected request to be allowed, got error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 OK, got: %d", rec.Code)
	}

	if rec.Header().Get("X-RateLimit-Limit") != "5" {
		t.Errorf("expected X-RateLimit-Limit '5', got '%s'", rec.Header().Get("X-RateLimit-Limit"))
	}
}

func TestRateLimitMiddleware_Exceeded_HTTP429(t *testing.T) {
	store := NewMockLimiterStore()
	// Limit is set to 2 requests
	middleware := customMiddleware.RateLimitMiddleware(store, 2, time.Minute)

	e := echo.New()
	dummyHandler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	// Request 1: Allowed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	_ = dummyHandler(e.NewContext(req1, rec1))

	// Request 2: Allowed
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	_ = dummyHandler(e.NewContext(req2, rec2))

	// Request 3: Exceeded -> Should return HTTP 429
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err := dummyHandler(c3)
	if err == nil {
		t.Fatalf("expected HTTP 429 error on 3rd request, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusTooManyRequests {
		t.Errorf("expected HTTP 429 Too Many Requests, got: %v", err)
	}
}
