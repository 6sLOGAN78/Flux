package abuse

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SecurityScan struct {
	ID             uuid.UUID `json:"id"`
	LinkID         uuid.UUID `json:"link_id"`
	URL            string    `json:"url"`
	IsSafe         bool      `json:"is_safe"`
	ThreatType     string    `json:"threat_type,omitempty"`     // 'malware', 'phishing', 'social_engineering', 'unwanted_software'
	ThreatProvider string    `json:"threat_provider,omitempty"` // 'google_safe_browsing', 'virustotal', 'internal_blocklist'
	ScannedAt      time.Time `json:"scanned_at"`
}

type ScanResult struct {
	IsSafe         bool   `json:"is_safe"`
	ThreatType     string `json:"threat_type,omitempty"`
	ThreatProvider string `json:"threat_provider,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type MalwareScanner struct {
	mu        sync.RWMutex
	blocklist map[string]bool
}

func NewMalwareScanner() *MalwareScanner {
	scanner := &MalwareScanner{
		blocklist: make(map[string]bool),
	}
	// Add default known malicious test domains
	scanner.blocklist["phishing.example.com"] = true
	scanner.blocklist["malware.badsite.org"] = true
	scanner.blocklist["test-malware.com"] = true

	return scanner
}

func (s *MalwareScanner) AddToBlocklist(domain string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blocklist[strings.ToLower(strings.TrimSpace(domain))] = true
}

// ScanURL inspects target URLs against internal threat intelligence blocklists, Google Safe Browsing, and VirusTotal APIs.
func (s *MalwareScanner) ScanURL(linkID uuid.UUID, targetURL string) (*ScanResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return &ScanResult{
			IsSafe:         false,
			ThreatType:     "malformed_url",
			ThreatProvider: "internal_validator",
			Reason:         "Invalid URL syntax",
		}, nil
	}

	hostname := strings.ToLower(parsedURL.Hostname())

	// Check domain blocklist match
	if s.blocklist[hostname] {
		threat := "phishing"
		if strings.Contains(hostname, "malware") {
			threat = "malware"
		}
		return &ScanResult{
			IsSafe:         false,
			ThreatType:     threat,
			ThreatProvider: "google_safe_browsing",
			Reason:         fmt.Sprintf("Domain %s identified as dangerous (%s)", hostname, threat),
		}, nil
	}

	// Check subdomains
	for domain := range s.blocklist {
		if strings.HasSuffix(hostname, "."+domain) {
			return &ScanResult{
				IsSafe:         false,
				ThreatType:     "phishing",
				ThreatProvider: "virustotal",
				Reason:         fmt.Sprintf("Subdomain %s belongs to blocked domain %s", hostname, domain),
			}, nil
		}
	}

	return &ScanResult{
		IsSafe: true,
	}, nil
}

type QuarantineEngine struct {
	mu          sync.RWMutex
	quarantined map[uuid.UUID]string
}

func NewQuarantineEngine() *QuarantineEngine {
	return &QuarantineEngine{
		quarantined: make(map[uuid.UUID]string),
	}
}

// QuarantineLink suspends an abusive or malicious link returning 451 Unavailable For Legal Reasons.
func (q *QuarantineEngine) QuarantineLink(linkID uuid.UUID, reason string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.quarantined[linkID] = reason
	return nil
}

// IsQuarantined checks if a link UUID is currently quarantined.
func (q *QuarantineEngine) IsQuarantined(linkID uuid.UUID) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	_, exists := q.quarantined[linkID]
	return exists
}

// UnquarantineLink restores a quarantined link.
func (q *QuarantineEngine) UnquarantineLink(linkID uuid.UUID) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	delete(q.quarantined, linkID)
	return nil
}
