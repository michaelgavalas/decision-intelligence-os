-- Migration 0002: decision core.
-- Adds the decision lifecycle table plus the assumptions and evidence that
-- support each decision's reasoning trail.

CREATE TABLE IF NOT EXISTS decisions (
    id uuid PRIMARY KEY,
    team_id uuid NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    owner_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title text NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
    description text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'decided', 'archived')),
    decided_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_decisions_team_created ON decisions(team_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_decisions_owner ON decisions(owner_id);

CREATE TABLE IF NOT EXISTS assumptions (
    id uuid PRIMARY KEY,
    decision_id uuid NOT NULL REFERENCES decisions(id) ON DELETE CASCADE,
    statement text NOT NULL CHECK (length(statement) BETWEEN 1 AND 2000),
    confidence numeric(4, 3) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_assumptions_decision_created ON assumptions(decision_id, created_at DESC);

CREATE TABLE IF NOT EXISTS evidence (
    id uuid PRIMARY KEY,
    assumption_id uuid NOT NULL REFERENCES assumptions(id) ON DELETE CASCADE,
    source_type text NOT NULL CHECK (source_type IN ('url', 'document', 'note', 'dataset')),
    source_url text,
    content text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_evidence_assumption_created ON evidence(assumption_id, created_at DESC);
