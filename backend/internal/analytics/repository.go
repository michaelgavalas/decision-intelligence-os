package analytics

import (
	"context"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Repository is the read-only persistence boundary for analytics. Every method
// takes a db.Querier so callers choose whether the aggregate runs against the
// pool or joins an in-flight transaction. All metrics are computed in SQL from
// the source tables; the repository owns no denormalized state.
type Repository interface {
	// ForecastMetrics returns the team's Brier score and the number of resolved
	// forecasts it covers.
	ForecastMetrics(ctx context.Context, q db.Querier, teamID uuid.UUID) (brier float64, count int, err error)
	// DecisionSuccessRate returns the team's decision success rate and the number
	// of resolved decisions.
	DecisionSuccessRate(ctx context.Context, q db.Querier, teamID uuid.UUID) (rate float64, resolved int, err error)
	// Calibration returns the team's calibration bins, ordered by bucket. The
	// result is empty (not an error) when the team has no resolved forecasts.
	Calibration(ctx context.Context, q db.Querier, teamID uuid.UUID) ([]CalibrationBin, error)
}

// repository is the sqlc-backed Repository implementation.
type repository struct{}

// NewRepository returns the default Repository.
func NewRepository() Repository {
	return repository{}
}

func (repository) ForecastMetrics(ctx context.Context, q db.Querier, teamID uuid.UUID) (float64, int, error) {
	row, err := sqlc.New(q).TeamForecastMetrics(ctx, teamID)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.KindInternal, "ANALYTICS_FORECAST_FAILED", "failed to compute forecast metrics")
	}
	return row.BrierScore, int(row.ForecastCount), nil
}

func (repository) DecisionSuccessRate(ctx context.Context, q db.Querier, teamID uuid.UUID) (float64, int, error) {
	row, err := sqlc.New(q).TeamDecisionSuccessRate(ctx, teamID)
	if err != nil {
		return 0, 0, errors.Wrap(err, errors.KindInternal, "ANALYTICS_SUCCESS_RATE_FAILED", "failed to compute decision success rate")
	}
	return row.SuccessRate, int(row.ResolvedCount), nil
}

func (repository) Calibration(ctx context.Context, q db.Querier, teamID uuid.UUID) ([]CalibrationBin, error) {
	rows, err := sqlc.New(q).TeamCalibration(ctx, teamID)
	if err != nil {
		return nil, errors.Wrap(err, errors.KindInternal, "ANALYTICS_CALIBRATION_FAILED", "failed to compute calibration")
	}
	bins := make([]CalibrationBin, 0, len(rows))
	for _, r := range rows {
		bins = append(bins, CalibrationBin{
			Bucket:            int(r.Bucket),
			MeanPredicted:     r.MeanPredicted,
			ObservedFrequency: r.ObservedFrequency,
			SampleSize:        int(r.SampleSize),
		})
	}
	return bins, nil
}
