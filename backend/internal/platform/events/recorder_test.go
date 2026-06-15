package events_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/db"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/dbtest"
	"github.com/michaelgavalas/decision-intelligence-os/backend/internal/platform/events"
	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/id"
)

func TestRecordPersistsEvent(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	ctx := context.Background()
	rec := events.NewRecorder()

	aggregateID := id.New()
	actorID := id.New()
	payload := map[string]any{"title": "Launch in EU", "count": float64(3)}

	err := rec.Record(ctx, pool, events.Event{
		AggregateID:   aggregateID,
		AggregateType: "decision",
		Type:          events.TypeDecisionCreated,
		Payload:       payload,
		ActorID:       &actorID,
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	var (
		gotType    string
		gotAgg     uuid.UUID
		gotActor   *uuid.UUID
		rawPayload []byte
	)
	row := pool.QueryRow(ctx,
		`SELECT event_type, aggregate_id, actor_id, payload FROM events WHERE aggregate_id = $1`,
		aggregateID,
	)
	if err := row.Scan(&gotType, &gotAgg, &gotActor, &rawPayload); err != nil {
		t.Fatalf("scan event row: %v", err)
	}

	if gotType != events.TypeDecisionCreated {
		t.Errorf("event_type = %q, want %q", gotType, events.TypeDecisionCreated)
	}
	if gotAgg != aggregateID {
		t.Errorf("aggregate_id = %v, want %v", gotAgg, aggregateID)
	}
	if gotActor == nil || *gotActor != actorID {
		t.Errorf("actor_id = %v, want %v", gotActor, actorID)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(rawPayload, &roundTripped); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if roundTripped["title"] != "Launch in EU" || roundTripped["count"] != float64(3) {
		t.Errorf("payload round-trip = %v, want %v", roundTripped, payload)
	}
}

func TestRecordNilPayloadStoredAsEmptyObject(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	ctx := context.Background()
	rec := events.NewRecorder()
	aggregateID := id.New()

	err := rec.Record(ctx, pool, events.Event{
		AggregateID:   aggregateID,
		AggregateType: "decision",
		Type:          events.TypeDecisionUpdated,
		Payload:       nil,
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	var raw []byte
	if err := pool.QueryRow(ctx,
		`SELECT payload FROM events WHERE aggregate_id = $1`, aggregateID,
	).Scan(&raw); err != nil {
		t.Fatalf("scan payload: %v", err)
	}
	if string(raw) != "{}" {
		t.Errorf("payload = %q, want %q", string(raw), "{}")
	}
}

func TestRecordRolledBackTransactionDoesNotPersist(t *testing.T) {
	pool := dbtest.NewPool(t)
	dbtest.TruncateAll(t, pool)

	ctx := context.Background()
	rec := events.NewRecorder()
	txm := db.NewTxManager(pool)
	aggregateID := id.New()

	sentinel := errors.New("force rollback")
	err := txm.WithinTx(ctx, func(q db.Querier) error {
		if err := rec.Record(ctx, q, events.Event{
			AggregateID:   aggregateID,
			AggregateType: "decision",
			Type:          events.TypeAssumptionAdded,
			Payload:       map[string]any{"k": "v"},
		}); err != nil {
			return err
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("WithinTx error = %v, want sentinel", err)
	}

	var count int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM events WHERE aggregate_id = $1`, aggregateID,
	).Scan(&count); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 0 {
		t.Errorf("event count after rollback = %d, want 0", count)
	}
}
