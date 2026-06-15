-- name: StoreRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE id = $1;

-- name: RevokeAllRefreshTokensForUser :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: MarkRefreshTokenReplaced :exec
UPDATE refresh_tokens
SET revoked_at = now(),
    replaced_by = $2
WHERE id = $1;
