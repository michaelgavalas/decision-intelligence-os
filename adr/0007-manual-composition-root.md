# 0007 - Manual dependency-injection composition root

## Status

Accepted

## Context

The backend has many small collaborators: a pgx pool, repositories per domain, services per domain, the auth/token machinery, DataLoaders, the GraphQL server, middleware, and config. These must be constructed in the right order and wired together - services receive their repository interface and the other domains' service interfaces they depend on (see the dependency rule in [ADR-0001](./0001-modular-monolith.md)).

Go has DI frameworks (Wire's code generation, Fx/dig's runtime reflection container). We had to decide whether to adopt one or wire dependencies by hand.

## Decision

Wire everything **by hand in a manual composition root** in `cmd/api`. There is **no DI framework**.

`main` (and a small `bootstrap`/`server` setup it calls) constructs dependencies bottom-up: load typed config → open the pgx pool → build repositories → build services (injecting repositories and the specific other-domain service interfaces each one needs) → build resolvers → assemble middleware and the GraphQL server. Dependencies are passed as **constructor arguments** and held as **interface fields**, so call sites depend on contracts, not concretes.

## Consequences

**Positive**

- **The wiring is just Go you can read top to bottom.** Anyone can open `cmd/api` and see the entire object graph and its construction order - no annotations, no generated graph to decode, no reflection container to reason about.
- **Compile-time safety.** A missing or mistyped dependency is a build error, immediately, with a normal stack location - not a runtime panic from a container at startup.
- **Zero added dependencies** and no framework lifecycle to learn - directly serving the simplicity and minimal-dependency goals.
- **It reinforces the dependency rule.** Because you wire each service's collaborators explicitly, an accidental cross-domain dependency is visible right there in the composition root.

**Negative / trade-offs**

- **The composition root grows** as domains are added; it's the one file that touches everything. Acceptable - and arguably desirable: it's a single, honest map of the system. We keep it organized with small per-area setup helpers.
- **No automatic resolution** - you must order construction yourself. With a modest, acyclic graph this is trivial and, again, makes the structure explicit.
- **Some repetitive constructor calls.** A small price for clarity over magic.

## Alternatives considered

- **Wire (compile-time codegen DI)** - rejected: less magic than runtime containers, but still indirection (generated wiring) over a graph small enough to read by hand; adds a generation step for little gain.
- **Fx / dig (runtime reflection container)** - rejected: hides the object graph, turns wiring mistakes into startup panics, and adds a framework lifecycle - the opposite of "explicit over clever."
