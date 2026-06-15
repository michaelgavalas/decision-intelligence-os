package decisions_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// seedTeam inserts a user and a team owned by that user, returning both ids so
// decision foreign keys are satisfied.
func seedTeam(t *testing.T, pool *pgxpool.Pool) (teamID, ownerID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	u, err := users.NewRepository().Create(ctx, pool, users.CreateParams{
		ID:           id.New(),
		Email:        "owner@example.com",
		Name:         "Owner",
		PasswordHash: "hash",
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
	return team.ID, u.ID
}

func TestDecisionRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := decisions.NewRepository()
	ctx := context.Background()

	t.Run("create, get, update, status", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		teamID, ownerID := seedTeam(t, pool)

		created, err := repo.Create(ctx, pool, decisions.Decision{
			ID:          id.New(),
			TeamID:      teamID,
			OwnerID:     ownerID,
			Title:       "Launch in EU",
			Description: "Expand to the EU market",
			Status:      decisions.StatusDraft,
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if created.Status != decisions.StatusDraft || created.DecidedAt != nil {
			t.Errorf("created = %+v, want draft with nil DecidedAt", created)
		}

		got, err := repo.GetByID(ctx, pool, created.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Title != "Launch in EU" {
			t.Errorf("GetByID title = %q, want Launch in EU", got.Title)
		}

		updated, err := repo.Update(ctx, pool, created.ID, "Launch in EU (revised)", "New description")
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Title != "Launch in EU (revised)" || updated.Description != "New description" {
			t.Errorf("Update = %+v, want revised title/description", updated)
		}

		now := got.CreatedAt
		decided, err := repo.UpdateStatus(ctx, pool, created.ID, decisions.StatusDecided, &now)
		if err != nil {
			t.Fatalf("UpdateStatus: %v", err)
		}
		if decided.Status != decisions.StatusDecided || decided.DecidedAt == nil {
			t.Errorf("UpdateStatus = %+v, want decided with DecidedAt set", decided)
		}
	})

	t.Run("get missing is not found", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		_, err := repo.GetByID(ctx, pool, uuid.New())
		if errors.KindOf(err) != errors.KindNotFound || errors.CodeOf(err) != "DECISION_NOT_FOUND" {
			t.Errorf("GetByID err = %v, want NotFound/DECISION_NOT_FOUND", err)
		}
	})

	t.Run("list and keyset pagination", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		teamID, ownerID := seedTeam(t, pool)

		const total = 5
		for i := 0; i < total; i++ {
			if _, err := repo.Create(ctx, pool, decisions.Decision{
				ID:      id.New(),
				TeamID:  teamID,
				OwnerID: ownerID,
				Title:   "Decision",
				Status:  decisions.StatusDraft,
			}); err != nil {
				t.Fatalf("Create %d: %v", i, err)
			}
		}

		count, err := repo.CountByTeam(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("CountByTeam: %v", err)
		}
		if count != total {
			t.Errorf("CountByTeam = %d, want %d", count, total)
		}

		// First page of 2.
		page1, err := repo.ListByTeam(ctx, pool, teamID, 2)
		if err != nil {
			t.Fatalf("ListByTeam: %v", err)
		}
		if len(page1) != 2 {
			t.Fatalf("page1 len = %d, want 2", len(page1))
		}

		// Page after the last item of page1 using keyset bounds.
		last := page1[len(page1)-1]
		page2, err := repo.ListByTeamAfter(ctx, pool, teamID, last.CreatedAt, last.ID, 2)
		if err != nil {
			t.Fatalf("ListByTeamAfter: %v", err)
		}
		if len(page2) != 2 {
			t.Fatalf("page2 len = %d, want 2", len(page2))
		}
		// Pages must not overlap.
		for _, a := range page1 {
			for _, b := range page2 {
				if a.ID == b.ID {
					t.Errorf("pages overlap on %v", a.ID)
				}
			}
		}
	})
}

func TestDecisionRepositoryListByIDs(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := decisions.NewRepository()
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	teamID, ownerID := seedTeam(t, pool)
	d1, err := repo.Create(ctx, pool, decisions.Decision{
		ID: id.New(), TeamID: teamID, OwnerID: ownerID, Title: "One", Status: decisions.StatusDraft,
	})
	if err != nil {
		t.Fatalf("Create d1: %v", err)
	}
	d2, err := repo.Create(ctx, pool, decisions.Decision{
		ID: id.New(), TeamID: teamID, OwnerID: ownerID, Title: "Two", Status: decisions.StatusDraft,
	})
	if err != nil {
		t.Fatalf("Create d2: %v", err)
	}

	rows, err := repo.ListByIDs(ctx, pool, []uuid.UUID{d1.ID, d2.ID})
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
	if !seen[d1.ID] || !seen[d2.ID] {
		t.Errorf("ListByIDs covered = %v, want both decisions", seen)
	}
}
