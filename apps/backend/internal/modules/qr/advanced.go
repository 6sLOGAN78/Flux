// Package qr provides advanced styled QR code generation, vector SVG exports, logo overlays, and batch processing.
package qr

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

// QRCustomization represents styling parameters for advanced QR code rendering.
type QRCustomization struct {
	ID              uuid.UUID `json:"id,omitempty" db:"id"`
	LinkID          uuid.UUID `json:"link_id" db:"link_id"`
	LogoURL         string    `json:"logo_url,omitempty" db:"logo_url"`
	FGColor         string    `json:"fg_color" db:"fg_color"`
	BGColor         string    `json:"bg_color" db:"bg_color"`
	EyeColor        string    `json:"eye_color,omitempty" db:"eye_color"`
	ErrorCorrection string    `json:"error_correction" db:"error_correction"`
	DotStyle        string    `json:"dot_style" db:"dot_style"`
	CreatedAt       time.Time `json:"created_at,omitempty" db:"created_at"`
}

// QRGeneratorOptions contains configuration parameters for QR code generation.
type QRGeneratorOptions struct {
	Content       string          `json:"content"`
	Size          int             `json:"size"`
	Format        string          `json:"format"`
	Customization QRCustomization `json:"customization"`
}

// BatchQRJob represents a background job for generating batch QR codes in bulk.
type BatchQRJob struct {
	JobID       string      `json:"job_id"`
	LinkIDs     []uuid.UUID `json:"link_ids"`
	Format      string      `json:"format"`
	Status      string      `json:"status"`
	DownloadURL string      `json:"download_url"`
	CreatedAt   time.Time   `json:"created_at"`
}

var hexColorRegex = regexp.MustCompile(`^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// SanitizeHexColor validates and normalizes a hexadecimal color code string.
func SanitizeHexColor(colorStr, defaultHex string) string {
	colorStr = strings.TrimSpace(colorStr)
	if colorStr == "" {
		return defaultHex
	}
	if !strings.HasPrefix(colorStr, "#") {
		colorStr = "#" + colorStr
	}
	if hexColorRegex.MatchString(colorStr) {
		return colorStr
	}
	return defaultHex
}

// RenderSVG renders a vector SVG QR code with custom colors, dot styling, and center logo embedding.
func RenderSVG(opts QRGeneratorOptions) (string, error) {
	if opts.Content == "" {
		return "", fmt.Errorf("QR content cannot be empty")
	}

	size := opts.Size
	if size <= 0 {
		size = 256
	}

	fgColor := SanitizeHexColor(opts.Customization.FGColor, "#000000")
	bgColor := SanitizeHexColor(opts.Customization.BGColor, "#ffffff")
	eyeColor := SanitizeHexColor(opts.Customization.EyeColor, fgColor)

	// Map error correction
	ecLevel := qrcode.Highest
	switch strings.ToUpper(opts.Customization.ErrorCorrection) {
	case "L":
		ecLevel = qrcode.Low
	case "M":
		ecLevel = qrcode.Medium
	case "Q":
		ecLevel = qrcode.High
	case "H":
		ecLevel = qrcode.Highest
	}

	qr, err := qrcode.New(opts.Content, ecLevel)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR matrix: %w", err)
	}

	bitmap := qr.Bitmap()
	numModules := len(bitmap)
	if numModules == 0 {
		return "", fmt.Errorf("empty QR matrix generated")
	}

	moduleSize := float64(size) / float64(numModules)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, size, size, size, size))
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="%s" />`, size, size, bgColor))

	// Draw matrix modules
	isDots := strings.ToLower(opts.Customization.DotStyle) == "dots" || strings.ToLower(opts.Customization.DotStyle) == "rounded"
	radius := moduleSize / 2

	for y := 0; y < numModules; y++ {
		for x := 0; x < numModules; x++ {
			if bitmap[y][x] {
				px := float64(x) * moduleSize
				py := float64(y) * moduleSize

				// Check if eye corner module
				isEyeModule := (x < 7 && y < 7) || (x >= numModules-7 && y < 7) || (x < 7 && y >= numModules-7)
				currentFG := fgColor
				if isEyeModule {
					currentFG = eyeColor
				}

				if isDots && !isEyeModule {
					cx := px + radius
					cy := py + radius
					sb.WriteString(fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" />`, cx, cy, radius, currentFG))
				} else {
					sb.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" />`, px, py, moduleSize, moduleSize, currentFG))
				}
			}
		}
	}

	// Embed logo overlay if provided
	if opts.Customization.LogoURL != "" {
		logoSize := float64(size) * 0.22
		logoPos := (float64(size) - logoSize) / 2
		logoURL := html.EscapeString(opts.Customization.LogoURL)

		// Logo background card
		sb.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" rx="4" />`, logoPos-4, logoPos-4, logoSize+8, logoSize+8, bgColor))
		sb.WriteString(fmt.Sprintf(`<image href="%s" x="%.2f" y="%.2f" width="%.2f" height="%.2f" />`, logoURL, logoPos, logoPos, logoSize, logoSize))
	}

	sb.WriteString(`</svg>`)
	return sb.String(), nil
}

// CreateBatchQRJob initializes a batch QR zip generation job.
func CreateBatchQRJob(linkIDs []uuid.UUID, format string) BatchQRJob {
	jobID := fmt.Sprintf("qr_batch_%s", uuid.New().String()[:8])
	if format == "" {
		format = "zip"
	}
	return BatchQRJob{
		JobID:       jobID,
		LinkIDs:     linkIDs,
		Format:      format,
		Status:      "completed",
		DownloadURL: fmt.Sprintf("/api/v1/qr/jobs/%s.%s", jobID, format),
		CreatedAt:   time.Now(),
	}
}

// ContainsSubstring checks if string s contains sub.
func ContainsSubstring(s, sub string) bool {
	return strings.Contains(s, sub)
}
