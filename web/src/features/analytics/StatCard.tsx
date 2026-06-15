import type { LucideIcon } from "lucide-react";

import { Card } from "@/components/ui/card";
import { cn } from "@/lib/cn";

interface StatCardProps {
  /** Short metric name shown above the value. */
  label: string;
  /** The formatted metric value. */
  value: React.ReactNode;
  /** A short clarifying sub-label rendered beneath the value. */
  hint?: string;
  /** Optional icon shown in the card corner. */
  icon?: LucideIcon;
  /** Optional badge or other annotation rendered next to the value. */
  badge?: React.ReactNode;
  className?: string;
}

/** A compact KPI card with a label, prominent value, and optional annotation. */
export function StatCard({
  label,
  value,
  hint,
  icon: Icon,
  badge,
  className,
}: StatCardProps) {
  return (
    <Card className={cn("flex flex-col gap-2 p-5", className)}>
      <div className="flex items-center justify-between gap-2">
        <span className="text-sm font-medium text-muted-foreground">
          {label}
        </span>
        {Icon ? (
          <Icon className="size-4 text-muted-foreground" aria-hidden="true" />
        ) : null}
      </div>
      <div className="flex flex-wrap items-baseline gap-2">
        <span className="text-3xl font-semibold tabular-nums tracking-tight text-foreground">
          {value}
        </span>
        {badge}
      </div>
      {hint ? <p className="text-xs text-muted-foreground">{hint}</p> : null}
    </Card>
  );
}
