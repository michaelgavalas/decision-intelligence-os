# 0006 - Relay connections and the userErrors mutation pattern

## Status

Accepted

## Context

With GraphQL chosen as the only public API (see [ADR-0002](./0002-graphql-only-api.md)), we still have to settle the conventions every type and mutation follows: how lists paginate, how IDs are shaped, and - crucially - how clients distinguish a *user's mistake* ("confidence must be between 0 and 1") from a *system failure* (database down) or an *authorization denial*. Conflating these makes client code defensive and brittle, with validation messages buried in transport-level error arrays.

We want a consistent, predictable contract that both the web and mobile clients can rely on without per-field special-casing.

## Decision

Adopt **Relay conventions** and a **`userErrors` mutation pattern**:

- **Relay-style cursor pagination.** Every list returns a **`Connection`/`Edge`** shape with opaque, stable cursors and `PageInfo`; no offset pagination. Entities are addressed by the GraphQL `ID` scalar.
- **Mutations take a single `input` object and return a typed `Payload`.** Each payload carries the affected entity *and* a `userErrors: [UserError!]!` list.
- **Expected, recoverable validation failures are returned as data** in `userErrors` (with a `field` path and a `message`), not as GraphQL transport errors. The mutation still "succeeds" at the transport level; the client renders inline messages from `userErrors`.
- **Unexpected and authorization failures are GraphQL errors** carrying a stable `extensions.code`: `UNAUTHENTICATED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `VALIDATION`, `INTERNAL`.
- **Per-request DataLoaders** batch nested lookups to prevent N+1.

## Consequences

**Positive**

- **Clean separation of concerns.** Clients branch on `userErrors` for form-level UX and on `extensions.code` for transport-level handling (redirect on `UNAUTHENTICATED`, toast on `INTERNAL`). No string-matching error messages.
- **Consistent, scalable pagination.** Cursors are stable under inserts/deletes, unlike offsets, and the same shape works everywhere.
- **Opaque cursors** decouple clients from the underlying keyset ordering and stay valid across inserts and deletes.
- **Mobile and web share identical handling logic.**

**Negative / trade-offs**

- **More schema boilerplate** - every mutation needs an `input` and a `Payload` type, every list a `Connection`. Mitigated by codegen and a consistent template.
- **Cursor pagination is harder to jump-to-page** than offsets (no "page 7"). Acceptable: our UIs are feed/scroll-oriented.
- **Two error channels** must be applied consistently, or the benefit erodes. Enforced in review.

## Alternatives considered

- **Throwing all errors as GraphQL transport errors** - rejected: forces clients to parse error arrays and string-match to tell validation from outages.
- **Offset/limit pagination** - rejected: unstable under concurrent writes and inconsistent across clients.
