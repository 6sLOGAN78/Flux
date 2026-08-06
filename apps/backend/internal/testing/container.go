// Package testing provides integration testing helpers, testcontainers, assertions, and mock database fixtures.
package testing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerConfig holds container configuration parameters.
type ContainerConfig struct {
	Image string
	Env   map[string]string
}

// PostgresContainer manages a containerized PostgreSQL instance for integration testing.
type PostgresContainer struct {
	Container testcontainers.Container
	Host      string
	Port      int
	DSN       string
}

// SetupPostgresContainer starts a Postgres testcontainer instance.
func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "flux_test",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres testcontainer: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres mapped port: %w", err)
	}

	portInt, _ := strconv.Atoi(port.Port())
	dsn := fmt.Sprintf("postgres://test_user:test_password@%s:%s/flux_test?sslmode=disable", host, port.Port())

	return &PostgresContainer{
		Container: container,
		Host:      host,
		Port:      portInt,
		DSN:       dsn,
	}, nil
}

// Terminate gracefully shuts down the container.
func (p *PostgresContainer) Terminate(ctx context.Context) error {
	if p.Container != nil {
		return p.Container.Terminate(ctx)
	}
	return nil
}

// RedisContainer manages a containerized Redis instance for integration testing.
type RedisContainer struct {
	Container testcontainers.Container
	Host      string
	Port      int
	Address   string
}

// SetupRedisContainer starts a Redis testcontainer instance.
func SetupRedisContainer(ctx context.Context) (*RedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start redis testcontainer: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis mapped port: %w", err)
	}

	portInt, _ := strconv.Atoi(port.Port())
	address := fmt.Sprintf("%s:%s", host, port.Port())

	return &RedisContainer{
		Container: container,
		Host:      host,
		Port:      portInt,
		Address:   address,
	}, nil
}

// Terminate gracefully shuts down the container.
func (r *RedisContainer) Terminate(ctx context.Context) error {
	if r.Container != nil {
		return r.Container.Terminate(ctx)
	}
	return nil
}
