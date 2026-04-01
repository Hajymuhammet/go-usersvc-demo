package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates and verifies a PostgreSQL connection pool.
func NewPool(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("postgres: failed to ping: %w", err)
	}

	return pool, nil
}
