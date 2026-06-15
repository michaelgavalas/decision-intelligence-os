package analytics

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Teams is the narrow slice of the teams domain the analytics domain depends on:
// resolving the caller's membership so reads can be authorized.
type Teams interface {
	GetMembership(ctx context.Context, teamID, userID uuid.UUID) (teams.Membership, error)
}

// Service is the analytics domain's application boundary. It authorizes the
// caller against the team and returns decision-quality metrics computed from
// source data. It performs reads only, so it needs no transaction manager.
type Service interface {
	// TeamMetrics returns the team's headline decision-quality summary. The
	// caller must be a member of the team.
	TeamMetrics(ctx context.Context, teamID uuid.UUID) (TeamMetrics, error)
	// Calibration returns the team's calibration report. The caller must be a
	// member of the team.
	Calibration(ctx context.Context, teamID uuid.UUID) (CalibrationReport, error)
}

// service is the default Service implementation. It reads through the pool.
type service struct {
	pool  *pgxpool.Pool
	repo  Repository
	teams Teams
}

// NewService wires a Service from its collaborators.
func NewService(pool *pgxpool.Pool, repo Repository, teamsDep Teams) Service {
	return &service{pool: pool, repo: repo, teams: teamsDep}
}

func (s *service) TeamMetrics(ctx context.Context, teamID uuid.UUID) (TeamMetrics, error) {
	if err := s.requireMember(ctx, teamID); err != nil {
		return TeamMetrics{}, err
	}

	brier, forecasts, err := s.repo.ForecastMetrics(ctx, s.pool, teamID)
	if err != nil {
		return TeamMetrics{}, err
	}
	rate, resolved, err := s.repo.DecisionSuccessRate(ctx, s.pool, teamID)
	if err != nil {
		return TeamMetrics{}, err
	}
	return TeamMetrics{
		BrierScore:            brier,
		ForecastCount:         forecasts,
		DecisionSuccessRate:   rate,
		ResolvedDecisionCount: resolved,
	}, nil
}

func (s *service) Calibration(ctx context.Context, teamID uuid.UUID) (CalibrationReport, error) {
	if err := s.requireMember(ctx, teamID); err != nil {
		return CalibrationReport{}, err
	}

	bins, err := s.repo.Calibration(ctx, s.pool, teamID)
	if err != nil {
		return CalibrationReport{}, err
	}
	return CalibrationReport{Bins: bins}, nil
}

// requireMember confirms the authenticated caller belongs to the team. A
// non-member (whether the membership lookup reports forbidden or not found) is
// surfaced uniformly as Forbidden so analytics never leaks team existence.
func (s *service) requireMember(ctx context.Context, teamID uuid.UUID) error {
	principal, err := authctx.Require(ctx)
	if err != nil {
		return err
	}
	if _, err := s.teams.GetMembership(ctx, teamID, principal.UserID); err != nil {
		switch errors.KindOf(err) {
		case errors.KindForbidden, errors.KindNotFound:
			return errors.Forbidden("NOT_TEAM_MEMBER", "caller is not a member of the team")
		default:
			return err
		}
	}
	return nil
}
