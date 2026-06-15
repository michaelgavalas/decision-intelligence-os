-- name: CreateUser :one
INSERT INTO users (id, email, name, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserName :one
UPDATE users
SET name = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListUsersByIDs :many
SELECT * FROM users
WHERE id = ANY(sqlc.arg(ids)::uuid[]);
