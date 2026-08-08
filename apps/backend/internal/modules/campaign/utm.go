// Package campaign provides campaign management and UTM parameter building utilities.
package campaign

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Campaign represents a marketing campaign entity.
type Campaign struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	UTMSource   *string   `json:"utm_source,omitempty"`
	UTMMedium   *string   `json:"utm_medium,omitempty"`
	UTMCampaign *string   `json:"utm_campaign,omitempty"`
	UTMTerm     *string   `json:"utm_term,omitempty"`
	UTMContent  *string   `json:"utm_content,omitempty"`
	Budget      float64   `json:"budget"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UTMParams holds UTM tracking parameters.
type UTMParams struct {
	UTMSource   string `json:"utm_source"`
	UTMMedium   string `json:"utm_medium"`
	UTMCampaign string `json:"utm_campaign"`
	UTMTerm     string `json:"utm_term,omitempty"`
	UTMContent  string `json:"utm_content,omitempty"`
}

// LinkStat represents metrics for a link within a campaign.
type LinkStat struct {
	LinkID string `json:"link_id"`
	Slug   string `json:"slug"`
	Clicks int64  `json:"clicks"`
}

// CampaignSummary represents aggregated metrics for a campaign.
type CampaignSummary struct {
	CampaignID  string `json:"campaign_id"`
	TotalLinks  int    `json:"total_links"`
	TotalClicks int64  `json:"total_clicks"`
	TopLink     string `json:"top_link"`
}

// UTMBuilder provides parameter injection and sanitization.
type UTMBuilder struct{}

func NewUTMBuilder() *UTMBuilder {
	return &UTMBuilder{}
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9_\-]+`)

// SanitizeParam cleans and normalizes a UTM parameter string.
func SanitizeParam(val string) string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return ""
	}
	lowered := strings.ToLower(trimmed)
	replacedSpaces := strings.ReplaceAll(lowered, " ", "_")
	cleaned := nonAlphanumericRegex.ReplaceAllString(replacedSpaces, "_")
	cleaned = regexp.MustCompile(`_+`).ReplaceAllString(cleaned, "_")
	return strings.Trim(cleaned, "_")
}

// BuildURL injects sanitized UTM parameters into a destination URL while preserving original query params and hash fragments.
func (b *UTMBuilder) BuildURL(rawURL string, params UTMParams) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid destination url: %w", err)
	}

	query := parsedURL.Query()

	if s := SanitizeParam(params.UTMSource); s != "" {
		query.Set("utm_source", s)
	}
	if m := SanitizeParam(params.UTMMedium); m != "" {
		query.Set("utm_medium", m)
	}
	if c := SanitizeParam(params.UTMCampaign); c != "" {
		query.Set("utm_campaign", c)
	}
	if t := SanitizeParam(params.UTMTerm); t != "" {
		query.Set("utm_term", t)
	}
	if cnt := SanitizeParam(params.UTMContent); cnt != "" {
		query.Set("utm_content", cnt)
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

// CampaignService provides business logic for grouping and summarizing campaign stats.
type CampaignService struct{}

func NewCampaignService() *CampaignService {
	return &CampaignService{}
}

// GroupCampaignStats aggregates total links, total clicks, and top link for a campaign.
func (s *CampaignService) GroupCampaignStats(campaignID string, links []LinkStat) CampaignSummary {
	var totalClicks int64
	var maxClicks int64
	topLink := ""

	for _, link := range links {
		totalClicks += link.Clicks
		if link.Clicks >= maxClicks {
			maxClicks = link.Clicks
			topLink = link.Slug
		}
	}

	return CampaignSummary{
		CampaignID:  campaignID,
		TotalLinks:  len(links),
		TotalClicks: totalClicks,
		TopLink:     topLink,
	}
}

// ContainsSubstring checks if string s contains sub.
func ContainsSubstring(s, sub string) bool {
	return strings.Contains(s, sub)
}
