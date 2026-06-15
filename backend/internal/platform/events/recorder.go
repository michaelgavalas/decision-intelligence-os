package events

import (
	"context"
	"encoding/json"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db/sqlc"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

// Recorder appends events to the append-only events table. It carries no state;
// each Record call uses the Querier the caller supplies so the insert joins the
// caller's transaction when one is in flight.
type Recorder struct{}

// NewRecorder returns a Recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Record marshals the event payload to JSON, generates a time-ordered id, and
// inserts the event using the supplied Querier. A nil payload is stored as the
// empty JSON object so the column's NOT NULL constraint is always satisfied.
func (r *Recorder) Record(ctx context.Context, q db.Querier, e Event) error {
	payload := []byte("{}")
	if e.Payload != nil {
		marshaled, err := json.Marshal(e.Payload)
		if err != nil {
			return errors.Wrap(err, errors.KindValidation, "EVENT_PAYLOAD_INVALID", "failed to marshal event payload")
		}
		payload = marshaled
	}

	queries := sqlc.New(q)
	if _, err := queries.InsertEvent(ctx, sqlc.InsertEventParams{
		ID:            id.New(),
		AggregateID:   e.AggregateID,
		AggregateType: e.AggregateType,
		EventType:     e.Type,
		Payload:       payload,
		ActorID:       e.ActorID,
	}); err != nil {
		return errors.Wrap(err, errors.KindInternal, "EVENT_INSERT_FAILED", "failed to record event")
	}

	return nil
}
