package webhook_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/modules/webhook"

	"github.com/google/uuid"
)

func TestComputeAndVerifySignature(t *testing.T) {
	secret := "whsec_test_secret_key_12345"
	payload := []byte(`{"event":"link.created","link_id":"123"}`)

	sig := webhook.ComputeSignature(payload, secret)
	if sig == "" {
		t.Fatal("expected non-empty HMAC signature")
	}

	if !webhook.VerifySignature(payload, secret, sig) {
		t.Error("expected signature verification to pass")
	}

	if webhook.VerifySignature(payload, "wrong_secret", sig) {
		t.Error("expected signature verification to fail with wrong secret")
	}
}

func TestDeliverWebhook_Success(t *testing.T) {
	var receivedSig, receivedEvent string

	// Create test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedSig = r.Header.Get("X-Flux-Signature")
		receivedEvent = r.Header.Get("X-Flux-Event")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	wb := webhook.Webhook{
		ID:          uuid.New(),
		WorkspaceID: uuid.New(),
		URL:         ts.URL,
		Secret:      "whsec_secret999",
		Events:      []string{"link.created"},
		IsActive:    true,
	}

	payload := map[string]any{"link_id": "xyz_999", "url": "https://example.com"}
	delivery, err := webhook.DeliverWebhook(wb, "link.created", payload)
	if err != nil {
		t.Fatalf("unexpected delivery error: %v", err)
	}

	if delivery.Status != "success" {
		t.Errorf("expected delivery status 'success', got %q", delivery.Status)
	}
	if delivery.ResponseStatus != 200 {
		t.Errorf("expected response status 200, got %d", delivery.ResponseStatus)
	}
	if receivedEvent != "link.created" {
		t.Errorf("expected X-Flux-Event header 'link.created', got %q", receivedEvent)
	}
	if receivedSig == "" {
		t.Error("expected X-Flux-Signature header to be set")
	}
}

func TestDeliverWebhook_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal_server_error"}`))
	}))
	defer ts.Close()

	wb := webhook.Webhook{
		ID:          uuid.New(),
		WorkspaceID: uuid.New(),
		URL:         ts.URL,
		Secret:      "whsec_secret999",
		Events:      []string{"click.recorded"},
		IsActive:    true,
	}

	payload := map[string]any{"click_id": "clk_123"}
	delivery, err := webhook.DeliverWebhook(wb, "click.recorded", payload)
	if err == nil {
		t.Error("expected error on HTTP 500 response")
	}

	if delivery.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", delivery.Status)
	}
	if delivery.ResponseStatus != 500 {
		t.Errorf("expected response status 500, got %d", delivery.ResponseStatus)
	}
}
