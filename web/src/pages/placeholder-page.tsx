import type { LucideIcon } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";

interface PlaceholderPageProps {
  title: string;
  icon: LucideIcon;
  description: string;
}

/**
 * Temporary scaffold page. Each feature route renders one of these until the
 * corresponding feature is implemented in a later phase.
 */
export function PlaceholderPage({
  title,
  icon,
  description,
}: PlaceholderPageProps) {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-bold tracking-tight text-foreground">
        {title}
      </h1>
      <Card>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <EmptyState
            icon={icon}
            title={`${title} coming soon`}
            description={description}
          />
        </CardContent>
      </Card>
    </div>
  );
}
