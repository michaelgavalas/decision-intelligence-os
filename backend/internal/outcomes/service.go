package outcomes

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
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// summaryMaxLen bounds an outcome summary, matching the database check.
const summaryMaxLen = 2000

// aggregateType labels outcome events in the audit log.
const aggregateType = "outcome"

// Decisions is the narrow slice of the decisions domain the outcomes domain
// depends on: authorizing access and atomically marking a decision decided when
// its outcome is recorded.
type Decisions interface {
	AuthorizeAccess(ctx context.Context, id uuid.UUID) (decisions.Decision, string, error)
	MarkDecided(ctx context.Context, q db.Querier, id uuid.UUID) (decisions.Decision, error)
}

// txRunner runs a unit of work inside a database transaction.
type txRunner interface {
	WithinTx(ctx context.Context, fn func(q db.Querier) error) error
}

// recorder appends an event to the audit log using the supplied querier.
type recorder interface {
	Record(ctx context.Context, q db.Querier, e events.Event) error
}

// RecordInput carries the fields needed to record a decision's outcome.
type RecordInput struct {
	DecisionID uuid.UUID
	Summary    string
	Success    bool
	ResolvedAt *time.Time
}

// Service is the outcomes domain's application boundary. It authorizes every
// operation through the parent decision and records audit events.
type Service interface {
	// Record upserts the decision's outcome. Requires the caller to be the
	// decision OWNER or a team ADMIN. The decision must be in active or decided
	// status. Within one transaction it upserts the outcome, emits OutcomeRecorded,
	// and (if the decision is still active) marks the decision decided.
	Record(ctx context.Context, in RecordInput) (Outcome, error)
	// GetForDecision returns a decision's outcome for any team member.
	GetForDecision(ctx context.Context, decisionID uuid.UUID) (Outcome, error)
}

// service is the default Service implementation.
type service struct {
	pool      *pgxpool.Pool
	tx        txRunner
	repo      Repository
	recorder  recorder
	decisions Decisions
	clk       clock.Clock
}

// NewService wires a Service from its collaborators.
func NewService(
	pool *pgxpool.Pool,
	tx *db.TxManager,
	repo Repository,
	rec *events.Recorder,
	decisionsDep Decisions,
	clk clock.Clock,
) Service {
	return &service{
		pool:      pool,
		tx:        tx,
		repo:      repo,
		recorder:  rec,
		decisions: decisionsDep,
		clk:       clk,
	}
}

func (s *service) Record(ctx context.Context, in RecordInput) (Outcome, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Outcome{}, err
	}
	decision, role, err := s.decisions.AuthorizeAccess(ctx, in.DecisionID)
	if err != nil {
		return Outcome{}, err
	}
	if decision.OwnerID != principal.UserID && role != teams.RoleAdmin {
		return Outcome{}, errors.Forbidden("NOT_DECISION_OWNER", "only the decision owner or a team admin may record an outcome")
	}
	if decision.Status != decisions.StatusActive && decision.Status != decisions.StatusDecided {
		return Outcome{}, errors.Validation("DECISION_NOT_DECIDABLE", "an outcome can only be recorded for an active or decided decision")
	}

	summary := strings.TrimSpace(in.Summary)
	if err := validate(summary); err != nil {
		return Outcome{}, err
	}

	resolvedAt := s.clk.Now()
	if in.ResolvedAt != nil {
		resolvedAt = *in.ResolvedAt
	}

	outcome := Outcome{
		ID:         id.New(),
		DecisionID: in.DecisionID,
		Summary:    summary,
		Success:    in.Success,
		ResolvedAt: resolvedAt,
	}

	var recorded Outcome
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		recorded, txErr = s.repo.Upsert(ctx, q, outcome)
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		if txErr := s.recorder.Record(ctx, q, events.Event{
			AggregateID:   recorded.ID,
			AggregateType: aggregateType,
			Type:          events.TypeOutcomeRecorded,
			Payload:       map[string]any{"decision_id": recorded.DecisionID.String(), "success": recorded.Success},
			ActorID:       &actor,
		}); txErr != nil {
			return txErr
		}
		if decision.Status == decisions.StatusActive {
			if _, txErr := s.decisions.MarkDecided(ctx, q, in.DecisionID); txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return Outcome{}, err
	}
	return recorded, nil
}

func (s *service) GetForDecision(ctx context.Context, decisionID uuid.UUID) (Outcome, error) {
	if _, _, err := s.decisions.AuthorizeAccess(ctx, decisionID); err != nil {
		return Outcome{}, err
	}
	return s.repo.GetByDecision(ctx, s.pool, decisionID)
}

// validate enforces the summary invariants.
func validate(summary string) error {
	if summary == "" {
		return errors.Validation("OUTCOME_SUMMARY_REQUIRED", "outcome summary is required")
	}
	if len(summary) > summaryMaxLen {
		return errors.Validation("OUTCOME_SUMMARY_TOO_LONG", "outcome summary must be at most 2000 characters")
	}
	return nil
}
