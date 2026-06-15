-- name: TeamForecastMetrics :one
-- Brier score across a team's resolved forecasts: the mean squared error
-- between each prediction's probability and the realized outcome (1 on success,
-- 0 otherwise). Lower is better; an empty team scores 0. A prediction is
-- resolved when its decision has an outcome.
SELECT
  COALESCE(AVG(POWER(p.probability - (CASE WHEN o.success THEN 1 ELSE 0 END), 2)), 0)::float8 AS brier_score,
  COUNT(*) AS forecast_count
FROM predictions p
JOIN decisions d ON d.id = p.decision_id
JOIN outcomes o ON o.decision_id = d.id
WHERE d.team_id = $1;

-- name: TeamDecisionSuccessRate :one
-- Fraction of a team's resolved decisions whose outcome succeeded, plus the
-- number of resolved decisions. An empty team scores 0.
SELECT
  COALESCE(AVG(CASE WHEN o.success THEN 1.0 ELSE 0.0 END), 0)::float8 AS success_rate,
  COUNT(*) AS resolved_count
FROM decisions d
JOIN outcomes o ON o.decision_id = d.id
WHERE d.team_id = $1;

-- name: TeamCalibration :many
-- Reliability bins for a team's resolved forecasts: predictions grouped into
-- deciles of predicted probability, with the mean predicted probability and the
-- observed success frequency in each bin. A well-calibrated team has observed
-- frequency tracking mean predicted probability across bins.
SELECT
  LEAST(width_bucket(p.probability, 0, 1, 10), 10)::int AS bucket,
  AVG(p.probability)::float8 AS mean_predicted,
  AVG(CASE WHEN o.success THEN 1.0 ELSE 0.0 END)::float8 AS observed_frequency,
  COUNT(*) AS sample_size
FROM predictions p
JOIN decisions d ON d.id = p.decision_id
JOIN outcomes o ON o.decision_id = d.id
WHERE d.team_id = $1
GROUP BY bucket
ORDER BY bucket;
