package assumptions

import (
	"context"
	"strings"

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

// statementMaxLen bounds an assumption statement, matching the database check.
const statementMaxLen = 2000

// aggregateType labels assumption events in the audit log.
const aggregateType = "assumption"

// Decisions is the narrow slice of the decisions domain the assumptions domain
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

// AddInput carries the fields needed to add an assumption to a decision.
type AddInput struct {
	DecisionID uuid.UUID
	Statement  string
	Confidence float64
}

// Service is the assumptions domain's application boundary. It authorizes every
// operation through the parent decision and records audit events.
type Service interface {
	// Add records a new assumption on a decision. The caller must be a non-viewer
	// member of the decision's team. Emits AssumptionAdded.
	Add(ctx context.Context, in AddInput) (Assumption, error)
	// GetByID returns an assumption visible to any member of its decision's team.
	GetByID(ctx context.Context, id uuid.UUID) (Assumption, error)
	// AuthorizeAccess returns an assumption and the caller's role on its decision,
	// or Forbidden when the caller is not a member. The evidence domain uses it.
	AuthorizeAccess(ctx context.Context, id uuid.UUID) (Assumption, string, error)
	// ListForDecision returns a decision's assumptions for any team member.
	ListForDecision(ctx context.Context, decisionID uuid.UUID) ([]Assumption, error)
	// Update edits an assumption. The caller must be a non-viewer member.
	Update(ctx context.Context, id uuid.UUID, statement string, confidence float64) (Assumption, error)
	// Remove deletes an assumption. The caller must be a non-viewer member.
	Remove(ctx context.Context, id uuid.UUID) error
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

func (s *service) Add(ctx context.Context, in AddInput) (Assumption, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Assumption{}, err
	}
	_, role, err := s.decisions.AuthorizeAccess(ctx, in.DecisionID)
	if err != nil {
		return Assumption{}, err
	}
	if role == teams.RoleViewer {
		return Assumption{}, errors.Forbidden("ASSUMPTION_FORBIDDEN", "viewers cannot add assumptions")
	}

	statement := strings.TrimSpace(in.Statement)
	if err := validate(statement, in.Confidence); err != nil {
		return Assumption{}, err
	}

	assumption := Assumption{
		ID:         id.New(),
		DecisionID: in.DecisionID,
		Statement:  statement,
		Confidence: in.Confidence,
	}

	var created Assumption
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, q, assumption)
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		return s.recorder.Record(ctx, q, events.Event{
			AggregateID:   created.ID,
			AggregateType: aggregateType,
			Type:          events.TypeAssumptionAdded,
			Payload:       map[string]any{"decision_id": created.DecisionID.String(), "confidence": created.Confidence},
			ActorID:       &actor,
		})
	})
	if err != nil {
		return Assumption{}, err
	}
	return created, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (Assumption, error) {
	assumption, _, err := s.AuthorizeAccess(ctx, id)
	return assumption, err
}

func (s *service) AuthorizeAccess(ctx context.Context, id uuid.UUID) (Assumption, string, error) {
	assumption, err := s.repo.GetByID(ctx, s.pool, id)
	if err != nil {
		return Assumption{}, "", err
	}
	_, role, err := s.decisions.AuthorizeAccess(ctx, assumption.DecisionID)
	if err != nil {
		return Assumption{}, "", err
	}
	return assumption, role, nil
}

func (s *service) ListForDecision(ctx context.Context, decisionID uuid.UUID) ([]Assumption, error) {
	if _, _, err := s.decisions.AuthorizeAccess(ctx, decisionID); err != nil {
		return nil, err
	}
	return s.repo.ListByDecision(ctx, s.pool, decisionID)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, statement string, confidence float64) (Assumption, error) {
	if _, err := authctx.Require(ctx); err != nil {
		return Assumption{}, err
	}
	_, role, err := s.AuthorizeAccess(ctx, id)
	if err != nil {
		return Assumption{}, err
	}
	if role == teams.RoleViewer {
		return Assumption{}, errors.Forbidden("ASSUMPTION_FORBIDDEN", "viewers cannot edit assumptions")
	}

	trimmed := strings.TrimSpace(statement)
	if err := validate(trimmed, confidence); err != nil {
		return Assumption{}, err
	}

	var updated Assumption
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, q, id, trimmed, confidence)
		return txErr
	})
	if err != nil {
		return Assumption{}, err
	}
	return updated, nil
}

func (s *service) Remove(ctx context.Context, id uuid.UUID) error {
	if _, err := authctx.Require(ctx); err != nil {
		return err
	}
	_, role, err := s.AuthorizeAccess(ctx, id)
	if err != nil {
		return err
	}
	if role == teams.RoleViewer {
		return errors.Forbidden("ASSUMPTION_FORBIDDEN", "viewers cannot remove assumptions")
	}
	return s.tx.WithinTx(ctx, func(q db.Querier) error {
		return s.repo.Delete(ctx, q, id)
	})
}

// validate enforces the statement length and confidence range invariants.
func validate(statement string, confidence float64) error {
	if statement == "" {
		return errors.Validation("ASSUMPTION_STATEMENT_REQUIRED", "assumption statement is required")
	}
	if len(statement) > statementMaxLen {
		return errors.Validation("ASSUMPTION_STATEMENT_TOO_LONG", "assumption statement must be at most 2000 characters")
	}
	if confidence < 0 || confidence > 1 {
		return errors.Validation("INVALID_CONFIDENCE", "confidence must be between 0 and 1")
	}
	return nil
}
