package testing

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/config"
	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Keep these immutable references identical to compose.yaml.
const (
	PostgresImage = "postgres:17.11-alpine@sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24"
	RedisImage    = "redis:8.10.2-alpine@sha256:3811787313eba226a2ef38658c6ccb91cd5e110edc89c37767de373120a0e5a0"
)

// loopbackPort prevents disposable test dependencies from being exposed on LAN interfaces.
func loopbackPort(port string) func(*container.HostConfig) {
	return func(cfg *container.HostConfig) {
		cfg.PortBindings = network.PortMap{
			network.MustParsePort(port): {{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "0"}},
		}
	}
}

type TestDB struct {
	Pool      *pgxpool.Pool
	Container testcontainers.Container
	Config    *config.Config
}

// SetupTestDB creates a Postgres container and applies application migrations.
func SetupTestDB(t *testing.T) (*TestDB, func()) {
	t.Helper()
	db, cleanup := SetupTestPostgres(t)
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	err := database.Migrate(context.Background(), &logger, db.Config)
	require.NoError(t, err, "failed to apply database migrations")
	return db, cleanup
}

// SetupTestPostgres creates an isolated real Postgres database without application migrations.
// Use SetupTestDB for tests that require the application schema.
func SetupTestPostgres(t *testing.T) (*TestDB, func()) {
	t.Helper()

	ctx := context.Background()
	dbName := fmt.Sprintf("test_db_%s", uuid.New().String()[:8])
	dbUser := "testuser"
	dbPassword := "testpassword"

	req := testcontainers.ContainerRequest{
		Image:        PostgresImage,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
			wait.ForListeningPort("5432/tcp"),
		).WithStartupTimeout(60 * time.Second),
		HostConfigModifier: loopbackPort("5432/tcp"),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, pgContainer)
	require.NoError(t, err, "failed to start postgres container")

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	mappedPort, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err, "failed to get mapped port")
	port, err := strconv.Atoi(mappedPort.Port())
	require.NoError(t, err)

	// Create configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:            host,
			Port:            port,
			User:            dbUser,
			Password:        dbPassword,
			Name:            dbName,
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 300,
			ConnMaxIdleTime: 300,
		},
		Primary: config.Primary{
			Env: "test",
		},
		Server: config.ServerConfig{
			Port:               "8080",
			ReadTimeout:        30,
			WriteTimeout:       30,
			IdleTimeout:        30,
			CORSAllowedOrigins: []string{"*"},
		},
		Integration: config.IntegrationConfig{
			ResendAPIKey: "test-key",
		},
		Redis: config.RedisConfig{
			Address: "localhost:6379",
		},
		Auth: config.AuthConfig{
			SecretKey: "test-secret",
		},
	}

	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	db, err := database.New(cfg, &logger, nil)
	require.NoError(t, err, "failed to connect to ready postgres container")
	// Registered immediately so an assertion failure cannot leak the pool.
	t.Cleanup(db.Pool.Close)

	testDB := &TestDB{
		Pool:      db.Pool,
		Container: pgContainer,
		Config:    cfg,
	}

	// Return cleanup function that just closes the pool (container is managed by t.Cleanup)
	cleanup := func() {
		if db.Pool != nil {
			db.Pool.Close()
		}
	}

	return testDB, cleanup
}

// TestRedis owns an isolated Redis client and its disposable test container.
type TestRedis struct {
	Client    *redis.Client
	Container testcontainers.Container
	Config    config.RedisConfig
}

// SetupTestRedis creates real Redis infrastructure using the local service pin.
func SetupTestRedis(t *testing.T) (*TestRedis, func()) {
	t.Helper()
	ctx := context.Background()
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:              RedisImage,
			ExposedPorts:       []string{"6379/tcp"},
			WaitingFor:         wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
			HostConfigModifier: loopbackPort("6379/tcp"),
		},
		Started: true,
	})
	testcontainers.CleanupContainer(t, redisContainer)
	require.NoError(t, err, "failed to start redis container")

	host, err := redisContainer.Host(ctx)
	require.NoError(t, err, "failed to get redis container host")
	port, err := redisContainer.MappedPort(ctx, "6379/tcp")
	require.NoError(t, err, "failed to get redis mapped port")
	address := net.JoinHostPort(host, port.Port())
	client := redis.NewClient(&redis.Options{Addr: address})
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			if err := client.Close(); err != nil {
				t.Errorf("failed to close test redis client: %v", err)
			}
		})
	}
	t.Cleanup(cleanup)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(pingCtx).Err(), "failed to ping test redis")
	return &TestRedis{
		Client: client, Container: redisContainer,
		Config: config.RedisConfig{Address: address},
	}, cleanup
}

// CleanupTestDB closes the database connection and terminates the container
func (db *TestDB) CleanupTestDB(ctx context.Context, logger *zerolog.Logger) error {
	logger.Info().Msg("cleaning up test database")

	if db.Pool != nil {
		db.Pool.Close()
	}

	if db.Container != nil {
		if err := db.Container.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}

	return nil
}
