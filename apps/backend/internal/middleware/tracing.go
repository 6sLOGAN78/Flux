package middleware

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// TracingMiddleware initializes New Relic APM tracing middleware for Echo if license key is configured.
func TracingMiddleware(licenseKey string, appName string) echo.MiddlewareFunc {
	if licenseKey == "" {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		log.Printf("warning: failed to initialize New Relic APM: %v", err)
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return nrecho.Middleware(app)
}
