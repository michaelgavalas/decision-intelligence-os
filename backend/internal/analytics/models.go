// Package analytics computes decision-quality metrics for a team directly from
// the source tables (predictions, outcomes, decisions). It owns no tables of its
// own: every metric is derived on read so the numbers can never drift from the
// facts they summarize.
package analytics

// TeamMetrics is the headline decision-quality summary for a team.
type TeamMetrics struct {
	// BrierScore is the mean squared error of the team's resolved forecasts.
	// It ranges over [0, 1]; 0 is a perfect score and lower is better.
	BrierScore float64
	// ForecastCount is the number of resolved forecasts the Brier score covers.
	ForecastCount int
	// DecisionSuccessRate is the fraction of resolved decisions whose outcome
	// succeeded, in [0, 1].
	DecisionSuccessRate float64
	// ResolvedDecisionCount is the number of decisions with a recorded outcome.
	ResolvedDecisionCount int
}

// CalibrationBin is one decile of a calibration (reliability) curve: it compares
// what the team predicted against what actually happened for forecasts whose
// probability fell in that decile.
type CalibrationBin struct {
	// Bucket is the decile index, 1..10, of predicted probability.
	Bucket int
	// MeanPredicted is the average predicted probability of forecasts in the bin.
	MeanPredicted float64
	// ObservedFrequency is the fraction of those forecasts that actually
	// succeeded, in [0, 1].
	ObservedFrequency float64
	// SampleSize is the number of forecasts in the bin.
	SampleSize int
}

// CalibrationReport is the full set of calibration bins for a team, ordered by
// bucket. It is empty when the team has no resolved forecasts.
type CalibrationReport struct {
	Bins []CalibrationBin
}
