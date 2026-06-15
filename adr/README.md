# Architecture Decision Records

This directory records the load-bearing architectural decisions for Decision Intelligence OS - the choices that are expensive to reverse and that newcomers most often ask "why?" about. Each record captures the context at the time, the decision, and the trade-offs we accepted, so that future contributors can revisit a decision with the original reasoning in hand rather than guessing at intent.

Records use the **[MADR](https://adr.github.io/madr/)** (Markdown Any Decision Records) format: a short, structured document with **Title**, **Status**, **Context**, **Decision**, **Consequences** (positive and negative), and, where relevant, **Alternatives considered**. A decision is never silently changed - if we reverse course, we add a new ADR that supersedes the old one and update its status.

## Index

| ADR | Decision | Summary |
| --- | --- | --- |
| [0001](./0001-modular-monolith.md) | Modular monolith over microservices | One deployable backend with compiler-enforced bounded contexts, not a distributed system. |
| [0002](./0002-graphql-only-api.md) | GraphQL as the only public API | A single schema-first contract for web and mobile; REST limited to ops health checks. |
| [0003](./0003-postgres-only-listen-notify.md) | PostgreSQL as the sole datastore | Postgres is the source of truth *and* the pub/sub bus via `LISTEN/NOTIFY` - no Redis or Kafka. |
| [0004](./0004-events-outbox-not-event-sourcing.md) | Append-only events: audit log + outbox | Events written in the same tx as state changes; normalized tables stay the source of truth. |
| [0005](./0005-sqlc-over-orm.md) | sqlc + golang-migrate over an ORM | Compile-time-checked hand-written SQL and versioned reversible migrations, no ORM. |
| [0006](./0006-relay-and-usererrors.md) | Relay connections + userErrors | Cursor-based pagination and validation failures returned as data, not transport errors. |
| [0007](./0007-manual-composition-root.md) | Manual DI composition root | Explicit wiring in `cmd/api`; no DI framework or reflection magic. |
| [0008](./0008-ai-optional-isolated.md) | AI optional, isolated, off by default | AI is an accelerant in `internal/ai`, never required for any core workflow. |
