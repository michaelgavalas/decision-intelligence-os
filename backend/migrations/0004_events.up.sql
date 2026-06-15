-- Migration 0004: audit event log.
-- Append-only event stream capturing domain changes for audit history. Rows
-- are never updated or deleted, so the table intentionally omits updated_at.

CREATE TABLE IF NOT EXISTS events (
    id uuid PRIMARY KEY,
    aggregate_id uuid NOT NULL,
    aggregate_type text NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    actor_id uuid,
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_events_aggregate ON events(aggregate_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(aggregate_type, occurred_at);
