package users

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/clock"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// nameMaxLen bounds a user's display name.
const nameMaxLen = 200

// ProvisionParams carries the fields needed to create a user during
// registration.
type ProvisionParams struct {
	Email        string
	Name         string
	PasswordHash string
}

// Service is the user domain's application boundary. It enforces authorization
// and business rules and delegates persistence to a Repository.
type Service interface {
	// Provision creates a user using the caller-supplied querier so it can join
	// an in-flight registration transaction. It performs no authorization and is
	// intended for the auth domain only.
	Provision(ctx context.Context, q db.Querier, p ProvisionParams) (User, error)
	// GetByID returns a user by id. It requires an authenticated caller.
	GetByID(ctx context.Context, userID uuid.UUID) (User, error)
	// FindByEmail returns a user by email for internal use by the auth domain.
	// It performs no authorization.
	FindByEmail(ctx context.Context, email string) (User, error)
	// UpdateProfile updates the caller's own display name. It is forbidden to
	// update another user's profile.
	UpdateProfile(ctx context.Context, userID uuid.UUID, name string) (User, error)
}

// service is the default Service implementation. It reads through the pool and
// performs multi-step writes through the transaction manager.
type service struct {
	pool *pgxpool.Pool
	tx   *db.TxManager
	repo Repository
	clk  clock.Clock
}

// NewService wires a Service from its collaborators.
func NewService(pool *pgxpool.Pool, tx *db.TxManager, repo Repository, clk clock.Clock) Service {
	return &service{pool: pool, tx: tx, repo: repo, clk: clk}
}

func (s *service) Provision(ctx context.Context, q db.Querier, p ProvisionParams) (User, error) {
	email := strings.TrimSpace(p.Email)
	if email == "" {
		return User{}, errors.Validation("EMAIL_REQUIRED", "email is required")
	}
	if err := validateName(p.Name); err != nil {
		return User{}, err
	}

	return s.repo.Create(ctx, q, CreateParams{
		ID:           id.New(),
		Email:        email,
		Name:         strings.TrimSpace(p.Name),
		PasswordHash: p.PasswordHash,
	})
}

func (s *service) GetByID(ctx context.Context, userID uuid.UUID) (User, error) {
	if _, err := authctx.Require(ctx); err != nil {
		return User{}, err
	}
	return s.repo.GetByID(ctx, s.pool, userID)
}

func (s *service) FindByEmail(ctx context.Context, email string) (User, error) {
	return s.repo.GetByEmail(ctx, s.pool, strings.TrimSpace(email))
}

func (s *service) UpdateProfile(ctx context.Context, userID uuid.UUID, name string) (User, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return User{}, err
	}
	if principal.UserID != userID {
		return User{}, errors.Forbidden("FORBIDDEN", "cannot update another user's profile")
	}
	if err := validateName(name); err != nil {
		return User{}, err
	}
	return s.repo.UpdateName(ctx, s.pool, userID, strings.TrimSpace(name))
}

// validateName enforces the shared display-name rules.
func validateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return errors.Validation("NAME_REQUIRED", "name is required")
	}
	if len(trimmed) > nameMaxLen {
		return errors.Validation("NAME_TOO_LONG", "name must be at most 200 characters")
	}
	return nil
}
