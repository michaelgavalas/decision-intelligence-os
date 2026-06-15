// Package events records domain changes into the append-only events table.
// Recording happens inside the caller's transaction (or directly against the
// pool) so that an event and the state change it describes commit atomically.
package events

import "github.com/google/uuid"

// Canonical event type names. These string values are stable and form part of
// the audit history contract, so they must not change once persisted.
const (
	TypeDecisionCreated   = "DecisionCreated"
	TypeDecisionUpdated   = "DecisionUpdated"
	TypeAssumptionAdded   = "AssumptionAdded"
	TypeEvidenceAttached  = "EvidenceAttached"
	TypePredictionCreated = "PredictionCreated"
	TypeOutcomeRecorded   = "OutcomeRecorded"
)

// Event describes a single domain occurrence to be appended to the audit log.
type Event struct {
	// AggregateID identifies the entity the event pertains to.
	AggregateID uuid.UUID
	// AggregateType names the entity kind, e.g. "decision".
	AggregateType string
	// Type is one of the Type* constants in this package.
	Type string
	// Payload is any JSON-serializable value. A nil payload is stored as {}.
	Payload any
	// ActorID is the user responsible for the change, when known.
	ActorID *uuid.UUID
}
