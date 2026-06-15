import { Lightbulb, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { useToast } from "@/components/ui/toaster";
import {
  type AssumptionFieldsFragment,
  useRemoveAssumptionMutation,
} from "@/graphql/generated/graphql";

import { AssumptionDialog } from "./AssumptionDialog";
import { ConfirmDialog } from "./ConfirmDialog";
import { EvidenceList } from "./EvidenceList";
import { toPercent, toPercentValue } from "./percent";

interface AssumptionsSectionProps {
  decisionId: string;
  assumptions: AssumptionFieldsFragment[];
  onChanged(): void;
}

/** Lists a decision's assumptions, each with its confidence and evidence. */
export function AssumptionsSection({
  decisionId,
  assumptions,
  onChanged,
}: AssumptionsSectionProps) {
  const toast = useToast();
  const [addOpen, setAddOpen] = useState(false);
  const [editing, setEditing] = useState<AssumptionFieldsFragment | null>(null);
  const [removing, setRemoving] = useState<AssumptionFieldsFragment | null>(
    null,
  );

  const [removeAssumption, { loading: removeLoading }] =
    useRemoveAssumptionMutation();

  async function handleRemove() {
    if (!removing) {
      return;
    }
    try {
      const { data } = await removeAssumption({
        variables: { id: removing.id },
      });
      const payload = data?.removeAssumption;
      if (payload && payload.userErrors.length > 0) {
        toast.error(payload.userErrors[0].message);
        return;
      }
      toast.success("Assumption removed");
      setRemoving(null);
      onChanged();
    } catch {
      toast.error("Something went wrong. Please try again.");
    }
  }

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between">
        <CardTitle>Assumptions</CardTitle>
        <Button size="sm" variant="outline" onClick={() => setAddOpen(true)}>
          <Plus className="size-4" />
          Add assumption
        </Button>
      </CardHeader>
      <CardContent className="flex flex-col gap-4">
        {assumptions.length === 0 ? (
          <EmptyState
            icon={Lightbulb}
            title="No assumptions yet"
            description="Capture what must be true for this decision to succeed."
          />
        ) : (
          <ul className="flex flex-col gap-4">
            {assumptions.map((assumption) => {
              const percentValue = toPercentValue(assumption.confidence);
              return (
                <li
                  key={assumption.id}
                  className="rounded-lg border border-border p-4"
                >
                  <div className="flex items-start justify-between gap-3">
                    <p className="whitespace-pre-wrap break-words text-sm font-medium text-foreground">
                      {assumption.statement}
                    </p>
                    <div className="flex shrink-0 items-center gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-8"
                        aria-label="Edit assumption"
                        onClick={() => setEditing(assumption)}
                      >
                        <Pencil className="size-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-8 text-destructive"
                        aria-label="Remove assumption"
                        onClick={() => setRemoving(assumption)}
                      >
                        <Trash2 className="size-4" />
                      </Button>
                    </div>
                  </div>

                  <div className="mt-3 flex items-center gap-3">
                    <span className="text-xs font-medium tabular-nums text-muted-foreground">
                      Confidence {toPercent(assumption.confidence)}
                    </span>
                    <div
                      className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted"
                      role="progressbar"
                      aria-valuenow={percentValue}
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-label="Confidence"
                    >
                      <div
                        className="h-full rounded-full bg-primary"
                        style={{ width: `${percentValue}%` }}
                      />
                    </div>
                  </div>

                  <EvidenceList
                    assumptionId={assumption.id}
                    evidence={assumption.evidence}
                    onChanged={onChanged}
                  />
                </li>
              );
            })}
          </ul>
        )}
      </CardContent>

      <AssumptionDialog
        decisionId={decisionId}
        open={addOpen}
        onOpenChange={setAddOpen}
        onSaved={onChanged}
      />

      {editing ? (
        <AssumptionDialog
          decisionId={decisionId}
          assumption={editing}
          open={editing !== null}
          onOpenChange={(open) => {
            if (!open) {
              setEditing(null);
            }
          }}
          onSaved={onChanged}
        />
      ) : null}

      <ConfirmDialog
        open={removing !== null}
        onOpenChange={(open) => {
          if (!open) {
            setRemoving(null);
          }
        }}
        title="Remove assumption?"
        description="This assumption and its attached evidence will be permanently removed."
        confirmLabel="Remove"
        destructive
        loading={removeLoading}
        onConfirm={() => void handleRemove()}
      />
    </Card>
  );
}
