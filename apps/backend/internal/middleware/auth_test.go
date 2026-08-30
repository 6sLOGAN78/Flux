package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/repository"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestClerkJWTMiddleware_IgnoresSpoofingHeaders(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	
	// Inject malicious attacker headers
	req.Header.Set("X-Test-Clerk-User-ID", "attacker")
	req.Header.Set("X-Test-Clerk-Org-ID", "another-org")
	req.Header.Set("X-Test-Clerk-Role", "org:admin")
	
	// Because there is no valid Authorization Bearer token, 
	// it should fail immediately with 401 Unauthorized,
	// proving that the test headers cannot bypass auth.
	
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := middleware.ClerkJWTMiddleware(&repository.UserRepository{})

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	assert.Error(t, err)

	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}
