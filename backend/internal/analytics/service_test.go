package analytics

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/authctx"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// fakeTeams is an in-memory Teams dependency. A user absent from members is
// treated as a non-member and reported as Forbidden, matching the real domain.
type fakeTeams struct {
	members map[uuid.UUID]bool
}

func (f *fakeTeams) GetMembership(_ context.Context, teamID, userID uuid.UUID) (teams.Membership, error) {
	if !f.members[userID] {
		return teams.Membership{}, errors.Forbidden("NOT_TEAM_MEMBER", "caller is not a member of the team")
	}
	return teams.Membership{TeamID: teamID, UserID: userID, Role: teams.RoleMember}, nil
}

// fakeRepo is an in-memory Repository returning canned aggregates.
type fakeRepo struct {
	brier     float64
	forecasts int
	rate      float64
	resolved  int
	bins      []CalibrationBin
}

func (f *fakeRepo) ForecastMetrics(_ context.Context, _ db.Querier, _ uuid.UUID) (float64, int, error) {
	return f.brier, f.forecasts, nil
}

func (f *fakeRepo) DecisionSuccessRate(_ context.Context, _ db.Querier, _ uuid.UUID) (float64, int, error) {
	return f.rate, f.resolved, nil
}

func (f *fakeRepo) Calibration(_ context.Context, _ db.Querier, _ uuid.UUID) ([]CalibrationBin, error) {
	return f.bins, nil
}

func newTestService(repo Repository, tm Teams) *service {
	return &service{pool: nil, repo: repo, teams: tm}
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return authctx.WithPrincipal(context.Background(), authctx.Principal{UserID: userID})
}

func TestTeamMetricsRequiresAuthentication(t *testing.T) {
	s := newTestService(&fakeRepo{}, &fakeTeams{members: map[uuid.UUID]bool{}})

	_, err := s.TeamMetrics(context.Background(), uuid.New())
	if errors.KindOf(err) != errors.KindUnauthenticated {
		t.Errorf("err = %v, want Unauthenticated", err)
	}
}

func TestTeamMetricsForbiddenForNonMember(t *testing.T) {
	stranger := uuid.New()
	s := newTestService(&fakeRepo{}, &fakeTeams{members: map[uuid.UUID]bool{}})

	_, err := s.TeamMetrics(ctxWithUser(stranger), uuid.New())
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "NOT_TEAM_MEMBER" {
		t.Errorf("err = %v, want Forbidden/NOT_TEAM_MEMBER", err)
	}
}

func TestTeamMetricsCombinesRepoResultsForMember(t *testing.T) {
	member := uuid.New()
	teamID := uuid.New()
	repo := &fakeRepo{brier: 0.03, forecasts: 3, rate: 0.6667, resolved: 3}
	s := newTestService(repo, &fakeTeams{members: map[uuid.UUID]bool{member: true}})

	got, err := s.TeamMetrics(ctxWithUser(member), teamID)
	if err != nil {
		t.Fatalf("TeamMetrics: %v", err)
	}
	want := TeamMetrics{
		BrierScore:            0.03,
		ForecastCount:         3,
		DecisionSuccessRate:   0.6667,
		ResolvedDecisionCount: 3,
	}
	if got != want {
		t.Errorf("TeamMetrics = %+v, want %+v", got, want)
	}
}

func TestCalibrationForbiddenForNonMember(t *testing.T) {
	stranger := uuid.New()
	s := newTestService(&fakeRepo{}, &fakeTeams{members: map[uuid.UUID]bool{}})

	_, err := s.Calibration(ctxWithUser(stranger), uuid.New())
	if errors.KindOf(err) != errors.KindForbidden || errors.CodeOf(err) != "NOT_TEAM_MEMBER" {
		t.Errorf("err = %v, want Forbidden/NOT_TEAM_MEMBER", err)
	}
}

func TestCalibrationReturnsBinsForMember(t *testing.T) {
	member := uuid.New()
	teamID := uuid.New()
	bins := []CalibrationBin{
		{Bucket: 3, MeanPredicted: 0.2, ObservedFrequency: 0, SampleSize: 1},
		{Bucket: 9, MeanPredicted: 0.85, ObservedFrequency: 1, SampleSize: 2},
	}
	repo := &fakeRepo{bins: bins}
	s := newTestService(repo, &fakeTeams{members: map[uuid.UUID]bool{member: true}})

	got, err := s.Calibration(ctxWithUser(member), teamID)
	if err != nil {
		t.Fatalf("Calibration: %v", err)
	}
	if len(got.Bins) != len(bins) {
		t.Fatalf("bins len = %d, want %d", len(got.Bins), len(bins))
	}
	for i := range bins {
		if got.Bins[i] != bins[i] {
			t.Errorf("bin %d = %+v, want %+v", i, got.Bins[i], bins[i])
		}
	}
}
