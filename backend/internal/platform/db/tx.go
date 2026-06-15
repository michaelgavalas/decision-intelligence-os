package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// TxManager runs units of work inside a database transaction.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager returns a TxManager backed by the given pool.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx begins a transaction and invokes fn with it as a Querier. It commits
// when fn returns nil and rolls back otherwise. A panic in fn triggers a
// rollback and is re-raised so callers observe the original failure.
func (m *TxManager) WithinTx(ctx context.Context, fn func(q Querier) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, errors.KindInternal, "DB_TX_BEGIN_FAILED", "failed to begin transaction")
	}

	committed := false
	defer func() {
		if !committed {
			// Rollback is best-effort; a committed tx returns ErrTxClosed which
			// is intentionally ignored here.
			_ = tx.Rollback(ctx)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, errors.KindInternal, "DB_TX_COMMIT_FAILED", "failed to commit transaction")
	}
	committed = true
	return nil
}
