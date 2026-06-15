// Package ratelimit provides a fixed-window rate limiter backed by the
// rate_limits table. Keeping the counter in PostgreSQL makes limits consistent
// across every backend instance without a separate cache or coordinator.
package ratelimit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// allowQuery atomically upserts the counter for a key. When the stored window
// has expired it resets the count to 1 and starts a fresh window; otherwise it
// increments. RETURNING yields the post-increment count so the caller can
// decide whether the request is within the limit.
const allowQuery = `
INSERT INTO rate_limits (key, count, window_start) VALUES ($1, 1, now())
ON CONFLICT (key) DO UPDATE SET
  count = CASE WHEN rate_limits.window_start < now() - ($2::bigint || ' milliseconds')::interval THEN 1 ELSE rate_limits.count + 1 END,
  window_start = CASE WHEN rate_limits.window_start < now() - ($2::bigint || ' milliseconds')::interval THEN now() ELSE rate_limits.window_start END
RETURNING count`

// Limiter enforces a fixed-window token bucket per key.
type Limiter struct {
	pool *pgxpool.Pool
}

// NewLimiter returns a Limiter backed by the given pool.
func NewLimiter(pool *pgxpool.Pool) *Limiter {
	return &Limiter{pool: pool}
}

// Allow records one request against key and reports whether it is permitted. It
// increments the counter for the current window, resetting to 1 when the prior
// window has elapsed, and returns true when the post-increment count is within
// limit.
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	windowMillis := window.Milliseconds()

	var count int
	if err := l.pool.QueryRow(ctx, allowQuery, key, windowMillis).Scan(&count); err != nil {
		return false, errors.Wrap(err, errors.KindInternal, "RATE_LIMIT_QUERY_FAILED", "failed to evaluate rate limit")
	}

	return count <= limit, nil
}
