package analytics

import (
	"strings"
)

// GeoIPResolver defines the interface for resolving IP addresses to Geo location info.
type GeoIPResolver interface {
	Lookup(ip string) (country string, city string, err error)
}

// MockGeoIPResolver provides a mock in-memory implementation of GeoIPResolver for testing and fallback.
type MockGeoIPResolver struct {
	Data map[string][2]string
}

func (m *MockGeoIPResolver) Lookup(ip string) (string, string, error) {
	if m.Data != nil {
		if val, ok := m.Data[ip]; ok {
			return val[0], val[1], nil
		}
	}
	return "Unknown", "Unknown", nil
}

// EventEnricher enriches raw ClickEvent payloads with GeoIP and User-Agent metadata.
type EventEnricher struct {
	geoResolver GeoIPResolver
}

// NewEventEnricher initializes an EventEnricher instance.
func NewEventEnricher(geoResolver GeoIPResolver) *EventEnricher {
	if geoResolver == nil {
		geoResolver = &MockGeoIPResolver{}
	}
	return &EventEnricher{geoResolver: geoResolver}
}

// Enrich parses the User-Agent and performs GeoIP lookup on a ClickEvent.
func (e *EventEnricher) Enrich(event *ClickEvent) {
	if event == nil {
		return
	}

	// 1. User-Agent Parsing
	if event.UserAgent != "" {
		event.Browser = parseBrowser(event.UserAgent)
		event.OS = parseOS(event.UserAgent)
		event.DeviceType = parseDeviceType(event.UserAgent)
	}

	// 2. GeoIP Lookup
	if event.IPAddress != "" && (event.Country == "" || event.City == "") {
		country, city, err := e.geoResolver.Lookup(event.IPAddress)
		if err == nil {
			if event.Country == "" {
				event.Country = country
			}
			if event.City == "" {
				event.City = city
			}
		}
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
