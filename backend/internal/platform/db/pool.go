// Package db provides PostgreSQL connection pooling and transaction management
// built on pgx. Repositories depend on the Querier interface so they can run
// against either a pool or an in-flight transaction.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Pool defaults applied when constructing a new connection pool.
const (
	maxConns          = 10
	minConns          = 2
	maxConnLifetime   = time.Hour
	maxConnIdleTime   = 30 * time.Minute
	healthCheckPeriod = time.Minute
)

// CommandTag is re-exported so callers of Querier need not import pgconn
// directly.
type CommandTag = pgconn.CommandTag

// NewPool opens a pgxpool with sane defaults and verifies connectivity with a
// ping before returning.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "DB_CONFIG_INVALID", "invalid database url")
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.HealthCheckPeriod = healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "DB_CONNECT_FAILED", "failed to create connection pool")
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.Wrap(err, errors.KindInternal, "DB_PING_FAILED", "failed to reach database")
	}

	return pool, nil
}
