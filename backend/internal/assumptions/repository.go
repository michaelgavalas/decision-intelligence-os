package assumptions

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the persistence boundary for assumptions. Every method takes a
// db.Querier so callers choose whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	Create(ctx context.Context, q db.Querier, a Assumption) (Assumption, error)
	GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Assumption, error)
	ListByDecision(ctx context.Context, q db.Querier, decisionID uuid.UUID) ([]Assumption, error)
	ListByDecisionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Assumption, error)
	ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Assumption, error)
	Update(ctx context.Context, q db.Querier, id uuid.UUID, statement string, confidence float64) (Assumption, error)
	Delete(ctx context.Context, q db.Querier, id uuid.UUID) error
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) Create(ctx context.Context, q db.Querier, a Assumption) (Assumption, error) {
	confidence, err := toNumeric(a.Confidence)
	if err != nil {
		return Assumption{}, err
	}
	row, err := sqlc.New(q).CreateAssumption(ctx, sqlc.CreateAssumptionParams{
		ID:         a.ID,
		DecisionID: a.DecisionID,
		Statement:  a.Statement,
		Confidence: confidence,
	})
	if err != nil {
		return Assumption{}, errors.Wrap(err, errors.KindInternal, "ASSUMPTION_CREATE_FAILED", "failed to create assumption")
	}
	return toAssumption(row), nil
}

func (repository) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Assumption, error) {
	row, err := sqlc.New(q).GetAssumptionByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Assumption{}, errors.NotFound("ASSUMPTION_NOT_FOUND", "assumption not found")
		}
		return Assumption{}, errors.Wrap(err, errors.KindInternal, "ASSUMPTION_GET_FAILED", "failed to load assumption")
	}
	return toAssumption(row), nil
}

func (repository) ListByDecision(ctx context.Context, q db.Querier, decisionID uuid.UUID) ([]Assumption, error) {
	rows, err := sqlc.New(q).ListAssumptionsByDecision(ctx, decisionID)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "ASSUMPTION_LIST_FAILED", "failed to list assumptions")
	}
	return toAssumptions(rows), nil
}

func (repository) ListByDecisionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Assumption, error) {
	rows, err := sqlc.New(q).ListAssumptionsByDecisionIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "ASSUMPTION_LIST_FAILED", "failed to list assumptions")
	}
	return toAssumptions(rows), nil
}

func (repository) ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Assumption, error) {
	rows, err := sqlc.New(q).ListAssumptionsByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "ASSUMPTION_LIST_FAILED", "failed to list assumptions")
	}
	return toAssumptions(rows), nil
}

func (repository) Update(ctx context.Context, q db.Querier, id uuid.UUID, statement string, confidence float64) (Assumption, error) {
	num, err := toNumeric(confidence)
	if err != nil {
		return Assumption{}, err
	}
	row, err := sqlc.New(q).UpdateAssumption(ctx, sqlc.UpdateAssumptionParams{
		ID:         id,
		Statement:  statement,
		Confidence: num,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Assumption{}, errors.NotFound("ASSUMPTION_NOT_FOUND", "assumption not found")
		}
		return Assumption{}, errors.Wrap(err, errors.KindInternal, "ASSUMPTION_UPDATE_FAILED", "failed to update assumption")
	}
	return toAssumption(row), nil
}

func (repository) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	if err := sqlc.New(q).DeleteAssumption(ctx, id); err != nil {
		return errors.Wrap(err, errors.KindInternal, "ASSUMPTION_DELETE_FAILED", "failed to delete assumption")
	}
	return nil
}

// toNumeric converts a float64 confidence into the pgtype.Numeric the generated
// query expects, formatted to the column's three-decimal scale.
func toNumeric(f float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%.3f", f)); err != nil {
		return pgtype.Numeric{}, errors.Wrap(err, errors.KindInternal, "CONFIDENCE_ENCODE_FAILED", "failed to encode confidence")
	}
	return n, nil
}

// fromNumeric converts a stored pgtype.Numeric confidence back into a float64.
func fromNumeric(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil {
		return 0
	}
	return f.Float64
}

// toAssumption maps a generated sqlc row to the domain entity.
func toAssumption(r sqlc.Assumption) Assumption {
	return Assumption{
		ID:         r.ID,
		DecisionID: r.DecisionID,
		Statement:  r.Statement,
		Confidence: fromNumeric(r.Confidence),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

// toAssumptions maps a slice of generated rows to domain entities.
func toAssumptions(rows []sqlc.Assumption) []Assumption {
	out := make([]Assumption, 0, len(rows))
	for _, r := range rows {
		out = append(out, toAssumption(r))
	}
	return out
}
