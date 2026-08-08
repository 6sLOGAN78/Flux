package notification_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/modules/notification"

	"github.com/google/uuid"
)

func TestBuildSlackBlockKit(t *testing.T) {
	payload := notification.NotificationPayload{
		Title:   "Link Click Threshold Alert",
		Message: "Your link 'xyz123' reached 10,000 clicks!",
		Type:    "warning",
		LinkURL: "https://flux.dev/dashboard/links/xyz123",
	}

	blockKit := notification.BuildSlackBlockKit(payload)
	if blockKit == nil {
		t.Fatal("expected non-nil Slack Block Kit map")
	}

	blocks, ok := blockKit["blocks"].([]any)
	if !ok || len(blocks) == 0 {
		t.Error("expected blocks array in Slack Block Kit payload")
	}
}

func TestBuildDiscordEmbed(t *testing.T) {
	payload := notification.NotificationPayload{
		Title:   "New Milestone Reached",
		Message: "1,000,000 total platform clicks logged!",
		Type:    "info",
	}

	embed := notification.BuildDiscordEmbed(payload)
	if embed == nil {
		t.Fatal("expected non-nil Discord embed map")
	}

	embeds, ok := embed["embeds"].([]any)
	if !ok || len(embeds) == 0 {
		t.Error("expected embeds array in Discord payload")
	}
}

func TestDispatcher_SendSlack(t *testing.T) {
	var receivedBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	dispatcher := notification.NewDispatcher()
	payload := notification.NotificationPayload{
		Channel:   "slack",
		Recipient: ts.URL,
		Title:     "System Alert",
		Message:   "High CPU usage detected",
		Type:      "alert",
	}

	err := dispatcher.Send(payload)
	if err != nil {
		t.Fatalf("unexpected error sending Slack notification: %v", err)
	}

	if receivedBody["blocks"] == nil {
		t.Error("expected Slack webhook body to contain 'blocks'")
	}
}

func TestDispatcher_SendInApp(t *testing.T) {
	dispatcher := notification.NewDispatcher()
	userID := uuid.New().String()

	payload := notification.NotificationPayload{
		Channel:   "in_app",
		Recipient: userID,
		Title:     "Welcome to Flux",
		Message:   "Workspace member invite accepted",
		Type:      "info",
	}

	err := dispatcher.Send(payload)
	if err != nil {
		t.Fatalf("unexpected error creating in-app notification: %v", err)
	}

	notifications := dispatcher.GetInAppNotifications(userID)
	if len(notifications) != 1 {
		t.Fatalf("expected 1 in-app notification for user, got %d", len(notifications))
	}
	if notifications[0].Title != "Welcome to Flux" {
		t.Errorf("unexpected notification title: %q", notifications[0].Title)
	}
}
