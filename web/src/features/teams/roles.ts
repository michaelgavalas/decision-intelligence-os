import type { Role } from "@/graphql/generated/graphql";
import type { BadgeProps } from "@/components/ui/badge";

/** Human-readable label for each membership role. */
export const ROLE_LABELS: Record<Role, string> = {
  ADMIN: "Admin",
  MEMBER: "Member",
  VIEWER: "Viewer",
};

/**
 * Badge variant used as a colour cue for each role. The label is always shown
 * alongside, so colour is never the sole signal.
 */
export const ROLE_BADGE_VARIANTS: Record<Role, BadgeProps["variant"]> = {
  ADMIN: "primary",
  MEMBER: "default",
  VIEWER: "outline",
};

/** Roles ordered from most to least privileged, for select options. */
export const ROLE_OPTIONS: Role[] = ["ADMIN", "MEMBER", "VIEWER"];
