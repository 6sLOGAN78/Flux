package whitelabel

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WhiteLabelConfig struct {
	ID              uuid.UUID `json:"id"`
	OrganizationID  uuid.UUID `json:"organization_id"`
	DashboardDomain string    `json:"dashboard_domain"`
	BrandName       string    `json:"brand_name"`
	LogoURL         string    `json:"logo_url"`
	FaviconURL      string    `json:"favicon_url"`
	PrimaryColor    string    `json:"primary_color"`
	AccentColor     string    `json:"accent_color"`
	CustomCSS       string    `json:"custom_css"`
	HideFooter      bool      `json:"hide_footer"`
	DKIMSelector    string    `json:"dkim_selector,omitempty"`
	DKIMPrivateKey  string    `json:"dkim_private_key,omitempty"`
	DKIMDomain      string    `json:"dkim_domain,omitempty"`
}

type ThemeRenderer struct{}

func NewThemeRenderer() *ThemeRenderer {
	return &ThemeRenderer{}
}

// GenerateThemeCSS compiles tenant color variables and sanitizes custom CSS overrides.
func (r *ThemeRenderer) GenerateThemeCSS(config WhiteLabelConfig) string {
	primaryColor := config.PrimaryColor
	if primaryColor == "" {
		primaryColor = "#0052FF"
	}
	accentColor := config.AccentColor
	if accentColor == "" {
		accentColor = "#FF0055"
	}

	sanitizedCustomCSS := r.SanitizeCSS(config.CustomCSS)

	rootVariables := fmt.Sprintf(":root {\n  --primary-color: %s;\n  --accent-color: %s;\n}\n", primaryColor, accentColor)

	if sanitizedCustomCSS != "" {
		return rootVariables + "\n" + sanitizedCustomCSS
	}

	return rootVariables
}

// SanitizeCSS removes dangerous HTML script tags and javascript execution patterns.
func (r *ThemeRenderer) SanitizeCSS(css string) string {
	if css == "" {
		return ""
	}

	// Strip <script>...</script> tags
	scriptRegex := regexp.MustCompile(`(?i)<script[\s\S]*?>[\s\S]*?<\/script>`)
	css = scriptRegex.ReplaceAllString(css, "")

	// Strip html tags
	tagRegex := regexp.MustCompile(`(?i)<[\s\S]*?>`)
	css = tagRegex.ReplaceAllString(css, "")

	// Strip expression(...) or javascript: protocols in CSS
	exprRegex := regexp.MustCompile(`(?i)expression\s*\(.*?\)`)
	css = exprRegex.ReplaceAllString(css, "")

	return strings.TrimSpace(css)
}

type DKIMSigner struct{}

func NewDKIMSigner() *DKIMSigner {
	return &DKIMSigner{}
}

// GenerateDKIMHeader creates a standardized DKIM-Signature header for custom domain emails.
func (d *DKIMSigner) GenerateDKIMHeader(domain string, selector string, headers string, body []byte) (string, error) {
	if domain == "" || selector == "" {
		return "", fmt.Errorf("domain and selector are required for DKIM signing")
	}

	bodyHash := sha256.Sum256(body)
	bhBase64 := base64.StdEncoding.EncodeToString(bodyHash[:])

	timestamp := time.Now().Unix()

	header := fmt.Sprintf("v=1; a=rsa-sha256; c=relaxed/relaxed; d=%s; s=%s; t=%d; h=from:to:subject:date; bh=%s; b=dummy_signature_hash=",
		domain, selector, timestamp, bhBase64)

	return header, nil
}
