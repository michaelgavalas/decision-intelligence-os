-- name: InsertEvent :one
INSERT INTO events (id, aggregate_id, aggregate_type, event_type, payload, actor_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListEventsByAggregate :many
SELECT * FROM events
WHERE aggregate_id = $1
ORDER BY occurred_at;
