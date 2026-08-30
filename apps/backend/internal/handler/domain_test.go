package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/rs/zerolog"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"
)

func TestDomainAPI(t *testing.T) {
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

	redisContainer, err := pkgtesting.SetupRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	redisClient := redis.NewClient(&redis.Options{Addr: redisContainer.Address})
	cache := repository.NewRedisRedirectCache(redisClient)

	domainRepo := repository.NewDomainRepository(pool)
	domainSvc := service.NewDomainService(domainRepo, cache, "flux.ly")
	domainHandler := handler.NewDomainHandler(domainSvc)

	e := echo.New()
	e.HTTPErrorHandler = errs.CustomHTTPErrorHandler
	
	// Setup user & workspaces
	clerkUserID := "user_" + uuid.NewString()
	userID := uuid.New()
	wsA := uuid.New()
	wsB := uuid.New()

	_, err = pool.Exec(ctx, "INSERT INTO workspaces (id, name) VALUES ($1, $2), ($3, $4)", wsA, "WS A", wsB, "WS B")
	require.NoError(t, err)

	createDomainRequest := func(tenantID, reqBody string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewBufferString(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set("x-tenant-id", tenantID)
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		c.Set("user_id", userID.String())
		c.Set("clerk_user_id", clerkUserID)
		c.Set("tenant_id", tenantID)

		if err := domainHandler.CreateDomain(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		return rec
	}

	getDomainsRequest := func(tenantID string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains", nil)
		req.Header.Set("x-tenant-id", tenantID)
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		
		c.Set("user_id", userID.String())
		c.Set("clerk_user_id", clerkUserID)
		c.Set("tenant_id", tenantID)

		if err := domainHandler.GetDomains(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		return rec
	}

	getDomainRequest := func(tenantID, id string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/domains/"+id, nil)
		req.Header.Set("x-tenant-id", tenantID)
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)
		
		c.Set("user_id", userID.String())
		c.Set("clerk_user_id", clerkUserID)
		c.Set("tenant_id", tenantID)

		if err := domainHandler.GetDomain(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		return rec
	}

	deleteDomainRequest := func(tenantID, id string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/"+id, nil)
		req.Header.Set("x-tenant-id", tenantID)
		
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id)
		
		c.Set("user_id", userID.String())
		c.Set("clerk_user_id", clerkUserID)
		c.Set("tenant_id", tenantID)

		if err := domainHandler.DeleteDomain(c); err != nil {
			e.HTTPErrorHandler(err, c)
		}
		return rec
	}


	t.Run("Create valid domain", func(t *testing.T) {
		rec := createDomainRequest(wsA.String(), `{"hostname":"valid-domain.com"}`)
		assert.Equal(t, http.StatusCreated, rec.Code)
		
		var res map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &res)
		assert.Equal(t, "valid-domain.com", res["hostname"])
		assert.Equal(t, "pending", res["status"])
		assert.Contains(t, res["verification_token"], "flux-verify=")
	})

	t.Run("Create invalid domain", func(t *testing.T) {
		rec := createDomainRequest(wsA.String(), `{"hostname":"invalid/host"}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Create duplicate domain", func(t *testing.T) {
		rec1 := createDomainRequest(wsA.String(), `{"hostname":"dup-domain.com"}`)
		assert.Equal(t, http.StatusCreated, rec1.Code)
		
		rec2 := createDomainRequest(wsB.String(), `{"hostname":"dup-domain.com"}`)
		assert.Equal(t, http.StatusConflict, rec2.Code)
	})

	t.Run("List domains isolated by workspace", func(t *testing.T) {
		recA := getDomainsRequest(wsA.String())
		var resA map[string][]map[string]interface{}
		json.Unmarshal(recA.Body.Bytes(), &resA)
		assert.Len(t, resA["data"], 2) // valid-domain.com, dup-domain.com

		recB := getDomainsRequest(wsB.String())
		var resB map[string][]map[string]interface{}
		json.Unmarshal(recB.Body.Bytes(), &resB)
		assert.Len(t, resB["data"], 0)
	})

	t.Run("Get domain cross-tenant fails safely", func(t *testing.T) {
		// Create in A
		recA := createDomainRequest(wsA.String(), `{"hostname":"cross-tenant.com"}`)
		var resA map[string]interface{}
		json.Unmarshal(recA.Body.Bytes(), &resA)
		id := resA["id"].(string)

		// Get from B
		recB := getDomainRequest(wsB.String(), id)
		assert.Equal(t, http.StatusNotFound, recB.Code) // Safe 404, not 403 or 500
	})

	t.Run("Delete domain cross-tenant fails safely", func(t *testing.T) {
		// Create in A
		recA := createDomainRequest(wsA.String(), `{"hostname":"delete-cross.com"}`)
		var resA map[string]interface{}
		json.Unmarshal(recA.Body.Bytes(), &resA)
		id := resA["id"].(string)

		// Delete from B
		recB := deleteDomainRequest(wsB.String(), id)
		assert.Equal(t, http.StatusNotFound, recB.Code) // Safe 404
	})

	t.Run("Delete domain successful", func(t *testing.T) {
		// Create in A
		recA := createDomainRequest(wsA.String(), `{"hostname":"delete-me.com"}`)
		var resA map[string]interface{}
		json.Unmarshal(recA.Body.Bytes(), &resA)
		id := resA["id"].(string)

		// Delete from A
		recDel := deleteDomainRequest(wsA.String(), id)
		assert.Equal(t, http.StatusNoContent, recDel.Code)

		// Verify deleted
		recGet := getDomainRequest(wsA.String(), id)
		assert.Equal(t, http.StatusNotFound, recGet.Code)
	})

	t.Run("Normalize hostname correctly", func(t *testing.T) {
		rec := createDomainRequest(wsA.String(), `{"hostname":"Upper-Case.COM:443."}`)
		assert.Equal(t, http.StatusCreated, rec.Code)
		
		var res map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &res)
		assert.Equal(t, "upper-case.com", res["hostname"])
	})

	t.Run("Reject platform domains", func(t *testing.T) {
		rec := createDomainRequest(wsA.String(), `{"hostname":"flux.ly"}`)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		rec2 := createDomainRequest(wsA.String(), `{"hostname":"app.flux.ly"}`)
		assert.Equal(t, http.StatusBadRequest, rec2.Code)
	})
}
