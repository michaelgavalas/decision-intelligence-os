package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/auth"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// seedUser inserts a user so refresh_tokens foreign keys are satisfied.
func seedUser(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	u, err := users.NewRepository().Create(context.Background(), pool, users.CreateParams{
		ID:           id.New(),
		Email:        email,
		Name:         "Seed",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u.ID
}

func TestRefreshTokenRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := auth.NewRepository()
	ctx := context.Background()

	t.Run("store and get by hash", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		user := seedUser(t, pool, "owner@example.com")

		stored, err := repo.Store(ctx, pool, auth.RefreshToken{
			ID:        id.New(),
			UserID:    user,
			TokenHash: "hash-1",
			ExpiresAt: time.Now().Add(time.Hour).UTC(),
		})
		if err != nil {
			t.Fatalf("Store: %v", err)
		}

		got, err := repo.GetByHash(ctx, pool, "hash-1")
		if err != nil {
			t.Fatalf("GetByHash: %v", err)
		}
		if got.ID != stored.ID || got.UserID != user {
			t.Errorf("GetByHash = %+v, want id %v user %v", got, stored.ID, user)
		}
		if got.Revoked() {
			t.Error("freshly stored token reports revoked")
		}
	})

	t.Run("missing hash is not found", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		_, err := repo.GetByHash(ctx, pool, "absent")
		if errors.KindOf(err) != errors.KindNotFound || errors.CodeOf(err) != "REFRESH_NOT_FOUND" {
			t.Errorf("GetByHash err = %v, want NotFound/REFRESH_NOT_FOUND", err)
		}
	})

	t.Run("revoke", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		user := seedUser(t, pool, "owner@example.com")
		stored, err := repo.Store(ctx, pool, auth.RefreshToken{
			ID: id.New(), UserID: user, TokenHash: "hash-r", ExpiresAt: time.Now().Add(time.Hour).UTC(),
		})
		if err != nil {
			t.Fatalf("Store: %v", err)
		}

		if err := repo.Revoke(ctx, pool, stored.ID); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		got, err := repo.GetByHash(ctx, pool, "hash-r")
		if err != nil {
			t.Fatalf("GetByHash: %v", err)
		}
		if !got.Revoked() {
			t.Error("token not revoked after Revoke")
		}
	})

	t.Run("revoke all for user", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		user := seedUser(t, pool, "owner@example.com")
		for _, h := range []string{"a", "b"} {
			if _, err := repo.Store(ctx, pool, auth.RefreshToken{
				ID: id.New(), UserID: user, TokenHash: h, ExpiresAt: time.Now().Add(time.Hour).UTC(),
			}); err != nil {
				t.Fatalf("Store %s: %v", h, err)
			}
		}

		if err := repo.RevokeAllForUser(ctx, pool, user); err != nil {
			t.Fatalf("RevokeAllForUser: %v", err)
		}
		for _, h := range []string{"a", "b"} {
			got, err := repo.GetByHash(ctx, pool, h)
			if err != nil {
				t.Fatalf("GetByHash %s: %v", h, err)
			}
			if !got.Revoked() {
				t.Errorf("token %s not revoked", h)
			}
		}
	})

	t.Run("mark replaced", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		user := seedUser(t, pool, "owner@example.com")
		old, err := repo.Store(ctx, pool, auth.RefreshToken{
			ID: id.New(), UserID: user, TokenHash: "old", ExpiresAt: time.Now().Add(time.Hour).UTC(),
		})
		if err != nil {
			t.Fatalf("Store old: %v", err)
		}
		next, err := repo.Store(ctx, pool, auth.RefreshToken{
			ID: id.New(), UserID: user, TokenHash: "new", ExpiresAt: time.Now().Add(time.Hour).UTC(),
		})
		if err != nil {
			t.Fatalf("Store new: %v", err)
		}

		if err := repo.MarkReplaced(ctx, pool, old.ID, next.ID); err != nil {
			t.Fatalf("MarkReplaced: %v", err)
		}
		got, err := repo.GetByHash(ctx, pool, "old")
		if err != nil {
			t.Fatalf("GetByHash: %v", err)
		}
		if !got.Revoked() {
			t.Error("replaced token not revoked")
		}
		if got.ReplacedBy == nil || *got.ReplacedBy != next.ID {
			t.Errorf("ReplacedBy = %v, want %v", got.ReplacedBy, next.ID)
		}
	})
}
