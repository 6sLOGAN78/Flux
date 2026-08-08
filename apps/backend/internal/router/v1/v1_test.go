package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "flux/apps/backend/internal/router/v1"
	"flux/apps/backend/internal/service"

	"github.com/labstack/echo/v4"
)

func TestV1Routes_MeUnauthorizedWithoutHeader(t *testing.T) {
	e := echo.New()
	authSvc := service.NewAuthService("jwt-test-secret", "")
	v1.RegisterV1Routes(e, authSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 Unauthorized without auth header, got %d", rec.Code)
	}
}

func TestV1Routes_MeAuthorizedWithValidToken(t *testing.T) {
	e := echo.New()
	authSvc := service.NewAuthService("jwt-test-secret", "")
	v1.RegisterV1Routes(e, authSvc, nil)

	token, _, err := authSvc.GenerateTokenPair("usr_12345", "user@example.com")
	if err != nil {
		t.Fatalf("unexpected error generating token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected HTTP 200 OK with valid bearer token, got %d", rec.Code)
	}
}
