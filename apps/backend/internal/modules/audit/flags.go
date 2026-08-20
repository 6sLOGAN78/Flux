package audit

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type FeatureFlag struct {
	ID                     uuid.UUID   `json:"id"`
	Key                    string      `json:"key"`
	Description            string      `json:"description,omitempty"`
	IsEnabledGlobally      bool        `json:"is_enabled_globally"`
	EnabledOrganizationIDs []uuid.UUID `json:"enabled_organization_ids,omitempty"`
	PercentageRollout      int         `json:"percentage_rollout"` // 0 to 100
	CreatedAt              time.Time   `json:"created_at"`
}

type FeatureFlagEngine struct {
	mu    sync.RWMutex
	flags map[string]*FeatureFlag
}

func NewFeatureFlagEngine() *FeatureFlagEngine {
	return &FeatureFlagEngine{
		flags: make(map[string]*FeatureFlag),
	}
}

// RegisterFlag adds or updates a feature flag configuration.
func (e *FeatureFlagEngine) RegisterFlag(flag *FeatureFlag) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if flag.ID == uuid.Nil {
		flag.ID = uuid.New()
	}
	if flag.CreatedAt.IsZero() {
		flag.CreatedAt = time.Now().UTC()
	}

	e.flags[flag.Key] = flag
	return nil
}

// IsEnabled evaluates whether a flag is enabled for the provided org and user context.
func (e *FeatureFlagEngine) IsEnabled(key string, orgID uuid.UUID, userID uuid.UUID) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	flag, exists := e.flags[key]
	if !exists {
		return false
	}

	// 1. Global toggle check
	if flag.IsEnabledGlobally {
		return true
	}

	// 2. Organization whitelist check
	if orgID != uuid.Nil && len(flag.EnabledOrganizationIDs) > 0 {
		for _, enabledOrgID := range flag.EnabledOrganizationIDs {
			if enabledOrgID == orgID {
				return true
			}
		}
	}

	// 3. Deterministic percentage rollout check
	if flag.PercentageRollout > 0 {
		if flag.PercentageRollout >= 100 {
			return true
		}

		targetID := userID.String()
		if targetID == uuid.Nil.String() && orgID != uuid.Nil {
			targetID = orgID.String()
		}

		if targetID != uuid.Nil.String() {
			hasher := fnv.New32a()
			_, _ = hasher.Write([]byte(key + ":" + targetID))
			hashVal := hasher.Sum32()
			bucket := int(hashVal % 100)
			if bucket < flag.PercentageRollout {
				return true
			}
		}
	}

	return false
}
