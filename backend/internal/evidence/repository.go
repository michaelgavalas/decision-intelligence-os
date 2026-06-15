package evidence

import (
	"context"
	stderrors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the persistence boundary for evidence. Every method takes a
// db.Querier so callers choose whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	Create(ctx context.Context, q db.Querier, e Evidence) (Evidence, error)
	GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Evidence, error)
	ListByAssumption(ctx context.Context, q db.Querier, assumptionID uuid.UUID) ([]Evidence, error)
	ListByAssumptionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Evidence, error)
	ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Evidence, error)
	Update(ctx context.Context, q db.Querier, id uuid.UUID, sourceType string, sourceURL *string, content string) (Evidence, error)
	Delete(ctx context.Context, q db.Querier, id uuid.UUID) error
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) Create(ctx context.Context, q db.Querier, e Evidence) (Evidence, error) {
	row, err := sqlc.New(q).CreateEvidence(ctx, sqlc.CreateEvidenceParams{
		ID:           e.ID,
		AssumptionID: e.AssumptionID,
		SourceType:   e.SourceType,
		SourceUrl:    e.SourceURL,
		Content:      e.Content,
	})
	if err != nil {
		return Evidence{}, errors.Wrap(err, errors.KindInternal, "EVIDENCE_CREATE_FAILED", "failed to create evidence")
	}
	return toEvidence(row), nil
}

func (repository) GetByID(ctx context.Context, q db.Querier, id uuid.UUID) (Evidence, error) {
	row, err := sqlc.New(q).GetEvidenceByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Evidence{}, errors.NotFound("EVIDENCE_NOT_FOUND", "evidence not found")
		}
		return Evidence{}, errors.Wrap(err, errors.KindInternal, "EVIDENCE_GET_FAILED", "failed to load evidence")
	}
	return toEvidence(row), nil
}

func (repository) ListByAssumption(ctx context.Context, q db.Querier, assumptionID uuid.UUID) ([]Evidence, error) {
	rows, err := sqlc.New(q).ListEvidenceByAssumption(ctx, assumptionID)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "EVIDENCE_LIST_FAILED", "failed to list evidence")
	}
	return toEvidenceList(rows), nil
}

func (repository) ListByAssumptionIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Evidence, error) {
	rows, err := sqlc.New(q).ListEvidenceByAssumptionIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "EVIDENCE_LIST_FAILED", "failed to list evidence")
	}
	return toEvidenceList(rows), nil
}

func (repository) ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Evidence, error) {
	rows, err := sqlc.New(q).ListEvidenceByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "EVIDENCE_LIST_FAILED", "failed to list evidence")
	}
	return toEvidenceList(rows), nil
}

func (repository) Update(ctx context.Context, q db.Querier, id uuid.UUID, sourceType string, sourceURL *string, content string) (Evidence, error) {
	row, err := sqlc.New(q).UpdateEvidence(ctx, sqlc.UpdateEvidenceParams{
		ID:         id,
		SourceType: sourceType,
		SourceUrl:  sourceURL,
		Content:    content,
	})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Evidence{}, errors.NotFound("EVIDENCE_NOT_FOUND", "evidence not found")
		}
		return Evidence{}, errors.Wrap(err, errors.KindInternal, "EVIDENCE_UPDATE_FAILED", "failed to update evidence")
	}
	return toEvidence(row), nil
}

func (repository) Delete(ctx context.Context, q db.Querier, id uuid.UUID) error {
	if err := sqlc.New(q).DeleteEvidence(ctx, id); err != nil {
		return errors.Wrap(err, errors.KindInternal, "EVIDENCE_DELETE_FAILED", "failed to delete evidence")
	}
	return nil
}

// toEvidence maps a generated sqlc row to the domain entity.
func toEvidence(r sqlc.Evidence) Evidence {
	return Evidence{
		ID:           r.ID,
		AssumptionID: r.AssumptionID,
		SourceType:   r.SourceType,
		SourceURL:    r.SourceUrl,
		Content:      r.Content,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

// toEvidenceList maps a slice of generated rows to domain entities.
func toEvidenceList(rows []sqlc.Evidence) []Evidence {
	out := make([]Evidence, 0, len(rows))
	for _, r := range rows {
		out = append(out, toEvidence(r))
	}
	return out
}
