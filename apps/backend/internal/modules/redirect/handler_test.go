package redirect_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/modules/redirect"

	"github.com/labstack/echo/v4"
)

// MockRepository implements redirect.RedirectRepository in memory
type MockRepository struct {
	links map[string]*redirect.LinkRedirectTarget
}

func (m *MockRepository) GetBySlug(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error) {
	link, exists := m.links[slug]
	if !exists {
		return nil, redirect.ErrNotFound
	}
	return link, nil
}

// MockCache implements redirect.RedirectCache in memory
type MockCache struct {
	store map[string]*redirect.LinkRedirectTarget
}

func (m *MockCache) Get(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error) {
	link, exists := m.store[slug]
	if !exists {
		return nil, errors.New("cache miss")
	}
	return link, nil
}

func (m *MockCache) Set(ctx context.Context, slug string, target *redirect.LinkRedirectTarget, ttl time.Duration) error {
	m.store[slug] = target
	return nil
}

func (m *MockCache) Delete(ctx context.Context, slug string) error {
	delete(m.store, slug)
	return nil
}

func TestRedirectHandler_ActiveLink_301(t *testing.T) {
	repo := &MockRepository{
		links: map[string]*redirect.LinkRedirectTarget{
			"openai": {
				Slug:           "openai",
				DestinationURL: "https://openai.com",
				Status:         "active",
				RedirectCode:   http.StatusMovedPermanently,
			},
		},
	}
	cache := &MockCache{store: make(map[string]*redirect.LinkRedirectTarget)}
	svc := redirect.NewRedirectService(repo, cache)
	handler := redirect.NewRedirectHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/openai", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("openai")

	err := handler.HandleRedirect(c)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("expected HTTP 301 Moved Permanently, got: %d", rec.Code)
	}

	if rec.Header().Get("Location") != "https://openai.com" {
		t.Errorf("expected Location 'https://openai.com', got '%s'", rec.Header().Get("Location"))
	}

	// Verify Cache Hit on second call
	cachedTarget, err := cache.Get(context.Background(), "openai")
	if err != nil || cachedTarget.DestinationURL != "https://openai.com" {
		t.Errorf("expected cache to be updated on miss")
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	repo := &MockRepository{links: make(map[string]*redirect.LinkRedirectTarget)}
	cache := &MockCache{store: make(map[string]*redirect.LinkRedirectTarget)}
	svc := redirect.NewRedirectService(repo, cache)
	handler := redirect.NewRedirectHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/unknownslug", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("unknownslug")

	err := handler.HandleRedirect(c)
	if err == nil {
		t.Fatalf("expected HTTP 404 error, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusNotFound {
		t.Errorf("expected HTTP 404 Not Found, got: %v", err)
	}
}

func TestRedirectHandler_ExpiredLink(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	repo := &MockRepository{
		links: map[string]*redirect.LinkRedirectTarget{
			"expiredlink": {
				Slug:           "expiredlink",
				DestinationURL: "https://example.com",
				Status:         "active",
				ExpiresAt:      &past,
				RedirectCode:   http.StatusFound,
			},
		},
	}
	cache := &MockCache{store: make(map[string]*redirect.LinkRedirectTarget)}
	svc := redirect.NewRedirectService(repo, cache)
	handler := redirect.NewRedirectHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/expiredlink", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("expiredlink")

	err := handler.HandleRedirect(c)
	if err == nil {
		t.Fatalf("expected HTTP 410 Gone error, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusGone {
		t.Errorf("expected HTTP 410 Gone, got: %v", err)
	}
}

func TestRedirectHandler_DisabledLink(t *testing.T) {
	repo := &MockRepository{
		links: map[string]*redirect.LinkRedirectTarget{
			"disabledlink": {
				Slug:           "disabledlink",
				DestinationURL: "https://example.com",
				Status:         "disabled",
				RedirectCode:   http.StatusFound,
			},
		},
	}
	cache := &MockCache{store: make(map[string]*redirect.LinkRedirectTarget)}
	svc := redirect.NewRedirectService(repo, cache)
	handler := redirect.NewRedirectHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/disabledlink", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("disabledlink")

	err := handler.HandleRedirect(c)
	if err == nil {
		t.Fatalf("expected HTTP 403 Forbidden error, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusForbidden {
		t.Errorf("expected HTTP 403 Forbidden, got: %v", err)
	}
}

func TestRedirectHandler_DeletedLink(t *testing.T) {
	repo := &MockRepository{
		links: map[string]*redirect.LinkRedirectTarget{
			"deletedlink": {
				Slug:           "deletedlink",
				DestinationURL: "https://example.com",
				Status:         "deleted",
				RedirectCode:   http.StatusMovedPermanently,
			},
		},
	}
	cache := &MockCache{store: make(map[string]*redirect.LinkRedirectTarget)}
	svc := redirect.NewRedirectService(repo, cache)
	handler := redirect.NewRedirectHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/deletedlink", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("deletedlink")

	err := handler.HandleRedirect(c)
	if err == nil {
		t.Fatalf("expected HTTP 410 Gone error, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusGone {
		t.Errorf("expected HTTP 410 Gone, got: %v", err)
	}
}

func TestRedirectHandler_ActiveLink_302(t *testing.T) {
	repo := &MockRepository{
		links: map[string]*redirect.LinkRedirectTarget{
			"temp": {
				Slug:           "temp",
				DestinationURL: "https://temp.example.com",
				Status:         "active",
				RedirectCode:   http.StatusFound,
			},
		},
	}
	cache := &MockCache{store: make(map[string]*redirect.LinkRedirectTarget)}
	svc := redirect.NewRedirectService(repo, cache)
	handler := redirect.NewRedirectHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/temp", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("temp")

	err := handler.HandleRedirect(c)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rec.Code != http.StatusFound {
		t.Errorf("expected HTTP 302 Found, got: %d", rec.Code)
	}
}

