package teams_test

import (
	"context"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
)

// TestCreateTeamMakesCallerAdmin exercises the real transactional path: the
// service creates a team and the caller's admin membership atomically.
func TestCreateTeamMakesCallerAdmin(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	caller := seedUser(t, pool, "founder@example.com")
	repo := teams.NewRepository()
	svc := teams.NewService(pool, db.NewTxManager(pool), repo, clock.System{})

	ctx := authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: caller})

	team, err := svc.CreateTeam(ctx, "Founders")
	if err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	m, err := repo.GetMembership(context.Background(), pool, team.ID, caller)
	if err != nil {
		t.Fatalf("GetMembership: %v", err)
	}
	if m.Role != teams.RoleAdmin {
		t.Errorf("caller role = %q, want admin", m.Role)
	}

	mine, err := svc.ListMyTeams(ctx)
	if err != nil {
		t.Fatalf("ListMyTeams: %v", err)
	}
	if len(mine) != 1 || mine[0].ID != team.ID {
		t.Errorf("ListMyTeams = %+v, want one team %v", mine, team.ID)
	}
}
