// Package outcomes owns the final result recorded for a decision: a summary, a
// success flag, and the time the decision was resolved. A decision has at most
// one outcome, so recording again updates the existing record. Outcomes are
// scoped to a decision and authorized through it.
package outcomes

import (
	"time"

	"github.com/google/uuid"
)

// Outcome is the canonical outcome entity. Success reports whether the decision
// achieved its intended result; ResolvedAt is when that result was known.
type Outcome struct {
	ID         uuid.UUID
	DecisionID uuid.UUID
	Summary    string
	Success    bool
	ResolvedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
