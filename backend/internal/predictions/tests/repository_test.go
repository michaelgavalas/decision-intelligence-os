package predictions_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// seedDecision inserts a user, team, membership, and decision so that
// prediction foreign keys are satisfied, returning the decision id.
func seedDecision(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	u, err := users.NewRepository().Create(ctx, pool, users.CreateParams{
		ID: id.New(), Email: email, Name: "Owner", PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	team, err := teams.NewRepository().CreateTeam(ctx, pool, id.New(), "Acme")
	if err != nil {
		t.Fatalf("seed team: %v", err)
	}
	if _, err := teams.NewRepository().AddMember(ctx, pool, team.ID, u.ID, teams.RoleAdmin); err != nil {
		t.Fatalf("seed membership: %v", err)
	}
	d, err := decisions.NewRepository().Create(ctx, pool, decisions.Decision{
		ID: id.New(), TeamID: team.ID, OwnerID: u.ID, Title: "T", Status: decisions.StatusActive,
	})
	if err != nil {
		t.Fatalf("seed decision: %v", err)
	}
	return d.ID
}

func TestPredictionRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := predictions.NewRepository()
	ctx := context.Background()

	t.Run("round-trip create, get, update, list", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		decisionID := seedDecision(t, pool, "owner@example.com")

		resolvesAt := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		created, err := repo.Create(ctx, pool, predictions.Prediction{
			ID: id.New(), DecisionID: decisionID, Statement: "Ships on time", Probability: 0.725, ResolvesAt: &resolvesAt,
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if created.Probability != 0.725 {
			t.Errorf("probability = %v, want 0.725", created.Probability)
		}
		if created.ResolvesAt == nil || !created.ResolvesAt.Equal(resolvesAt) {
			t.Errorf("ResolvesAt = %v, want %v", created.ResolvesAt, resolvesAt)
		}

		got, err := repo.GetByID(ctx, pool, created.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Statement != "Ships on time" {
			t.Errorf("statement = %q, want Ships on time", got.Statement)
		}

		updated, err := repo.Update(ctx, pool, created.ID, "Ships late", 0.2, nil)
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Statement != "Ships late" || updated.Probability != 0.2 {
			t.Errorf("Update = %+v, want late/0.2", updated)
		}
		if updated.ResolvesAt != nil {
			t.Errorf("ResolvesAt = %v, want nil after clearing", updated.ResolvesAt)
		}

		list, err := repo.ListByDecision(ctx, pool, decisionID)
		if err != nil {
			t.Fatalf("ListByDecision: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("ListByDecision len = %d, want 1", len(list))
		}

		if _, err := repo.GetByID(ctx, pool, uuid.New()); errors.KindOf(err) != errors.KindNotFound {
			t.Errorf("GetByID unknown err = %v, want NotFound", err)
		}
	})

	t.Run("list by multiple decision ids", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		d1 := seedDecision(t, pool, "a@example.com")
		d2 := seedDecision(t, pool, "b@example.com")

		for _, d := range []uuid.UUID{d1, d2} {
			if _, err := repo.Create(ctx, pool, predictions.Prediction{
				ID: id.New(), DecisionID: d, Statement: "s", Probability: 0.5,
			}); err != nil {
				t.Fatalf("Create for %v: %v", d, err)
			}
		}

		rows, err := repo.ListByDecisionIDs(ctx, pool, []uuid.UUID{d1, d2})
		if err != nil {
			t.Fatalf("ListByDecisionIDs: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("ListByDecisionIDs len = %d, want 2", len(rows))
		}
		seen := map[uuid.UUID]bool{}
		for _, r := range rows {
			seen[r.DecisionID] = true
		}
		if !seen[d1] || !seen[d2] {
			t.Errorf("ListByDecisionIDs covered = %v, want both decisions", seen)
		}
	})
}

func TestPredictionRepositoryListByIDs(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := predictions.NewRepository()
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	decisionID := seedDecision(t, pool, "list@example.com")
	p1, err := repo.Create(ctx, pool, predictions.Prediction{
		ID: id.New(), DecisionID: decisionID, Statement: "one", Probability: 0.4,
	})
	if err != nil {
		t.Fatalf("Create p1: %v", err)
	}
	p2, err := repo.Create(ctx, pool, predictions.Prediction{
		ID: id.New(), DecisionID: decisionID, Statement: "two", Probability: 0.6,
	})
	if err != nil {
		t.Fatalf("Create p2: %v", err)
	}

	rows, err := repo.ListByIDs(ctx, pool, []uuid.UUID{p1.ID, p2.ID})
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
	if !seen[p1.ID] || !seen[p2.ID] {
		t.Errorf("ListByIDs covered = %v, want both predictions", seen)
	}
}
