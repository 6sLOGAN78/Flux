// Package webhook provides outbound HTTP callback delivery, HMAC-SHA256 signatures, and delivery log tracking.
package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Webhook represents a registered webhook endpoint configuration.
type Webhook struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id" db:"workspace_id"`
	URL         string    `json:"url" db:"url"`
	Secret      string    `json:"secret" db:"secret"`
	Events      []string  `json:"events" db:"events"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// WebhookDelivery represents an execution audit log entry for a delivered webhook payload.
type WebhookDelivery struct {
	ID              uuid.UUID      `json:"id" db:"id"`
	WebhookID       uuid.UUID      `json:"webhook_id" db:"webhook_id"`
	EventType       string         `json:"event_type" db:"event_type"`
	Payload         map[string]any `json:"payload" db:"payload"`
	ResponseStatus  int            `json:"response_status" db:"response_status"`
	ResponseBody    string         `json:"response_body,omitempty" db:"response_body"`
	ExecutionTimeMS int64          `json:"execution_time_ms" db:"execution_time_ms"`
	Status          string         `json:"status" db:"status"` // "success", "failed", "retrying"
	CreatedAt       time.Time      `json:"created_at" db:"created_at"`
}

// ComputeSignature calculates the HMAC-SHA256 signature for a webhook payload.
func ComputeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies if an incoming HMAC-SHA256 signature matches expected payload signature.
func VerifySignature(payload []byte, secret, signature string) bool {
	expectedSig := ComputeSignature(payload, secret)
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// DeliverWebhook executes an outbound HTTP POST request delivering signed webhook payload to subscriber URL.
func DeliverWebhook(wb Webhook, eventType string, payload map[string]any) (WebhookDelivery, error) {
	delivery := WebhookDelivery{
		ID:        uuid.New(),
		WebhookID: wb.ID,
		EventType: eventType,
		Payload:   payload,
		Status:    "failed",
		CreatedAt: time.Now(),
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		delivery.ResponseBody = fmt.Sprintf("marshal error: %v", err)
		return delivery, err
	}

	sig := ComputeSignature(jsonBytes, wb.Secret)

	req, err := http.NewRequest("POST", wb.URL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		delivery.ResponseBody = fmt.Sprintf("request creation error: %v", err)
		return delivery, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Flux-Signature", sig)
	req.Header.Set("X-Flux-Event", eventType)
	req.Header.Set("User-Agent", "Flux-WebhookDispatcher/2.0")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	resp, err := client.Do(req)
	delivery.ExecutionTimeMS = time.Since(start).Milliseconds()

	if err != nil {
		delivery.ResponseBody = fmt.Sprintf("http dispatch error: %v", err)
		return delivery, err
	}
	defer resp.Body.Close()

	delivery.ResponseStatus = resp.StatusCode
	respBody, _ := io.ReadAll(resp.Body)
	delivery.ResponseBody = string(respBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = "success"
		return delivery, nil
	}

	return delivery, fmt.Errorf("webhook endpoint returned HTTP %d", resp.StatusCode)
}
