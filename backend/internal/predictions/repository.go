package predictions

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the persistence boundary for predictions. Every method takes a
// db.Querier so callers choose whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	Create(ctx context.Context, q db.Querier, p Prediction) (Prediction, error)
	GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Prediction, error)
	ListByDecision(ctx context.Context, q db.Querier, decisionID uuid.UUID) ([]Prediction, error)
	ListByDecisionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Prediction, error)
	ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Prediction, error)
	Update(ctx context.Context, q db.Querier, id uuid.UUID, statement string, probability float64, resolvesAt *time.Time) (Prediction, error)
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) Create(ctx context.Context, q db.Querier, p Prediction) (Prediction, error) {
	probability, err := toNumeric(p.Probability)
	if err != nil {
		return Prediction{}, err
	}
	row, err := sqlc.New(q).CreatePrediction(ctx, sqlc.CreatePredictionParams{
		ID:          p.ID,
		DecisionID:  p.DecisionID,
		Statement:   p.Statement,
		Probability: probability,
		ResolvesAt:  toTimestamptz(p.ResolvesAt),
	})
	if err != nil {
		return Prediction{}, errors.Wrap(err, errors.KindInternal, "PREDICTION_CREATE_FAILED", "failed to create prediction")
	}
	return toPrediction(row), nil
}

func (repository) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Prediction, error) {
	row, err := sqlc.New(q).GetPredictionByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Prediction{}, errors.NotFound("PREDICTION_NOT_FOUND", "prediction not found")
		}
		return Prediction{}, errors.Wrap(err, errors.KindInternal, "PREDICTION_GET_FAILED", "failed to load prediction")
	}
	return toPrediction(row), nil
}

func (repository) ListByDecision(ctx context.Context, q db.Querier, decisionID uuid.UUID) ([]Prediction, error) {
	rows, err := sqlc.New(q).ListPredictionsByDecision(ctx, decisionID)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "PREDICTION_LIST_FAILED", "failed to list predictions")
	}
	return toPredictions(rows), nil
}

func (repository) ListByDecisionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Prediction, error) {
	rows, err := sqlc.New(q).ListPredictionsByDecisionIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "PREDICTION_LIST_FAILED", "failed to list predictions")
	}
	return toPredictions(rows), nil
}

func (repository) ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Prediction, error) {
	rows, err := sqlc.New(q).ListPredictionsByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "PREDICTION_LIST_FAILED", "failed to list predictions")
	}
	return toPredictions(rows), nil
}

func (repository) Update(ctx context.Context, q db.Querier, id uuid.UUID, statement string, probability float64, resolvesAt *time.Time) (Prediction, error) {
	num, err := toNumeric(probability)
	if err != nil {
		return Prediction{}, err
	}
	row, err := sqlc.New(q).UpdatePrediction(ctx, sqlc.UpdatePredictionParams{
		ID:          id,
		Statement:   statement,
		Probability: num,
		ResolvesAt:  toTimestamptz(resolvesAt),
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Prediction{}, errors.NotFound("PREDICTION_NOT_FOUND", "prediction not found")
		}
		return Prediction{}, errors.Wrap(err, errors.KindInternal, "PREDICTION_UPDATE_FAILED", "failed to update prediction")
	}
	return toPrediction(row), nil
}

// toNumeric converts a float64 probability into the pgtype.Numeric the generated
// query expects, formatted to the column's three-decimal scale.
func toNumeric(f float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%.3f", f)); err != nil {
		return pgtype.Numeric{}, errors.Wrap(err, errors.KindInternal, "PROBABILITY_ENCODE_FAILED", "failed to encode probability")
	}
	return n, nil
}

// fromNumeric converts a stored pgtype.Numeric probability back into a float64.
func fromNumeric(n pgtype.Numeric) float64 {
	f, err := n.Float64Value()
	if err != nil {
		return 0
	}
	return f.Float64
}

// toTimestamptz converts an optional time into the nullable pgtype.Timestamptz
// the generated query expects.
func toTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// toPrediction maps a generated sqlc row to the domain entity.
func toPrediction(r sqlc.Prediction) Prediction {
	var resolvesAt *time.Time
	if r.ResolvesAt.Valid {
		t := r.ResolvesAt.Time
		resolvesAt = &t
	}
	return Prediction{
		ID:          r.ID,
		DecisionID:  r.DecisionID,
		Statement:   r.Statement,
		Probability: fromNumeric(r.Probability),
		ResolvesAt:  resolvesAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// toPredictions maps a slice of generated rows to domain entities.
func toPredictions(rows []sqlc.Prediction) []Prediction {
	out := make([]Prediction, 0, len(rows))
	for _, r := range rows {
		out = append(out, toPrediction(r))
	}
	return out
}
