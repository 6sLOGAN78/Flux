package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flux/apps/backend/internal/repository"
)

func TestRedirectSecurity_CustomDomains(t *testing.T) {
	ctx := context.Background()
	pool, cleanupDB := setupParityTestDB(t) // reuse the postgres setup from parity test
	defer cleanupDB()

	pgRepo := repository.NewPostgresRedirectRepository(pool)

	// Setup Workspaces
	wsA_ID := uuid.New()
	wsB_ID := uuid.New()
	_, err := pool.Exec(ctx, "INSERT INTO workspaces (id, name) VALUES ($1, $2), ($3, $4)", wsA_ID, "Workspace A", wsB_ID, "Workspace B")
	require.NoError(t, err)

	// Setup Domains
	domainA_ID := uuid.New() // ACTIVE
	domainB_ID := uuid.New() // ACTIVE
	domainPending_ID := uuid.New() // PENDING
	domainDisabled_ID := uuid.New() // DISABLED

	_, err = pool.Exec(ctx, `
		INSERT INTO custom_domains (id, tenant_id, hostname, status, verification_token) VALUES 
		($1, $2, 'domain-a.com', 'active', 'tokA'),
		($3, $4, 'domain-b.com', 'active', 'tokB'),
		($5, $2, 'pending.com', 'pending', 'tokP'),
		($6, $2, 'disabled.com', 'disabled', 'tokD')
	`, domainA_ID, wsA_ID, domainB_ID, wsB_ID, domainPending_ID, domainDisabled_ID)
	require.NoError(t, err)

	// Setup Links
	// Link A belongs to Domain A
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, custom_domain_id) 
		VALUES (gen_random_uuid(), $1, 'abc', 'https://a.com', $2)
	`, wsA_ID, domainA_ID)
	require.NoError(t, err)

	// Link B belongs to Domain B (different slug because globally unique)
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, custom_domain_id) 
		VALUES (gen_random_uuid(), $1, 'abc2', 'https://b.com', $2)
	`, wsB_ID, domainB_ID)
	require.NoError(t, err)

	// Link P belongs to Pending Domain
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, custom_domain_id) 
		VALUES (gen_random_uuid(), $1, 'abc-p', 'https://p.com', $2)
	`, wsA_ID, domainPending_ID)
	require.NoError(t, err)

	// Link D belongs to Disabled Domain
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, custom_domain_id) 
		VALUES (gen_random_uuid(), $1, 'abc-d', 'https://d.com', $2)
	`, wsA_ID, domainDisabled_ID)
	require.NoError(t, err)

	// Link NoDomain (platform default)
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url) 
		VALUES (gen_random_uuid(), $1, 'xyz', 'https://platform.com')
	`, wsA_ID)
	require.NoError(t, err)

	t.Run("Correct routing", func(t *testing.T) {
		target, err := pgRepo.GetByHostAndSlug(ctx, "domain-a.com", "abc")
		require.NoError(t, err)
		assert.Equal(t, "https://a.com", target.DestinationURL)
		
		target, err = pgRepo.GetByHostAndSlug(ctx, "domain-b.com", "abc2")
		require.NoError(t, err)
		assert.Equal(t, "https://b.com", target.DestinationURL)
	})

	t.Run("Cross-domain isolation (Domain A trying to get Domain B's link)", func(t *testing.T) {
		// Even though "abc" exists, Domain A shouldn't get it if it belongs to B. 
		// Actually "abc" belongs to both A and B, but they are different DB rows.
		// Let's test a slug that only belongs to B.
		_, err = pool.Exec(ctx, `
			INSERT INTO links (id, tenant_id, short_code, destination_url, custom_domain_id) 
			VALUES (gen_random_uuid(), $1, 'only-b', 'https://b2.com', $2)
		`, wsB_ID, domainB_ID)
		require.NoError(t, err)

		_, err = pgRepo.GetByHostAndSlug(ctx, "domain-a.com", "only-b")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Unknown domain", func(t *testing.T) {
		_, err = pgRepo.GetByHostAndSlug(ctx, "unknown.com", "abc")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Unverified domain", func(t *testing.T) {
		_, err = pgRepo.GetByHostAndSlug(ctx, "pending.com", "abc-p")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Disabled domain", func(t *testing.T) {
		_, err = pgRepo.GetByHostAndSlug(ctx, "disabled.com", "abc-d")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Platform default domain routing", func(t *testing.T) {
		// Empty hostname means platform domain
		target, err := pgRepo.GetByHostAndSlug(ctx, "", "xyz")
		require.NoError(t, err)
		assert.Equal(t, "https://platform.com", target.DestinationURL)
	})
	
	t.Run("Prevent platform domain from accessing custom domain link", func(t *testing.T) {
		// Empty hostname but trying to access link attached to domain A
		_, err = pgRepo.GetByHostAndSlug(ctx, "", "only-b")
		assert.ErrorIs(t, err, repository.ErrNotFound)
	})

	t.Run("Allow custom domain to access platform link owned by same workspace", func(t *testing.T) {
		// domain-a.com trying to access link 'xyz' which has NO custom domain but belongs to wsA
		target, err := pgRepo.GetByHostAndSlug(ctx, "domain-a.com", "xyz")
		require.NoError(t, err)
		assert.Equal(t, "https://platform.com", target.DestinationURL)
	})
}
