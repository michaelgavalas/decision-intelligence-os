# 0002 - GraphQL as the only public API

## Status

Accepted

## Context

Two clients consume the backend - a React web app and an Expo mobile app - and they need to share a contract. Their data needs differ: a desktop decision board pulls deep, nested graphs, while a mobile review screen wants a slim slice. The domain is naturally a graph (decision → assumptions → evidence; decision → predictions → outcome), and we want real-time updates without a separate broker.

We had to choose the shape of the public API up front, because it dictates client tooling, codegen, and how teams reason about every feature.

## Decision

**GraphQL is the only public API.** The schema is the contract of record (schema-first, via gqlgen), shared identically by web and mobile. Queries, mutations, and subscriptions are all GraphQL.

**REST is reserved exclusively for operations:** `/healthz` (liveness) and `/readyz` (readiness with a DB ping), plus optional env-gated `/metrics`. These are for orchestrators and monitoring, not for clients. No other REST endpoints are added without an explicit decision.

All types are strongly typed; generic `JSON` scalars for domain data are disallowed, so the schema stays self-documenting and the generated client types stay precise.

## Consequences

**Positive**

- **One contract, two clients.** Web and mobile generate types from the same schema; drift is caught at build time.
- **Clients fetch exactly what they need,** avoiding the over/under-fetching that plagues fixed REST resources - valuable on mobile networks.
- **Built-in subscriptions** give us real-time updates over one transport (`graphql-ws`) without bolting on a second protocol.
- **The schema is living documentation** and powers tooling (codegen, IDE autocomplete, type-checked queries).

**Negative / trade-offs**

- **GraphQL invites expensive queries.** Mitigated with per-request DataLoaders (N+1 control), and query depth + complexity limits; introspection is disabled in production.
- **No free HTTP caching** the way REST GETs get from CDNs/proxies. Acceptable: this is an authenticated, personalized product, so HTTP caching was never going to do much.
- **More upfront tooling** (codegen on schema change). Worth it for the type safety and shared contract.

## Alternatives considered

- **REST/JSON** - rejected: a rigid resource shape forces over/under-fetching across two divergent clients and gives us nothing for real-time.
- **gRPC** - rejected: excellent service-to-service, awkward for browsers and unnecessary for a single backend with web/mobile clients.
