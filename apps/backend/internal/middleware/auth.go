package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"flux/apps/backend/internal/repository"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/labstack/echo/v4"
)

// ClerkJWTMiddleware validates Bearer JWT tokens using Clerk SDK and syncs identity.
func ClerkJWTMiddleware(userRepo *repository.UserRepository) echo.MiddlewareFunc {
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
			
			token := parts[1]

			// Verify Clerk token (assuming clerk-sdk is initialized with CLERK_SECRET_KEY env var in main)
			claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
				Token: token,
			})
			if err != nil {
				fmt.Println("Clerk Token validation error:", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired clerk access token")
			}

			// JIT Identity Sync
			// Email isn't always in claims, but typically we can use Subject as Clerk ID.
			// Let's extract email/name if we have them. 
			// We can parse the standard claims.
			email := ""
			clerkUserID := claims.Subject

			u, err := userRepo.SyncUser(c.Request().Context(), clerkUserID, email, "")
			if err != nil {
				fmt.Println("Failed to sync user:", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "identity resolution failed")
			}
			
			c.Set("user_id", u.ID)
			c.Set("clerk_user_id", clerkUserID)

			// Resolve Organization
			// By default, org_id is available if the token was issued for an organization context
			activeOrgID := claims.ActiveOrganizationID

			var orgName string
			if activeOrgID != "" {
				orgName = claims.ActiveOrganizationSlug
				if orgName == "" {
					orgName = "Organization"
				}
			} else {
				orgName = "Personal Workspace"
			}

			w, err := userRepo.SyncWorkspace(c.Request().Context(), activeOrgID, orgName, u.ID)
			if err != nil {
				fmt.Println("Failed to sync workspace:", err)
				return echo.NewHTTPError(http.StatusInternalServerError, "workspace resolution failed")
			}

			c.Set("tenant_id", w.ID)

			activeRole := claims.ActiveOrganizationRole
			if activeRole != "" {
				c.Set("tenant_role", activeRole)
			}

			return next(c)
		}
	}
}
