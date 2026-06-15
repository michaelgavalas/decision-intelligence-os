// Package assumptions owns the assumptions that underpin a decision: the
// beliefs a decision rests on, each carrying a confidence score. Assumptions are
// scoped to a decision and authorized through it.
package assumptions

import (
	"time"

	"github.com/google/uuid"
)

// Assumption is the canonical assumption entity. Confidence is a probability in
// the closed interval [0, 1].
type Assumption struct {
	ID         uuid.UUID
	DecisionID uuid.UUID
	Statement  string
	Confidence float64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
