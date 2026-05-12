package testhelpers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
	container testcontainers.Container
	pool      *pgxpool.Pool
	DSN       string
}

func StartPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "123456",
			"POSTGRES_DB":       "userdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	dsn := fmt.Sprintf("postgres://postgres:123456@%s:%s/userdb?sslmode=disable",
		host, port.Port())

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresContainer{
		container: container,
		pool:      pool,
		DSN:       dsn,
	}, nil
}

func (pc *PostgresContainer) GetPool() *pgxpool.Pool {
	return pc.pool
}
func (pc *PostgresContainer) RunMigrations(ctx context.Context, migrations ...string) error {
	for _, migration := range migrations {
		if _, err := pc.pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}
	return nil
}

func (pc *PostgresContainer) Terminate(ctx context.Context) error {
	if pc.pool != nil {
		pc.pool.Close()
	}
	if pc.container != nil {
		return pc.container.Terminate(ctx)
	}
	return nil
}
