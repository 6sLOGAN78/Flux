package logger_test

import (
	"testing"

	"flux/apps/backend/internal/config"
	"flux/apps/backend/internal/logger"

	"github.com/rs/zerolog"
)

func TestNewLoggerService(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		ServiceName: "test-service",
		Environment: "test",
		LogLevel:    "debug",
		NewRelic: config.NewRelicConfig{
			LicenseKey: "",
		},
	}

	service := logger.NewLoggerService(cfg)
	if service == nil {
		t.Fatalf("expected non-nil LoggerService")
	}
	if service.GetApplication() != nil {
		t.Errorf("expected nil NewRelic Application when license key is empty")
	}
	service.Shutdown()
}

func TestNewLoggerWithService(t *testing.T) {
	cfg := &config.ObservabilityConfig{
		ServiceName: "test-service",
		Environment: "development",
		LogLevel:    "debug",
		Logging:     config.LoggingConfig{Format: "console"},
	}

	service := logger.NewLoggerService(cfg)
	l := logger.NewLoggerWithService(cfg, service)

	if l.GetLevel() != zerolog.DebugLevel {
		t.Errorf("expected debug level, got %v", l.GetLevel())
	}
}

func TestNewPgxLogger(t *testing.T) {
	l := logger.NewPgxLogger(zerolog.InfoLevel)
	if l.GetLevel() != zerolog.InfoLevel {
		t.Errorf("expected info level for pgx logger, got %v", l.GetLevel())
	}
}

func TestGetPgxTraceLogLevel(t *testing.T) {
	tests := []struct {
		level    zerolog.Level
		expected int
	}{
		{zerolog.DebugLevel, 6},
		{zerolog.InfoLevel, 4},
		{zerolog.WarnLevel, 3},
		{zerolog.ErrorLevel, 2},
		{zerolog.Disabled, 0},
	}

	for _, tt := range tests {
		got := logger.GetPgxTraceLogLevel(tt.level)
		if got != tt.expected {
			t.Errorf("GetPgxTraceLogLevel(%v) = %d, expected %d", tt.level, got, tt.expected)
		}
	}
}
