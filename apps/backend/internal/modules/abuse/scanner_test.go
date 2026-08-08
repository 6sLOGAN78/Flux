package abuse

import (
	"testing"

	"github.com/google/uuid"
)

func TestMalwareScanner_CleanURL(t *testing.T) {
	scanner := NewMalwareScanner()

	linkID := uuid.New()
	cleanURL := "https://example.com/safe-page"

	res, err := scanner.ScanURL(linkID, cleanURL)
	if err != nil {
		t.Fatalf("unexpected error scanning clean URL: %v", err)
	}

	if !res.IsSafe {
		t.Errorf("expected clean URL to be safe")
	}

	if res.ThreatType != "" {
		t.Errorf("expected empty threat type for safe URL, got '%s'", res.ThreatType)
	}
}

func TestMalwareScanner_PhishingMalwareDetection(t *testing.T) {
	scanner := NewMalwareScanner()
	scanner.AddToBlocklist("phishing-login.com")
	scanner.AddToBlocklist("malware-distributor.net")

	linkID := uuid.New()
	phishingURL := "https://phishing-login.com/bank/login"

	res, err := scanner.ScanURL(linkID, phishingURL)
	if err != nil {
		t.Fatalf("unexpected error scanning phishing URL: %v", err)
	}

	if res.IsSafe {
		t.Errorf("expected phishing URL to be marked unsafe")
	}

	if res.ThreatType != "phishing" && res.ThreatType != "malware" {
		t.Errorf("expected threat type 'phishing' or 'malware', got '%s'", res.ThreatType)
	}

	if res.ThreatProvider == "" {
		t.Errorf("expected non-empty threat provider")
	}
}

func TestQuarantineEngine_Takedown(t *testing.T) {
	engine := NewQuarantineEngine()

	linkID := uuid.New()
	err := engine.QuarantineLink(linkID, "Google Safe Browsing flagged phishing site")
	if err != nil {
		t.Fatalf("unexpected error quarantining link: %v", err)
	}

	if !engine.IsQuarantined(linkID) {
		t.Errorf("expected link %s to be quarantined", linkID)
	}
}
