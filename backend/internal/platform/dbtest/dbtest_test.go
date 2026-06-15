package dbtest_test

import (
	"context"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
)

func TestNewPoolPings(t *testing.T) {
	pool := dbtest.NewPool(t)

	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestNewPoolAppliesMigrations(t *testing.T) {
	pool := dbtest.NewPool(t)

	var exists bool
	err := pool.QueryRow(context.Background(),
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'users'
		)`,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("query information_schema: %v", err)
	}
	if !exists {
		t.Fatal("users table does not exist after migrations")
	}
}

func TestTruncateAll(t *testing.T) {
	pool := dbtest.NewPool(t)

	// Should run without error even on an empty schema.
	dbtest.TruncateAll(t, pool)
}
