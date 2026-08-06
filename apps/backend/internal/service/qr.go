package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"flux/apps/backend/internal/model/qr"
	"flux/apps/backend/internal/repository"

	goqrcode "github.com/skip2/go-qrcode"
)

// QRGenerator encapsulates QR matrix rendering.
type QRGenerator struct{}

func NewQRGenerator() *QRGenerator {
	return &QRGenerator{}
}

func (g *QRGenerator) GeneratePNG(content string, opts qr.QROptions) ([]byte, error) {
	if opts.Size <= 0 {
		opts.Size = 256
	}
	level := parseRecoveryLevel(opts.ErrorCorrection)
	return goqrcode.Encode(content, level, opts.Size)
}

func (g *QRGenerator) GenerateSVG(content string, opts qr.QROptions) (string, error) {
	if opts.Size <= 0 {
		opts.Size = 256
	}
	level := parseRecoveryLevel(opts.ErrorCorrection)
	qrCode, err := goqrcode.New(content, level)
	if err != nil {
		return "", err
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

func parseRecoveryLevel(levelStr string) goqrcode.RecoveryLevel {
	switch strings.ToUpper(levelStr) {
	case "L":
		return goqrcode.Low
	case "M":
		return goqrcode.Medium
	case "Q":
		return goqrcode.High
	case "H":
		return goqrcode.Highest
	default:
		return goqrcode.Highest
	}
}

// QRService orchestrates QR code generation and caching.
type QRService struct {
	generator *QRGenerator
	cache     repository.QRCache
}

func NewQRService(cache repository.QRCache) *QRService {
	return &QRService{
		generator: NewQRGenerator(),
		cache:     cache,
	}
}

func (s *QRService) Generate(content string, opts qr.QROptions) ([]byte, error) {
	cacheKey := fmt.Sprintf("qr:%s:%d:%s:%s", content, opts.Size, opts.Format, opts.ErrorCorrection)

	if s.cache != nil {
		cached, err := s.cache.Get(context.Background(), cacheKey)
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
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

	if s.cache != nil {
		_ = s.cache.Set(context.Background(), cacheKey, result)
	}

	return result, nil
}
