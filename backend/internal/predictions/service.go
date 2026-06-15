package predictions

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// statementMaxLen bounds a prediction statement, matching the database check.
const statementMaxLen = 2000

// aggregateType labels prediction events in the audit log.
const aggregateType = "prediction"

// Decisions is the narrow slice of the decisions domain the predictions domain
// depends on: authorizing access to a decision and learning the caller's role.
type Decisions interface {
	AuthorizeAccess(ctx context.Context, id uuid.UUID) (decisions.Decision, string, error)
}

// txRunner runs a unit of work inside a database transaction.
type txRunner interface {
	WithinTx(ctx context.Context, fn func(q db.Querier) error) error
}

// recorder appends an event to the audit log using the supplied querier.
type recorder interface {
	Record(ctx context.Context, q db.Querier, e events.Event) error
}

// CreateInput carries the fields needed to add a prediction to a decision.
type CreateInput struct {
	DecisionID  uuid.UUID
	Statement   string
	Probability float64
	ResolvesAt  *time.Time
}

// Service is the predictions domain's application boundary. It authorizes every
// operation through the parent decision and records audit events.
type Service interface {
	// Create records a new prediction on a decision. The caller must be a
	// non-viewer member of the decision's team. Emits PredictionCreated.
	Create(ctx context.Context, in CreateInput) (Prediction, error)
	// GetByID returns a prediction visible to any member of its decision's team.
	GetByID(ctx context.Context, id uuid.UUID) (Prediction, error)
	// ListForDecision returns a decision's predictions for any team member.
	ListForDecision(ctx context.Context, decisionID uuid.UUID) ([]Prediction, error)
	// Update edits a prediction. The caller must be a non-viewer member.
	Update(ctx context.Context, id uuid.UUID, statement string, probability float64, resolvesAt *time.Time) (Prediction, error)
}

// service is the default Service implementation.
type service struct {
	pool      *pgxpool.Pool
	tx        txRunner
	repo      Repository
	recorder  recorder
	decisions Decisions
}

// NewService wires a Service from its collaborators.
func NewService(
	pool *pgxpool.Pool,
	tx *db.TxManager,
	repo Repository,
	rec *events.Recorder,
	decisionsDep Decisions,
) Service {
	return &service{
		pool:      pool,
		tx:        tx,
		repo:      repo,
		recorder:  rec,
		decisions: decisionsDep,
	}
}

func (s *service) Create(ctx context.Context, in CreateInput) (Prediction, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Prediction{}, err
	}
	_, role, err := s.decisions.AuthorizeAccess(ctx, in.DecisionID)
	if err != nil {
		return Prediction{}, err
	}
	if role == teams.RoleViewer {
		return Prediction{}, errors.Forbidden("PREDICTION_FORBIDDEN", "viewers cannot add predictions")
	}

	statement := strings.TrimSpace(in.Statement)
	if err := validate(statement, in.Probability); err != nil {
		return Prediction{}, err
	}

	prediction := Prediction{
		ID:          id.New(),
		DecisionID:  in.DecisionID,
		Statement:   statement,
		Probability: in.Probability,
		ResolvesAt:  in.ResolvesAt,
	}

	var created Prediction
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, q, prediction)
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		return s.recorder.Record(ctx, q, events.Event{
			AggregateID:   created.ID,
			AggregateType: aggregateType,
			Type:          events.TypePredictionCreated,
			Payload:       map[string]any{"decision_id": created.DecisionID.String(), "probability": created.Probability},
			ActorID:       &actor,
		})
	})
	if err != nil {
		return Prediction{}, err
	}
	return created, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (Prediction, error) {
	prediction, err := s.repo.GetByID(ctx, s.pool, id)
	if err != nil {
		return Prediction{}, err
	}
	if _, _, err := s.decisions.AuthorizeAccess(ctx, prediction.DecisionID); err != nil {
		return Prediction{}, err
	}
	return prediction, nil
}

func (s *service) ListForDecision(ctx context.Context, decisionID uuid.UUID) ([]Prediction, error) {
	if _, _, err := s.decisions.AuthorizeAccess(ctx, decisionID); err != nil {
		return nil, err
	}
	return s.repo.ListByDecision(ctx, s.pool, decisionID)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, statement string, probability float64, resolvesAt *time.Time) (Prediction, error) {
	if _, err := authctx.Require(ctx); err != nil {
		return Prediction{}, err
	}
	prediction, err := s.repo.GetByID(ctx, s.pool, id)
	if err != nil {
		return Prediction{}, err
	}
	_, role, err := s.decisions.AuthorizeAccess(ctx, prediction.DecisionID)
	if err != nil {
		return Prediction{}, err
	}
	if role == teams.RoleViewer {
		return Prediction{}, errors.Forbidden("PREDICTION_FORBIDDEN", "viewers cannot edit predictions")
	}

	trimmed := strings.TrimSpace(statement)
	if err := validate(trimmed, probability); err != nil {
		return Prediction{}, err
	}

	var updated Prediction
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, q, id, trimmed, probability, resolvesAt)
		return txErr
	})
	if err != nil {
		return Prediction{}, err
	}
	return updated, nil
}

// validate enforces the statement length and probability range invariants.
func validate(statement string, probability float64) error {
	if statement == "" {
		return errors.Validation("PREDICTION_STATEMENT_REQUIRED", "prediction statement is required")
	}
	if len(statement) > statementMaxLen {
		return errors.Validation("PREDICTION_STATEMENT_TOO_LONG", "prediction statement must be at most 2000 characters")
	}
	if probability < 0 || probability > 1 {
		return errors.Validation("INVALID_PROBABILITY", "probability must be between 0 and 1")
	}
	return nil
}
