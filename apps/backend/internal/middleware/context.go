package middleware

import (
	"github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// UserIDKey and the other keys identify values stored in Echo request context.
const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	LoggerKey   = "logger"
)

// ContextEnhancer attaches the request-scoped contextual logger.
type ContextEnhancer struct{ server *server.Server }

// NewContextEnhancer constructs request context middleware.
func NewContextEnhancer(s *server.Server) *ContextEnhancer { return &ContextEnhancer{server: s} }

// EnhanceContext attaches safe request identity to the contextual logger.
func (ce *ContextEnhancer) EnhanceContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			contextLogger := logger.WithContext(*ce.server.Logger, c.Request().Context())
			c.Set(LoggerKey, &contextLogger)
			return next(c)
		}
	}
}

// GetUserID returns the authenticated request user identifier.
func GetUserID(c echo.Context) string {
	value, _ := c.Get(UserIDKey).(string)
	return value
}

// GetLogger returns the contextual request logger.
func GetLogger(c echo.Context) *zerolog.Logger {
	if log, ok := c.Get(LoggerKey).(*zerolog.Logger); ok {
		return log
	}
	log := zerolog.Nop()
	return &log
}
