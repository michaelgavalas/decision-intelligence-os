-- Migration 0003: forecasting and resolution.
-- Adds predictions made against a decision and the single recorded outcome
-- per decision used to score forecast quality.

CREATE TABLE IF NOT EXISTS predictions (
    id uuid PRIMARY KEY,
    decision_id uuid NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    statement text NOT NULL CHECK (length(statement) BETWEEN 1 AND 2000),
    probability numeric(4, 3) NOT NULL CHECK (probability >= 0 AND probability <= 1),
    resolves_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_predictions_decision_created ON predictions(decision_id, created_at DESC);

CREATE TABLE IF NOT EXISTS outcomes (
    id uuid PRIMARY KEY,
    decision_id uuid NOT NULL UNIQUE REFERENCES decisions(id) ON DELETE CASCADE,
    summary text NOT NULL DEFAULT '',
    success boolean NOT NULL,
    resolved_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
