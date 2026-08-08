package ogmeta_test

import (
	"testing"

	"flux/apps/backend/internal/modules/ogmeta"
)

func TestIsSocialBot(t *testing.T) {
	botUserAgents := []string{
		"Mozilla/5.0 (compatible; Twitterbot/1.0)",
		"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uinject.html)",
		"Slackbot-LinkExpanding 1.0 (+https://api.slack.com/robots)",
		"LinkedInBot/1.0 (compatible; Mozilla/5.0; Jakarta Commons-HttpClient/3.1)",
		"WhatsApp/2.21.12.21 A",
		"TelegramBot (like TwitterBot)",
		"Discordbot/2.0",
	}

	for _, ua := range botUserAgents {
		if !ogmeta.IsSocialBot(ua) {
			t.Errorf("expected IsSocialBot to return true for UA %q", ua)
		}
	}

	humanUserAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
	}

	for _, ua := range humanUserAgents {
		if ogmeta.IsSocialBot(ua) {
			t.Errorf("expected IsSocialBot to return false for human UA %q", ua)
		}
	}
}

func TestGenerateOGHTML(t *testing.T) {
	meta := ogmeta.OGMeta{
		Title:           "Exclusive Summer Sale",
		Description:     "Get 50% off all products today!",
		ImageURL:        "https://cdn.acme.com/banner.jpg",
		TwitterCardType: "summary_large_image",
		TargetURL:       "https://example.com/sale",
	}

	htmlOutput := ogmeta.GenerateOGHTML(meta)

	expectedTags := []string{
		`<meta property="og:title" content="Exclusive Summer Sale" />`,
		`<meta property="og:description" content="Get 50% off all products today!" />`,
		`<meta property="og:image" content="https://cdn.acme.com/banner.jpg" />`,
		`<meta name="twitter:card" content="summary_large_image" />`,
	}

	for _, tag := range expectedTags {
		if !containsSubstring(htmlOutput, tag) {
			t.Errorf("expected HTML output to contain tag %q", tag)
		}
	}
}

func TestExtractOGMetaFromHTML(t *testing.T) {
	rawHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Original Page Title</title>
    <meta property="og:title" content="OG Title Override" />
    <meta property="og:description" content="OG Description Text" />
    <meta property="og:image" content="https://example.com/og.png" />
</head>
<body></body>
</html>`

	meta := ogmeta.ExtractOGMetaFromHTML(rawHTML)

	if meta.Title != "OG Title Override" {
		t.Errorf("expected title 'OG Title Override', got %q", meta.Title)
	}
	if meta.Description != "OG Description Text" {
		t.Errorf("expected description 'OG Description Text', got %q", meta.Description)
	}
	if meta.ImageURL != "https://example.com/og.png" {
		t.Errorf("expected image_url 'https://example.com/og.png', got %q", meta.ImageURL)
	}
}

func containsSubstring(s, sub string) bool {
	return ogmeta.ContainsSubstring(s, sub)
}
