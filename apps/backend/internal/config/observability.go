package config

// ObservabilityConfig holds parameters for tracing, logging, and metrics.
type ObservabilityConfig struct {
	ServiceName        string `koanf:"SERVICE_NAME"`
	Environment        string `koanf:"ENVIRONMENT"`
	NewRelicLicenseKey string `koanf:"NEW_RELIC_LICENSE_KEY"`
	LogLevel           string `koanf:"LOG_LEVEL"`
}
