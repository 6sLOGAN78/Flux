package repository_test

import (
	"context"
	"strings"
	"testing"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/model/webhook"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"
	"flux/apps/backend/internal/lib/utils"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping webhook repository integration test in short mode")
	}

	ctx := context.Background()
	pgContainer, err := pkgtesting.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	dbPool, err := database.InitDBPool(ctx, pgContainer.DSN)
	require.NoError(t, err)
	defer dbPool.Close()

	logger := zerolog.Nop()
	err = database.MigrateDSN(ctx, &logger, pgContainer.DSN)
	require.NoError(t, err)

	repo := repository.NewWebhookRepository(dbPool)

	// Create workspace A and B
	workspaceA := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Workspace A')`, workspaceA)
	require.NoError(t, err)

	workspaceB := uuid.New()
	_, err = dbPool.Exec(ctx, `INSERT INTO workspaces (id, name) VALUES ($1, 'Workspace B')`, workspaceB)
	require.NoError(t, err)

	t.Run("Security: Random Secret Generation", func(t *testing.T) {
		sec1, err := utils.GenerateWebhookSecret()
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(sec1, "whsec_"))
		assert.Equal(t, len("whsec_")+64, len(sec1))

		sec2, err := utils.GenerateWebhookSecret()
		require.NoError(t, err)
		assert.NotEqual(t, sec1, sec2) // Must be distinct!
	})

	t.Run("Security: Tenant Isolation", func(t *testing.T) {
		secA, _ := utils.GenerateWebhookSecret()
		
		whA := &webhook.Webhook{
			WorkspaceID: workspaceA,
			EndpointURL: "https://example.com/webhook",
			Secret:      secA,
			Active:      true,
			Events:      []string{"link.redirect"},
		}

		createdA, err := repo.CreateWebhook(ctx, whA)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, createdA.ID)
		assert.Equal(t, "https://example.com/webhook", createdA.EndpointURL)

		// Workspace B attempts to read Workspace A's webhook
		_, err = repo.GetWebhook(ctx, workspaceB, createdA.ID)
		require.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err) // Safely returns not found, leaking no info

		// Workspace B attempts to update Workspace A's webhook
		falseVal := false
		_, err = repo.UpdateWebhook(ctx, workspaceB, createdA.ID, &falseVal, nil, nil)
		require.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)

		// Workspace B attempts to delete Workspace A's webhook
		err = repo.DeleteWebhook(ctx, workspaceB, createdA.ID)
		require.Error(t, err)
		assert.Equal(t, repository.ErrNotFound, err)

		// Workspace A can successfully read it
		fetchedA, err := repo.GetWebhook(ctx, workspaceA, createdA.ID)
		require.NoError(t, err)
		assert.Equal(t, secA, fetchedA.Secret)

		// Workspace A can update it
		urlB := "https://example.com/new"
		updatedA, err := repo.UpdateWebhook(ctx, workspaceA, createdA.ID, nil, &urlB, nil)
		require.NoError(t, err)
		assert.Equal(t, urlB, updatedA.EndpointURL)
		
		// Ensure events persist
		assert.Contains(t, updatedA.Events, "link.redirect")
	})

	t.Run("Security: Invalid URL Validation", func(t *testing.T) {
		assert.False(t, utils.IsValidURL("not-a-url"))
		assert.False(t, utils.IsValidURL("ftp://bad-scheme.com"))
		assert.True(t, utils.IsValidURL("https://webhook.site/123"))
		assert.True(t, utils.IsValidURL("http://localhost:8080")) // Development allowed
	})

	t.Run("Security: Foreign Workspace IDs Rejected", func(t *testing.T) {
		sec, _ := utils.GenerateWebhookSecret()
		wh := &webhook.Webhook{
			WorkspaceID: uuid.New(), // non-existent
			EndpointURL: "https://example.com/webhook",
			Secret:      sec,
			Active:      true,
			Events:      []string{"link.redirect"},
		}
		
		_, err := repo.CreateWebhook(ctx, wh)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "violates foreign key constraint")
	})
}
