package loaders_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/graph/loaders"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/evidence"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// newLoaders builds a loader set wired to fresh repositories over pool, plus a
// context carrying it, mirroring what the HTTP middleware does per request.
func newLoaders(pool *pgxpool.Pool) (*loaders.Loaders, context.Context) {
	l := loaders.NewLoaders(
		pool,
		users.NewRepository(),
		teams.NewRepository(),
		decisions.NewRepository(),
		assumptions.NewRepository(),
		evidence.NewRepository(),
		predictions.NewRepository(),
		outcomes.NewRepository(),
	)
	return l, loaders.WithLoaders(context.Background(), l)
}

func TestLoadersIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	// Seed a user, team, decision, and two assumptions on that decision.
	u, err := users.NewRepository().Create(ctx, pool, users.CreateParams{
		ID: id.New(), Email: "owner@example.com", Name: "Owner", PasswordHash: "hash",
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
		ID: id.New(), TeamID: team.ID, OwnerID: u.ID, Title: "Launch", Status: decisions.StatusDraft,
	})
	if err != nil {
		t.Fatalf("seed decision: %v", err)
	}
	for _, s := range []string{"a", "b"} {
		if _, err := assumptions.NewRepository().Create(ctx, pool, assumptions.Assumption{
			ID: id.New(), DecisionID: d.ID, Statement: s, Confidence: 0.5,
		}); err != nil {
			t.Fatalf("seed assumption %q: %v", s, err)
		}
	}

	t.Run("GetUser returns the seeded user", func(t *testing.T) {
		_, lctx := newLoaders(pool)
		got, err := loaders.GetUser(lctx, u.ID)
		if err != nil {
			t.Fatalf("GetUser: %v", err)
		}
		if got.ID != u.ID || got.Email != "owner@example.com" {
			t.Errorf("GetUser = %+v, want %v / owner@example.com", got, u.ID)
		}
	})

	t.Run("GetUser returns NotFound for an unknown id", func(t *testing.T) {
		_, lctx := newLoaders(pool)
		_, err := loaders.GetUser(lctx, uuid.New())
		if errors.KindOf(err) != errors.KindNotFound {
			t.Errorf("GetUser unknown err = %v, want NotFound", err)
		}
	})

	t.Run("AssumptionsByDecision batches the decision's assumptions", func(t *testing.T) {
		_, lctx := newLoaders(pool)
		got, err := loaders.AssumptionsByDecision(lctx, d.ID)
		if err != nil {
			t.Fatalf("AssumptionsByDecision: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("AssumptionsByDecision len = %d, want 2", len(got))
		}
		for _, a := range got {
			if a.DecisionID != d.ID {
				t.Errorf("assumption decision = %v, want %v", a.DecisionID, d.ID)
			}
		}
	})

	t.Run("OutcomeByDecision returns nil when none exists", func(t *testing.T) {
		_, lctx := newLoaders(pool)
		got, err := loaders.OutcomeByDecision(lctx, d.ID)
		if err != nil {
			t.Fatalf("OutcomeByDecision: %v", err)
		}
		if got != nil {
			t.Errorf("OutcomeByDecision = %+v, want nil", got)
		}
	})

	t.Run("MembersByTeam returns the team's members", func(t *testing.T) {
		_, lctx := newLoaders(pool)
		got, err := loaders.MembersByTeam(lctx, team.ID)
		if err != nil {
			t.Fatalf("MembersByTeam: %v", err)
		}
		if len(got) != 1 || got[0].UserID != u.ID {
			t.Errorf("MembersByTeam = %+v, want one membership for %v", got, u.ID)
		}
	})

	t.Run("accessors error without loaders in context", func(t *testing.T) {
		if _, err := loaders.GetUser(context.Background(), u.ID); err == nil {
			t.Error("GetUser without loaders: expected error, got nil")
		}
	})
}
