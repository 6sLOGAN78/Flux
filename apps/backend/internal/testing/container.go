package testing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainerConfig struct {
	Image string
	Env   map[string]string
}

type PostgresContainer struct {
	Container testcontainers.Container
	Host      string
	Port      int
	DSN       string
}

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
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	portInt, _ := strconv.Atoi(port.Port())
	dsn := fmt.Sprintf("postgres://test_user:test_password@%s:%s/flux_test?sslmode=disable", host, port.Port())

	return &PostgresContainer{Container: container, Host: host, Port: portInt, DSN: dsn}, nil
}

func (p *PostgresContainer) Terminate(ctx context.Context) error {
	if p.Container != nil {
		return p.Container.Terminate(ctx)
	}
	return nil
}

type RedisContainer struct {
	Container testcontainers.Container
	Host      string
	Port      int
	Address   string
}

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
		return nil, err
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	portInt, _ := strconv.Atoi(port.Port())
	address := fmt.Sprintf("%s:%s", host, port.Port())

	return &RedisContainer{Container: container, Host: host, Port: portInt, Address: address}, nil
}

func (r *RedisContainer) Terminate(ctx context.Context) error {
	if r.Container != nil {
		return r.Container.Terminate(ctx)
	}
	return nil
}

type ClickHouseContainer struct {
	Container testcontainers.Container
	Host      string
	Port      int
	Address   string
}

func SetupClickHouseContainer(ctx context.Context) (*ClickHouseContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"CLICKHOUSE_USER": "default",
			"CLICKHOUSE_PASSWORD": "testpassword",
			"CLICKHOUSE_DB": "default",
		},
		WaitingFor:   wait.ForListeningPort("9000/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start clickhouse testcontainer: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		return nil, err
	}

	portInt, _ := strconv.Atoi(port.Port())
	// Use DSN string for ClickHouse
	address := fmt.Sprintf("clickhouse://default:testpassword@%s:%s/default", host, port.Port())

	return &ClickHouseContainer{Container: container, Host: host, Port: portInt, Address: address}, nil
}

func (c *ClickHouseContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}
