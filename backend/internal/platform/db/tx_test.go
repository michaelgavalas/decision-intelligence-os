package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// Compile-time assertions that the standard pgx types satisfy Querier.
var (
	_ Querier = (*pgxpool.Pool)(nil)
	_ Querier = (pgx.Tx)(nil)
)

// startPostgres boots a disposable postgres:18-alpine container and returns a
// connected pool plus a cleanup function. The test is skipped when Docker is
// unavailable.
func startPostgres(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Skipf("skipping: cannot start postgres container (docker unavailable?): %v", err)
	}

	connCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dsn, err := container.ConnectionString(connCtx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("connection string: %v", err)
	}

	pool, err := NewPool(connCtx, dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("NewPool: %v", err)
	}

	cleanup := func() {
		pool.Close()
		_ = container.Terminate(context.Background())
	}
	return pool, cleanup
}

// rowCount returns the number of rows in table t.
func rowCount(ctx context.Context, t *testing.T, q Querier) int {
	t.Helper()
	var n int
	if err := q.QueryRow(ctx, "SELECT count(*) FROM t").Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestWithinTx(t *testing.T) {
	pool, cleanup := startPostgres(t)
	defer cleanup()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, "CREATE TABLE t (id int)"); err != nil {
		t.Fatalf("create table: %v", err)
	}

	mgr := NewTxManager(pool)

	t.Run("commits on nil error", func(t *testing.T) {
		err := mgr.WithinTx(ctx, func(q Querier) error {
			_, err := q.Exec(ctx, "INSERT INTO t (id) VALUES (1)")
			return err
		})
		if err != nil {
			t.Fatalf("WithinTx: %v", err)
		}
		if got := rowCount(ctx, t, pool); got != 1 {
			t.Errorf("row count after commit = %d, want 1", got)
		}
	})

	t.Run("rolls back on returned error", func(t *testing.T) {
		before := rowCount(ctx, t, pool)
		sentinel := errors.New("boom")
		err := mgr.WithinTx(ctx, func(q Querier) error {
			if _, err := q.Exec(ctx, "INSERT INTO t (id) VALUES (2)"); err != nil {
				return err
			}
			return sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Fatalf("WithinTx err = %v, want sentinel", err)
		}
		if got := rowCount(ctx, t, pool); got != before {
			t.Errorf("row count after rollback = %d, want %d", got, before)
		}
	})

	t.Run("rolls back and re-panics", func(t *testing.T) {
		before := rowCount(ctx, t, pool)
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("expected panic to propagate")
				}
			}()
			_ = mgr.WithinTx(ctx, func(q Querier) error {
				if _, err := q.Exec(ctx, "INSERT INTO t (id) VALUES (3)"); err != nil {
					return err
				}
				panic("kaboom")
			})
		}()
		if got := rowCount(ctx, t, pool); got != before {
			t.Errorf("row count after panic rollback = %d, want %d", got, before)
		}
	})
}
