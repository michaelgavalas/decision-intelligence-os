package users_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

func seedUser(t *testing.T, pool *pgxpool.Pool, repo users.Repository, email, name string) users.User {
	t.Helper()
	u, err := repo.Create(context.Background(), pool, users.CreateParams{
		ID:           id.New(),
		Email:        email,
		Name:         name,
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

func TestRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := users.NewRepository()
	ctx := context.Background()

	t.Run("create then get round-trip", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		created := seedUser(t, pool, repo, "ada@example.com", "Ada")

		got, err := repo.GetByID(ctx, pool, created.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Email != "ada@example.com" || got.Name != "Ada" {
			t.Errorf("GetByID = %+v, want email/name to match", got)
		}
		if got.CreatedAt.IsZero() {
			t.Error("GetByID: expected CreatedAt to be set")
		}
	})

	t.Run("get by email", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		created := seedUser(t, pool, repo, "grace@example.com", "Grace")

		got, err := repo.GetByEmail(ctx, pool, "grace@example.com")
		if err != nil {
			t.Fatalf("GetByEmail: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("GetByEmail id = %v, want %v", got.ID, created.ID)
		}
	})

	t.Run("duplicate email is a conflict", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		seedUser(t, pool, repo, "dup@example.com", "First")

		_, err := repo.Create(ctx, pool, users.CreateParams{
			ID:           id.New(),
			Email:        "dup@example.com",
			Name:         "Second",
			PasswordHash: "hash",
		})
		if errors.KindOf(err) != errors.KindConflict || errors.CodeOf(err) != "EMAIL_TAKEN" {
			t.Fatalf("duplicate email err = %v, want Conflict/EMAIL_TAKEN", err)
		}
	})

	t.Run("missing user is not found", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)

		_, err := repo.GetByID(ctx, pool, uuid.New())
		if errors.KindOf(err) != errors.KindNotFound || errors.CodeOf(err) != "USER_NOT_FOUND" {
			t.Fatalf("missing user err = %v, want NotFound/USER_NOT_FOUND", err)
		}
	})

	t.Run("update name", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		created := seedUser(t, pool, repo, "rename@example.com", "Old")

		updated, err := repo.UpdateName(ctx, pool, created.ID, "New")
		if err != nil {
			t.Fatalf("UpdateName: %v", err)
		}
		if updated.Name != "New" {
			t.Errorf("UpdateName name = %q, want New", updated.Name)
		}
	})
}

func TestRepositoryListByIDs(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := users.NewRepository()
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	a := seedUser(t, pool, repo, "list-a@example.com", "A")
	b := seedUser(t, pool, repo, "list-b@example.com", "B")

	rows, err := repo.ListByIDs(ctx, pool, []uuid.UUID{a.ID, b.ID})
	if err != nil {
		t.Fatalf("ListByIDs: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("ListByIDs len = %d, want 2", len(rows))
	}
	seen := map[uuid.UUID]bool{}
	for _, r := range rows {
		seen[r.ID] = true
	}
	if !seen[a.ID] || !seen[b.ID] {
		t.Errorf("ListByIDs covered = %v, want both users", seen)
	}
}
