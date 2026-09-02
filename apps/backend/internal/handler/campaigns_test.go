package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/errs"
	"flux/apps/backend/internal/handler"
	"flux/apps/backend/internal/model/campaign"
	"flux/apps/backend/internal/model/link"
	"flux/apps/backend/internal/repository"
	"flux/apps/backend/internal/service"
	pkgtesting "flux/apps/backend/internal/testing"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()
	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)

	pool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	return pool, func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}
}

// mockAuthMiddleware injects a fake tenant into the request context
func mockAuthMiddleware(tenantID uuid.UUID) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("tenant_id", tenantID)
			return next(c)
		}
	}
}

func TestCampaignAPI_MultiTenantIsolation(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	campRepo := repository.NewCampaignRepository(pool)
	
	linkRepo := repository.NewLinkRepository(pool)
	campSvc := service.NewCampaignService(campRepo, linkRepo, nil)
	campHandler := handler.NewCampaignHandler(campSvc)

	
	linkSvc := service.NewLinkService(linkRepo, nil, campRepo, nil)
	linkHandler := handler.NewLinksHandler(linkSvc)

	e := echo.New()
	e.HTTPErrorHandler = errs.CustomHTTPErrorHandler
	
	// Setup workspaces (we just need arbitrary UUIDs for testing isolation)
	workspaceA := uuid.New()
	workspaceB := uuid.New()

	// Insert workspaces directly to satisfy FK
	_, err := pool.Exec(context.Background(), "INSERT INTO workspaces (id, name) VALUES ($1, $2), ($3, $4)",
		workspaceA, "Workspace A",
		workspaceB, "Workspace B")
	require.NoError(t, err)

	// Create APIs for A and B
	apiA := e.Group("/a")
	apiA.Use(mockAuthMiddleware(workspaceA))
	apiA.POST("/campaigns", campHandler.CreateCampaign)
	apiA.GET("/campaigns/:id", campHandler.GetCampaign)
	apiA.PATCH("/campaigns/:id", campHandler.UpdateCampaign)
	apiA.DELETE("/campaigns/:id", campHandler.DeleteCampaign)
	apiA.POST("/links", linkHandler.CreateLink)

	apiB := e.Group("/b")
	apiB.Use(mockAuthMiddleware(workspaceB))
	apiB.GET("/campaigns/:id", campHandler.GetCampaign)
	apiB.POST("/links", linkHandler.CreateLink)

	// --- Step 1: Workspace A creates a campaign ---
	payloadA := campaign.CreateCampaignPayload{Name: "Campaign A"}
	bodyA, _ := json.Marshal(payloadA)
	reqA := httptest.NewRequest(http.MethodPost, "/a/campaigns", bytes.NewReader(bodyA))
	reqA.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recA := httptest.NewRecorder()
	e.ServeHTTP(recA, reqA)

	assert.Equal(t, http.StatusCreated, recA.Code)
	var createdCampA campaign.Campaign
	json.Unmarshal(recA.Body.Bytes(), &createdCampA)
	assert.Equal(t, "Campaign A", createdCampA.Name)
	assert.Equal(t, workspaceA, createdCampA.WorkspaceID)

	// --- Step 2: Workspace B tries to GET Workspace A's campaign ---
	reqB := httptest.NewRequest(http.MethodGet, "/b/campaigns/"+createdCampA.ID.String(), nil)
	recB := httptest.NewRecorder()
	e.ServeHTTP(recB, reqB)

	assert.Equal(t, http.StatusNotFound, recB.Code) // B cannot access A's campaign

	// --- Step 3: Workspace B attempts to create a Link assigned to Workspace A's campaign ---
	payloadLinkB := link.CreateLinkPayload{
		DestinationURL: "https://b.com",
		CampaignID:     &createdCampA.ID, // Forging the campaign ID!
	}
	bodyLinkB, _ := json.Marshal(payloadLinkB)
	reqLinkB := httptest.NewRequest(http.MethodPost, "/b/links", bytes.NewReader(bodyLinkB))
	reqLinkB.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLinkB := httptest.NewRecorder()
	e.ServeHTTP(recLinkB, reqLinkB)

	// Should be rejected because Campaign A does not belong to Workspace B
	assert.Equal(t, http.StatusBadRequest, recLinkB.Code)
	assert.Contains(t, recLinkB.Body.String(), "cross-workspace association denied")

	// --- Step 4: Workspace A creates a Link assigned to Workspace A's campaign ---
	payloadLinkA := link.CreateLinkPayload{
		DestinationURL: "https://a.com",
		CampaignID:     &createdCampA.ID, // Valid
	}
	bodyLinkA, _ := json.Marshal(payloadLinkA)
	reqLinkA := httptest.NewRequest(http.MethodPost, "/a/links", bytes.NewReader(bodyLinkA))
	reqLinkA.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recLinkA := httptest.NewRecorder()
	e.ServeHTTP(recLinkA, reqLinkA)

	assert.Equal(t, http.StatusCreated, recLinkA.Code)
	var createdLinkA link.Link
	json.Unmarshal(recLinkA.Body.Bytes(), &createdLinkA)
	assert.Equal(t, createdCampA.ID, *createdLinkA.CampaignID)

	// --- Step 5: Verify ON DELETE SET NULL behavior ---
	reqDel := httptest.NewRequest(http.MethodDelete, "/a/campaigns/"+createdCampA.ID.String(), nil)
	recDel := httptest.NewRecorder()
	e.ServeHTTP(recDel, reqDel)
	assert.Equal(t, http.StatusNoContent, recDel.Code)

	// Verify the Link still exists and its campaign_id is NULL
	dbLink, err := linkRepo.GetLinkByID(context.Background(), &workspaceA, createdLinkA.ID)
	require.NoError(t, err)
	assert.NotNil(t, dbLink)
	assert.Nil(t, dbLink.CampaignID) // Must be NULL now
}
