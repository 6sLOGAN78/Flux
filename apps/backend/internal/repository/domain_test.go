package repository_test

import (
	"context"
	"testing"
	"time"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
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

func TestDomainRepository(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pool, cleanup := setupTestDB(t)
	defer cleanup()

	domainRepo := repository.NewDomainRepository(pool)

	wsA_ID := uuid.New().String()
	wsB_ID := uuid.New().String()

	// Create Workspace A
	_, err := pool.Exec(ctx, "INSERT INTO workspaces (id, clerk_org_id, name) VALUES ($1, $2, $3)", wsA_ID, "org_a", "Workspace A")
	require.NoError(t, err)

	// Create Workspace B
	_, err = pool.Exec(ctx, "INSERT INTO workspaces (id, clerk_org_id, name) VALUES ($1, $2, $3)", wsB_ID, "org_b", "Workspace B")
	require.NoError(t, err)

	// 1. Domain creation
	d1, err := domainRepo.CreateDomain(ctx, wsA_ID, "example.com", "token123")
	if err != nil {
		t.Fatalf("failed to create domain d1: %v", err)
	}
	if d1.Hostname != "example.com" {
		t.Errorf("expected hostname example.com, got %s", d1.Hostname)
	}

	// 2. Duplicate hostname (different workspace) should fail
	_, err = domainRepo.CreateDomain(ctx, wsB_ID, "example.com", "token456")
	if err == nil {
		t.Errorf("expected error when creating duplicate domain in different workspace")
	}

	// 3. Same workspace duplicate should fail
	_, err = domainRepo.CreateDomain(ctx, wsA_ID, "example.com", "token789")
	if err == nil {
		t.Errorf("expected error when creating duplicate domain in same workspace")
	}

	// 4. Case normalization & trailing dots check
	_, err = domainRepo.CreateDomain(ctx, wsB_ID, "EXAMPLE.com", "token000")
	if err == nil {
		t.Errorf("expected error when creating differently-cased duplicate domain")
	}

	_, err = domainRepo.CreateDomain(ctx, wsA_ID, "test.com.", "token_dot")
	if err == nil {
		t.Errorf("expected error when creating domain with trailing dot")
	}

	// 5. Workspace isolation
	_, err = domainRepo.GetDomainByID(ctx, wsB_ID, d1.ID)
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound when requesting domain from wrong workspace, got %v", err)
	}

	d1_fetched, err := domainRepo.GetDomainByID(ctx, wsA_ID, d1.ID)
	if err != nil {
		t.Fatalf("failed to fetch domain with correct workspace: %v", err)
	}
	if d1_fetched.Hostname != "example.com" {
		t.Errorf("hostname mismatch")
	}

	// 6. Link relationship and Domain deletion behavior
	linkID := uuid.New().String()
	_, err = pool.Exec(ctx, "INSERT INTO links (id, short_code, destination_url, tenant_id, custom_domain_id) VALUES ($1, $2, $3, $4, $5)", linkID, "domtest", "https://google.com", wsA_ID, d1.ID)
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	// Delete domain
	err = domainRepo.DeleteDomain(ctx, wsA_ID, d1.ID)
	if err != nil {
		t.Fatalf("failed to delete domain: %v", err)
	}

	// Check if link still exists and custom_domain_id is NULL
	var linkDomainID *string
	err = pool.QueryRow(ctx, "SELECT custom_domain_id FROM links WHERE id = $1", linkID).Scan(&linkDomainID)
	if err != nil {
		t.Fatalf("failed to fetch link after domain deletion: %v", err)
	}
	if linkDomainID != nil {
		t.Errorf("expected custom_domain_id to be NULL after domain deletion, got %v", *linkDomainID)
	}
}
