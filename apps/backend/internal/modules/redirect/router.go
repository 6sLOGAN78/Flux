// Package redirect provides high-speed dynamic smart routing and rules evaluation.
package redirect

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RedirectRule defines a conditional routing rule entity.
type RedirectRule struct {
	ID           uuid.UUID `json:"id" db:"id"`
	LinkID       uuid.UUID `json:"link_id" db:"link_id"`
	RuleType     string    `json:"rule_type" db:"rule_type"`       // 'geo', 'device', 'language', 'time'
	Priority     int       `json:"priority" db:"priority"`         // 1..N (evaluated ascending)
	ConditionKey string    `json:"condition_key" db:"condition_key"` // 'US', 'iOS', 'en', ISO range 'start|end'
	TargetURL    string    `json:"target_url" db:"target_url"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// RequestMetadata holds extracted request attributes for rule matching.
type RequestMetadata struct {
	IPAddress      string    `json:"ip_address"`
	CountryCode    string    `json:"country_code"`
	UserAgent      string    `json:"user_agent"`
	OS             string    `json:"os"`
	DeviceType     string    `json:"device_type"`
	AcceptLanguage string    `json:"accept_language"`
	Timestamp      time.Time `json:"timestamp"`
}

// SmartRouter evaluates conditional routing rules for incoming redirect requests.
type SmartRouter struct{}

func NewSmartRouter() *SmartRouter {
	return &SmartRouter{}
}

// EvaluateRules evaluates rules ordered by Priority ASC and returns matching target_url, or fallback URL.
func (r *SmartRouter) EvaluateRules(rules []RedirectRule, meta RequestMetadata, defaultTargetURL string) string {
	if len(rules) == 0 {
		return defaultTargetURL
	}

	// Filter active rules and sort by Priority ascending
	activeRules := make([]RedirectRule, 0, len(rules))
	for _, rule := range rules {
		if rule.IsActive {
			activeRules = append(activeRules, rule)
		}
	}

	if len(activeRules) == 0 {
		return defaultTargetURL
	}

	sort.Slice(activeRules, func(i, j int) bool {
		return activeRules[i].Priority < activeRules[j].Priority
	})

	if meta.Timestamp.IsZero() {
		meta.Timestamp = time.Now()
	}

	for _, rule := range activeRules {
		if r.matchRule(rule, meta) {
			return rule.TargetURL
		}
	}

	return defaultTargetURL
}

func (r *SmartRouter) matchRule(rule RedirectRule, meta RequestMetadata) bool {
	switch strings.ToLower(rule.RuleType) {
	case "geo":
		return strings.EqualFold(strings.TrimSpace(rule.ConditionKey), strings.TrimSpace(meta.CountryCode))
	case "device":
		cond := strings.ToLower(strings.TrimSpace(rule.ConditionKey))
		os := strings.ToLower(strings.TrimSpace(meta.OS))
		dev := strings.ToLower(strings.TrimSpace(meta.DeviceType))
		ua := strings.ToLower(strings.TrimSpace(meta.UserAgent))

		return cond == os || cond == dev || strings.Contains(ua, cond)
	case "language":
		cond := strings.ToLower(strings.TrimSpace(rule.ConditionKey))
		langHeader := strings.ToLower(strings.TrimSpace(meta.AcceptLanguage))
		return strings.Contains(langHeader, cond)
	case "time":
		return r.matchTimeWindow(rule.ConditionKey, meta.Timestamp)
	default:
		return false
	}
}

func (r *SmartRouter) matchTimeWindow(conditionKey string, now time.Time) bool {
	parts := strings.Split(conditionKey, "|")
	if len(parts) == 2 {
		start, errStart := time.Parse(time.RFC3339, strings.TrimSpace(parts[0]))
		end, errEnd := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))

		if errStart == nil && errEnd == nil {
			return !now.Before(start) && !now.After(end)
		}
	}
	return false
}
