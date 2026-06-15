# 0004 - Append-only events: audit log + transactional outbox, not event sourcing

## Status

Accepted

## Context

The product is, fundamentally, about *decision history*: who believed what, when, and how a decision evolved before the outcome was known. That history must be trustworthy and immutable. Separately, we need a reliable mechanism to drive real-time subscriptions and possible future projections without a second datastore (see [ADR-0003](./0003-postgres-only-listen-notify.md)).

Two patterns are tempting here. Full **event sourcing** makes the event stream the source of truth and derives all state by replay. The **transactional outbox** keeps normalized tables authoritative and uses an events table for reliable publication. We need to be explicit about which one we are doing, because they have very different consequences.

## Decision

Maintain a single **append-only `events` table** - `(id, aggregate_id, event_type, payload, created_at)` - that serves two roles simultaneously:

1. **Audit log:** a permanent, ordered record of every meaningful change (`DecisionCreated`, `AssumptionAdded`, `EvidenceAttached`, `PredictionCreated`, `OutcomeRecorded`, ...). **Events are never deleted.**
2. **Transactional outbox:** the same rows are the basis for asynchronous fan-out (subscriptions, future read models).

The defining invariant: **the state change and its event are written in the same database transaction.** Either both commit or neither does.

This is explicitly **not event sourcing.** The **normalized tables remain the source of truth.** We never reconstruct current state by replaying events, and **analytics are computed directly from the normalized `predictions` and `outcomes` tables**, not from the event stream.

## Consequences

**Positive**

- **The audit log cannot drift from reality** - it is written atomically with the data it describes.
- **No dual-write problem.** Because the event commits in the same transaction, there is no window where a change is saved but its notification/event is lost (or vice versa).
- **Queries stay simple.** Current state is a plain `SELECT` against normalized tables, not a fold over a stream.
- **A complete, immutable decision history** - a core product feature - falls out of the design for free.

**Negative / trade-offs**

- **Some duplication of intent** between the normalized write and the event insert. We accept this; the events carry change semantics the tables don't.
- **No free time-travel of arbitrary state.** Because we don't event-source, "reconstruct the exact entity at time T" requires interpreting events rather than a built-in replay. We don't need that capability today.
- **The events table grows unbounded** (by design - events are never deleted). Managed with partitioning/archival if it ever becomes a size concern; not a problem at current scale.

## Alternatives considered

- **Full event sourcing** - rejected: replay-derived state and projection rebuilds add large complexity for a benefit (perfect temporal reconstruction) the product doesn't require; our analytics want current normalized data anyway.
- **No events; audit via triggers or app logs** - rejected: triggers hide logic in the database and log-based audit isn't transactional or queryable as first-class data.
