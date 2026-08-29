package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/redirect"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"

	"github.com/labstack/echo/v4"
)

type mockRedirectRepo struct{}

func (m *mockRedirectRepo) GetBySlug(ctx context.Context, slug string) (*redirect.LinkRedirectTarget, error) {
	if slug == "exists" {
		return &redirect.LinkRedirectTarget{
			Slug:           "exists",
			LinkID:         "link-123",
			TenantID:       "tenant-456",
			DestinationURL: "https://example.com",
			Status:         "active",
		}, nil
	}
	return nil, repository.ErrNotFound
}

type mockPublisher struct {
	eventChan chan *analytics.AnalyticsEvent
	fail      bool
}

func (m *mockPublisher) PublishEvent(ctx context.Context, event *analytics.AnalyticsEvent) error {
	if m.fail {
		// Even if failing, we can push to channel to signal it ran
		m.eventChan <- event
		return errors.New("simulated publisher failure")
	}
	m.eventChan <- event
	return nil
}

func TestRedirectHandler_AnalyticsEventGeneration(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/exists", nil)
	req.Header.Set("User-Agent", "TestAgent")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("exists")

	repo := &mockRedirectRepo{}
	svc := service.NewRedirectService(repo, nil)
	
	// Create mock publisher with a channel to catch the async event
	pub := &mockPublisher{eventChan: make(chan *analytics.AnalyticsEvent, 1)}
	h := handler.NewRedirectHandler(svc, pub)

	err := h.HandleRedirect(c)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if rec.Code != http.StatusMovedPermanently && rec.Code != http.StatusFound {
		t.Errorf("expected redirect status, got: %d", rec.Code)
	}

	// Verify the async event
	select {
	case event := <-pub.eventChan:
		if event.LinkID != "link-123" {
			t.Errorf("expected LinkID 'link-123', got %s", event.LinkID)
		}
		if event.WorkspaceID != "tenant-456" {
			t.Errorf("expected WorkspaceID 'tenant-456', got %s", event.WorkspaceID)
		}
		if event.EventType != analytics.EventTypeLinkRedirect {
			t.Errorf("expected EventType '%s', got %s", analytics.EventTypeLinkRedirect, event.EventType)
		}
		if event.UserAgent != "TestAgent" {
			t.Errorf("expected UserAgent 'TestAgent', got %s", event.UserAgent)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for analytics event to be published")
	}
}

func TestRedirectHandler_AnalyticsPublisherFailureIsolation(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/exists", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("slug")
	c.SetParamValues("exists")

	repo := &mockRedirectRepo{}
	svc := service.NewRedirectService(repo, nil)
	
	pub := &mockPublisher{
		eventChan: make(chan *analytics.AnalyticsEvent, 1),
		fail:      true, // Simulate Redis/ClickHouse failure
	}
	h := handler.NewRedirectHandler(svc, pub)

	err := h.HandleRedirect(c)
	if err != nil {
		t.Fatalf("expected redirect to succeed despite analytics failure, got error: %v", err)
	}

	if rec.Code != http.StatusMovedPermanently && rec.Code != http.StatusFound {
		t.Errorf("expected redirect status despite analytics failure, got: %d", rec.Code)
	}

	// Wait for async call to finish
	select {
	case <-pub.eventChan:
		// Passed
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for analytics publisher call")
	}
}
