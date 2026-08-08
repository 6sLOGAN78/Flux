package whitelabel

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestThemeRenderer_GenerateThemeCSS(t *testing.T) {
	renderer := NewThemeRenderer()

	config := WhiteLabelConfig{
		ID:              uuid.New(),
		OrganizationID:  uuid.New(),
		DashboardDomain: "analytics.acme.com",
		BrandName:       "Acme Analytics",
		LogoURL:         "https://acme.com/logo.png",
		FaviconURL:      "https://acme.com/favicon.ico",
		PrimaryColor:    "#0052FF",
		AccentColor:     "#FF0055",
		CustomCSS:       "body { background-color: #f4f4f4; } <script>alert('xss')</script>",
		HideFooter:      true,
	}

	css := renderer.GenerateThemeCSS(config)

	if !strings.Contains(css, "--primary-color: #0052FF;") {
		t.Errorf("expected generated CSS to include --primary-color: #0052FF;")
	}

	if !strings.Contains(css, "--accent-color: #FF0055;") {
		t.Errorf("expected generated CSS to include --accent-color: #FF0055;")
	}

	if strings.Contains(css, "<script>") {
		t.Errorf("expected custom CSS sanitizer to strip <script> tags")
	}

	if !strings.Contains(css, "body { background-color: #f4f4f4; }") {
		t.Errorf("expected custom CSS rules to be preserved")
	}
}

func TestDKIMSigner_GenerateHeader(t *testing.T) {
	signer := NewDKIMSigner()

	domain := "acme.com"
	selector := "flux"
	header, err := signer.GenerateDKIMHeader(domain, selector, "Subject: Test\r\nFrom: info@acme.com", []byte("Hello World"))
	if err != nil {
		t.Fatalf("unexpected DKIM header generation error: %v", err)
	}

	if !strings.Contains(header, "v=1; a=rsa-sha256;") {
		t.Errorf("expected DKIM header to start with v=1; a=rsa-sha256;")
	}

	if !strings.Contains(header, "d=acme.com;") {
		t.Errorf("expected DKIM header to contain d=acme.com;")
	}

	if !strings.Contains(header, "s=flux;") {
		t.Errorf("expected DKIM header to contain s=flux;")
	}
}
