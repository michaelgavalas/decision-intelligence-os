import type { BadgeProps } from "@/components/ui/badge";

export interface BrierQuality {
  label: string;
  variant: NonNullable<BadgeProps["variant"]>;
}

/**
 * Maps a Brier score (0 is perfect, 1 is worst) to a qualitative rating.
 * Thresholds: <=0.1 Excellent, <=0.2 Good, <=0.25 Fair, otherwise Needs work.
 */
export function brierQuality(score: number): BrierQuality {
  if (score <= 0.1) {
    return { label: "Excellent", variant: "success" };
  }
  if (score <= 0.2) {
    return { label: "Good", variant: "primary" };
  }
  if (score <= 0.25) {
    return { label: "Fair", variant: "warning" };
  }
  return { label: "Needs work", variant: "destructive" };
}
