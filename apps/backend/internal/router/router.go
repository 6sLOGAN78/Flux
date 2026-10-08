// Package router registers role-specific HTTP routes and middleware.
package router

import (
	"net/http"

	"github.com/6sLOGAN78/flux/internal/handler"
	"github.com/6sLOGAN78/flux/internal/middleware"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

const (
	requestsPerSecond = 20
)

// NewRouter installs middleware and registers the role-owned routes.
func NewRouter(s *server.Server, h *handler.Handlers, services *service.Services) *echo.Echo {
	middlewares := middleware.NewMiddlewares(s)

	router := echo.New()

	router.HTTPErrorHandler = middlewares.Global.GlobalErrorHandler

	// Correlation, tracing and recovery wrap early CORS/rate-limit responses.
	router.Use(
		middleware.RequestID(),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if c.Path() == "/api/v1/me" {
					c.Response().Header().Set("Cache-Control", "no-store")
				}
				return next(c)
			}
		},
		middleware.NewTracingMiddleware(s, s.Telemetry).EnhanceTracing(),
		middlewares.ContextEnhancer.EnhanceContext(),
		middlewares.Global.RequestLogger(),
		middlewares.Global.Recover(),
		middlewares.Global.CORS(),
		middlewares.Global.Secure(),
		echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
			Skipper: func(c echo.Context) bool { return c.Path() == "/live" || c.Path() == "/ready" },
			Store:   echoMiddleware.NewRateLimiterMemoryStore(rate.Limit(requestsPerSecond)),
			DenyHandler: func(c echo.Context, _ string, _ error) error {
				middleware.GetLogger(c).
					Warn().
					Str("operation",
						"http.request").
					Str("http.route",
						middleware.SafeRoute(c)).
					Str("http.request.method",
						middleware.SafeMethod(c.Request().
							Method)).
					Msg("http.request")
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
			},
		}),
	)

	// register system routes
	registerSystemRoutes(router, h)

	// register versioned routes
	var auth *service.AuthService
	if services != nil {
		auth = services.Auth
	}
	product := handler.NewProductHandler(auth)
	router.Group("/api/v1").GET("/me", product.Me)

	return router
}
