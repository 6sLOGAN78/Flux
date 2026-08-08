package campaign_test

import (
	"testing"

	"flux/apps/backend/internal/modules/campaign"
)

func TestBuildURL_InjectsSanitizedUTMParams(t *testing.T) {
	builder := campaign.NewUTMBuilder()

	baseURL := "https://example.com/pricing?ref=promo#details"
	params := campaign.UTMParams{
		UTMSource:   " twitter ",
		UTMMedium:   "cpc ",
		UTMCampaign: "summer sale 2026",
		UTMTerm:     "link shortener",
		UTMContent:  "banner_ad",
	}

	resultURL, err := builder.BuildURL(baseURL, params)
	if err != nil {
		t.Fatalf("expected BuildURL to succeed, got error: %v", err)
	}

	expectedSubstring := "utm_source=twitter"
	if !containsSubstring(resultURL, expectedSubstring) {
		t.Errorf("expected URL to contain '%s', got '%s'", expectedSubstring, resultURL)
	}

	expectedCampaign := "utm_campaign=summer_sale_2026"
	if !containsSubstring(resultURL, expectedCampaign) {
		t.Errorf("expected URL to contain '%s', got '%s'", expectedCampaign, resultURL)
	}

	// Verify original query param is preserved
	if !containsSubstring(resultURL, "ref=promo") {
		t.Errorf("expected original query param 'ref=promo' to be preserved, got '%s'", resultURL)
	}

	// Verify URL fragment/hash is preserved at end
	if !containsSubstring(resultURL, "#details") {
		t.Errorf("expected URL hash '#details' to be preserved, got '%s'", resultURL)
	}
}

func TestSanitizeParam(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Twitter  ", "twitter"},
		{"Summer Sale 2026!", "summer_sale_2026"},
		{"cpc-ad@v2", "cpc-ad_v2"},
		{"", ""},
	}

	for _, tt := range tests {
		actual := campaign.SanitizeParam(tt.input)
		if actual != tt.expected {
			t.Errorf("SanitizeParam(%q) = %q, expected %q", tt.input, actual, tt.expected)
		}
	}
}

func TestGroupCampaignStats(t *testing.T) {
	service := campaign.NewCampaignService()
	links := []campaign.LinkStat{
		{LinkID: "link-1", Slug: "q3-pricing", Clicks: 500},
		{LinkID: "link-2", Slug: "q3-signup", Clicks: 1200},
		{LinkID: "link-3", Slug: "q3-docs", Clicks: 300},
	}

	summary := service.GroupCampaignStats("camp-123", links)

	if summary.TotalLinks != 3 {
		t.Errorf("expected TotalLinks = 3, got %d", summary.TotalLinks)
	}
	if summary.TotalClicks != 2000 {
		t.Errorf("expected TotalClicks = 2000, got %d", summary.TotalClicks)
	}
	if summary.TopLink != "q3-signup" {
		t.Errorf("expected TopLink = 'q3-signup', got '%s'", summary.TopLink)
	}
}

func containsSubstring(s, sub string) bool {
	return campaign.ContainsSubstring(s, sub)
}
