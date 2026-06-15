package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/ratelimit"
)

func TestAllowEnforcesLimitAndResetsWindow(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	ctx := context.Background()
	limiter := ratelimit.NewLimiter(pool)

	const (
		key    = "user:42"
		limit  = 2
		window = 300 * time.Millisecond
	)

	first, err := limiter.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow #1: %v", err)
	}
	if !first {
		t.Error("Allow #1 = false, want true")
	}

	second, err := limiter.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow #2: %v", err)
	}
	if !second {
		t.Error("Allow #2 = false, want true")
	}

	third, err := limiter.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow #3: %v", err)
	}
	if third {
		t.Error("Allow #3 = true, want false (limit exceeded)")
	}

	// Wait for the window to elapse, then the counter should reset.
	time.Sleep(350 * time.Millisecond)

	after, err := limiter.Allow(ctx, key, limit, window)
	if err != nil {
		t.Fatalf("Allow after window: %v", err)
	}
	if !after {
		t.Error("Allow after window reset = false, want true")
	}
}

func TestAllowKeysAreIndependent(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	ctx := context.Background()
	limiter := ratelimit.NewLimiter(pool)

	const (
		limit  = 1
		window = time.Second
	)

	a, err := limiter.Allow(ctx, "key-a", limit, window)
	if err != nil {
		t.Fatalf("Allow key-a: %v", err)
	}
	b, err := limiter.Allow(ctx, "key-b", limit, window)
	if err != nil {
		t.Fatalf("Allow key-b: %v", err)
	}
	if !a || !b {
		t.Errorf("independent keys both within limit want true,true got %v,%v", a, b)
	}
}
