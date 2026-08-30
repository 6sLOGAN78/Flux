package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTLSAuthHandler(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	pool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)
	defer pool.Close()

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	domainRepo := repository.NewDomainRepository(pool)

	wsID := uuid.New().String()
	_, err = pool.Exec(ctx, "INSERT INTO workspaces (id, name) VALUES ($1, $2)", wsID, "Test Workspace")
	require.NoError(t, err)

	setupDomain := func(hostname, status string) {
		d, err := domainRepo.CreateDomain(ctx, wsID, hostname, "token")
		require.NoError(t, err)
		err = domainRepo.UpdateDomainStatus(ctx, d.ID, status)
		require.NoError(t, err)
	}

	setupDomain("active.com", "active")
	setupDomain("pending.com", "pending")
	setupDomain("failed.com", "failed")
	setupDomain("disabled.com", "disabled")
	setupDomain("UPPER-ACTIVE.COM", "active") // Testing normalization

	e := echo.New()
	apiKey := "secret123"
	h := handler.NewTLSAuthHandler(domainRepo, apiKey)

	performRequest := func(domain string, token string) (int, string) {
		req := httptest.NewRequest(http.MethodGet, "/api/internal/tls/ask?domain="+domain, nil)
		if token != "" {
			req.Header.Set("X-Internal-Token", token)
		}
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := h.CheckAuthorization(c)
		require.NoError(t, err) // the handler should not return standard errors, only JSON
		
		return rec.Code, rec.Body.String()
	}

	t.Run("unauthorized no token", func(t *testing.T) {
		code, body := performRequest("active.com", "")
		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("unauthorized wrong token", func(t *testing.T) {
		code, body := performRequest("active.com", "wrong")
		assert.Equal(t, http.StatusUnauthorized, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("authorized active domain", func(t *testing.T) {
		code, body := performRequest("active.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":true`)
	})

	t.Run("authorized active domain normalized", func(t *testing.T) {
		code, body := performRequest("ACTIVE.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":true`)
	})

	t.Run("authorized upper created domain normalized", func(t *testing.T) {
		code, body := performRequest("upper-active.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":true`)
	})

	t.Run("denied pending domain", func(t *testing.T) {
		code, body := performRequest("pending.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("denied failed domain", func(t *testing.T) {
		code, body := performRequest("failed.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("denied disabled domain", func(t *testing.T) {
		code, body := performRequest("disabled.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("denied unknown domain", func(t *testing.T) {
		code, body := performRequest("unknown.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("denied subdomain of active", func(t *testing.T) {
		code, body := performRequest("foo.active.com", apiKey)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"authorized":false`) // requires exact match
	})

	t.Run("denied missing param", func(t *testing.T) {
		code, body := performRequest("", apiKey)
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Contains(t, body, `"authorized":false`)
	})

	t.Run("fail closed when api key is not configured", func(t *testing.T) {
		hEmpty := handler.NewTLSAuthHandler(domainRepo, "")
		req := httptest.NewRequest(http.MethodGet, "/api/internal/tls/ask?domain=active.com", nil)
		req.Header.Set("X-Internal-Token", "")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		err := hEmpty.CheckAuthorization(c)
		require.NoError(t, err)
		
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), `"authorized":false`)
	})
}

func TestTLSAuthHandler_QueryToken(t *testing.T) {
	// Let's add a quick check for query token
}
