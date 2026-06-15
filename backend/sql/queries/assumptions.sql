-- name: CreateAssumption :one
INSERT INTO assumptions (id, decision_id, statement, confidence)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAssumptionByID :one
SELECT * FROM assumptions
WHERE id = $1;

-- name: ListAssumptionsByDecision :many
SELECT * FROM assumptions
WHERE decision_id = $1
ORDER BY created_at DESC, id DESC;

-- name: ListAssumptionsByDecisionIDs :many
SELECT * FROM assumptions
WHERE decision_id = ANY(sqlc.arg(ids)::uuid[])
ORDER BY decision_id, created_at DESC;

-- name: UpdateAssumption :one
UPDATE assumptions
SET statement = $2,
    confidence = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteAssumption :exec
DELETE FROM assumptions
WHERE id = $1;

-- name: ListAssumptionsByIDs :many
SELECT * FROM assumptions
WHERE id = ANY(sqlc.arg(ids)::uuid[]);
