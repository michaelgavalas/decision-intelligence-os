package evidence_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/evidence"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// seedAssumption inserts the user, team, decision, and assumption chain so that
// evidence foreign keys are satisfied, returning the assumption id.
func seedAssumption(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
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
		ID: id.New(), TeamID: team.ID, OwnerID: u.ID, Title: "T", Status: decisions.StatusDraft,
	})
	if err != nil {
		t.Fatalf("seed decision: %v", err)
	}
	a, err := assumptions.NewRepository().Create(ctx, pool, assumptions.Assumption{
		ID: id.New(), DecisionID: d.ID, Statement: "s", Confidence: 0.5,
	})
	if err != nil {
		t.Fatalf("seed assumption: %v", err)
	}
	return a.ID
}

func strPtr(s string) *string { return &s }

func TestEvidenceRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := evidence.NewRepository()
	ctx := context.Background()

	t.Run("round-trip create, get, update, list, delete", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		assumptionID := seedAssumption(t, pool, "owner@example.com")

		created, err := repo.Create(ctx, pool, evidence.Evidence{
			ID:           id.New(),
			AssumptionID: assumptionID,
			SourceType:   evidence.SourceURL,
			SourceURL:    strPtr("https://example.com"),
			Content:      "Report",
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if created.SourceURL == nil || *created.SourceURL != "https://example.com" {
			t.Errorf("source url = %v, want example.com", created.SourceURL)
		}

		got, err := repo.GetByID(ctx, pool, created.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.SourceType != evidence.SourceURL {
			t.Errorf("source type = %q, want url", got.SourceType)
		}

		// Update to a note with no URL, exercising the nullable source_url column.
		updated, err := repo.Update(ctx, pool, created.ID, evidence.SourceNote, nil, "A written note")
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.SourceType != evidence.SourceNote || updated.SourceURL != nil {
			t.Errorf("Update = %+v, want note with nil url", updated)
		}

		list, err := repo.ListByAssumption(ctx, pool, assumptionID)
		if err != nil {
			t.Fatalf("ListByAssumption: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("ListByAssumption len = %d, want 1", len(list))
		}

		if err := repo.Delete(ctx, pool, created.ID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repo.GetByID(ctx, pool, created.ID); errors.KindOf(err) != errors.KindNotFound {
			t.Errorf("GetByID after delete err = %v, want NotFound", err)
		}
	})

	t.Run("list by multiple assumption ids", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		a1 := seedAssumption(t, pool, "a@example.com")
		a2 := seedAssumption(t, pool, "b@example.com")

		for _, a := range []uuid.UUID{a1, a2} {
			if _, err := repo.Create(ctx, pool, evidence.Evidence{
				ID: id.New(), AssumptionID: a, SourceType: evidence.SourceNote, Content: "c",
			}); err != nil {
				t.Fatalf("Create for %v: %v", a, err)
			}
		}

		rows, err := repo.ListByAssumptionIDs(ctx, pool, []uuid.UUID{a1, a2})
		if err != nil {
			t.Fatalf("ListByAssumptionIDs: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("ListByAssumptionIDs len = %d, want 2", len(rows))
		}
		seen := map[uuid.UUID]bool{}
		for _, r := range rows {
			seen[r.AssumptionID] = true
		}
		if !seen[a1] || !seen[a2] {
			t.Errorf("ListByAssumptionIDs covered = %v, want both assumptions", seen)
		}
	})
}

func TestEvidenceRepositoryListByIDs(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := evidence.NewRepository()
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	assumptionID := seedAssumption(t, pool, "list@example.com")
	e1, err := repo.Create(ctx, pool, evidence.Evidence{
		ID: id.New(), AssumptionID: assumptionID, SourceType: evidence.SourceNote, Content: "one",
	})
	if err != nil {
		t.Fatalf("Create e1: %v", err)
	}
	e2, err := repo.Create(ctx, pool, evidence.Evidence{
		ID: id.New(), AssumptionID: assumptionID, SourceType: evidence.SourceNote, Content: "two",
	})
	if err != nil {
		t.Fatalf("Create e2: %v", err)
	}

	rows, err := repo.ListByIDs(ctx, pool, []uuid.UUID{e1.ID, e2.ID})
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
	if !seen[e1.ID] || !seen[e2.ID] {
		t.Errorf("ListByIDs covered = %v, want both evidence", seen)
	}
}
