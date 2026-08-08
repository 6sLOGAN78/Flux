// Package publicapi provides API Key generation, SHA-256 key hashing, OAuth 2.0 token issuance, and scope authorization.
package publicapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// APIKey represents an API Key credential entity.
type APIKey struct {
	ID              uuid.UUID  `json:"id,omitempty" db:"id"`
	WorkspaceID     uuid.UUID  `json:"workspace_id" db:"workspace_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	Name            string     `json:"name" db:"name"`
	KeyPrefix       string     `json:"key_prefix" db:"key_prefix"`
	KeyHash         string     `json:"key_hash" db:"key_hash"`
	Scopes          []string   `json:"scopes" db:"scopes"`
	RateLimitPerMin int        `json:"rate_limit_per_min" db:"rate_limit_per_min"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt       time.Time  `json:"created_at,omitempty" db:"created_at"`
}

// OAuthTokenResponse represents an OAuth 2.0 Bearer Token response payload.
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// HashKey computes SHA-256 hash of a raw API key.
func HashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}

// ValidateKey compares a raw API key against a stored SHA-256 hash.
func ValidateKey(rawKey, storedHash string) bool {
	return HashKey(rawKey) == storedHash
}

// GenerateAPIKey generates a new raw API key (e.g. "flx_live_...") and its hashed APIKey entity.
func GenerateAPIKey(workspaceID, userID uuid.UUID, name string, scopes []string, rateLimit int) (string, APIKey, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", APIKey{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	secret := hex.EncodeToString(bytes)
	prefix := "flx_live_"
	rawKey := prefix + secret
	keyHash := HashKey(rawKey)

	if len(scopes) == 0 {
		scopes = []string{"links:read", "links:write"}
	}
	if rateLimit <= 0 {
		rateLimit = 100
	}

	entity := APIKey{
		ID:              uuid.New(),
		WorkspaceID:     workspaceID,
		UserID:          userID,
		Name:            name,
		KeyPrefix:       prefix,
		KeyHash:         keyHash,
		Scopes:          scopes,
		RateLimitPerMin: rateLimit,
		CreatedAt:       time.Now(),
	}

	return rawKey, entity, nil
}

// IssueOAuthToken creates a new OAuth 2.0 access token response for a client.
func IssueOAuthToken(clientID string, scopes []string, expiresInSeconds int) (OAuthTokenResponse, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return OAuthTokenResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	tokenSecret := hex.EncodeToString(bytes)
	accessToken := fmt.Sprintf("flx_oauth_%s", tokenSecret)

	if expiresInSeconds <= 0 {
		expiresInSeconds = 3600
	}

	return OAuthTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresInSeconds,
		Scope:       strings.Join(scopes, " "),
	}, nil
}

// HasScope verifies if a given scope exists in the allowed scopes list.
func HasScope(scopes []string, targetScope string) bool {
	targetScope = strings.ToLower(strings.TrimSpace(targetScope))
	for _, s := range scopes {
		if strings.ToLower(strings.TrimSpace(s)) == targetScope {
			return true
		}
	}
	return false
}

// HasPrefix checks if string s starts with prefix.
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}
