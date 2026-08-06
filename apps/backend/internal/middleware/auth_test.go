package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/middleware"
	"flux/apps/backend/internal/modules/auth"

	"github.com/labstack/echo/v4"
)

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	authSvc := auth.NewAuthService("test-secret", "")
	mw := middleware.JWTMiddleware(authSvc)

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	if err == nil {
		t.Fatalf("expected error for missing Authorization header, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 Unauthorized, got: %v", err)
	}
}

func TestJWTMiddleware_InvalidHeaderFormat(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic invalidtokenformat")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	authSvc := auth.NewAuthService("test-secret", "")
	mw := middleware.JWTMiddleware(authSvc)

	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	if err == nil {
		t.Fatalf("expected error for invalid Bearer format, got nil")
	}

	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 Unauthorized, got: %v", err)
	}
}

func TestJWTMiddleware_ValidBearerToken(t *testing.T) {
	authSvc := auth.NewAuthService("test-secret", "")
	accessToken, _, err := authSvc.GenerateTokenPair("user_999", "user999@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := middleware.JWTMiddleware(authSvc)

	var retrievedUserID, retrievedEmail string
	handler := mw(func(c echo.Context) error {
		retrievedUserID, _ = c.Get("user_id").(string)
		retrievedEmail, _ = c.Get("email").(string)
		return c.String(http.StatusOK, "success")
	})

	err = handler(c)
	if err != nil {
		t.Fatalf("expected no error for valid token, got: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 OK, got: %d", rec.Code)
	}

	if retrievedUserID != "user_999" {
		t.Errorf("expected user_id 'user_999', got '%s'", retrievedUserID)
	}

	if retrievedEmail != "user999@example.com" {
		t.Errorf("expected email 'user999@example.com', got '%s'", retrievedEmail)
	}
}
