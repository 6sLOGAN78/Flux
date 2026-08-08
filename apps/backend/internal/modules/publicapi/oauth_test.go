package publicapi_test

import (
	"testing"

	"flux/apps/backend/internal/modules/publicapi"

	"github.com/google/uuid"
)

func TestGenerateAndValidateAPIKey(t *testing.T) {
	wsID := uuid.New()
	userID := uuid.New()

	rawKey, keyEntity, err := publicapi.GenerateAPIKey(wsID, userID, "CI Pipeline Key", []string{"links:read", "links:write"}, 100)
	if err != nil {
		t.Fatalf("unexpected error generating API key: %v", err)
	}

	if !publicapi.HasPrefix(rawKey, "flx_live_") {
		t.Errorf("expected raw API key to start with 'flx_live_', got %q", rawKey)
	}

	if keyEntity.KeyHash == "" {
		t.Error("expected KeyHash to be populated")
	}

	// Validate raw key against stored hash
	if !publicapi.ValidateKey(rawKey, keyEntity.KeyHash) {
		t.Error("expected valid raw key to match stored hash")
	}

	// Invalid key must fail
	if publicapi.ValidateKey("flx_live_invalidkey123", keyEntity.KeyHash) {
		t.Error("expected invalid raw key to fail validation")
	}
}

func TestHasScope(t *testing.T) {
	scopes := []string{"links:read", "links:write", "analytics:read"}

	if !publicapi.HasScope(scopes, "links:read") {
		t.Error("expected has 'links:read' scope to return true")
	}

	if publicapi.HasScope(scopes, "admin:delete") {
		t.Error("expected missing 'admin:delete' scope to return false")
	}
}

func TestIssueOAuthToken(t *testing.T) {
	token, err := publicapi.IssueOAuthToken("client_app_123", []string{"links:read"}, 3600)
	if err != nil {
		t.Fatalf("unexpected error issuing OAuth token: %v", err)
	}

	if token.AccessToken == "" {
		t.Error("expected AccessToken to be non-empty")
	}
	if token.TokenType != "Bearer" {
		t.Errorf("expected TokenType 'Bearer', got %q", token.TokenType)
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("expected ExpiresIn 3600, got %d", token.ExpiresIn)
	}
}
