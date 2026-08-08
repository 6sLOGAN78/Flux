package qr_test

import (
	"testing"

	"flux/apps/backend/internal/modules/qr"

	"github.com/google/uuid"
)

func TestRenderSVG_CustomizationAndLogo(t *testing.T) {
	opts := qr.QRGeneratorOptions{
		Content: "https://flux.dev/test123",
		Size:    512,
		Format:  "svg",
		Customization: qr.QRCustomization{
			LinkID:          uuid.New(),
			LogoURL:         "https://cdn.acme.com/logo.png",
			FGColor:         "#3b82f6",
			BGColor:         "#ffffff",
			EyeColor:        "#1e40af",
			ErrorCorrection: "H",
			DotStyle:        "dots",
		},
	}

	svgOutput, err := qr.RenderSVG(opts)
	if err != nil {
		t.Fatalf("unexpected error rendering SVG QR: %v", err)
	}

	expectedElements := []string{
		`<svg`,
		`xmlns="http://www.w3.org/2000/svg"`,
		`fill="#ffffff"`, // BG Color
		`fill="#3b82f6"`, // FG Color
		`href="https://cdn.acme.com/logo.png"`, // Embedded Logo
		`</svg>`,
	}

	for _, elem := range expectedElements {
		if !qr.ContainsSubstring(svgOutput, elem) {
			t.Errorf("expected SVG output to contain %q", elem)
		}
	}
}

func TestSanitizeHexColor(t *testing.T) {
	tests := []struct {
		input    string
		def      string
		expected string
	}{
		{"#3b82f6", "#000000", "#3b82f6"},
		{"3b82f6", "#000000", "#3b82f6"},
		{"invalid", "#000000", "#000000"},
		{"", "#ffffff", "#ffffff"},
	}

	for _, tc := range tests {
		actual := qr.SanitizeHexColor(tc.input, tc.def)
		if actual != tc.expected {
			t.Errorf("SanitizeHexColor(%q, %q) = %q; expected %q", tc.input, tc.def, actual, tc.expected)
		}
	}
}

func TestCreateBatchQRJob(t *testing.T) {
	linkIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	job := qr.CreateBatchQRJob(linkIDs, "zip")

	if job.JobID == "" {
		t.Error("expected non-empty JobID")
	}
	if len(job.LinkIDs) != 3 {
		t.Errorf("expected 3 link_ids, got %d", len(job.LinkIDs))
	}
	if job.Status != "completed" {
		t.Errorf("expected status 'completed', got %q", job.Status)
	}
}
