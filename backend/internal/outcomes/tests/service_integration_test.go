package outcomes_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// seedActiveDecision inserts a user, team, membership, and an active decision so
// that an outcome can be recorded against it. It returns the decision and the
// owner id.
func seedActiveDecision(t *testing.T, pool *pgxpool.Pool) (decisions.Decision, uuid.UUID) {
	t.Helper()
	ctx := context.Background()

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
		ID: id.New(), TeamID: team.ID, OwnerID: u.ID, Title: "T", Status: decisions.StatusActive,
	})
	if err != nil {
		t.Fatalf("seed decision: %v", err)
	}
	return d, u.ID
}

func TestServiceRecordIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	decision, ownerID := seedActiveDecision(t, pool)

	txMgr := db.NewTxManager(pool)
	teamSvc := teams.NewService(pool, txMgr, teams.NewRepository(), clock.System{})
	decisionSvc := decisions.NewService(pool, txMgr, decisions.NewRepository(), events.NewRecorder(), teamSvc, clock.System{})
	svc := outcomes.NewService(pool, txMgr, outcomes.NewRepository(), events.NewRecorder(), decisionSvc, clock.System{})

	ctx := authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: ownerID})

	recorded, err := svc.Record(ctx, outcomes.RecordInput{DecisionID: decision.ID, Summary: "Shipped successfully", Success: true})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if !recorded.Success || recorded.Summary != "Shipped successfully" {
		t.Errorf("recorded = %+v, want success/Shipped successfully", recorded)
	}

	// The outcome row is readable back through the service.
	got, err := svc.GetForDecision(ctx, decision.ID)
	if err != nil {
		t.Fatalf("GetForDecision: %v", err)
	}
	if got.ID != recorded.ID {
		t.Errorf("GetForDecision id = %v, want %v", got.ID, recorded.ID)
	}

	// Recording an outcome on an active decision marks it decided.
	after, err := decisions.NewRepository().GetByID(ctx, pool, decision.ID)
	if err != nil {
		t.Fatalf("reload decision: %v", err)
	}
	if after.Status != decisions.StatusDecided {
		t.Errorf("decision status = %q, want decided", after.Status)
	}
	if after.DecidedAt == nil {
		t.Error("decided_at is nil, want set")
	}
}
