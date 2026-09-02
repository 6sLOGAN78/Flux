package config_test

import (
	"os"
	"testing"

	"flux/apps/backend/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/go-playground/validator/v10"
)

func TestConfig_Validation(t *testing.T) {
	validate := validator.New()
	
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
	}{
		{
			name: "valid config",
			cfg: config.Config{
				Primary: config.PrimaryConfig{Env: "local"},
				Server: config.ServerConfig{Port: "8080"},
				Database: config.DatabaseConfig{
					Host: "localhost", Port: "5432", User: "pg", Password: "pwd", Name: "db",
				},
				Redis: config.RedisConfig{Address: "localhost"},
				Clerk: config.ClerkConfig{SecretKey: "sk_test_123"},
			},
			expectError: false,
		},
		{
			name: "missing clerk secret",
			cfg: config.Config{
				Primary: config.PrimaryConfig{Env: "local"},
				Server: config.ServerConfig{Port: "8080"},
				Database: config.DatabaseConfig{
					Host: "localhost", Port: "5432", User: "pg", Password: "pwd", Name: "db",
				},
				Redis: config.RedisConfig{Address: "localhost"},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.Struct(tc.cfg)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_LoadEnvMapping(t *testing.T) {
	// Setup env
	os.Setenv("FLUX_PRIMARY.ENV", "production")
	os.Setenv("FLUX_SERVER.PORT", "9090")
	os.Setenv("FLUX_DATABASE.HOST", "db.example.com")
	os.Setenv("FLUX_DATABASE.PORT", "5432")
	os.Setenv("FLUX_DATABASE.USER", "dbuser")
	os.Setenv("FLUX_DATABASE.PASSWORD", "dbpass")
	os.Setenv("FLUX_DATABASE.NAME", "dbname")
	os.Setenv("FLUX_REDIS.ADDRESS", "redis.example.com:6379")
	os.Setenv("FLUX_AUTH.SECRET_KEY", "sk_test_clerk")
	os.Setenv("FLUX_STRIPE.SECRET_KEY", "sk_live_stripe")
	os.Setenv("FLUX_STRIPE.WEBHOOK_SECRET", "whsec_live_stripe")
	
	defer func() {
		os.Clearenv()
	}()

	cfg, err := config.LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	
	assert.Equal(t, "production", cfg.Primary.Env)
	assert.Equal(t, "9090", cfg.Server.Port)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, "redis.example.com:6379", cfg.Redis.Address)
	assert.Equal(t, "sk_test_clerk", cfg.Clerk.SecretKey)
}

func TestConfig_ProductionFailClosed(t *testing.T) {
	os.Setenv("FLUX_PRIMARY.ENV", "production")
	os.Setenv("FLUX_SERVER.PORT", "9090")
	os.Setenv("FLUX_DATABASE.HOST", "db.example.com")
	os.Setenv("FLUX_DATABASE.PORT", "5432")
	os.Setenv("FLUX_DATABASE.USER", "dbuser")
	os.Setenv("FLUX_DATABASE.PASSWORD", "dbpass")
	os.Setenv("FLUX_DATABASE.NAME", "dbname")
	os.Setenv("FLUX_REDIS.ADDRESS", "redis.example.com:6379")
	os.Setenv("FLUX_AUTH.SECRET_KEY", "sk_test_clerk")
	os.Setenv("FLUX_STRIPE.SECRET_KEY", "sk_test_secret") // unsafe default!
	
	defer os.Clearenv()

	_, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stripe")
}
