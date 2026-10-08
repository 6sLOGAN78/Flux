// Package middleware applies request identity, authentication, tracing, and HTTP safeguards.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/sqlerr"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// GlobalMiddlewares collects HTTP boundary middleware.
type GlobalMiddlewares struct {
	server *server.Server
}

// NewGlobalMiddlewares constructs HTTP middleware from server settings.
func NewGlobalMiddlewares(s *server.Server) *GlobalMiddlewares {
	return &GlobalMiddlewares{
		server: s,
	}
}

// CORS applies the configured cross-origin policy.
func (global *GlobalMiddlewares) CORS() echo.MiddlewareFunc {
	origins := global.server.Config.Server.CORSAllowedOrigins
	cors := middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) { return ExactOrigin(origin, origins), nil },
		AllowMethods: []string{
			http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions,
		},
		AllowHeaders:  []string{"Authorization", "Content-Type", "Idempotency-Key", "If-Match"},
		ExposeHeaders: []string{"X-Request-ID", "Retry-After", "X-RateLimit-Limit"},
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		allowed := cors(next)
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/api/v1/") {
				c.Response().Header().Set("Cache-Control", "no-store")
				if c.Request().Method == http.MethodOptions {
					values := c.Request().Header.Values("Origin")
					if len(values) != 1 || !ExactOrigin(values[0], origins) {
						return echo.NewHTTPError(http.StatusForbidden)
					}
				}
			}
			return allowed(c)
		}
	}
}

// RequestLogger logs safe request metadata after handler execution.
func (global *GlobalMiddlewares) RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			status := c.Response().Status
			log := GetLogger(c)
			var event *zerolog.Event
			switch {
			case status >= http.StatusInternalServerError:
				event = log.Error()
			case status >= http.StatusBadRequest:
				event = log.Warn()
			default:
				event = log.Info()
			}
			event.Str("operation",
				"http.request").
				Str("http.request.method",
					SafeMethod(c.Request().
						Method)).
				Str("http.route",
					SafeRoute(c)).
				Int("http.response.status_code",
					status).
				Msg("http.request")
			return nil
		}
	}
}

// Recover converts panics to the HTTP error boundary.
func (global *GlobalMiddlewares) Recover() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisablePrintStack:   true,
		DisableErrorHandler: true,
		LogErrorFunc: func(_ echo.Context, _ error, _ []byte) error {
			return errors.New("operation failed")
		},
	})
}

// Secure applies secure HTTP response headers.
func (global *GlobalMiddlewares) Secure() echo.MiddlewareFunc {
	return middleware.Secure()
}

// GlobalErrorHandler serializes safe public errors and logs their category.
func (global *GlobalMiddlewares) GlobalErrorHandler(err error, c echo.Context) {
	// First try to handle database errors and convert them to appropriate HTTP errors

	// Try to handle known database errors
	// Only do this for errors that haven't already been converted to HTTPError
	var httpErr *errs.HTTPError
	if !errors.As(err, &httpErr) {
		var echoErr *echo.HTTPError
		if errors.As(err, &echoErr) {
			if echoErr.Code == http.StatusNotFound {
				err = errs.NewNotFoundError("Route not found", false, nil)
			}
		} else {
			// Here we call our sqlerr handler which will convert database errors
			// to appropriate application errors
			err = sqlerr.HandleError(err)
		}
	}

	// Now process the possibly converted error
	var echoErr *echo.HTTPError
	var status int
	var code string
	var message string
	var fieldErrors []errs.FieldError
	var action *errs.Action

	switch {
	case errors.As(err, &httpErr):
		status = httpErr.Status
		code = httpErr.Code
		message = httpErr.Message
		fieldErrors = httpErr.Errors
		action = httpErr.Action

	case errors.As(err, &echoErr):
		status = echoErr.Code
		code = errs.MakeUpperCaseWithUnderscores(http.StatusText(status))
		message = http.StatusText(status)
		if status == http.StatusTooManyRequests && echoErr.Message == "Rate limit exceeded" {
			message = "Rate limit exceeded"
		}

	default:
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(
			http.StatusText(http.StatusInternalServerError))
		message = http.StatusText(http.StatusInternalServerError)
	}

	// Telemetry records fixed operational fields. Public typed validation stays intact.
	if status == http.StatusServiceUnavailable {
		c.Response().Header().Set("Retry-After", "1")
	}
	GetLogger(c).
		Error().
		Str("operation",
			"http.request").
		Str("error.category",
			"unknown").
		Int("http.response.status_code",
			status).
		Msg("http.request")

	if !c.Response().Committed {
		_ = c.JSON(status, errs.HTTPError{
			Code:     code,
			Message:  message,
			Status:   status,
			Override: httpErr != nil && httpErr.Override,
			Errors:   fieldErrors,
			Action:   action,
		})
	}
}
