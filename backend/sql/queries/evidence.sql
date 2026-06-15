-- name: CreateEvidence :one
INSERT INTO evidence (id, assumption_id, source_type, source_url, content)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetEvidenceByID :one
SELECT * FROM evidence
WHERE id = $1;

-- name: ListEvidenceByAssumption :many
SELECT * FROM evidence
WHERE assumption_id = $1
ORDER BY created_at DESC, id DESC;

-- name: ListEvidenceByAssumptionIDs :many
SELECT * FROM evidence
WHERE assumption_id = ANY(sqlc.arg(ids)::uuid[])
ORDER BY assumption_id, created_at DESC;

-- name: UpdateEvidence :one
UPDATE evidence
SET source_type = $2,
    source_url = $3,
    content = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteEvidence :exec
DELETE FROM evidence
WHERE id = $1;

-- name: ListEvidenceByIDs :many
SELECT * FROM evidence
WHERE id = ANY(sqlc.arg(ids)::uuid[]);
