package pagination_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/pagination"
)

type node struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

func cursorOf(n node) (time.Time, uuid.UUID) {
	return n.CreatedAt, n.ID
}

func makeNodes(n int) []node {
	base := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	nodes := make([]node, n)
	for i := range nodes {
		nodes[i] = node{ID: id.New(), CreatedAt: base.Add(time.Duration(i) * time.Minute)}
	}
	return nodes
}

func TestBuildConnectionHasNextPage(t *testing.T) {
	// Limit 20, supply 21 items (Limit+1) to signal another page.
	nodes := makeNodes(21)
	args := pagination.PageArgs{First: intPtr(20)}

	conn := pagination.BuildConnection(nodes, args, 100, cursorOf)

	if len(conn.Edges) != 20 {
		t.Fatalf("len(Edges) = %d, want 20 (trimmed)", len(conn.Edges))
	}
	if !conn.PageInfo.HasNextPage {
		t.Error("HasNextPage = false, want true")
	}
	if conn.TotalCount != 100 {
		t.Errorf("TotalCount = %d, want 100", conn.TotalCount)
	}
	if conn.PageInfo.StartCursor == nil || conn.PageInfo.EndCursor == nil {
		t.Fatal("Start/End cursor must be non-nil for non-empty page")
	}

	wantStart := pagination.EncodeCursor(nodes[0].CreatedAt, nodes[0].ID)
	wantEnd := pagination.EncodeCursor(nodes[19].CreatedAt, nodes[19].ID)
	if *conn.PageInfo.StartCursor != wantStart {
		t.Errorf("StartCursor = %q, want %q", *conn.PageInfo.StartCursor, wantStart)
	}
	if *conn.PageInfo.EndCursor != wantEnd {
		t.Errorf("EndCursor = %q, want %q", *conn.PageInfo.EndCursor, wantEnd)
	}
	if conn.Edges[0].Cursor != wantStart {
		t.Errorf("Edges[0].Cursor = %q, want %q", conn.Edges[0].Cursor, wantStart)
	}
}

func TestBuildConnectionNoNextPage(t *testing.T) {
	nodes := makeNodes(5)
	args := pagination.PageArgs{First: intPtr(20)}

	conn := pagination.BuildConnection(nodes, args, 5, cursorOf)

	if len(conn.Edges) != 5 {
		t.Fatalf("len(Edges) = %d, want 5", len(conn.Edges))
	}
	if conn.PageInfo.HasNextPage {
		t.Error("HasNextPage = true, want false")
	}
}

func TestBuildConnectionEmpty(t *testing.T) {
	conn := pagination.BuildConnection([]node{}, pagination.PageArgs{}, 0, cursorOf)
	if len(conn.Edges) != 0 {
		t.Errorf("len(Edges) = %d, want 0", len(conn.Edges))
	}
	if conn.PageInfo.HasNextPage {
		t.Error("HasNextPage = true on empty, want false")
	}
	if conn.PageInfo.StartCursor != nil || conn.PageInfo.EndCursor != nil {
		t.Error("cursors must be nil on empty page")
	}
}

func TestBuildConnectionAfterSetsHasPreviousPage(t *testing.T) {
	nodes := makeNodes(3)
	after := pagination.EncodeCursor(nodes[0].CreatedAt, nodes[0].ID)
	args := pagination.PageArgs{First: intPtr(20), After: &after}

	conn := pagination.BuildConnection(nodes, args, 10, cursorOf)
	if !conn.PageInfo.HasPreviousPage {
		t.Error("HasPreviousPage = false, want true when After is set")
	}
}
