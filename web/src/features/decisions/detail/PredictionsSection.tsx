import { Pencil, Plus, TrendingUp } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import type { PredictionFieldsFragment } from "@/graphql/generated/graphql";
import { formatDate } from "@/lib/format";

import { PredictionDialog } from "./PredictionDialog";
import { toPercent } from "./percent";

interface PredictionsSectionProps {
  decisionId: string;
  predictions: PredictionFieldsFragment[];
  onChanged(): void;
}

/** Lists a decision's forecasts with their probability and resolve date. */
export function PredictionsSection({
  decisionId,
  predictions,
  onChanged,
}: PredictionsSectionProps) {
  const [addOpen, setAddOpen] = useState(false);
  const [editing, setEditing] = useState<PredictionFieldsFragment | null>(null);

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle>Predictions</CardTitle>
        <Button size="sm" variant="outline" onClick={() => setAddOpen(true)}>
          <Plus className="size-4" />
          Add prediction
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {predictions.length === 0 ? (
          <EmptyState
            icon={TrendingUp}
            title="No predictions yet"
            description="Record forecasts so you can measure how well-calibrated they were."
          />
        ) : (
          <ul className="flex flex-col gap-3">
            {predictions.map((prediction) => (
              <li
                key={prediction.id}
                className="flex items-start justify-between gap-3 rounded-lg border border-border p-4"
              >
                <div className="flex min-w-0 flex-col gap-2">
                  <p className="whitespace-pre-wrap break-words text-sm font-medium text-foreground">
                    {prediction.statement}
                  </p>
                  <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                    <span className="font-medium tabular-nums">
                      Probability {toPercent(prediction.probability)}
                    </span>
                    <span className="tabular-nums">
                      Resolves {formatDate(prediction.resolvesAt)}
                    </span>
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-8 shrink-0"
                  aria-label="Edit prediction"
                  onClick={() => setEditing(prediction)}
                >
                  <Pencil className="size-4" />
                </Button>
              </li>
            ))}
          </ul>
        )}
      </CardContent>

      <PredictionDialog
        decisionId={decisionId}
        open={addOpen}
        onOpenChange={setAddOpen}
        onSaved={onChanged}
      />

      {editing ? (
        <PredictionDialog
          decisionId={decisionId}
          prediction={editing}
          open={editing !== null}
          onOpenChange={(open) => {
            if (!open) {
              setEditing(null);
            }
          }}
          onSaved={onChanged}
        />
      ) : null}
    </Card>
  );
}
