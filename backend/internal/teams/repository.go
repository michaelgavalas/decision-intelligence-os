package teams

import (
	"context"
	stderrors "errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// pgUniqueViolation is the PostgreSQL SQLSTATE for a unique-constraint failure.
const pgUniqueViolation = "23505"

// Repository is the persistence boundary for teams and membership. Every method
// takes a db.Querier so callers choose whether the operation joins an in-flight
// transaction (pass a tx) or runs standalone (pass the pool).
type Repository interface {
	CreateTeam(ctx context.Context, q db.Querier, id uuid.UUID, name string) (Team, error)
	GetTeam(ctx context.Context, q db.Querier, id uuid.UUID) (Team, error)
	AddMember(ctx context.Context, q db.Querier, teamID, userID uuid.UUID, role string) (Membership, error)
	GetMembership(ctx context.Context, q db.Querier, teamID, userID uuid.UUID) (Membership, error)
	ListMembers(ctx context.Context, q db.Querier, teamID uuid.UUID) ([]Membership, error)
	ListTeamsForUser(ctx context.Context, q db.Querier, userID uuid.UUID) ([]Team, error)
	UpdateMemberRole(ctx context.Context, q db.Querier, teamID, userID uuid.UUID, role string) (Membership, error)
	RemoveMember(ctx context.Context, q db.Querier, teamID, userID uuid.UUID) error
	CountAdmins(ctx context.Context, q db.Querier, teamID uuid.UUID) (int, error)
	ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Team, error)
	ListMembersByTeamIDs(ctx context.Context, q db.Querier, teamIDs []uuid.UUID) ([]Membership, error)
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) CreateTeam(ctx context.Context, q db.Querier, id uuid.UUID, name string) (Team, error) {
	row, err := sqlc.New(q).CreateTeam(ctx, sqlc.CreateTeamParams{ID: id, Name: name})
	if err != nil {
		return Team{}, errors.Wrap(err, errors.KindInternal, "TEAM_CREATE_FAILED", "failed to create team")
	}
	return toTeam(row), nil
}

func (repository) GetTeam(ctx context.Context, q db.Querier, id uuid.UUID) (Team, error) {
	row, err := sqlc.New(q).GetTeamByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Team{}, errors.NotFound("TEAM_NOT_FOUND", "team not found")
		}
		return Team{}, errors.Wrap(err, errors.KindInternal, "TEAM_GET_FAILED", "failed to load team")
	}
	return toTeam(row), nil
}

func (repository) AddMember(ctx context.Context, q db.Querier, teamID, userID uuid.UUID, role string) (Membership, error) {
	row, err := sqlc.New(q).AddTeamMember(ctx, sqlc.AddTeamMemberParams{TeamID: teamID, UserID: userID, Role: role})
	if err != nil {
		if isUniqueViolation(err) {
			return Membership{}, errors.Conflict("ALREADY_MEMBER", "user is already a member of the team")
		}
		return Membership{}, errors.Wrap(err, errors.KindInternal, "MEMBER_ADD_FAILED", "failed to add team member")
	}
	return toMembership(row), nil
}

func (repository) GetMembership(ctx context.Context, q db.Querier, teamID, userID uuid.UUID) (Membership, error) {
	row, err := sqlc.New(q).GetMembership(ctx, sqlc.GetMembershipParams{TeamID: teamID, UserID: userID})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Membership{}, errors.NotFound("MEMBERSHIP_NOT_FOUND", "membership not found")
		}
		return Membership{}, errors.Wrap(err, errors.KindInternal, "MEMBERSHIP_GET_FAILED", "failed to load membership")
	}
	return toMembership(row), nil
}

func (repository) ListMembers(ctx context.Context, q db.Querier, teamID uuid.UUID) ([]Membership, error) {
	rows, err := sqlc.New(q).ListMembersByTeam(ctx, teamID)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "MEMBER_LIST_FAILED", "failed to list team members")
	}
	out := make([]Membership, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMembership(r))
	}
	return out, nil
}

func (repository) ListTeamsForUser(ctx context.Context, q db.Querier, userID uuid.UUID) ([]Team, error) {
	rows, err := sqlc.New(q).ListTeamsForUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "TEAM_LIST_FAILED", "failed to list teams")
	}
	out := make([]Team, 0, len(rows))
	for _, r := range rows {
		out = append(out, toTeam(r))
	}
	return out, nil
}

func (repository) UpdateMemberRole(ctx context.Context, q db.Querier, teamID, userID uuid.UUID, role string) (Membership, error) {
	row, err := sqlc.New(q).UpdateMemberRole(ctx, sqlc.UpdateMemberRoleParams{TeamID: teamID, UserID: userID, Role: role})
	if err != nil {
		if stderrors.Is(err, pgx.ErrNoRows) {
			return Membership{}, errors.NotFound("MEMBERSHIP_NOT_FOUND", "membership not found")
		}
		return Membership{}, errors.Wrap(err, errors.KindInternal, "MEMBER_UPDATE_FAILED", "failed to update member role")
	}
	return toMembership(row), nil
}

func (repository) RemoveMember(ctx context.Context, q db.Querier, teamID, userID uuid.UUID) error {
	if err := sqlc.New(q).RemoveTeamMember(ctx, sqlc.RemoveTeamMemberParams{TeamID: teamID, UserID: userID}); err != nil {
		return errors.Wrap(err, errors.KindInternal, "MEMBER_REMOVE_FAILED", "failed to remove team member")
	}
	return nil
}

func (repository) CountAdmins(ctx context.Context, q db.Querier, teamID uuid.UUID) (int, error) {
	n, err := sqlc.New(q).CountTeamAdmins(ctx, teamID)
	if err != nil {
		return 0, errors.Wrap(err, errors.KindInternal, "ADMIN_COUNT_FAILED", "failed to count team admins")
	}
	return int(n), nil
}

func (repository) ListByIDs(ctx context.Context, q db.Querier, ids []uuid.UUID) ([]Team, error) {
	rows, err := sqlc.New(q).ListTeamsByIDs(ctx, ids)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "TEAM_LIST_FAILED", "failed to list teams")
	}
	out := make([]Team, 0, len(rows))
	for _, r := range rows {
		out = append(out, toTeam(r))
	}
	return out, nil
}

func (repository) ListMembersByTeamIDs(ctx context.Context, q db.Querier, teamIDs []uuid.UUID) ([]Membership, error) {
	rows, err := sqlc.New(q).ListMembersByTeamIDs(ctx, teamIDs)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "MEMBER_LIST_FAILED", "failed to list team members")
	}
	out := make([]Membership, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMembership(r))
	}
	return out, nil
}

// toTeam maps a generated sqlc row to the domain entity.
func toTeam(r sqlc.Team) Team {
	return Team{ID: r.ID, Name: r.Name, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// toMembership maps a generated sqlc row to the domain entity.
func toMembership(r sqlc.TeamMember) Membership {
	return Membership{TeamID: r.TeamID, UserID: r.UserID, Role: r.Role, CreatedAt: r.CreatedAt}
}

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint
// violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return stderrors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
