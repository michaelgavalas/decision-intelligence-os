-- name: UpsertOutcome :one
INSERT INTO outcomes (id, decision_id, summary, success, resolved_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (decision_id) DO UPDATE
SET summary = EXCLUDED.summary,
    success = EXCLUDED.success,
    resolved_at = EXCLUDED.resolved_at,
    updated_at = now()
RETURNING *;

-- name: GetOutcomeByDecision :one
SELECT * FROM outcomes
WHERE decision_id = $1;

-- name: ListOutcomesByDecisionIDs :many
SELECT * FROM outcomes
WHERE decision_id = ANY(sqlc.arg(ids)::uuid[]);

-- name: ListOutcomesByIDs :many
SELECT * FROM outcomes
WHERE id = ANY(sqlc.arg(ids)::uuid[]);
