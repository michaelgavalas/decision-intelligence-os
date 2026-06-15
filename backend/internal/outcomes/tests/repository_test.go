package outcomes_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

func TestOutcomeRepositoryListByIDs(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := outcomes.NewRepository()
	ctx := context.Background()
	dbtest.TruncateAll(t, pool)

	decision, _ := seedActiveDecision(t, pool)
	o, err := repo.Upsert(ctx, pool, outcomes.Outcome{
		ID:         id.New(),
		DecisionID: decision.ID,
		Summary:    "shipped",
		Success:    true,
		ResolvedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	rows, err := repo.ListByIDs(ctx, pool, []uuid.UUID{o.ID})
	if err != nil {
		t.Fatalf("ListByIDs: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListByIDs len = %d, want 1", len(rows))
	}
	if rows[0].ID != o.ID || rows[0].DecisionID != decision.ID {
		t.Errorf("ListByIDs = %+v, want id %v decision %v", rows[0], o.ID, decision.ID)
	}
}
