# 0003 - PostgreSQL as the sole datastore; LISTEN/NOTIFY for pub/sub

## Status

Accepted

## Context

The system needs durable relational storage, transactional integrity for the audit/outbox model, and a way to push real-time updates to GraphQL subscriptions. The reflexive answer in many stacks is "Postgres for data, Redis for cache/pub-sub, maybe Kafka for events." Each additional datastore is another thing to provision, secure, monitor, back up, and keep consistent with the others.

Our constraints already forbid Redis, MongoDB, Elasticsearch, Kafka, RabbitMQ, and NATS unless explicitly justified. This ADR records *why* a single datastore is sufficient.

## Decision

**PostgreSQL 18 is the only datastore,** and it serves three jobs:

1. **Source of truth** - all normalized state and the append-only `events` table.
2. **Pub/sub bus** - GraphQL subscriptions are driven by Postgres **`LISTEN/NOTIFY`**. When a write transaction commits an event, it issues a `NOTIFY`; subscription resolvers `LISTEN` and fan out to connected `graphql-ws` clients.
3. **Auth rate limiting** - counters live in Postgres, not Redis.

No second datastore is introduced unless profiling proves Postgres cannot do the job.

## Consequences

**Positive**

- **One thing to operate, secure, and back up.** Dramatically smaller operational surface and failure-mode space.
- **No cross-store consistency problem.** Because the event and the state change commit together, the same transaction that durably records a change is the one that triggers the notification.
- **`LISTEN/NOTIFY` is transactional:** notifications fire only on commit, so subscribers never see phantom events from rolled-back transactions.
- **Simpler local dev and CI** - one container.

**Negative / trade-offs**

- **`NOTIFY` payloads are size-limited and not durable.** We treat them as lightweight signals ("something changed for aggregate X"); the durable record is the `events` row, which a subscriber can read back. We do not rely on the payload as the only delivery.
- **`LISTEN/NOTIFY` does not buffer for disconnected clients.** A client that drops a subscription may miss notifications; on reconnect it refetches current state. Acceptable for a UI-update channel; this is not a guaranteed-delivery work queue.
- **Throughput ceiling is Postgres's.** At very high fan-out a dedicated broker would scale further - but we are nowhere near that, and adding one now would be premature.

## Alternatives considered

- **Redis pub/sub + cache** - rejected: a second datastore for a problem Postgres already solves, with no transactional tie to our writes.
- **Kafka/NATS event bus** - rejected: heavy infrastructure for a single-binary product; our needs are an audit log and UI notifications, both of which Postgres covers.
