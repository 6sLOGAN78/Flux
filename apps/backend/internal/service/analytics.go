package service

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"flux/apps/backend/internal/model/analytics"
	"flux/apps/backend/internal/repository"
)

// AsyncCollector collects click events asynchronously without blocking the caller thread.
type AsyncCollector struct {
	producer  repository.EventProducer
	eventChan chan *analytics.ClickEvent
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewAsyncCollector initializes a new AsyncCollector instance.
func NewAsyncCollector(producer repository.EventProducer, bufferSize int) *AsyncCollector {
	if bufferSize <= 0 {
		bufferSize = 5000
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncCollector{
		producer:  producer,
		eventChan: make(chan *analytics.ClickEvent, bufferSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start launches the background worker goroutine loop.
func (c *AsyncCollector) Start(ctx context.Context) {
	c.wg.Add(1)
	go c.workerLoop()
}

// Stop gracefully shuts down the background worker and drains pending events.
func (c *AsyncCollector) Stop() {
	c.cancel()
	close(c.eventChan)
	c.wg.Wait()
}

// CollectAsync pushes a click event to the internal buffered channel non-blockingly.
func (c *AsyncCollector) CollectAsync(event *analytics.ClickEvent) {
	if event == nil {
		return
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	select {
	case c.eventChan <- event:
	default:
		log.Printf("warning: click event buffer full, dropping event for slug '%s'", event.Slug)
	}
}

func (c *AsyncCollector) workerLoop() {
	defer c.wg.Done()

	for event := range c.eventChan {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if c.producer != nil {
			_ = c.producer.Publish(ctx, event)
		}
		cancel()
	}
}

// EventEnricher enriches raw ClickEvent payloads with GeoIP and User-Agent metadata.
type EventEnricher struct{}

func NewEventEnricher() *EventEnricher {
	return &EventEnricher{}
}

func (e *EventEnricher) Enrich(event *analytics.ClickEvent) {
	if event == nil {
		return
	}
	if event.UserAgent != "" {
		event.Browser = parseBrowser(event.UserAgent)
		event.OS = parseOS(event.UserAgent)
		event.DeviceType = parseDeviceType(event.UserAgent)
	}
}

func parseBrowser(ua string) string {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "edg"):
		return "Edge"
	case strings.Contains(lower, "firefox"):
		return "Firefox"
	case strings.Contains(lower, "opr") || strings.Contains(lower, "opera"):
		return "Opera"
	case strings.Contains(lower, "chrome"):
		return "Chrome"
	case strings.Contains(lower, "safari"):
		return "Safari"
	default:
		return "Unknown"
	}
}

func parseOS(ua string) string {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "cpu os"):
		return "iOS"
	case strings.Contains(lower, "android"):
		return "Android"
	case strings.Contains(lower, "windows"):
		return "Windows"
	case strings.Contains(lower, "macintosh") || strings.Contains(lower, "mac os x"):
		return "macOS"
	case strings.Contains(lower, "linux"):
		return "Linux"
	default:
		return "Unknown"
	}
}

func parseDeviceType(ua string) string {
	lower := strings.ToLower(ua)
	switch {
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet"):
		return "Tablet"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "mobile") || strings.Contains(lower, "android"):
		return "Mobile"
	default:
		return "Desktop"
	}
}
