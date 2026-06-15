package outcomes

import (
	"context"
	stderrors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the persistence boundary for outcomes. Every method takes a
// db.Querier so callers choose whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	Upsert(ctx context.Context, q db.Querier, o Outcome) (Outcome, error)
	GetByDecision(ctx context.Context, q db.Querier, decisionID uuid.UUID) (Outcome, error)
	ListByDecisionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Outcome, error)
	ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Outcome, error)
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) Upsert(ctx context.Context, q db.Querier, o Outcome) (Outcome, error) {
	row, err := sqlc.New(q).UpsertOutcome(ctx, sqlc.UpsertOutcomeParams{
		ID:         o.ID,
		DecisionID: o.DecisionID,
		Summary:    o.Summary,
		Success:    o.Success,
		ResolvedAt: o.ResolvedAt,
	})
	if err != nil {
		return Outcome{}, errors.Wrap(err, errors.KindInternal, "OUTCOME_UPSERT_FAILED", "failed to record outcome")
	}
	return toOutcome(row), nil
}

func (repository) GetByDecision(ctx context.Context, q db.Querier, decisionID uuid.UUID) (Outcome, error) {
	row, err := sqlc.New(q).GetOutcomeByDecision(ctx, decisionID)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Outcome{}, errors.NotFound("OUTCOME_NOT_FOUND", "outcome not found")
		}
		return Outcome{}, errors.Wrap(err, errors.KindInternal, "OUTCOME_GET_FAILED", "failed to load outcome")
	}
	return toOutcome(row), nil
}

func (repository) ListByDecisionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Outcome, error) {
	rows, err := sqlc.New(q).ListOutcomesByDecisionIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "OUTCOME_LIST_FAILED", "failed to list outcomes")
	}
	out := make([]Outcome, 0, len(rows))
	for _, r := range rows {
		out = append(out, toOutcome(r))
	}
	return out, nil
}

func (repository) ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Outcome, error) {
	rows, err := sqlc.New(q).ListOutcomesByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "OUTCOME_LIST_FAILED", "failed to list outcomes")
	}
	out := make([]Outcome, 0, len(rows))
	for _, r := range rows {
		out = append(out, toOutcome(r))
	}
	return out, nil
}

// toOutcome maps a generated sqlc row to the domain entity.
func toOutcome(r sqlc.Outcome) Outcome {
	return Outcome{
		ID:         r.ID,
		DecisionID: r.DecisionID,
		Summary:    r.Summary,
		Success:    r.Success,
		ResolvedAt: r.ResolvedAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
