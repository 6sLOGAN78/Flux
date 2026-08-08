// Package integration implements 3rd-party SaaS integrations including GA4 Measurement Protocol, Zapier, and Shopify webhooks.
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// GA4Event holds Google Analytics 4 Measurement Protocol event attributes.
type GA4Event struct {
	MeasurementID string         `json:"measurement_id"`
	APISecret     string         `json:"api_secret"`
	ClientID      string         `json:"client_id"`
	EventName     string         `json:"event_name"`
	Params        map[string]any `json:"params,omitempty"`
}

// GA4Client sends server-side measurement events to GA4 API endpoint.
type GA4Client struct {
	Endpoint string
	HTTP     *http.Client
}

// NewGA4Client creates a new GA4 Client instance.
func NewGA4Client(endpoint string) *GA4Client {
	if endpoint == "" {
		endpoint = "https://www.google-analytics.com/mp/collect"
	}
	return &GA4Client{
		Endpoint: endpoint,
		HTTP: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// BuildGA4Payload constructs GA4 Measurement Protocol v2 compliant JSON payload.
func BuildGA4Payload(event GA4Event) map[string]any {
	return map[string]any{
		"client_id": event.ClientID,
		"events": []map[string]any{
			{
				"name":   event.EventName,
				"params": event.Params,
			},
		},
	}
}

// SendGA4Event sends event payload to GA4 Measurement Protocol HTTP API.
func (c *GA4Client) SendGA4Event(event GA4Event) error {
	payload := BuildGA4Payload(event)
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GA4 payload: %w", err)
	}

	u, err := url.Parse(c.Endpoint)
	if err != nil {
		return fmt.Errorf("invalid GA4 endpoint URL: %w", err)
	}

	q := u.Query()
	if event.MeasurementID != "" {
		q.Set("measurement_id", event.MeasurementID)
	}
	if event.APISecret != "" {
		q.Set("api_secret", event.APISecret)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create GA4 request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("failed to dispatch GA4 HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("GA4 Measurement Protocol responded with status: %d", resp.StatusCode)
	}

	return nil
}

// FormatZapierHook formats link creation or click events into Zapier REST hook compatible JSON structure.
func FormatZapierHook(eventType string, shortCode string, destinationURL string) map[string]any {
	return map[string]any{
		"event_type":      eventType,
		"short_code":      shortCode,
		"destination_url": destinationURL,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}
}
