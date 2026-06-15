# Contributing

Thanks for your interest in Decision Intelligence OS. This guide covers everything you need to set up the project, make a change that meets the bar, and get it merged. Please skim [ARCHITECTURE.md](./ARCHITECTURE.md) first - the design has firm invariants, and PRs that violate them won't merge regardless of code quality.

## Prerequisites

| Tool | Version | Purpose |
| --- | --- | --- |
| **Go** | 1.25+ | Backend. |
| **Docker** + Compose | recent | Local stack and integration tests (`testcontainers-go`). |
| **pnpm** | latest | Monorepo workspace manager (web/mobile). |
| **sqlc** | latest | Generate type-safe DB code from SQL. |
| **golang-migrate** | latest | Create and run migrations (`migrate` CLI). |
| **golangci-lint** | latest | Linting. |

A working Docker daemon is required for the integration suite; the tests start a real PostgreSQL container.

## Getting started

```bash
git clone https://github.com/michaelgavalas/decision-intelligence-os.git
cd decision-intelligence-os

cp .env.example .env          # set JWT keys and CSRF secret for local dev
make generate                 # sqlc + gqlgen code generation
make up                       # full stack via docker compose
```

For an iterative backend loop without Docker, run Postgres in a container (or via `make up`), then:

```bash
make migrate-up
make run
```

## Make targets

```text
make help            # list all targets
make generate        # sqlc generate + gqlgen generate
make lint            # gofmt check + golangci-lint
make test            # go test ./...
make test-race       # go test -race -coverprofile=coverage.out
make build           # build the API binary
make run             # run the API locally
make migrate-up      # apply all migrations
make migrate-down    # roll back one migration
make migrate-create name=add_widgets   # scaffold a new migration pair
make up / make down  # start / stop the docker compose stack
```

## Coding standards

- **Formatting:** code must be `gofmt`/`goimports`-clean. `make lint` fails on unformatted code.
- **Linting:** `golangci-lint run ./...` must pass with no new findings.
- **Explicit over clever.** If a junior engineer can't understand the code in one reading, simplify it. Avoid magic abstractions.
- **`context.Context` is the first argument** of any function that does I/O or crosses a boundary.
- **Wrap errors with `%w`** to preserve the chain (`fmt.Errorf("create decision: %w", err)`), and define **typed/sentinel errors** for conditions callers branch on. Map them to GraphQL `extensions.code` at the resolver edge.
- **Respect the dependency rule.** Resolvers depend on service interfaces; services depend on their own repository interface and on *other domains' service interfaces only* - never another domain's repository or structs. No import cycles. All wiring lives in the `cmd/api` composition root.
- **No new infrastructure dependencies** (Redis, Kafka, an ORM, etc.) without an accompanying ADR and explicit approval.
- **Logs are structured** (`slog`) and must never contain passwords, secrets, or tokens.

## Testing and TDD

We expect a test-first workflow. Write the failing test, make it pass, refactor.

- **Service logic:** table-driven unit tests against hand-written fakes (no mocking framework).
- **Repositories and GraphQL critical paths:** integration tests against a real Postgres via `testcontainers-go`.
- Run `make test-race` before pushing; the race detector must be clean.
- **Coverage targets:** services ≥ **90%**, repositories ≥ **80%**, GraphQL critical paths covered.
- **Every bug fix must include a regression test** that fails before the fix and passes after.

## Migrations

Database schema changes follow strict rules so the history stays trustworthy and reversible:

- **One new migration per change.** Scaffold with `make migrate-create name=<snake_case>`, which produces paired `*.up.sql` and `*.down.sql` files.
- **Every migration is reversible** - write a real `down` that undoes the `up`.
- **Make it idempotent where possible** (`IF NOT EXISTS`, guarded constraint adds).
- **Never edit a migration after it has merged.** Fix forward with a new migration. Editing merged migrations breaks every environment that already applied them.
- Honor the data-model conventions: UUIDv7 PKs, `timestamptz` UTC, explicit FK `ON DELETE`, `CHECK` constraints, indexes on FKs and listing paths.

## GraphQL and code generation

The GraphQL schema is the contract of record. After editing the schema or SQL:

```bash
make generate     # regenerates sqlc models/queries and gqlgen resolvers
```

Commit the generated code alongside your change. Follow the API conventions in [ADR-0006](./adr/0006-relay-and-usererrors.md): Relay connections for lists, `input`/`Payload` mutations with a `userErrors` list for expected validation failures, and stable `extensions.code` values for transport-level errors.

## Commit conventions

We use **[Conventional Commits](https://www.conventionalcommits.org/)**, enforced in CI (commitlint):

```text
feat(decisions): add status transition guard
fix(auth): detect refresh-token reuse across sessions
docs(adr): record events-as-outbox decision
test(predictions): cover calibration scoring edge cases
refactor(analytics): extract accuracy computation
```

Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`, `build`, `perf`.

## Branch naming

Branch off `main` using the matching prefix:

```text
feat/decision-templates
fix/refresh-token-reuse
docs/architecture-diagrams
```

## Pull request process

1. Keep PRs focused and small enough to review in one sitting.
2. Ensure `make lint` and `make test-race` pass locally. CI runs **Lint → Test → Build → Security** (`govulncheck`, `gosec`, `trivy`, `gitleaks`).
3. Fill out the PR template: what changed, why, and how it was tested. Link any issue.
4. New or changed architecture-level decisions need an **ADR** in `adr/` (MADR format - see the [ADR index](./adr/README.md)).
5. Update docs (`README.md` / `ARCHITECTURE.md`) when behavior or structure changes.
6. At least one approving review is required. Address feedback with follow-up commits; we squash-merge on `main`.
7. **No deploys from local machines** - deployment happens only through the GitHub Actions pipeline on `main`.

By contributing, you agree your contributions are licensed under the project's [MIT License](./LICENSE).
