package v1_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v1 "flux/apps/backend/internal/router/v1"

	"github.com/labstack/echo/v4"
)

func TestV1Routes_MeUnauthorizedWithoutHeader(t *testing.T) {
	e := echo.New()
	v1.RegisterV1Routes(e, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 Unauthorized without auth header, got %d", rec.Code)
	}
}
