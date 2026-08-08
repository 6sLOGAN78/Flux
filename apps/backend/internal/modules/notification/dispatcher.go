// Package notification provides multi-channel alert dispatching across Email, Slack Block Kit, Discord embeds, and In-App sessions.
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Notification represents an in-app notification entity.
type Notification struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Message   string    `json:"message" db:"message"`
	Type      string    `json:"type" db:"type"` // "info", "warning", "alert"
	LinkURL   string    `json:"link_url,omitempty" db:"link_url"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NotificationPayload represents a multi-channel notification dispatch payload.
type NotificationPayload struct {
	Channel   string `json:"channel"`   // "email", "slack", "discord", "in_app"
	Recipient string `json:"recipient"` // Email, UserID UUID string, or Webhook URL
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type,omitempty"` // "info", "warning", "alert"
	LinkURL   string `json:"link_url,omitempty"`
}

// NotificationDispatcher dispatches notifications to configured channels.
type NotificationDispatcher struct {
	mu         sync.RWMutex
	inAppStore map[string][]Notification
	httpClient *http.Client
}

// NewDispatcher initializes a NotificationDispatcher.
func NewDispatcher() *NotificationDispatcher {
	return &NotificationDispatcher{
		inAppStore: make(map[string][]Notification),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// BuildSlackBlockKit constructs a Slack Block Kit formatted JSON payload.
func BuildSlackBlockKit(payload NotificationPayload) map[string]any {
	headerText := fmt.Sprintf("*%s*", payload.Title)
	if payload.Type == "alert" {
		headerText = fmt.Sprintf("🚨 *%s*", payload.Title)
	} else if payload.Type == "warning" {
		headerText = fmt.Sprintf("⚠️ *%s*", payload.Title)
	}

	blocks := []any{
		map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": headerText,
			},
		},
		map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": payload.Message,
			},
		},
	}

	if payload.LinkURL != "" {
		blocks = append(blocks, map[string]any{
			"type": "actions",
			"elements": []any{
				map[string]any{
					"type": "button",
					"text": map[string]any{
						"type": "plain_text",
						"text": "View in Dashboard",
					},
					"url": payload.LinkURL,
				},
			},
		})
	}

	return map[string]any{"blocks": blocks}
}

// BuildDiscordEmbed constructs a Discord webhook embed JSON payload.
func BuildDiscordEmbed(payload NotificationPayload) map[string]any {
	color := 0x3b82f6 // Default blue
	if payload.Type == "alert" {
		color = 0xef4444 // Red
	} else if payload.Type == "warning" {
		color = 0xf59e0b // Amber
	}

	embed := map[string]any{
		"title":       payload.Title,
		"description": payload.Message,
		"color":       color,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	if payload.LinkURL != "" {
		embed["url"] = payload.LinkURL
	}

	return map[string]any{"embeds": []any{embed}}
}

// Send dispatches a notification to its target channel.
func (d *NotificationDispatcher) Send(payload NotificationPayload) error {
	channel := strings.ToLower(strings.TrimSpace(payload.Channel))

	switch channel {
	case "slack":
		bodyMap := BuildSlackBlockKit(payload)
		return d.postJSON(payload.Recipient, bodyMap)
	case "discord":
		bodyMap := BuildDiscordEmbed(payload)
		return d.postJSON(payload.Recipient, bodyMap)
	case "in_app":
		return d.storeInApp(payload)
	case "email":
		// Transactional email dispatch simulation
		return nil
	default:
		return fmt.Errorf("unsupported notification channel %q", payload.Channel)
	}
}

func (d *NotificationDispatcher) postJSON(url string, data map[string]any) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := d.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned HTTP status %d", resp.StatusCode)
	}
	return nil
}

func (d *NotificationDispatcher) storeInApp(payload NotificationPayload) error {
	uID, err := uuid.Parse(payload.Recipient)
	if err != nil {
		uID = uuid.New()
	}

	n := Notification{
		ID:        uuid.New(),
		UserID:    uID,
		Title:     payload.Title,
		Message:   payload.Message,
		Type:      payload.Type,
		LinkURL:   payload.LinkURL,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.inAppStore[payload.Recipient] = append(d.inAppStore[payload.Recipient], n)
	return nil
}

// GetInAppNotifications returns all in-app notifications stored for a specific user ID.
func (d *NotificationDispatcher) GetInAppNotifications(userID string) []Notification {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.inAppStore[userID]
}
