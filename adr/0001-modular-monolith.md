# 0001 - Modular monolith over microservices

## Status

Accepted

## Context

Decision Intelligence OS is built and operated by a very small team. Its domain is highly relational: a single decision view stitches together assumptions, evidence, predictions, outcomes, and team membership, and the analytics engine reads across `predictions` and `outcomes` together. The product needs strong consistency - the audit trail must never disagree with the normalized data - and it needs to be cheap to run and easy to reason about.

Microservices are frequently the default reach for "serious" backends, so the choice deserves an explicit decision rather than a default.

## Decision

Build the backend as a **modular monolith**: a single Go binary, deployed alongside a single PostgreSQL database, with bounded contexts (`auth`, `users`, `teams`, `decisions`, `assumptions`, `evidence`, `predictions`, `outcomes`, `analytics`, `ai`) enforced **in code** rather than over the network.

Boundaries are kept honest by a strict dependency rule: resolvers depend on service interfaces; a service depends on its own repository interface and on *other domains' service interfaces only* - never another domain's repository or structs - with no import cycles. Cross-context communication is by ID through those interfaces. If a context ever genuinely needs to be extracted into its own service, the seam already exists.

## Consequences

**Positive**

- **Strong consistency for free.** A state change and its audit event commit in one local transaction; no distributed-transaction or saga machinery.
- **Drastically lower operational cost.** One binary, one database, one deploy pipeline. No service mesh, no inter-service auth, no cross-service tracing just to debug a single request.
- **Refactoring across boundaries is a compiler-checked change,** not a multi-repo schema negotiation.
- **Fast to develop and easy to hold in one head** - directly serving our simplicity goal.

**Negative / trade-offs**

- **No independent per-service scaling.** We scale the whole binary. Acceptable: the workload is database-bound, and Postgres is the bottleneck long before the Go process is.
- **No independent deploys per domain.** A change anywhere ships the whole binary. Acceptable for one team.
- **Boundary discipline is a social/CI contract, not a network firewall.** A careless import can violate it, so we enforce the dependency rule in review and linting.

## Alternatives considered

- **Microservices per domain** - rejected: distributed-systems complexity with none of the organizational payoff at our team size.
- **A single unstructured monolith** - rejected: without enforced boundaries it rots into a big ball of mud, and we'd lose the option to extract a service later.
