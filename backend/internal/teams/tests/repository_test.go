package teams_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// seedUser inserts a user so that team_member foreign keys are satisfied.
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

func TestTeamRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := teams.NewRepository()
	ctx := context.Background()

	t.Run("create team, add and list members", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		owner := seedUser(t, pool, "owner@example.com")
		other := seedUser(t, pool, "other@example.com")

		team, err := repo.CreateTeam(ctx, pool, id.New(), "Acme")
		if err != nil {
			t.Fatalf("CreateTeam: %v", err)
		}

		if _, err := repo.AddMember(ctx, pool, team.ID, owner, teams.RoleAdmin); err != nil {
			t.Fatalf("AddMember owner: %v", err)
		}
		if _, err := repo.AddMember(ctx, pool, team.ID, other, teams.RoleMember); err != nil {
			t.Fatalf("AddMember other: %v", err)
		}

		members, err := repo.ListMembers(ctx, pool, team.ID)
		if err != nil {
			t.Fatalf("ListMembers: %v", err)
		}
		if len(members) != 2 {
			t.Fatalf("ListMembers len = %d, want 2", len(members))
		}

		mine, err := repo.ListTeamsForUser(ctx, pool, owner)
		if err != nil {
			t.Fatalf("ListTeamsForUser: %v", err)
		}
		if len(mine) != 1 || mine[0].ID != team.ID {
			t.Errorf("ListTeamsForUser = %+v, want one team %v", mine, team.ID)
		}
	})

	t.Run("get team and membership", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		owner := seedUser(t, pool, "owner@example.com")
		team, err := repo.CreateTeam(ctx, pool, id.New(), "Acme")
		if err != nil {
			t.Fatalf("CreateTeam: %v", err)
		}
		if _, err := repo.AddMember(ctx, pool, team.ID, owner, teams.RoleAdmin); err != nil {
			t.Fatalf("AddMember: %v", err)
		}

		got, err := repo.GetTeam(ctx, pool, team.ID)
		if err != nil {
			t.Fatalf("GetTeam: %v", err)
		}
		if got.Name != "Acme" {
			t.Errorf("GetTeam name = %q, want Acme", got.Name)
		}

		m, err := repo.GetMembership(ctx, pool, team.ID, owner)
		if err != nil {
			t.Fatalf("GetMembership: %v", err)
		}
		if m.Role != teams.RoleAdmin {
			t.Errorf("GetMembership role = %q, want admin", m.Role)
		}
	})

	t.Run("missing team and membership are not found", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)

		_, err := repo.GetTeam(ctx, pool, uuid.New())
		if errors.KindOf(err) != errors.KindNotFound || errors.CodeOf(err) != "TEAM_NOT_FOUND" {
			t.Errorf("GetTeam err = %v, want NotFound/TEAM_NOT_FOUND", err)
		}

		_, err = repo.GetMembership(ctx, pool, uuid.New(), uuid.New())
		if errors.KindOf(err) != errors.KindNotFound || errors.CodeOf(err) != "MEMBERSHIP_NOT_FOUND" {
			t.Errorf("GetMembership err = %v, want NotFound/MEMBERSHIP_NOT_FOUND", err)
		}
	})

	t.Run("update role, count admins, remove member", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		owner := seedUser(t, pool, "owner@example.com")
		other := seedUser(t, pool, "other@example.com")
		team, err := repo.CreateTeam(ctx, pool, id.New(), "Acme")
		if err != nil {
			t.Fatalf("CreateTeam: %v", err)
		}
		if _, err := repo.AddMember(ctx, pool, team.ID, owner, teams.RoleAdmin); err != nil {
			t.Fatalf("AddMember owner: %v", err)
		}
		if _, err := repo.AddMember(ctx, pool, team.ID, other, teams.RoleMember); err != nil {
			t.Fatalf("AddMember other: %v", err)
		}

		if got, _ := repo.CountAdmins(ctx, pool, team.ID); got != 1 {
			t.Errorf("CountAdmins = %d, want 1", got)
		}

		if _, err := repo.UpdateMemberRole(ctx, pool, team.ID, other, teams.RoleAdmin); err != nil {
			t.Fatalf("UpdateMemberRole: %v", err)
		}
		if got, _ := repo.CountAdmins(ctx, pool, team.ID); got != 2 {
			t.Errorf("CountAdmins after promote = %d, want 2", got)
		}

		if err := repo.RemoveMember(ctx, pool, team.ID, other); err != nil {
			t.Fatalf("RemoveMember: %v", err)
		}
		if got, _ := repo.CountAdmins(ctx, pool, team.ID); got != 1 {
			t.Errorf("CountAdmins after remove = %d, want 1", got)
		}
	})

	t.Run("duplicate membership is a conflict", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		owner := seedUser(t, pool, "owner@example.com")
		team, err := repo.CreateTeam(ctx, pool, id.New(), "Acme")
		if err != nil {
			t.Fatalf("CreateTeam: %v", err)
		}
		if _, err := repo.AddMember(ctx, pool, team.ID, owner, teams.RoleAdmin); err != nil {
			t.Fatalf("AddMember: %v", err)
		}

		_, err = repo.AddMember(ctx, pool, team.ID, owner, teams.RoleMember)
		if errors.KindOf(err) != errors.KindConflict || errors.CodeOf(err) != "ALREADY_MEMBER" {
			t.Errorf("duplicate member err = %v, want Conflict/ALREADY_MEMBER", err)
		}
	})
}

func TestTeamRepositoryListByIDs(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := teams.NewRepository()
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	owner := seedUser(t, pool, "list-owner@example.com")
	t1, err := repo.CreateTeam(ctx, pool, id.New(), "Alpha")
	if err != nil {
		t.Fatalf("CreateTeam Alpha: %v", err)
	}
	t2, err := repo.CreateTeam(ctx, pool, id.New(), "Beta")
	if err != nil {
		t.Fatalf("CreateTeam Beta: %v", err)
	}
	if _, err := repo.AddMember(ctx, pool, t1.ID, owner, teams.RoleAdmin); err != nil {
		t.Fatalf("AddMember t1: %v", err)
	}
	if _, err := repo.AddMember(ctx, pool, t2.ID, owner, teams.RoleMember); err != nil {
		t.Fatalf("AddMember t2: %v", err)
	}

	got, err := repo.ListByIDs(ctx, pool, []uuid.UUID{t1.ID, t2.ID})
	if err != nil {
		t.Fatalf("ListByIDs: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListByIDs len = %d, want 2", len(got))
	}

	members, err := repo.ListMembersByTeamIDs(ctx, pool, []uuid.UUID{t1.ID, t2.ID})
	if err != nil {
		t.Fatalf("ListMembersByTeamIDs: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("ListMembersByTeamIDs len = %d, want 2", len(members))
	}
	byTeam := map[uuid.UUID]bool{}
	for _, m := range members {
		byTeam[m.TeamID] = true
	}
	if !byTeam[t1.ID] || !byTeam[t2.ID] {
		t.Errorf("ListMembersByTeamIDs covered = %v, want both teams", byTeam)
	}
}
