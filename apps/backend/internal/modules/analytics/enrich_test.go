package analytics_test

import (
	"testing"

	"flux/apps/backend/internal/modules/analytics"
)

func TestEventEnricher_UserAgentParsing(t *testing.T) {
	enricher := analytics.NewEventEnricher(nil)

	testCases := []struct {
		name         string
		ua           string
		wantBrowser  string
		wantOS       string
		wantDevice   string
	}{
		{
			name:        "Chrome on Windows Desktop",
			ua:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			wantBrowser: "Chrome",
			wantOS:      "Windows",
			wantDevice:  "Desktop",
		},
		{
			name:        "Safari on iPhone",
			ua:          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			wantBrowser: "Safari",
			wantOS:      "iOS",
			wantDevice:  "Mobile",
		},
		{
			name:        "Firefox on Mac",
			ua:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/119.0",
			wantBrowser: "Firefox",
			wantOS:      "macOS",
			wantDevice:  "Desktop",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &analytics.ClickEvent{
				UserAgent: tc.ua,
				IPAddress: "8.8.8.8",
			}
			enricher.Enrich(event)

			if event.Browser != tc.wantBrowser {
				t.Errorf("expected Browser '%s', got '%s'", tc.wantBrowser, event.Browser)
			}
			if event.OS != tc.wantOS {
				t.Errorf("expected OS '%s', got '%s'", tc.wantOS, event.OS)
			}
			if event.DeviceType != tc.wantDevice {
				t.Errorf("expected DeviceType '%s', got '%s'", tc.wantDevice, event.DeviceType)
			}
		})
	}
}

func TestEventEnricher_GeoIPLookup(t *testing.T) {
	mockGeo := &analytics.MockGeoIPResolver{
		Data: map[string][2]string{
			"8.8.8.8": {"US", "Mountain View"},
			"1.1.1.1": {"AU", "Sydney"},
		},
	}
	enricher := analytics.NewEventEnricher(mockGeo)

	event := &analytics.ClickEvent{
		IPAddress: "8.8.8.8",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0",
	}

	enricher.Enrich(event)

	if event.Country != "US" {
		t.Errorf("expected Country 'US', got '%s'", event.Country)
	}

	if event.City != "Mountain View" {
		t.Errorf("expected City 'Mountain View', got '%s'", event.City)
	}
}
