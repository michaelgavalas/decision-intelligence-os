package teams

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

// teamNameMaxLen bounds a team's name.
const teamNameMaxLen = 200

// Service is the team domain's application boundary. It enforces authorization
// and membership invariants and delegates persistence to a Repository.
type Service interface {
	// ProvisionPersonalTeam creates a team and makes ownerID its admin, all
	// within the caller-supplied querier so it can join a registration
	// transaction. It performs no authorization.
	ProvisionPersonalTeam(ctx context.Context, q db.Querier, ownerID uuid.UUID, name string) (Team, error)
	// CreateTeam creates a team in its own transaction and makes the caller its
	// admin. It requires an authenticated caller.
	CreateTeam(ctx context.Context, name string) (Team, error)
	// GetMembership returns a membership. The caller must be a member of the
	// team.
	GetMembership(ctx context.Context, teamID, userID uuid.UUID) (Membership, error)
	// ListMembers returns a team's members. The caller must be a member.
	ListMembers(ctx context.Context, teamID uuid.UUID) ([]Membership, error)
	// ListMyTeams returns the caller's teams.
	ListMyTeams(ctx context.Context) ([]Team, error)
	// AddMember adds a user to a team with the given role. The caller must be an
	// admin and the role must be valid.
	AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) (Membership, error)
	// ChangeRole updates a member's role. The caller must be an admin. Demoting
	// the last admin is rejected.
	ChangeRole(ctx context.Context, teamID, userID uuid.UUID, role string) (Membership, error)
	// RemoveMember removes a user from a team. The caller must be an admin.
	// Removing the last admin is rejected.
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
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

func (s *service) ProvisionPersonalTeam(ctx context.Context, q db.Querier, ownerID uuid.UUID, name string) (Team, error) {
	if err := validateTeamName(name); err != nil {
		return Team{}, err
	}

	team, err := s.repo.CreateTeam(ctx, q, id.New(), strings.TrimSpace(name))
	if err != nil {
		return Team{}, err
	}
	if _, err := s.repo.AddMember(ctx, q, team.ID, ownerID, RoleAdmin); err != nil {
		return Team{}, err
	}
	return team, nil
}

func (s *service) CreateTeam(ctx context.Context, name string) (Team, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Team{}, err
	}
	if err := validateTeamName(name); err != nil {
		return Team{}, err
	}

	var team Team
	err = s.tx.WithinTx(ctx, func(q db.Querier) error {
		var txErr error
		team, txErr = s.ProvisionPersonalTeam(ctx, q, principal.UserID, name)
		return txErr
	})
	if err != nil {
		return Team{}, err
	}
	return team, nil
}

func (s *service) GetMembership(ctx context.Context, teamID, userID uuid.UUID) (Membership, error) {
	if _, err := s.requireMember(ctx, teamID); err != nil {
		return Membership{}, err
	}
	return s.repo.GetMembership(ctx, s.pool, teamID, userID)
}

func (s *service) ListMembers(ctx context.Context, teamID uuid.UUID) ([]Membership, error) {
	if _, err := s.requireMember(ctx, teamID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, s.pool, teamID)
}

func (s *service) ListMyTeams(ctx context.Context) ([]Team, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListTeamsForUser(ctx, s.pool, principal.UserID)
}

func (s *service) AddMember(ctx context.Context, teamID, userID uuid.UUID, role string) (Membership, error) {
	if err := s.requireAdmin(ctx, teamID); err != nil {
		return Membership{}, err
	}
	if !ValidRole(role) {
		return Membership{}, errors.Validation("INVALID_ROLE", "role must be admin, member, or viewer")
	}
	return s.repo.AddMember(ctx, s.pool, teamID, userID, role)
}

func (s *service) ChangeRole(ctx context.Context, teamID, userID uuid.UUID, role string) (Membership, error) {
	if err := s.requireAdmin(ctx, teamID); err != nil {
		return Membership{}, err
	}
	if !ValidRole(role) {
		return Membership{}, errors.Validation("INVALID_ROLE", "role must be admin, member, or viewer")
	}

	// Guard the last admin: demoting the only admin would leave the team
	// unmanageable.
	if role != RoleAdmin {
		if err := s.guardLastAdmin(ctx, teamID, userID); err != nil {
			return Membership{}, err
		}
	}
	return s.repo.UpdateMemberRole(ctx, s.pool, teamID, userID, role)
}

func (s *service) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	if err := s.requireAdmin(ctx, teamID); err != nil {
		return err
	}
	if err := s.guardLastAdmin(ctx, teamID, userID); err != nil {
		return err
	}
	return s.repo.RemoveMember(ctx, s.pool, teamID, userID)
}

// requireMember resolves the caller and confirms they belong to the team,
// returning their membership. A non-member is forbidden.
func (s *service) requireMember(ctx context.Context, teamID uuid.UUID) (Membership, error) {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return Membership{}, err
	}
	m, err := s.repo.GetMembership(ctx, s.pool, teamID, principal.UserID)
	if err != nil {
		if errors.KindOf(err) == errors.KindNotFound {
			return Membership{}, errors.Forbidden("NOT_TEAM_MEMBER", "caller is not a member of the team")
		}
		return Membership{}, err
	}
	return m, nil
}

// requireAdmin resolves the caller and confirms they are an admin of the team.
func (s *service) requireAdmin(ctx context.Context, teamID uuid.UUID) error {
	m, err := s.requireMember(ctx, teamID)
	if err != nil {
		return err
	}
	if m.Role != RoleAdmin {
		return errors.Forbidden("NOT_TEAM_ADMIN", "caller is not a team admin")
	}
	return nil
}

// guardLastAdmin rejects demoting or removing the team's only admin. It is a
// no-op when the target is not an admin or other admins remain.
func (s *service) guardLastAdmin(ctx context.Context, teamID, userID uuid.UUID) error {
	target, err := s.repo.GetMembership(ctx, s.pool, teamID, userID)
	if err != nil {
		return err
	}
	if target.Role != RoleAdmin {
		return nil
	}
	admins, err := s.repo.CountAdmins(ctx, s.pool, teamID)
	if err != nil {
		return err
	}
	if admins <= 1 {
		return errors.Conflict("LAST_ADMIN", "team must retain at least one admin")
	}
	return nil
}

// validateTeamName enforces the shared team-name rules.
func validateTeamName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return errors.Validation("TEAM_NAME_REQUIRED", "team name is required")
	}
	if len(trimmed) > teamNameMaxLen {
		return errors.Validation("TEAM_NAME_TOO_LONG", "team name must be at most 200 characters")
	}
	return nil
}
