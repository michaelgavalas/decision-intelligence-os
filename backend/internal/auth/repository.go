package auth

import (
	"context"
	stderrors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the persistence boundary for refresh tokens. Every method takes
// a db.Querier so callers decide whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	// Store persists a refresh token. The caller sets ID, UserID, TokenHash, and
	// ExpiresAt.
	Store(ctx context.Context, q db.Querier, t RefreshToken) (RefreshToken, error)
	// GetByHash returns the token with the given hash, or NotFound when absent.
	GetByHash(ctx context.Context, q db.Querier, hash string) (RefreshToken, error)
	// Revoke marks a token revoked by id.
	Revoke(ctx context.Context, q db.Querier, id uuid.UUID) error
	// RevokeAllForUser revokes every outstanding token for a user.
	RevokeAllForUser(ctx context.Context, q db.Querier, userID uuid.UUID) error
	// MarkReplaced revokes a token and records the token that superseded it.
	MarkReplaced(ctx context.Context, q db.Querier, id, replacedBy uuid.UUID) error
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) Store(ctx context.Context, q db.Querier, t RefreshToken) (RefreshToken, error) {
	row, err := sqlc.New(q).StoreRefreshToken(ctx, sqlc.StoreRefreshTokenParams{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
	})
	if err != nil {
		return RefreshToken{}, errors.Wrap(err, errors.KindInternal, "REFRESH_STORE_FAILED", "failed to store refresh token")
	}
	return toRefreshToken(row), nil
}

func (repository) GetByHash(ctx context.Context, q db.Querier, hash string) (RefreshToken, error) {
	row, err := sqlc.New(q).GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return RefreshToken{}, errors.NotFound("REFRESH_NOT_FOUND", "refresh token not found")
		}
		return RefreshToken{}, errors.Wrap(err, errors.KindInternal, "REFRESH_GET_FAILED", "failed to load refresh token")
	}
	return toRefreshToken(row), nil
}

func (repository) Revoke(ctx context.Context, q db.Querier, id uuid.UUID) error {
	if err := sqlc.New(q).RevokeRefreshToken(ctx, id); err != nil {
		return errors.Wrap(err, errors.KindInternal, "REFRESH_REVOKE_FAILED", "failed to revoke refresh token")
	}
	return nil
}

func (repository) RevokeAllForUser(ctx context.Context, q db.Querier, userID uuid.UUID) error {
	if err := sqlc.New(q).RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
		return errors.Wrap(err, errors.KindInternal, "REFRESH_REVOKE_ALL_FAILED", "failed to revoke refresh tokens")
	}
	return nil
}

func (repository) MarkReplaced(ctx context.Context, q db.Querier, id, replacedBy uuid.UUID) error {
	replacement := replacedBy
	err := sqlc.New(q).MarkRefreshTokenReplaced(ctx, sqlc.MarkRefreshTokenReplacedParams{
		ID:         id,
		ReplacedBy: &replacement,
	})
	if err != nil {
		return errors.Wrap(err, errors.KindInternal, "REFRESH_REPLACE_FAILED", "failed to mark refresh token replaced")
	}
	return nil
}

// toRefreshToken maps a generated sqlc row to the domain entity.
func toRefreshToken(r sqlc.RefreshToken) RefreshToken {
	t := RefreshToken{
		ID:         r.ID,
		UserID:     r.UserID,
		TokenHash:  r.TokenHash,
		ExpiresAt:  r.ExpiresAt,
		ReplacedBy: r.ReplacedBy,
		CreatedAt:  r.CreatedAt,
	}
	if r.RevokedAt.Valid {
		revoked := r.RevokedAt.Time
		t.RevokedAt = &revoked
	}
	return t
}
