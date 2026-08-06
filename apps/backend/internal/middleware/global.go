package middleware

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// RegisterGlobalMiddlewares attaches standard CORS, Recover, and Logger middlewares.
func RegisterGlobalMiddlewares(e *echo.Echo) {
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(RequestIDMiddleware())
}
