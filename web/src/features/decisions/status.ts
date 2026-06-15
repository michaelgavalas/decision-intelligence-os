import type { BadgeProps } from "@/components/ui/badge";
import type { DecisionStatus } from "@/graphql/generated/graphql";

/** Maps a decision status enum to its Badge variant and human label. */
export const STATUS_META: Record<
  DecisionStatus,
  { variant: BadgeProps["variant"]; label: string }
> = {
  DRAFT: { variant: "draft", label: "Draft" },
  ACTIVE: { variant: "active", label: "Active" },
  DECIDED: { variant: "decided", label: "Decided" },
  ARCHIVED: { variant: "archived", label: "Archived" },
};
