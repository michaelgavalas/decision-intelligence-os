package pagination

import (
	"time"

	"github.com/google/uuid"
)

// PageInfo describes the bounds of a page within a Connection.
type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool
	StartCursor     *string
	EndCursor       *string
}

// Edge pairs a node with its cursor.
type Edge[T any] struct {
	Node   T
	Cursor string
}

// Connection is a Relay-style paginated result set.
type Connection[T any] struct {
	Edges      []Edge[T]
	PageInfo   PageInfo
	TotalCount int
}

// BuildConnection assembles a forward-pagination Connection. The items slice
// should hold up to Limit+1 rows: the extra row, when present, signals that a
// further page exists and is trimmed from the result. cursorOf extracts the
// (createdAt, id) position used to build each cursor.
func BuildConnection[T any](items []T, args PageArgs, totalCount int, cursorOf func(T) (time.Time, uuid.UUID)) Connection[T] {
	limit := args.Limit()

	hasNextPage := len(items) > limit
	if hasNextPage {
		items = items[:limit]
	}

	edges := make([]Edge[T], len(items))
	for i, item := range items {
		createdAt, nodeID := cursorOf(item)
		edges[i] = Edge[T]{Node: item, Cursor: EncodeCursor(createdAt, nodeID)}
	}

	pageInfo := PageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: args.After != nil,
	}
	if len(edges) > 0 {
		start := edges[0].Cursor
		end := edges[len(edges)-1].Cursor
		pageInfo.StartCursor = &start
		pageInfo.EndCursor = &end
	}

	return Connection[T]{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: totalCount,
	}
}
