package middleware

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

const productBodyLimit = 64 * 1024

// ExactOrigin rejects wildcards, null, credentials, paths and nonlocal cleartext.
func ExactOrigin(origin string, allowed []string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || origin == "" || parsed.Host == "" || parsed.User != nil ||
		parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Opaque != "" ||
		origin != parsed.Scheme+"://"+parsed.Host {
		return false
	}
	secure := parsed.Scheme == "https"
	local := parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1"
	if !secure && (parsed.Scheme != "http" || !local) {
		return false
	}
	for _, candidate := range allowed {
		if origin == candidate {
			return true
		}
	}
	return false
}

// BrowserMutation checks the browser boundary before any mutation handler runs.
func BrowserMutation(origins []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-store")
			request := c.Request()
			if request.Method == http.MethodGet || request.Method == http.MethodHead {
				return next(c)
			}
			values := request.Header.Values("Origin")
			if len(values) != 1 || !ExactOrigin(values[0], origins) {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			contentTypes := request.Header.Values("Content-Type")
			if len(contentTypes) != 1 {
				return echo.NewHTTPError(http.StatusUnsupportedMediaType)
			}
			contentType, _, err := mime.ParseMediaType(contentTypes[0])
			if err != nil || !strings.EqualFold(contentType, "application/json") {
				return echo.NewHTTPError(http.StatusUnsupportedMediaType)
			}
			body, err := io.ReadAll(http.MaxBytesReader(c.Response(), request.Body, productBodyLimit))
			if err != nil {
				return echo.NewHTTPError(http.StatusRequestEntityTooLarge)
			}
			request.Body = io.NopCloser(bytes.NewReader(body))
			return next(c)
		}
	}
}
