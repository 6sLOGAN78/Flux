package redirect_test

import (
	"testing"
	"time"

	"flux/apps/backend/internal/modules/redirect"

	"github.com/google/uuid"
)

func TestSmartRouter_GeoMatching(t *testing.T) {
	router := redirect.NewSmartRouter()

	linkID := uuid.New()
	defaultURL := "https://example.com/global"

	rules := []redirect.RedirectRule{
		{
			ID:           uuid.New(),
			LinkID:       linkID,
			RuleType:     "geo",
			Priority:     1,
			ConditionKey: "US",
			TargetURL:    "https://example.com/us",
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			LinkID:       linkID,
			RuleType:     "geo",
			Priority:     2,
			ConditionKey: "DE",
			TargetURL:    "https://example.com/de",
			IsActive:     true,
		},
	}

	// Match US
	metaUS := redirect.RequestMetadata{
		CountryCode: "US",
	}
	target := router.EvaluateRules(rules, metaUS, defaultURL)
	if target != "https://example.com/us" {
		t.Errorf("expected US target 'https://example.com/us', got '%s'", target)
	}

	// Match DE
	metaDE := redirect.RequestMetadata{
		CountryCode: "de", // test case-insensitivity
	}
	targetDE := router.EvaluateRules(rules, metaDE, defaultURL)
	if targetDE != "https://example.com/de" {
		t.Errorf("expected DE target 'https://example.com/de', got '%s'", targetDE)
	}

	// Fallback to default for unmatched country
	metaFR := redirect.RequestMetadata{
		CountryCode: "FR",
	}
	targetFallback := router.EvaluateRules(rules, metaFR, defaultURL)
	if targetFallback != defaultURL {
		t.Errorf("expected fallback default target '%s', got '%s'", defaultURL, targetFallback)
	}
}

func TestSmartRouter_DeviceMatching(t *testing.T) {
	router := redirect.NewSmartRouter()

	linkID := uuid.New()
	defaultURL := "https://example.com/web"

	rules := []redirect.RedirectRule{
		{
			ID:           uuid.New(),
			LinkID:       linkID,
			RuleType:     "device",
			Priority:     1,
			ConditionKey: "iOS",
			TargetURL:    "https://apps.apple.com/app/id123",
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			LinkID:       linkID,
			RuleType:     "device",
			Priority:     2,
			ConditionKey: "Android",
			TargetURL:    "https://play.google.com/store/apps/id123",
			IsActive:     true,
		},
	}

	// Match iOS
	metaIOS := redirect.RequestMetadata{
		OS:         "iOS",
		DeviceType: "mobile",
	}
	targetIOS := router.EvaluateRules(rules, metaIOS, defaultURL)
	if targetIOS != "https://apps.apple.com/app/id123" {
		t.Errorf("expected iOS target URL, got '%s'", targetIOS)
	}

	// Match Android
	metaAndroid := redirect.RequestMetadata{
		OS:         "Android",
		DeviceType: "mobile",
	}
	targetAndroid := router.EvaluateRules(rules, metaAndroid, defaultURL)
	if targetAndroid != "https://play.google.com/store/apps/id123" {
		t.Errorf("expected Android target URL, got '%s'", targetAndroid)
	}
}

func TestSmartRouter_TimeWindowMatching(t *testing.T) {
	router := redirect.NewSmartRouter()

	linkID := uuid.New()
	defaultURL := "https://example.com/regular"

	now := time.Now()
	startTime := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := now.Add(1 * time.Hour).Format(time.RFC3339)

	rules := []redirect.RedirectRule{
		{
			ID:           uuid.New(),
			LinkID:       linkID,
			RuleType:     "time",
			Priority:     1,
			ConditionKey: startTime + "|" + endTime,
			TargetURL:    "https://example.com/limited-promo",
			IsActive:     true,
		},
	}

	metaNow := redirect.RequestMetadata{
		Timestamp: now,
	}

	target := router.EvaluateRules(rules, metaNow, defaultURL)
	if target != "https://example.com/limited-promo" {
		t.Errorf("expected promo URL during active time window, got '%s'", target)
	}
}
