// Package dbtest provides a shared integration-test helper that boots an
// ephemeral PostgreSQL container, applies the project migrations, and hands
// back a ready connection pool. It exists so every package's integration tests
// exercise the same real schema instead of mocks.
//
// The package deliberately imports the standard testing package and accepts
// *testing.T so callers get the usual t.Skip / t.Cleanup ergonomics; this is
// the conventional Go test-helper shape and the file is shared (non-_test) so
// other packages can import it.
package dbtest

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// pgx5 database driver and file source for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
)

// migrationsDir resolves the absolute path to backend/migrations relative to
// this source file, so the helper works regardless of the test's working
// directory.
func migrationsDir(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("dbtest: cannot determine caller path")
	}
	// thisFile lives at internal/platform/dbtest/dbtest.go; migrations are at
	// backend/migrations, i.e. three directories up.
	abs, err := filepath.Abs(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations"))
	if err != nil {
		t.Fatalf("dbtest: resolve migrations dir: %v", err)
	}
	return abs
}

// NewPool starts an ephemeral postgres:18-alpine container, applies all
// migrations from backend/migrations, and returns a ready pool. It registers
// cleanup (terminate container, close pool) via t.Cleanup and skips the test
// when Docker is unavailable.
func NewPool(t *testing.T) *pgxpool.Pool { //nolint:thelper // shared integration test helper; callers expect direct t handling
	t.Helper()

	pool, _ := NewPoolWithURL(t)
	return pool
}

// NewPoolWithURL behaves like NewPool but additionally returns the container's
// connection string. It is useful for tests that construct their own pool (for
// example, exercising a composition root that opens its own pool against the
// same database).
func NewPoolWithURL(t *testing.T) (*pgxpool.Pool, string) { //nolint:thelper // shared integration test helper; callers expect direct t handling
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
		t.Skipf("docker unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	connCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dsn, err := container.ConnectionString(connCtx, "sslmode=disable")
	if err != nil {
		t.Fatalf("dbtest: connection string: %v", err)
	}

	applyMigrations(t, dsn)

	pool, err := db.NewPool(connCtx, dsn)
	if err != nil {
		t.Fatalf("dbtest: open pool: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool, dsn
}

// applyMigrations runs every up migration against the given database using the
// golang-migrate library. ErrNoChange is treated as success.
func applyMigrations(t *testing.T, dsn string) {
	t.Helper()

	source := "file://" + filepath.ToSlash(migrationsDir(t))

	// testcontainers returns a postgres:// URL; the golang-migrate pgx/v5
	// driver registers under the pgx5 scheme, so swap the scheme while keeping
	// the rest of the URL (credentials, host, port, query) intact.
	migrateURL := "pgx5://" + strings.TrimPrefix(dsn, "postgres://")

	m, err := migrate.New(source, migrateURL)
	if err != nil {
		t.Fatalf("dbtest: init migrate: %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			t.Errorf("dbtest: close migrate source: %v", srcErr)
		}
		if dbErr != nil {
			t.Errorf("dbtest: close migrate db: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("dbtest: apply migrations: %v", err)
	}
}

// TruncateAll truncates every data table (keeping the schema) so subtests can
// share a single container without leaking state between them. The
// schema_migrations bookkeeping table is intentionally left untouched.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) { //nolint:thelper // shared integration test helper; callers expect direct t handling
	t.Helper()

	ctx := context.Background()

	rows, err := pool.Query(ctx,
		`SELECT tablename FROM pg_tables
		 WHERE schemaname = 'public' AND tablename <> 'schema_migrations'`,
	)
	if err != nil {
		t.Fatalf("dbtest: list tables: %v", err)
	}

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			t.Fatalf("dbtest: scan table name: %v", err)
		}
		tables = append(tables, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatalf("dbtest: iterate tables: %v", err)
	}

	if len(tables) == 0 {
		return
	}

	stmt := "TRUNCATE TABLE "
	for i, name := range tables {
		if i > 0 {
			stmt += ", "
		}
		stmt += pgx.Identifier{name}.Sanitize()
	}
	stmt += " RESTART IDENTITY CASCADE"

	if _, err := pool.Exec(ctx, stmt); err != nil {
		t.Fatalf("dbtest: truncate: %v", err)
	}
}
