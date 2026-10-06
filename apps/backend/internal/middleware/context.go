package middleware

import (
	"github.com/6sLOGAN78/flux/internal/logger"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	LoggerKey   = "logger"
)

type ContextEnhancer struct{ server *server.Server }

func NewContextEnhancer(s *server.Server) *ContextEnhancer { return &ContextEnhancer{server: s} }
func (ce *ContextEnhancer) EnhanceContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			contextLogger := logger.WithContext(*ce.server.Logger, c.Request().Context())
			c.Set(LoggerKey, &contextLogger)
			return next(c)
		}
	}
}
func GetUserID(c echo.Context) string {
	value, _ := c.Get(UserIDKey).(string)
	return value
}
func GetLogger(c echo.Context) *zerolog.Logger {
	if log, ok := c.Get(LoggerKey).(*zerolog.Logger); ok {
		return log
	}
	log := zerolog.Nop()
	return &log
}
