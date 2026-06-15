// Package predictions owns the forecasts attached to a decision: probability
// estimates about future outcomes, each optionally bounded by a resolution
// time. Predictions are scoped to a decision and authorized through it.
package predictions

import (
	"time"

	"github.com/google/uuid"
)

// Prediction is the canonical prediction entity. Probability is a value in the
// closed interval [0, 1]. ResolvesAt is the optional time by which the
// prediction is expected to be settled.
type Prediction struct {
	ID          uuid.UUID
	DecisionID  uuid.UUID
	Statement   string
	Probability float64
	ResolvesAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
