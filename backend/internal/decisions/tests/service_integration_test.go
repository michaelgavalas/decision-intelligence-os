package decisions_test

import (
	"context"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/pagination"
)

func intPtr(n int) *int { return &n }

// TestServiceListPaginationIntegration exercises the real transactional create
// path and cursor pagination through the service against PostgreSQL.
func TestServiceListPaginationIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	teamID, ownerID := seedTeam(t, pool)

	teamSvc := teams.NewService(pool, db.NewTxManager(pool), teams.NewRepository(), clock.System{})
	svc := decisions.NewService(
		pool,
		db.NewTxManager(pool),
		decisions.NewRepository(),
		events.NewRecorder(),
		teamSvc,
		clock.System{},
	)

	ctx := authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: ownerID})

	const total = 5
	for i := 0; i < total; i++ {
		if _, err := svc.Create(ctx, decisions.CreateInput{TeamID: teamID, Title: "Decision"}); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	first := 2
	page1, err := svc.List(ctx, teamID, pagination.PageArgs{First: intPtr(first)})
	if err != nil {
		t.Fatalf("List page1: %v", err)
	}
	if page1.TotalCount != total {
		t.Errorf("TotalCount = %d, want %d", page1.TotalCount, total)
	}
	if len(page1.Edges) != first {
		t.Fatalf("page1 edges = %d, want %d", len(page1.Edges), first)
	}
	if !page1.PageInfo.HasNextPage {
		t.Error("page1 HasNextPage = false, want true")
	}

	page2, err := svc.List(ctx, teamID, pagination.PageArgs{First: intPtr(first), After: page1.PageInfo.EndCursor})
	if err != nil {
		t.Fatalf("List page2: %v", err)
	}
	if len(page2.Edges) != first {
		t.Fatalf("page2 edges = %d, want %d", len(page2.Edges), first)
	}
	if !page2.PageInfo.HasPreviousPage {
		t.Error("page2 HasPreviousPage = false, want true")
	}

	// Pages must not overlap.
	for _, a := range page1.Edges {
		for _, b := range page2.Edges {
			if a.Node.ID == b.Node.ID {
				t.Errorf("pages overlap on %v", a.Node.ID)
			}
		}
	}

	// Final page holds the remaining row and reports no further pages.
	page3, err := svc.List(ctx, teamID, pagination.PageArgs{First: intPtr(first), After: page2.PageInfo.EndCursor})
	if err != nil {
		t.Fatalf("List page3: %v", err)
	}
	if len(page3.Edges) != 1 {
		t.Errorf("page3 edges = %d, want 1", len(page3.Edges))
	}
	if page3.PageInfo.HasNextPage {
		t.Error("page3 HasNextPage = true, want false")
	}
}
