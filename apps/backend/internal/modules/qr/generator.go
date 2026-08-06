// Package qr provides high-performance QR code generation and image caching.
package qr

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/skip2/go-qrcode"
)

// QROptions defines configuration parameters for QR code generation.
type QROptions struct {
	Size            int    `json:"size"`             // Dimensions in pixels (default 256)
	Format          string `json:"format"`           // "png" or "svg"
	ErrorCorrection string `json:"error_correction"` // "L", "M", "Q", "H"
	FGColor         string `json:"fg_color,omitempty"`
	BGColor         string `json:"bg_color,omitempty"`
}

// QRGenerator encapsulates QR matrix rendering.
type QRGenerator struct{}

// NewQRGenerator initializes a QRGenerator instance.
func NewQRGenerator() *QRGenerator {
	return &QRGenerator{}
}

// GeneratePNG generates raster PNG image bytes for a target URL.
func (g *QRGenerator) GeneratePNG(content string, opts QROptions) ([]byte, error) {
	if opts.Size <= 0 {
		opts.Size = 256
	}

	level := parseRecoveryLevel(opts.ErrorCorrection)
	pngBytes, err := qrcode.Encode(content, level, opts.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG QR code: %w", err)
	}

	return pngBytes, nil
}

// GenerateSVG generates XML vector SVG markup for a target URL.
func (g *QRGenerator) GenerateSVG(content string, opts QROptions) (string, error) {
	if opts.Size <= 0 {
		opts.Size = 256
	}

	level := parseRecoveryLevel(opts.ErrorCorrection)
	qrCode, err := qrcode.New(content, level)
	if err != nil {
		return "", fmt.Errorf("failed to initialize SVG QR code: %w", err)
	}

	bitmap := qrCode.Bitmap()
	gridSize := len(bitmap)
	cellSize := float64(opts.Size) / float64(gridSize)

	fgColor := opts.FGColor
	if fgColor == "" {
		fgColor = "#000000"
	}

	bgColor := opts.BGColor
	if bgColor == "" {
		bgColor = "#ffffff"
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, opts.Size, opts.Size, opts.Size, opts.Size))
	buf.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="%s"/>`, opts.Size, opts.Size, bgColor))
	buf.WriteString(fmt.Sprintf(`<g fill="%s">`, fgColor))

	for y, row := range bitmap {
		for x, isBlack := range row {
			if isBlack {
				rx := float64(x) * cellSize
				ry := float64(y) * cellSize
				buf.WriteString(fmt.Sprintf(`<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f"/>`, rx, ry, cellSize, cellSize))
			}
		}
	}

	buf.WriteString(`</g></svg>`)
	return buf.String(), nil
}

func parseRecoveryLevel(levelStr string) qrcode.RecoveryLevel {
	switch strings.ToUpper(levelStr) {
	case "L":
		return qrcode.Low
	case "M":
		return qrcode.Medium
	case "Q":
		return qrcode.High
	case "H":
		return qrcode.Highest
	default:
		return qrcode.Highest
	}
}

// QRCache defines caching operations for generated QR images.
type QRCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}

// MockQRCache provides an in-memory QRCache implementation for testing.
type MockQRCache struct {
	store map[string][]byte
}

// NewMockQRCache initializes an in-memory QRCache.
func NewMockQRCache() *MockQRCache {
	return &MockQRCache{store: make(map[string][]byte)}
}

func (m *MockQRCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, exists := m.store[key]
	if !exists {
		return nil, fmt.Errorf("cache miss")
	}
	return val, nil
}

func (m *MockQRCache) Set(ctx context.Context, key string, data []byte) error {
	m.store[key] = data
	return nil
}

// QRService orchestrates QR generation and Redis/memory caching.
type QRService struct {
	generator *QRGenerator
	cache     QRCache
}

// NewQRService initializes a QRService instance.
func NewQRService(cache QRCache) *QRService {
	if cache == nil {
		cache = NewMockQRCache()
	}
	return &QRService{
		generator: NewQRGenerator(),
		cache:     cache,
	}
}

// Generate resolves or creates a cached QR image payload.
func (s *QRService) Generate(content string, opts QROptions) ([]byte, error) {
	cacheKey := fmt.Sprintf("qr:%s:%d:%s:%s", content, opts.Size, opts.Format, opts.ErrorCorrection)

	cached, err := s.cache.Get(context.Background(), cacheKey)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	var result []byte
	if strings.ToLower(opts.Format) == "svg" {
		svgStr, err := s.generator.GenerateSVG(content, opts)
		if err != nil {
			return nil, err
		}
		result = []byte(svgStr)
	} else {
		pngBytes, err := s.generator.GeneratePNG(content, opts)
		if err != nil {
			return nil, err
		}
		result = pngBytes
	}

	_ = s.cache.Set(context.Background(), cacheKey, result)
	return result, nil
}

func (g *QRGenerator) dummyUsage() {
	_ = strconv.Itoa(0)
}
