package config

import "strings"

// NewRelicConfig holds New Relic APM & Logging configuration.
type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key"`
	AppLogForwardingEnabled   bool   `koanf:"app_log_forwarding_enabled"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled"`
	DebugLogging              bool   `koanf:"debug_logging"`
}

// LoggingConfig holds format preferences.
type LoggingConfig struct {
	Format string `koanf:"format"`
}

// ObservabilityConfig holds parameters for tracing, logging, and metrics.
type ObservabilityConfig struct {
	ServiceName string         `koanf:"service_name"`
	Environment string         `koanf:"environment"`
	LogLevel    string         `koanf:"log_level"`
	Logging     LoggingConfig  `koanf:"logging"`
	NewRelic    NewRelicConfig `koanf:"new_relic"`
}

// GetLogLevel returns normal lowercase log level string.
func (c *ObservabilityConfig) GetLogLevel() string {
	if c == nil || c.LogLevel == "" {
		return "info"
	}
	return strings.ToLower(c.LogLevel)
}

// IsProduction checks if environment is set to production.
func (c *ObservabilityConfig) IsProduction() bool {
	if c == nil {
		return false
	}
	env := strings.ToLower(c.Environment)
	return env == "production" || env == "prod"
}
