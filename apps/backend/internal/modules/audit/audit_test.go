package audit_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"flux/apps/backend/internal/modules/audit"
)

func TestAuditLogger_LogAndQuery(t *testing.T) {
	logger := audit.NewAuditLogger()

	orgID := uuid.New()
	userID := uuid.New()

	entry := &audit.AuditLog{
		OrganizationID: &orgID,
		UserID:         userID,
		Action:         "link.created",
		ResourceType:   "link",
		ResourceID:     "ln_12345",
		IPAddress:      "198.51.100.42",
		UserAgent:      "Mozilla/5.0 (Macintosh; Apple Silicon)",
		Metadata: map[string]interface{}{
			"destination": "https://example.com/promo",
			"slug":        "promo2026",
		},
		CreatedAt: time.Now(),
	}

	err := logger.Log(entry)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, entry.ID)

	logs, err := logger.GetLogsByOrg(orgID)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "link.created", logs[0].Action)
	assert.Equal(t, "link", logs[0].ResourceType)
	assert.Equal(t, "198.51.100.42", logs[0].IPAddress)
}

func TestFeatureFlagEngine_Evaluations(t *testing.T) {
	engine := audit.NewFeatureFlagEngine()

	// 1. Global Flag
	err := engine.RegisterFlag(&audit.FeatureFlag{
		Key:               "feature_global",
		IsEnabledGlobally: true,
	})
	require.NoError(t, err)

	org1 := uuid.New()
	user1 := uuid.New()
	assert.True(t, engine.IsEnabled("feature_global", org1, user1))

	// 2. Organization Targeted Flag
	targetOrg := uuid.New()
	otherOrg := uuid.New()
	err = engine.RegisterFlag(&audit.FeatureFlag{
		Key:                    "feature_targeted_org",
		IsEnabledGlobally:      false,
		EnabledOrganizationIDs: []uuid.UUID{targetOrg},
	})
	require.NoError(t, err)

	assert.True(t, engine.IsEnabled("feature_targeted_org", targetOrg, user1))
	assert.False(t, engine.IsEnabled("feature_targeted_org", otherOrg, user1))

	// 3. Percentage Rollout Flag (deterministic hash)
	err = engine.RegisterFlag(&audit.FeatureFlag{
		Key:               "feature_rollout_50",
		PercentageRollout: 50,
	})
	require.NoError(t, err)

	// User IDs should evaluate deterministically
	eval1 := engine.IsEnabled("feature_rollout_50", org1, user1)
	eval2 := engine.IsEnabled("feature_rollout_50", org1, user1)
	assert.Equal(t, eval1, eval2)
}
