package middleware

import (
	"github.com/6sLOGAN78/flux/internal/observability"
	"github.com/labstack/echo/v4"
	"strings"
)

const (
	RequestIDHeader     = "X-Request-ID"
	RequestIDKey        = "request_id"
	CorrelationIDHeader = "X-Correlation-ID"
	CorrelationIDKey    = "correlation_id"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			// Check length before parsing to bound work on attacker-controlled headers.
			requestID, correlationID := req.Header.Get(RequestIDHeader), req.Header.Get(CorrelationIDHeader)
			if len(requestID) != 36 || len(req.Header.Values(RequestIDHeader)) != 1 {
				requestID = ""
			}
			if len(correlationID) != 36 || len(req.Header.Values(CorrelationIDHeader)) != 1 {
				correlationID = ""
			}
			ctx := observability.WithCorrelation(req.Context(), requestID, correlationID)
			ids := observability.CorrelationFromContext(ctx)
			c.Set(RequestIDKey, ids.RequestID)
			c.Set(CorrelationIDKey, ids.CorrelationID)
			for _, pair := range [][2]string{{RequestIDHeader, ids.RequestID}, {CorrelationIDHeader, ids.CorrelationID}} {
				req.Header.Set(pair[0], pair[1])
				c.Response().Header().Set(pair[0], pair[1])
			}
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}
func GetRequestID(c echo.Context) string {
	value, _ := c.Get(RequestIDKey).(string)
	return value
}
func GetCorrelationID(c echo.Context) string {
	value, _ := c.Get(CorrelationIDKey).(string)
	return value
}
func equalPropagationHeader(key string) bool {
	return strings.EqualFold(key, "tracestate") || strings.EqualFold(key, "baggage")
}
