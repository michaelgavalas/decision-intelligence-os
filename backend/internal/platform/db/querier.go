package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Querier is the subset of pgx used by repositories. It is satisfied by both
// *pgxpool.Pool and pgx.Tx, allowing the same repository code to run inside or
// outside a transaction.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
