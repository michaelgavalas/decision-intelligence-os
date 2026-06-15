# 0005 - sqlc + golang-migrate instead of an ORM

## Status

Accepted

## Context

The backend is data-intensive and relies on PostgreSQL features deliberately: foreign keys with explicit `ON DELETE`, `CHECK` constraints, carefully chosen indexes, `LISTEN/NOTIFY`, and transactional writes that touch a domain table and the `events` table together. We need a data-access approach that makes these first-class rather than abstracting them away, and a migration approach that keeps schema history reversible and reviewable.

ORMs (in Go, e.g. GORM) trade SQL control for convenience. Our project values correctness, simplicity, and type safety over that convenience, and our constraints explicitly forbid an ORM.

## Decision

Use **sqlc** for data access and **golang-migrate** for schema migrations. No ORM.

- **sqlc** generates type-safe Go from hand-written SQL queries. We write the SQL we want; sqlc checks it against the schema at generation time and produces typed methods. Queries live in `backend/sql/`, generated code is committed.
- **golang-migrate** manages **versioned, reversible** SQL migrations (`*.up.sql` / `*.down.sql`) in `backend/migrations/`. Migrations are idempotent where possible and are **never edited after merge** - fixes go forward as new migrations.
- Migrations run as a one-shot step in deployment and via `make migrate-*` locally.

## Consequences

**Positive**

- **Full control over SQL.** We use Postgres features directly - CTEs, `RETURNING`, partial indexes, constraint-aware upserts - with no leaky abstraction in the way.
- **Compile-time safety without runtime reflection.** sqlc catches a renamed column or a type mismatch at code-gen time, not in production. The generated code is plain, readable Go.
- **Predictable performance.** What runs against the database is exactly what we wrote; no surprise N+1 from lazy associations, no opaque query builder.
- **Reviewable schema history.** Each migration is a small, explicit, reversible SQL diff.

**Negative / trade-offs**

- **More boilerplate for trivial CRUD** than an ORM - you write the SQL yourself. Acceptable: the cost is small and the clarity is worth it.
- **A code-generation step** (`make generate`) must run after query/schema changes, and generated files are committed. A minor workflow tax.
- **No automatic schema migration from models.** We hand-author migrations. This is a feature for us - it forces deliberate, reviewed schema evolution.

## Alternatives considered

- **An ORM (GORM, ent, etc.)** - rejected: hides SQL, encourages N+1 and over-fetching, and fights the explicit-Postgres-features design; also disallowed by project constraints.
- **Raw `database/sql` with hand-mapped rows** - rejected: same SQL control as sqlc but with manual, error-prone row scanning and no compile-time query checking.
