package decisions

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the persistence boundary for decisions. Every method takes a
// db.Querier so callers choose whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	Create(ctx context.Context, q db.Querier, d Decision) (Decision, error)
	GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Decision, error)
	ListByTeam(ctx context.Context, q db.Querier, teamID uuid.UUID, limit int) ([]Decision, error)
	ListByTeamAfter(ctx context.Context, q db.Querier, teamID uuid.UUID, afterCreatedAt time.Time, afterID uuid.UUID, limit int) ([]Decision, error)
	CountByTeam(ctx context.Context, q db.Querier, teamID uuid.UUID) (int, error)
	Update(ctx context.Context, q db.Querier, id uuid.UUID, title, description string) (Decision, error)
	UpdateStatus(ctx context.Context, q db.Querier, id uuid.UUID, status string, decidedAt *time.Time) (Decision, error)
	ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Decision, error)
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) Create(ctx context.Context, q db.Querier, d Decision) (Decision, error) {
	row, err := sqlc.New(q).CreateDecision(ctx, sqlc.CreateDecisionParams{
		ID:          d.ID,
		TeamID:      d.TeamID,
		OwnerID:     d.OwnerID,
		Title:       d.Title,
		Description: d.Description,
		Status:      d.Status,
	})
	if err != nil {
		return Decision{}, errors.Wrap(err, errors.KindInternal, "DECISION_CREATE_FAILED", "failed to create decision")
	}
	return toDecision(row), nil
}

func (repository) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Decision, error) {
	row, err := sqlc.New(q).GetDecisionByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Decision{}, errors.NotFound("DECISION_NOT_FOUND", "decision not found")
		}
		return Decision{}, errors.Wrap(err, errors.KindInternal, "DECISION_GET_FAILED", "failed to load decision")
	}
	return toDecision(row), nil
}

func (repository) ListByTeam(ctx context.Context, q db.Querier, teamID uuid.UUID, limit int) ([]Decision, error) {
	rows, err := sqlc.New(q).ListDecisionsByTeam(ctx, sqlc.ListDecisionsByTeamParams{
		TeamID: teamID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "DECISION_LIST_FAILED", "failed to list decisions")
	}
	return toDecisions(rows), nil
}

func (repository) ListByTeamAfter(ctx context.Context, q db.Querier, teamID uuid.UUID, afterCreatedAt time.Time, afterID uuid.UUID, limit int) ([]Decision, error) {
	rows, err := sqlc.New(q).ListDecisionsByTeamAfter(ctx, sqlc.ListDecisionsByTeamAfterParams{
		TeamID:    teamID,
		CreatedAt: afterCreatedAt,
		ID:        afterID,
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "DECISION_LIST_FAILED", "failed to list decisions")
	}
	return toDecisions(rows), nil
}

func (repository) CountByTeam(ctx context.Context, q db.Querier, teamID uuid.UUID) (int, error) {
	n, err := sqlc.New(q).CountDecisionsByTeam(ctx, teamID)
	if err != nil {
		return 0, errors.Wrap(err, errors.KindInternal, "DECISION_COUNT_FAILED", "failed to count decisions")
	}
	return int(n), nil
}

func (repository) Update(ctx context.Context, q db.Querier, id uuid.UUID, title, description string) (Decision, error) {
	row, err := sqlc.New(q).UpdateDecision(ctx, sqlc.UpdateDecisionParams{
		ID:          id,
		Title:       title,
		Description: description,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Decision{}, errors.NotFound("DECISION_NOT_FOUND", "decision not found")
		}
		return Decision{}, errors.Wrap(err, errors.KindInternal, "DECISION_UPDATE_FAILED", "failed to update decision")
	}
	return toDecision(row), nil
}

func (repository) UpdateStatus(ctx context.Context, q db.Querier, id uuid.UUID, status string, decidedAt *time.Time) (Decision, error) {
	var decided pgtype.Timestamptz
	if decidedAt != nil {
		decided = pgtype.Timestamptz{Time: *decidedAt, Valid: true}
	}
	row, err := sqlc.New(q).UpdateDecisionStatus(ctx, sqlc.UpdateDecisionStatusParams{
		ID:        id,
		Status:    status,
		DecidedAt: decided,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Decision{}, errors.NotFound("DECISION_NOT_FOUND", "decision not found")
		}
		return Decision{}, errors.Wrap(err, errors.KindInternal, "DECISION_STATUS_FAILED", "failed to update decision status")
	}
	return toDecision(row), nil
}

func (repository) ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Decision, error) {
	rows, err := sqlc.New(q).ListDecisionsByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "DECISION_LIST_FAILED", "failed to list decisions")
	}
	return toDecisions(rows), nil
}

// toDecision maps a generated sqlc row to the domain entity.
func toDecision(r sqlc.Decision) Decision {
	var decidedAt *time.Time
	if r.DecidedAt.Valid {
		t := r.DecidedAt.Time
		decidedAt = &t
	}
	return Decision{
		ID:          r.ID,
		TeamID:      r.TeamID,
		OwnerID:     r.OwnerID,
		Title:       r.Title,
		Description: r.Description,
		Status:      r.Status,
		DecidedAt:   decidedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// toDecisions maps a slice of generated rows to domain entities.
func toDecisions(rows []sqlc.Decision) []Decision {
	out := make([]Decision, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDecision(r))
	}
	return out
}
