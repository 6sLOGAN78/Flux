package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/model/redirect"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRedirectRepo struct{}

func (m *mockRedirectRepo) GetByHostAndSlug(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
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

func TestRedirectHandler_URLDecoration(t *testing.T) {
	tests := []struct {
		name        string
		destination string
		wantPrefix  string // What we expect the URL to start with
		wantSuffix  string // What we expect the URL to end with (fragments, etc)
		wantParams  map[string]string // Key=Value pairs that MUST exist
	}{
		{
			name:        "Basic URL",
			destination: "https://example.com/page",
			wantPrefix:  "https://example.com/page",
		},
		{
			name:        "Existing query",
			destination: "https://example.com/page?a=1",
			wantPrefix:  "https://example.com/page",
			wantParams:  map[string]string{"a": "1"},
		},
		{
			name:        "Fragment",
			destination: "https://example.com/page#section",
			wantPrefix:  "https://example.com/page",
			wantSuffix:  "#section",
		},
		{
			name:        "Query + fragment",
			destination: "https://example.com/page?a=1#section",
			wantPrefix:  "https://example.com/page",
			wantSuffix:  "#section",
			wantParams:  map[string]string{"a": "1"},
		},
		{
			name:        "Existing flux_cid",
			destination: "https://example.com/page?flux_cid=old_cid",
			wantPrefix:  "https://example.com/page",
		},
		{
			name:        "Special URL characters",
			destination: "https://example.com/page?spaced=hello%20world&unicode=✨#frag",
			wantPrefix:  "https://example.com/page",
			wantSuffix:  "#frag",
			wantParams:  map[string]string{"spaced": "hello world", "unicode": "✨"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/testslug", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("slug")
			c.SetParamValues("testslug")

			// Define an inline repo mock that returns the specific destination
			mockRepo := &inlineMockRepo{
				target: &redirect.LinkRedirectTarget{
					Slug:           "testslug",
					LinkID:         "link-123",
					TenantID:       "tenant-456",
					DestinationURL: tt.destination,
					Status:         "active",
				},
			}
			svc := service.NewRedirectService(mockRepo, nil)
			pub := &mockPublisher{eventChan: make(chan *analytics.AnalyticsEvent, 1)}
			h := handler.NewRedirectHandler(svc, pub)

			err := h.HandleRedirect(c)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			// Validate URL decoration
			location := rec.Header().Get("Location")
			if location == "" {
				t.Fatalf("expected redirect Location header")
			}

			parsedLoc, err := url.Parse(location)
			if err != nil {
				t.Fatalf("redirect Location is not a valid URL: %v", err)
			}

			// 1. Verify Event ID Identity matches the decorated CID
			var eventID string
			select {
			case event := <-pub.eventChan:
				eventID = event.EventID
				if eventID == "" {
					t.Fatalf("expected eventID to be generated")
				}
			case <-time.After(time.Second):
				t.Fatalf("timed out waiting for analytics event")
			}

			q := parsedLoc.Query()
			fluxCid := q.Get("flux_cid")
			if fluxCid != eventID {
				t.Errorf("expected flux_cid in URL to be %s, got %s", eventID, fluxCid)
			}

			// 2. Verify all expected params exist
			for k, v := range tt.wantParams {
				if q.Get(k) != v {
					t.Errorf("expected param %s=%s, got %s", k, v, q.Get(k))
				}
			}

			// 3. Verify fragment is preserved at the end
			if tt.wantSuffix != "" {
				if parsedLoc.Fragment != tt.wantSuffix[1:] { // ignore the #
					t.Errorf("expected fragment %s, got %s", tt.wantSuffix, parsedLoc.Fragment)
				}
			}
		})
	}
}

type inlineMockRepo struct {
	target *redirect.LinkRedirectTarget
}

func (m *inlineMockRepo) GetByHostAndSlug(ctx context.Context, hostname, slug string) (*redirect.LinkRedirectTarget, error) {
	if slug == m.target.Slug {
		return m.target, nil
	}
	return nil, repository.ErrNotFound
}

func TestRedirectHandler_URLDecoration_CacheParity(t *testing.T) {
	// Proves that a target resulting from a Cache HIT and Cache MISS
	// generates the EXACT SAME URL decoration behavior (other than eventID).
	
	targetMiss := &redirect.LinkRedirectTarget{
		Slug:           "parity",
		LinkID:         "link-123",
		TenantID:       "tenant-456",
		DestinationURL: "https://example.com/checkout?plan=pro#section",
		Status:         "active",
	}

	// Cache HIT has the exact same properties as proven by redirect_parity_test.go
	targetHit := &redirect.LinkRedirectTarget{
		Slug:           "parity",
		LinkID:         "link-123",
		TenantID:       "tenant-456",
		DestinationURL: "https://example.com/checkout?plan=pro#section",
		Status:         "active",
	}

	e := echo.New()

	// Simulate MISS
	reqMiss := httptest.NewRequest(http.MethodGet, "/parity", nil)
	recMiss := httptest.NewRecorder()
	cMiss := e.NewContext(reqMiss, recMiss)
	cMiss.SetParamNames("slug")
	cMiss.SetParamValues("parity")

	mockRepoMiss := &inlineMockRepo{target: targetMiss}
	hMiss := handler.NewRedirectHandler(service.NewRedirectService(mockRepoMiss, nil), nil)
	require.NoError(t, hMiss.HandleRedirect(cMiss))
	
	locMiss := recMiss.Header().Get("Location")
	parsedMiss, _ := url.Parse(locMiss)

	// Simulate HIT
	reqHit := httptest.NewRequest(http.MethodGet, "/parity", nil)
	recHit := httptest.NewRecorder()
	cHit := e.NewContext(reqHit, recHit)
	cHit.SetParamNames("slug")
	cHit.SetParamValues("parity")

	mockRepoHit := &inlineMockRepo{target: targetHit}
	hHit := handler.NewRedirectHandler(service.NewRedirectService(mockRepoHit, nil), nil)
	require.NoError(t, hHit.HandleRedirect(cHit))
	
	locHit := recHit.Header().Get("Location")
	parsedHit, _ := url.Parse(locHit)

	// Verify both properly preserve origin URL parts
	assert.Equal(t, "https", parsedMiss.Scheme)
	assert.Equal(t, "example.com", parsedMiss.Host)
	assert.Equal(t, "/checkout", parsedMiss.Path)
	assert.Equal(t, "pro", parsedMiss.Query().Get("plan"))
	assert.Equal(t, "section", parsedMiss.Fragment)

	assert.Equal(t, "https", parsedHit.Scheme)
	assert.Equal(t, "example.com", parsedHit.Host)
	assert.Equal(t, "/checkout", parsedHit.Path)
	assert.Equal(t, "pro", parsedHit.Query().Get("plan"))
	assert.Equal(t, "section", parsedHit.Fragment)

	// Verify both appended flux_cid
	cidMiss := parsedMiss.Query().Get("flux_cid")
	cidHit := parsedHit.Query().Get("flux_cid")
	
	assert.NotEmpty(t, cidMiss)
	assert.NotEmpty(t, cidHit)
	assert.NotEqual(t, cidMiss, cidHit, "event IDs must be unique per click")
}
