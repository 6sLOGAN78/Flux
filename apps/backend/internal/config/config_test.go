package config_test

import (
	"testing"

	"flux/apps/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			cfg: config.Config{
				ServerPort:     "8080",
				DatabaseURL:    "postgres://localhost",
				RedisURL:       "localhost:6379",
				ClerkSecretKey: "sk_test_123",
			},
			expectError: false,
		},
		{
			name: "missing clerk secret",
			cfg: config.Config{
				ServerPort:  "8080",
				DatabaseURL: "postgres://localhost",
				RedisURL:    "localhost:6379",
			},
			expectError: true,
			errorMsg:    "clerk secret key cannot be empty",
		},
		{
			name: "missing db url",
			cfg: config.Config{
				ServerPort:     "8080",
				RedisURL:       "localhost:6379",
				ClerkSecretKey: "sk_test_123",
			},
			expectError: true,
			errorMsg:    "database URL cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
