package decisions

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/pagination"
)

// titleMaxLen bounds a decision title; descriptions are unbounded by the domain
// beyond what the database enforces.
const titleMaxLen = 200

// aggregateType labels decision events in the audit log.
const aggregateType = "decision"

// Teams is the narrow slice of the teams domain the decisions domain depends on:
// resolving the caller's membership so authorization decisions can be made.
type Teams interface {
	GetMembership(ctx context.Context, teamID, userID uuid.UUID) (teams.Membership, error)
}

// txRunner runs a unit of work inside a database transaction. It is satisfied by
// *db.TxManager in production and by a fake in unit tests.
type txRunner interface {
	WithinTx(ctx context.Context, fn func(q db.Querier) error) error
}

// recorder appends an event to the audit log using the supplied querier. It is
// satisfied by *events.Recorder in production and by a fake in unit tests.
type recorder interface {
	Record(ctx context.Context, q db.Querier, e events.Event) error
}

// CreateInput carries the fields needed to create a decision.
type CreateInput struct {
	TeamID      uuid.UUID
	Title       string
	Description string
}

// UpdateInput carries the editable fields of a decision.
type UpdateInput struct {
	Title       string
	Description string
}

// Service is the decisions domain's application boundary. It enforces
// authorization and lifecycle invariants and records audit events.
type Service interface {
	// Create records a new draft decision. The caller must be a member or admin
	// of the team (viewers cannot create). Emits DecisionCreated.
	Create(ctx context.Context, in CreateInput) (Decision, error)
	// GetByID returns a decision visible to any team member.
	GetByID(ctx context.Context, id uuid.UUID) (Decision, error)
	// AuthorizeAccess returns a decision and the caller's role, or Forbidden when
	// the caller is not a member of the decision's team. Child domains use it to
	// authorize operations on decision-scoped entities.
	AuthorizeAccess(ctx context.Context, id uuid.UUID) (Decision, string, error)
	// List returns a team's decisions as a paginated connection. The caller must
	// be a member.
	List(ctx context.Context, teamID uuid.UUID, args pagination.PageArgs) (pagination.Connection[Decision], error)
	// Update edits a decision's title and description. The caller must be the
	// owner or a team admin. Emits DecisionUpdated.
	Update(ctx context.Context, id uuid.UUID, in UpdateInput) (Decision, error)
	// Transition moves a decision to a new status, validating the transition. The
	// caller must be the owner or a team admin. Moving to decided records
	// decided_at. Emits DecisionUpdated.
	Transition(ctx context.Context, id uuid.UUID, status string) (Decision, error)
	// MarkDecided sets the decision's status to decided (decided_at=now) within the
	// caller's transaction and emits a DecisionUpdated event. Authorization is the
	// CALLER's responsibility (it accepts a db.Querier and performs no authz),
	// mirroring the Provision pattern. It is a no-op returning the current decision
	// if the status is already decided.
	MarkDecided(ctx context.Context, q db.Querier, id uuid.UUID) (Decision, error)
}

// service is the default Service implementation.
type service struct {
	pool     *pgxpool.Pool
	tx       txRunner
	repo     Repository
	recorder recorder
	teams    Teams
	clk      clock.Clock
}

// NewService wires a Service from its collaborators.
func NewService(
	pool *pgxpool.Pool,
	tx *db.TxManager,
	repo Repository,
	rec *events.Recorder,
	teamsDep Teams,
	clk clock.Clock,
) Service {
	return &service{
		pool:     pool,
		tx:       tx,
		repo:     repo,
		recorder: rec,
		teams:    teamsDep,
		clk:      clk,
	}
}

func (s *service) Create(ctx context.Context, in CreateInput) (Decision, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Decision{}, err
	}
	role, err := s.role(ctx, in.TeamID)
	if err != nil {
		return Decision{}, err
	}
	if role != teams.RoleAdmin && role != teams.RoleMember {
		return Decision{}, errors.Forbidden("DECISION_CREATE_FORBIDDEN", "viewers cannot create decisions")
	}

	title := strings.TrimSpace(in.Title)
	if err := validateTitle(title); err != nil {
		return Decision{}, err
	}

	decision := Decision{
		ID:          id.New(),
		TeamID:      in.TeamID,
		OwnerID:     principal.UserID,
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Status:      StatusDraft,
	}

	var created Decision
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, q, decision)
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		return s.recorder.Record(ctx, q, events.Event{
			AggregateID:   created.ID,
			AggregateType: aggregateType,
			Type:          events.TypeDecisionCreated,
			Payload:       map[string]any{"title": created.Title, "team_id": created.TeamID.String()},
			ActorID:       &actor,
		})
	})
	if err != nil {
		return Decision{}, err
	}
	return created, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (Decision, error) {
	decision, _, err := s.AuthorizeAccess(ctx, id)
	return decision, err
}

func (s *service) AuthorizeAccess(ctx context.Context, id uuid.UUID) (Decision, string, error) {
	decision, err := s.repo.GetByID(ctx, s.pool, id)
	if err != nil {
		return Decision{}, "", err
	}
	role, err := s.role(ctx, decision.TeamID)
	if err != nil {
		return Decision{}, "", err
	}
	return decision, role, nil
}

func (s *service) List(ctx context.Context, teamID uuid.UUID, args pagination.PageArgs) (pagination.Connection[Decision], error) {
	if _, err := s.role(ctx, teamID); err != nil {
		return pagination.Connection[Decision]{}, err
	}
	if err := args.Validate(); err != nil {
		return pagination.Connection[Decision]{}, err
	}

	limit := args.Limit()
	// Fetch one extra row so BuildConnection can detect a following page.
	fetch := limit + 1

	var (
		items []Decision
		err   error
	)
	if args.After != nil {
		createdAt, afterID, decErr := pagination.DecodeCursor(*args.After)
		if decErr != nil {
			return pagination.Connection[Decision]{}, decErr
		}
		items, err = s.repo.ListByTeamAfter(ctx, s.pool, teamID, createdAt, afterID, fetch)
	} else {
		items, err = s.repo.ListByTeam(ctx, s.pool, teamID, fetch)
	}
	if err != nil {
		return pagination.Connection[Decision]{}, err
	}

	total, err := s.repo.CountByTeam(ctx, s.pool, teamID)
	if err != nil {
		return pagination.Connection[Decision]{}, err
	}

	conn := pagination.BuildConnection(items, args, total, func(d Decision) (time.Time, uuid.UUID) {
		return d.CreatedAt, d.ID
	})
	return conn, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (Decision, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Decision{}, err
	}
	decision, role, err := s.AuthorizeAccess(ctx, id)
	if err != nil {
		return Decision{}, err
	}
	if err := requireOwnerOrAdmin(decision, principal.UserID, role); err != nil {
		return Decision{}, err
	}

	title := strings.TrimSpace(in.Title)
	if err := validateTitle(title); err != nil {
		return Decision{}, err
	}

	var updated Decision
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, q, id, title, strings.TrimSpace(in.Description))
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		return s.recorder.Record(ctx, q, events.Event{
			AggregateID:   updated.ID,
			AggregateType: aggregateType,
			Type:          events.TypeDecisionUpdated,
			Payload:       map[string]any{"title": updated.Title},
			ActorID:       &actor,
		})
	})
	if err != nil {
		return Decision{}, err
	}
	return updated, nil
}

func (s *service) Transition(ctx context.Context, id uuid.UUID, status string) (Decision, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Decision{}, err
	}
	decision, role, err := s.AuthorizeAccess(ctx, id)
	if err != nil {
		return Decision{}, err
	}
	if err := requireOwnerOrAdmin(decision, principal.UserID, role); err != nil {
		return Decision{}, err
	}
	if !canTransition(decision.Status, status) {
		return Decision{}, errors.Validation("INVALID_TRANSITION", "decision cannot move from "+decision.Status+" to "+status)
	}

	var decidedAt *time.Time
	if status == StatusDecided {
		now := s.clk.Now()
		decidedAt = &now
	}

	var updated Decision
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		updated, txErr = s.repo.UpdateStatus(ctx, q, id, status, decidedAt)
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		return s.recorder.Record(ctx, q, events.Event{
			AggregateID:   updated.ID,
			AggregateType: aggregateType,
			Type:          events.TypeDecisionUpdated,
			Payload:       map[string]any{"status": updated.Status},
			ActorID:       &actor,
		})
	})
	if err != nil {
		return Decision{}, err
	}
	return updated, nil
}

func (s *service) MarkDecided(ctx context.Context, q db.Querier, id uuid.UUID) (Decision, error) {
	decision, err := s.repo.GetByID(ctx, q, id)
	if err != nil {
		return Decision{}, err
	}
	if decision.Status == StatusDecided {
		return decision, nil
	}

	now := s.clk.Now()
	updated, err := s.repo.UpdateStatus(ctx, q, id, StatusDecided, &now)
	if err != nil {
		return Decision{}, err
	}

	var actor *uuid.UUID
	if principal, ok := authctx.From(ctx); ok {
		a := principal.UserID
		actor = &a
	}
	if err := s.recorder.Record(ctx, q, events.Event{
		AggregateID:   updated.ID,
		AggregateType: aggregateType,
		Type:          events.TypeDecisionUpdated,
		Payload:       map[string]any{"status": StatusDecided},
		ActorID:       actor,
	}); err != nil {
		return Decision{}, err
	}
	return updated, nil
}

// role resolves the caller's role within teamID, returning Forbidden when the
// caller is not a member.
func (s *service) role(ctx context.Context, teamID uuid.UUID) (string, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return "", err
	}
	m, err := s.teams.GetMembership(ctx, teamID, principal.UserID)
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

// requireOwnerOrAdmin permits the change only when the caller owns the decision
// or is a team admin.
func requireOwnerOrAdmin(d Decision, userID uuid.UUID, role string) error {
	if d.OwnerID == userID || role == teams.RoleAdmin {
		return nil
	}
	return errors.Forbidden("DECISION_FORBIDDEN", "only the owner or a team admin may modify this decision")
}

// validateTitle enforces the decision title rules.
func validateTitle(title string) error {
	if title == "" {
		return errors.Validation("DECISION_TITLE_REQUIRED", "decision title is required")
	}
	if len(title) > titleMaxLen {
		return errors.Validation("DECISION_TITLE_TOO_LONG", "decision title must be at most 200 characters")
	}
	return nil
}
