package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/modules/integration"
)

func TestBuildGA4Payload(t *testing.T) {
	event := integration.GA4Event{
		MeasurementID: "G-TEST123456",
		APISecret:     "secret_abc999",
		ClientID:      "client_ip_123.456",
		EventName:     "click_redirect",
		Params: map[string]any{
			"short_code":      "xyz123",
			"destination_url": "https://example.com",
		},
	}

	payload := integration.BuildGA4Payload(event)
	if payload == nil {
		t.Fatal("expected non-nil GA4 payload map")
	}

	if payload["client_id"] != "client_ip_123.456" {
		t.Errorf("expected client_id 'client_ip_123.456', got %v", payload["client_id"])
	}

	events, ok := payload["events"].([]map[string]any)
	if !ok || len(events) == 0 {
		t.Fatal("expected non-empty events array in GA4 payload")
	}

	if events[0]["name"] != "click_redirect" {
		t.Errorf("expected event name 'click_redirect', got %v", events[0]["name"])
	}
}

func TestSendGA4Event_Success(t *testing.T) {
	var receivedBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusNoContent) // GA4 MP returns 204 No Content
	}))
	defer ts.Close()

	client := integration.NewGA4Client(ts.URL)
	event := integration.GA4Event{
		MeasurementID: "G-TEST123456",
		APISecret:     "secret_abc999",
		ClientID:      "client_ip_123.456",
		EventName:     "click_redirect",
		Params: map[string]any{
			"short_code": "xyz123",
		},
	}

	err := client.SendGA4Event(event)
	if err != nil {
		t.Fatalf("unexpected error sending GA4 event: %v", err)
	}

	if receivedBody["client_id"] != "client_ip_123.456" {
		t.Errorf("expected GA4 client_id payload to match, got %v", receivedBody["client_id"])
	}
}

func TestFormatZapierHook(t *testing.T) {
	zapPayload := integration.FormatZapierHook("link.created", "xyz123", "https://example.com/target")
	if zapPayload["event_type"] != "link.created" {
		t.Errorf("expected event_type 'link.created', got %v", zapPayload["event_type"])
	}
	if zapPayload["short_code"] != "xyz123" {
		t.Errorf("expected short_code 'xyz123', got %v", zapPayload["short_code"])
	}
}
