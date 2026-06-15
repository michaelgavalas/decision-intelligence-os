package evidence

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/assumptions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// aggregateType labels evidence events in the audit log.
const aggregateType = "evidence"

// Assumptions is the narrow slice of the assumptions domain the evidence domain
// depends on: authorizing access to an assumption and learning the caller's role.
type Assumptions interface {
	AuthorizeAccess(ctx context.Context, id uuid.UUID) (assumptions.Assumption, string, error)
}

// txRunner runs a unit of work inside a database transaction.
type txRunner interface {
	WithinTx(ctx context.Context, fn func(q db.Querier) error) error
}

// recorder appends an event to the audit log using the supplied querier.
type recorder interface {
	Record(ctx context.Context, q db.Querier, e events.Event) error
}

// AttachInput carries the fields needed to attach evidence to an assumption.
type AttachInput struct {
	AssumptionID uuid.UUID
	SourceType   string
	SourceURL    *string
	Content      string
}

// Service is the evidence domain's application boundary. It authorizes every
// operation through the parent assumption and records audit events.
type Service interface {
	// Attach records new evidence on an assumption. The caller must be a
	// non-viewer member of the assumption's team. Emits EvidenceAttached.
	Attach(ctx context.Context, in AttachInput) (Evidence, error)
	// GetByID returns evidence visible to any member of its team.
	GetByID(ctx context.Context, id uuid.UUID) (Evidence, error)
	// ListForAssumption returns an assumption's evidence for any team member.
	ListForAssumption(ctx context.Context, assumptionID uuid.UUID) ([]Evidence, error)
	// Update edits evidence. The caller must be a non-viewer member.
	Update(ctx context.Context, id uuid.UUID, sourceType string, sourceURL *string, content string) (Evidence, error)
	// Remove deletes evidence. The caller must be a non-viewer member.
	Remove(ctx context.Context, id uuid.UUID) error
}

// service is the default Service implementation.
type service struct {
	pool        *pgxpool.Pool
	tx          txRunner
	repo        Repository
	recorder    recorder
	assumptions Assumptions
}

// NewService wires a Service from its collaborators.
func NewService(
	pool *pgxpool.Pool,
	tx *db.TxManager,
	repo Repository,
	rec *events.Recorder,
	assumptionsDep Assumptions,
) Service {
	return &service{
		pool:        pool,
		tx:          tx,
		repo:        repo,
		recorder:    rec,
		assumptions: assumptionsDep,
	}
}

func (s *service) Attach(ctx context.Context, in AttachInput) (Evidence, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Evidence{}, err
	}
	_, role, err := s.assumptions.AuthorizeAccess(ctx, in.AssumptionID)
	if err != nil {
		return Evidence{}, err
	}
	if role == teams.RoleViewer {
		return Evidence{}, errors.Forbidden("EVIDENCE_FORBIDDEN", "viewers cannot attach evidence")
	}
	if !ValidSourceType(in.SourceType) {
		return Evidence{}, errors.Validation("INVALID_SOURCE_TYPE", "source type must be url, document, note, or dataset")
	}

	ev := Evidence{
		ID:           id.New(),
		AssumptionID: in.AssumptionID,
		SourceType:   in.SourceType,
		SourceURL:    trimURL(in.SourceURL),
		Content:      strings.TrimSpace(in.Content),
	}

	var created Evidence
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, q, ev)
		if txErr != nil {
			return txErr
		}
		actor := principal.UserID
		return s.recorder.Record(ctx, q, events.Event{
			AggregateID:   created.ID,
			AggregateType: aggregateType,
			Type:          events.TypeEvidenceAttached,
			Payload:       map[string]any{"assumption_id": created.AssumptionID.String(), "source_type": created.SourceType},
			ActorID:       &actor,
		})
	})
	if err != nil {
		return Evidence{}, err
	}
	return created, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (Evidence, error) {
	ev, _, err := s.authorize(ctx, id)
	return ev, err
}

func (s *service) ListForAssumption(ctx context.Context, assumptionID uuid.UUID) ([]Evidence, error) {
	if _, _, err := s.assumptions.AuthorizeAccess(ctx, assumptionID); err != nil {
		return nil, err
	}
	return s.repo.ListByAssumption(ctx, s.pool, assumptionID)
}

func (s *service) Update(ctx context.Context, id uuid.UUID, sourceType string, sourceURL *string, content string) (Evidence, error) {
	if _, err := authctx.Require(ctx); err != nil {
		return Evidence{}, err
	}
	_, role, err := s.authorize(ctx, id)
	if err != nil {
		return Evidence{}, err
	}
	if role == teams.RoleViewer {
		return Evidence{}, errors.Forbidden("EVIDENCE_FORBIDDEN", "viewers cannot edit evidence")
	}
	if !ValidSourceType(sourceType) {
		return Evidence{}, errors.Validation("INVALID_SOURCE_TYPE", "source type must be url, document, note, or dataset")
	}

	var updated Evidence
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, q, id, sourceType, trimURL(sourceURL), strings.TrimSpace(content))
		return txErr
	})
	if err != nil {
		return Evidence{}, err
	}
	return updated, nil
}

func (s *service) Remove(ctx context.Context, id uuid.UUID) error {
	if _, err := authctx.Require(ctx); err != nil {
		return err
	}
	_, role, err := s.authorize(ctx, id)
	if err != nil {
		return err
	}
	if role == teams.RoleViewer {
		return errors.Forbidden("EVIDENCE_FORBIDDEN", "viewers cannot remove evidence")
	}
	return s.tx.WithinTx(ctx, func(q db.Querier) error {
		return s.repo.Delete(ctx, q, id)
	})
}

// authorize loads evidence and resolves the caller's role through the parent
// assumption, returning Forbidden when the caller is not a member.
func (s *service) authorize(ctx context.Context, id uuid.UUID) (Evidence, string, error) {
	ev, err := s.repo.GetByID(ctx, s.pool, id)
	if err != nil {
		return Evidence{}, "", err
	}
	_, role, err := s.assumptions.AuthorizeAccess(ctx, ev.AssumptionID)
	if err != nil {
		return Evidence{}, "", err
	}
	return ev, role, nil
}

// trimURL trims surrounding whitespace from an optional source URL, leaving nil
// untouched and treating an empty result as absent.
func trimURL(u *string) *string {
	if u == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*u)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
