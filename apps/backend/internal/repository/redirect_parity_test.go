package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flux/apps/backend/internal/database"
	"flux/apps/backend/internal/repository"
	pkgtesting "flux/apps/backend/internal/testing"
	"github.com/rs/zerolog"
)

func setupParityTestDB(t *testing.T) (*pgxpool.Pool, func()) {
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

func TestRedirectParity_CacheHitMiss_UTMResolution(t *testing.T) {
	ctx := context.Background()
	
	pool, cleanupDB := setupParityTestDB(t)
	defer cleanupDB()

	redisContainer, err := pkgtesting.SetupRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisContainer.Address,
	})
	defer redisClient.Close()

	pgRepo := repository.NewPostgresRedirectRepository(pool)
	cacheRepo := repository.NewRedisRedirectCache(redisClient)

	// Setup Test Data
	workspaceID := uuid.New()
	_, err = pool.Exec(ctx, "INSERT INTO workspaces (id, name) VALUES ($1, $2)", workspaceID, "Parity Workspace")
	require.NoError(t, err)

	// Create Campaign with default utm_campaign
	campaignID := uuid.New()
	defaultUTMCampaign := "summer-sale"
	_, err = pool.Exec(ctx, `
		INSERT INTO campaigns (id, workspace_id, name, utm_campaign) 
		VALUES ($1, $2, $3, $4)
	`, campaignID, workspaceID, "Summer Campaign", defaultUTMCampaign)
	require.NoError(t, err)

	// Case 1: Link with Campaign, NO link-level UTMs -> Should inherit Campaign's default
	shortCode1 := "link1"
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, campaign_id) 
		VALUES (gen_random_uuid(), $1, $2, 'https://test1.com', $3)
	`, workspaceID, shortCode1, campaignID)
	require.NoError(t, err)

	// Case 2: Link with Campaign AND link-level UTMs -> Should override Campaign's default
	shortCode2 := "link2"
	overrideUTMCampaign := "summer-special"
	overrideSource := "twitter"
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, campaign_id, utm_campaign, utm_source) 
		VALUES (gen_random_uuid(), $1, $2, 'https://test2.com', $3, $4, $5)
	`, workspaceID, shortCode2, campaignID, overrideUTMCampaign, overrideSource)
	require.NoError(t, err)
	
	// Case 3: Link with NO Campaign -> Should just use Link's UTMs
	shortCode3 := "link3"
	noCampSource := "newsletter"
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, utm_source) 
		VALUES (gen_random_uuid(), $1, $2, 'https://test3.com', $3)
	`, workspaceID, shortCode3, noCampSource)
	require.NoError(t, err)


	// --- Test Case 1: Default resolution ---
	// 1. Cache MISS (fetch from Postgres)
	target1Miss, err := pgRepo.GetByHostAndSlug(ctx, "", shortCode1)
	require.NoError(t, err)
	assert.Equal(t, campaignID.String(), *target1Miss.CampaignID)
	assert.Equal(t, defaultUTMCampaign, *target1Miss.UTMCampaign)
	assert.Nil(t, target1Miss.UTMSource)

	// Cache it
	err = cacheRepo.Set(ctx, "", shortCode1, target1Miss, time.Minute)
	require.NoError(t, err)

	// 2. Cache HIT
	target1Hit, err := cacheRepo.Get(ctx, "", shortCode1)
	require.NoError(t, err)

	// Ensure Parity
	assert.Equal(t, target1Miss.CampaignID, target1Hit.CampaignID)
	assert.Equal(t, target1Miss.UTMCampaign, target1Hit.UTMCampaign)
	assert.Equal(t, target1Miss.UTMSource, target1Hit.UTMSource)


	// --- Test Case 2: Override resolution ---
	target2Miss, err := pgRepo.GetByHostAndSlug(ctx, "", shortCode2)
	require.NoError(t, err)
	assert.Equal(t, campaignID.String(), *target2Miss.CampaignID)
	assert.Equal(t, overrideUTMCampaign, *target2Miss.UTMCampaign)
	assert.Equal(t, overrideSource, *target2Miss.UTMSource)

	err = cacheRepo.Set(ctx, "", shortCode2, target2Miss, time.Minute)
	require.NoError(t, err)

	target2Hit, err := cacheRepo.Get(ctx, "", shortCode2)
	require.NoError(t, err)

	assert.Equal(t, target2Miss.CampaignID, target2Hit.CampaignID)
	assert.Equal(t, target2Miss.UTMCampaign, target2Hit.UTMCampaign)
	assert.Equal(t, target2Miss.UTMSource, target2Hit.UTMSource)


	// --- Test Case 3: No Campaign resolution ---
	target3Miss, err := pgRepo.GetByHostAndSlug(ctx, "", shortCode3)
	require.NoError(t, err)
	assert.Nil(t, target3Miss.CampaignID)
	assert.Nil(t, target3Miss.UTMCampaign)
	assert.Equal(t, noCampSource, *target3Miss.UTMSource)

	err = cacheRepo.Set(ctx, "", shortCode3, target3Miss, time.Minute)
	require.NoError(t, err)

	target3Hit, err := cacheRepo.Get(ctx, "", shortCode3)
	require.NoError(t, err)

	assert.Equal(t, target3Miss.CampaignID, target3Hit.CampaignID)
	assert.Equal(t, target3Miss.UTMCampaign, target3Hit.UTMCampaign)
	assert.Equal(t, target3Miss.UTMSource, target3Hit.UTMSource)
}

func TestRedirectParity_CacheHitMiss_CustomDomains(t *testing.T) {
	ctx := context.Background()
	
	pool, cleanupDB := setupParityTestDB(t)
	defer cleanupDB()

	redisContainer, err := pkgtesting.SetupRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisContainer.Address,
	})
	defer redisClient.Close()

	pgRepo := repository.NewPostgresRedirectRepository(pool)
	cacheRepo := repository.NewRedisRedirectCache(redisClient)

	// Setup Test Data
	workspaceID := uuid.New()
	_, err = pool.Exec(ctx, "INSERT INTO workspaces (id, name) VALUES ($1, $2)", workspaceID, "Domain Workspace")
	require.NoError(t, err)

	domainID := uuid.New()
	hostname := "analytics-parity.com"
	_, err = pool.Exec(ctx, `
		INSERT INTO custom_domains (id, tenant_id, hostname, status, verification_token) 
		VALUES ($1, $2, $3, 'active', 'test')
	`, domainID, workspaceID, hostname)
	require.NoError(t, err)

	shortCode := "domainlink"
	_, err = pool.Exec(ctx, `
		INSERT INTO links (id, tenant_id, short_code, destination_url, custom_domain_id) 
		VALUES (gen_random_uuid(), $1, $2, 'https://test1.com', $3)
	`, workspaceID, shortCode, domainID)
	require.NoError(t, err)

	// 1. Cache MISS
	targetMiss, err := pgRepo.GetByHostAndSlug(ctx, hostname, shortCode)
	require.NoError(t, err)
	assert.Equal(t, domainID.String(), *targetMiss.CustomDomainID)
	assert.Equal(t, hostname, targetMiss.Hostname)

	// Cache it
	err = cacheRepo.Set(ctx, hostname, shortCode, targetMiss, time.Minute)
	require.NoError(t, err)

	// 2. Cache HIT
	targetHit, err := cacheRepo.Get(ctx, hostname, shortCode)
	require.NoError(t, err)

	// Ensure Parity
	assert.Equal(t, targetMiss.CustomDomainID, targetHit.CustomDomainID)
	assert.Equal(t, targetMiss.Hostname, targetHit.Hostname)
}
