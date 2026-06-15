-- name: CreateDecision :one
INSERT INTO decisions (id, team_id, owner_id, title, description, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDecisionByID :one
SELECT * FROM decisions
WHERE id = $1;

-- name: ListDecisionsByTeam :many
SELECT * FROM decisions
WHERE team_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListDecisionsByTeamAfter :many
SELECT * FROM decisions
WHERE team_id = $1
  AND (created_at < $2 OR (created_at = $2 AND id < $3))
ORDER BY created_at DESC, id DESC
LIMIT $4;

-- name: CountDecisionsByTeam :one
SELECT count(*) FROM decisions
WHERE team_id = $1;

-- name: UpdateDecision :one
UPDATE decisions
SET title = $2,
    description = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateDecisionStatus :one
UPDATE decisions
SET status = $2,
    decided_at = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListDecisionsByIDs :many
SELECT * FROM decisions
WHERE id = ANY(sqlc.arg(ids)::uuid[]);
