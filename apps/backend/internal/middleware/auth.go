package middleware

import (
	"net/http"
	"strings"

	"flux/apps/backend/internal/service"

	"github.com/labstack/echo/v4"
)

// JWTMiddleware validates Bearer JWT tokens using AuthService.
func JWTMiddleware(authSvc *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing Authorization header")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid Authorization header format")
			}

			claims, err := authSvc.ValidateToken(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired access token")
			}

			if sub, ok := claims["sub"].(string); ok {
				c.Set("user_id", sub)
			}
			if email, ok := claims["email"].(string); ok {
				c.Set("email", email)
			}

			c.Set("user_claims", claims)
			return next(c)
		}
	}
}
