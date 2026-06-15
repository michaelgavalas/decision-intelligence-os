# 0008 - AI as an optional, isolated, disabled-by-default module

## Status

Accepted

## Context

There is obvious product value in assistance features: summarizing a pile of evidence, critiquing whether an assumption is well-formed, and flagging likely cognitive biases in how a decision was framed. At the same time, the integrity of the product rests on it being a dependable instrument for decision quality. A reliance on an external, probabilistic, sometimes-unavailable service in the *core* path would undermine that - and would make the product impossible to run or test offline, in regulated settings, or by users who simply don't want it.

We need a position on where assistance features live and how much the system depends on them.

## Decision

Treat AI as an **optional, isolated module**, **disabled by default**.

- All AI code lives behind interfaces in a single bounded context, **`internal/ai`**. No other domain imports a concrete AI implementation.
- It is gated by configuration - **`AI_ENABLED=false` by default.** When disabled, the module is effectively absent and the product is **fully functional**: every core workflow (create decision, record assumptions, attach evidence, forecast, record outcomes, view analytics) works end to end with no AI involved.
- AI may **summarize evidence, critique assumptions, and detect bias** - strictly as advisory overlays a user can request.
- AI **must never control or be required for a core workflow.** It cannot create, gate, or block a decision; it cannot be on the critical path of any write.

## Consequences

**Positive**

- **The product's reliability never depends on a probabilistic external service.** Outages, latency, cost spikes, or a deliberate "AI off" deployment cannot break decision-making.
- **Clean testability.** Core domains are tested without any AI; the AI module is tested in isolation behind its interface.
- **Deployment flexibility.** Privacy-sensitive or air-gapped deployments simply leave AI disabled with zero feature loss in the core.
- **Strong isolation** means the assistance provider/implementation can change without touching core domains.

**Negative / trade-offs**

- **Assistance features are off out of the box,** so their value isn't visible until an operator opts in. Acceptable, and the right default for an integrity-first product.
- **Some duplicated effort** to keep the boundary clean (interfaces, gating, graceful absence) rather than calling a service inline. That discipline is exactly what guarantees the "fully functional without AI" property.
- **Two code paths to consider** (enabled/disabled). Managed by making "disabled" the simple default and treating AI output as advisory-only in the UI.

## Alternatives considered

- **AI woven into core workflows** (e.g. required assumption critique before saving) - rejected: makes an optional accelerant a hard dependency and compromises reliability and testability.
- **No AI at all** - rejected: forgoes real product value that can be delivered safely as an isolated, opt-in overlay.
