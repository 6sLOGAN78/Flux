// Package middleware provides Echo v4 middleware handlers for authentication, tracing, and logging.
package middleware

import (
	"net/http"
	"strings"

	"flux/apps/backend/internal/modules/auth"

	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates incoming Bearer tokens using AuthService and sets user context.
func JWTMiddleware(authSvc *auth.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			tokenStr := parts[1]

			// Validate token using AuthService
			claims, err := authSvc.ValidateToken(tokenStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			// Store claims and user identity in Echo context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("claims", claims)

			return next(c)
		}
	}
}
