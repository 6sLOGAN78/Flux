package middleware

import (
	"net/http"
	"strings"

	"github.com/6sLOGAN78/flux/internal/errs"
	"github.com/6sLOGAN78/flux/internal/server"
	"github.com/6sLOGAN78/flux/internal/service"
	"github.com/labstack/echo/v4"
)

// AuthMiddleware verifies bearer identity without provider organization authority.
type AuthMiddleware struct{ auth *service.AuthService }

// NewAuthMiddleware constructs fixed-endpoint authentication dependencies.
func NewAuthMiddleware(s *server.Server) *AuthMiddleware {
	return &AuthMiddleware{auth: service.NewAuthService(s)}
}

// RequireAuth requires a verified actor; ambient cookies never authenticate requests.
func (auth *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return RequireActor(auth.auth)(next)
}

// RequireActor installs verified identity, never workspace permissions.
func RequireActor(auth *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-store")
			values := c.Request().Header.Values("Authorization")
			if len(values) != 1 {
				return errs.NewUnauthorizedError("Authentication required", false)
			}
			scheme, token, ok := strings.Cut(values[0], " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") {
				return errs.NewUnauthorizedError("Authentication required", false)
			}
			actor, err := auth.Authenticate(c.Request().Context(), token)
			if err != nil {
				return err
			}
			c.Set("actor", actor)
			return next(c)
		}
	}
}

// ProductSession protects installed versioned routes, including future mutations.
func ProductSession(auth *service.AuthService, origins []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		protected := BrowserMutation(origins)(RequireActor(auth)(next))
		return func(c echo.Context) error {
			if !strings.HasPrefix(c.Request().URL.Path, "/api/v1/") || c.Request().Method == http.MethodOptions {
				return next(c)
			}
			return protected(c)
		}
	}
}
