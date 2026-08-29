package service

import (
	"context"
	"encoding/json"
	"log"

	"flux/apps/backend/internal/model/analytics"
)

// LogAnalyticsPublisher is a dummy publisher that logs events to stdout.
// This satisfies the Phase 1 Analytics Pipeline Wiring requirement without requiring Redis or ClickHouse.
type LogAnalyticsPublisher struct{}

func NewLogAnalyticsPublisher() *LogAnalyticsPublisher {
	return &LogAnalyticsPublisher{}
}

func (p *LogAnalyticsPublisher) PublishEvent(ctx context.Context, event *analytics.AnalyticsEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to marshal analytics event: %v", err)
		return err
	}
	log.Printf("ANALYTICS EVENT PUBLISHED: %s", string(data))
	return nil
}
