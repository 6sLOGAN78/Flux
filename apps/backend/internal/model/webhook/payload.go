package webhook

import "time"

// WebhookEventPayload represents the stable outbound webhook envelope.
// We conceptually align with the Stripe webhook envelope structure.
type WebhookEventPayload struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	Data      interface{} `json:"data"`
}
