package analytics_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/analytics"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/decisions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/outcomes"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/predictions"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/teams"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/users"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// tolerance bounds floating-point comparisons of computed metrics.
const tolerance = 1e-6

// seedTeam inserts a user and a team with that user as admin, returning both ids
// so decision foreign keys are satisfied.
func seedTeam(t *testing.T, pool *pgxpool.Pool) (teamID, ownerID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	u, err := users.NewRepository().Create(ctx, pool, users.CreateParams{
		ID:           id.New(),
		Email:        "owner@example.com",
		Name:         "Owner",
		PasswordHash: "hash",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	team, err := teams.NewRepository().CreateTeam(ctx, pool, id.New(), "Acme")
	if err != nil {
		t.Fatalf("seed team: %v", err)
	}
	if _, err := teams.NewRepository().AddMember(ctx, pool, team.ID, u.ID, teams.RoleAdmin); err != nil {
		t.Fatalf("seed membership: %v", err)
	}
	return team.ID, u.ID
}

// seedResolvedDecision creates a decision with a single prediction and a
// recorded outcome, the unit of the analytics fixture.
func seedResolvedDecision(t *testing.T, pool *pgxpool.Pool, teamID, ownerID uuid.UUID, probability float64, success bool) {
	t.Helper()
	ctx := context.Background()

	d, err := decisions.NewRepository().Create(ctx, pool, decisions.Decision{
		ID:      id.New(),
		TeamID:  teamID,
		OwnerID: ownerID,
		Title:   "Decision",
		Status:  decisions.StatusActive,
	})
	if err != nil {
		t.Fatalf("seed decision: %v", err)
	}
	if _, err := predictions.NewRepository().Create(ctx, pool, predictions.Prediction{
		ID:          id.New(),
		DecisionID:  d.ID,
		Statement:   "It will work out",
		Probability: probability,
	}); err != nil {
		t.Fatalf("seed prediction: %v", err)
	}
	if _, err := outcomes.NewRepository().Upsert(ctx, pool, outcomes.Outcome{
		ID:         id.New(),
		DecisionID: d.ID,
		Summary:    "Resolved",
		Success:    success,
		ResolvedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed outcome: %v", err)
	}
}

func TestAnalyticsRepositoryIntegration(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := analytics.NewRepository()
	ctx := context.Background()

	t.Run("metrics over a known fixture", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		teamID, ownerID := seedTeam(t, pool)

		// Decision A: predicted 0.9, succeeded.
		seedResolvedDecision(t, pool, teamID, ownerID, 0.9, true)
		// Decision B: predicted 0.2, failed.
		seedResolvedDecision(t, pool, teamID, ownerID, 0.2, false)
		// Decision C: predicted 0.8, succeeded.
		seedResolvedDecision(t, pool, teamID, ownerID, 0.8, true)

		// Brier = AVG((0.9-1)^2, (0.2-0)^2, (0.8-1)^2)
		//       = AVG(0.01, 0.04, 0.04) = 0.03
		brier, count, err := repo.ForecastMetrics(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("ForecastMetrics: %v", err)
		}
		if count != 3 {
			t.Errorf("forecast count = %d, want 3", count)
		}
		if math.Abs(brier-0.03) > tolerance {
			t.Errorf("brier = %v, want 0.03", brier)
		}

		// Success rate = 2 of 3 ~ 0.6667.
		rate, resolved, err := repo.DecisionSuccessRate(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("DecisionSuccessRate: %v", err)
		}
		if resolved != 3 {
			t.Errorf("resolved count = %d, want 3", resolved)
		}
		if math.Abs(rate-2.0/3.0) > tolerance {
			t.Errorf("success rate = %v, want %v", rate, 2.0/3.0)
		}

		// Calibration: 0.2 -> bucket 3, 0.8 -> bucket 9, 0.9 -> bucket 10.
		bins, err := repo.Calibration(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("Calibration: %v", err)
		}
		byBucket := map[int]analytics.CalibrationBin{}
		var totalSamples int
		for _, b := range bins {
			byBucket[b.Bucket] = b
			totalSamples += b.SampleSize
		}
		for _, bucket := range []int{3, 9, 10} {
			if _, ok := byBucket[bucket]; !ok {
				t.Errorf("missing calibration bucket %d (bins=%+v)", bucket, bins)
			}
		}
		if totalSamples != 3 {
			t.Errorf("calibration sample sizes sum = %d, want 3", totalSamples)
		}
		// The 0.9 forecast succeeded, so its bucket observes frequency 1.0.
		if b := byBucket[10]; math.Abs(b.ObservedFrequency-1.0) > tolerance {
			t.Errorf("bucket 10 observed frequency = %v, want 1.0", b.ObservedFrequency)
		}
	})

	t.Run("empty team yields zero metrics and no calibration bins", func(t *testing.T) {
		dbtest.TruncateAll(t, pool)
		teamID, _ := seedTeam(t, pool)

		brier, count, err := repo.ForecastMetrics(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("ForecastMetrics: %v", err)
		}
		if brier != 0 || count != 0 {
			t.Errorf("empty forecast metrics = (%v, %d), want (0, 0)", brier, count)
		}

		rate, resolved, err := repo.DecisionSuccessRate(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("DecisionSuccessRate: %v", err)
		}
		if rate != 0 || resolved != 0 {
			t.Errorf("empty success rate = (%v, %d), want (0, 0)", rate, resolved)
		}

		bins, err := repo.Calibration(ctx, pool, teamID)
		if err != nil {
			t.Fatalf("Calibration: %v", err)
		}
		if len(bins) != 0 {
			t.Errorf("empty calibration = %+v, want no bins", bins)
		}
	})
}
