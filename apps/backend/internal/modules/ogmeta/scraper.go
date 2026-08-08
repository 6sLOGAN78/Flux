// Package ogmeta provides social crawler bot detection, Open Graph metadata extraction, and HTML preview rendering.
package ogmeta

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// OGMeta represents Open Graph metadata for link social card previews.
type OGMeta struct {
	LinkID          uuid.UUID `json:"link_id,omitempty" db:"link_id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	ImageURL        string    `json:"image_url" db:"image_url"`
	TwitterCardType string    `json:"twitter_card_type" db:"twitter_card_type"`
	SiteName        string    `json:"site_name,omitempty"`
	TargetURL       string    `json:"target_url,omitempty"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

var socialBotRegex = regexp.MustCompile(`(?i)(bot|facebookexternalhit|twitterbot|slackbot|linkedinbot|whatsapp|telegrambot|discordbot|pinterest|skypeuripreview|googlebot|bingbot|facebot|ia_archiver)`)

// IsSocialBot checks if the visitor User-Agent matches known social media crawlers & unfurlers.
func IsSocialBot(userAgent string) bool {
	if strings.TrimSpace(userAgent) == "" {
		return false
	}
	return socialBotRegex.MatchString(userAgent)
}

// GenerateOGHTML generates a complete HTML5 page containing Open Graph and Twitter Card meta tags for social crawlers.
func GenerateOGHTML(meta OGMeta) string {
	title := html.EscapeString(meta.Title)
	description := html.EscapeString(meta.Description)
	imageURL := html.EscapeString(meta.ImageURL)
	targetURL := html.EscapeString(meta.TargetURL)
	twitterCard := meta.TwitterCardType
	if twitterCard == "" {
		twitterCard = "summary_large_image"
	}
	twitterCard = html.EscapeString(twitterCard)

	siteName := meta.SiteName
	if siteName == "" {
		siteName = "Flux"
	}
	siteName = html.EscapeString(siteName)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="%s" />
    <meta property="og:title" content="%s" />
    <meta property="og:description" content="%s" />
    <meta property="og:image" content="%s" />
    <meta property="og:url" content="%s" />
    <meta name="twitter:card" content="%s" />
    <meta name="twitter:title" content="%s" />
    <meta name="twitter:description" content="%s" />
    <meta name="twitter:image" content="%s" />
</head>
<body>
    <p>Redirecting to <a href="%s">%s</a>...</p>
</body>
</html>`, title, siteName, title, description, imageURL, targetURL, twitterCard, title, description, imageURL, targetURL, targetURL)
}

var (
	ogTitleRegex    = regexp.MustCompile(`(?i)<meta\s+property=["']og:title["']\s+content=["']([^"']+)["']`)
	ogTitleRegexAlt = regexp.MustCompile(`(?i)<meta\s+content=["']([^"']+)["']\s+property=["']og:title["']`)
	ogDescRegex     = regexp.MustCompile(`(?i)<meta\s+property=["']og:description["']\s+content=["']([^"']+)["']`)
	ogDescRegexAlt  = regexp.MustCompile(`(?i)<meta\s+content=["']([^"']+)["']\s+property=["']og:description["']`)
	ogImageRegex    = regexp.MustCompile(`(?i)<meta\s+property=["']og:image["']\s+content=["']([^"']+)["']`)
	ogImageRegexAlt = regexp.MustCompile(`(?i)<meta\s+content=["']([^"']+)["']\s+property=["']og:image["']`)
	htmlTitleRegex  = regexp.MustCompile(`(?i)<title>(.*?)</title>`)
)

// ExtractOGMetaFromHTML parses raw HTML string and extracts Open Graph tags or title fallbacks.
func ExtractOGMetaFromHTML(rawHTML string) OGMeta {
	meta := OGMeta{
		TwitterCardType: "summary_large_image",
	}

	if match := ogTitleRegex.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.Title = html.UnescapeString(match[1])
	} else if match := ogTitleRegexAlt.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.Title = html.UnescapeString(match[1])
	} else if match := htmlTitleRegex.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.Title = html.UnescapeString(match[1])
	}

	if match := ogDescRegex.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.Description = html.UnescapeString(match[1])
	} else if match := ogDescRegexAlt.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.Description = html.UnescapeString(match[1])
	}

	if match := ogImageRegex.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.ImageURL = html.UnescapeString(match[1])
	} else if match := ogImageRegexAlt.FindStringSubmatch(rawHTML); len(match) > 1 {
		meta.ImageURL = html.UnescapeString(match[1])
	}

	return meta
}

// ContainsSubstring checks if string s contains sub.
func ContainsSubstring(s, sub string) bool {
	return strings.Contains(s, sub)
}
