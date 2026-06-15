// Package decisions owns the decision lifecycle: a decision moves through a
// small set of statuses from draft to a final, archived state. Decisions are
// the root aggregate the assumptions, evidence, predictions, and outcomes
// domains hang from.
package decisions

import (
	"time"

	"github.com/google/uuid"
)

// Decision lifecycle statuses.
const (
	// StatusDraft is a newly created decision still being shaped.
	StatusDraft = "draft"
	// StatusActive is a decision under active deliberation.
	StatusActive = "active"
	// StatusDecided is a decision whose course has been chosen.
	StatusDecided = "decided"
	// StatusArchived is a closed decision retained only for history.
	StatusArchived = "archived"
)

// Decision is the canonical decision entity.
type Decision struct {
	ID          uuid.UUID
	TeamID      uuid.UUID
	OwnerID     uuid.UUID
	Title       string
	Description string
	Status      string
	DecidedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// canTransition reports whether a decision may move from status from to status
// to. The lifecycle is intentionally narrow: a decision flows forward and may
// always be archived, but never moves backward.
func canTransition(from, to string) bool {
	switch from {
	case StatusDraft:
		return to == StatusActive || to == StatusArchived
	case StatusActive:
		return to == StatusDecided || to == StatusArchived
	case StatusDecided:
		return to == StatusArchived
	default:
		return false
	}
}
