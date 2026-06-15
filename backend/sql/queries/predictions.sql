-- name: CreatePrediction :one
INSERT INTO predictions (id, decision_id, statement, probability, resolves_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPredictionByID :one
SELECT * FROM predictions
WHERE id = $1;

-- name: ListPredictionsByDecision :many
SELECT * FROM predictions
WHERE decision_id = $1
ORDER BY created_at DESC, id DESC;

-- name: ListPredictionsByDecisionIDs :many
SELECT * FROM predictions
WHERE decision_id = ANY(sqlc.arg(ids)::uuid[])
ORDER BY decision_id, created_at DESC;

-- name: UpdatePrediction :one
UPDATE predictions
SET statement = $2,
    probability = $3,
    resolves_at = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListPredictionsByIDs :many
SELECT * FROM predictions
WHERE id = ANY(sqlc.arg(ids)::uuid[]);
