import type { DecisionStatus } from "@/graphql/generated/graphql";

/**
 * The lifecycle statuses a decision may move to from each current status.
 * DECIDED is normally reached by recording an outcome, but the explicit
 * transition is still offered.
 */
export const ALLOWED_TRANSITIONS: Record<DecisionStatus, DecisionStatus[]> = {
  DRAFT: ["ACTIVE", "ARCHIVED"],
  ACTIVE: ["DECIDED", "ARCHIVED"],
  DECIDED: ["ARCHIVED"],
  ARCHIVED: [],
};

/** The statuses a decision in `status` may transition to. */
export function allowedTransitions(status: DecisionStatus): DecisionStatus[] {
  return ALLOWED_TRANSITIONS[status];
}
