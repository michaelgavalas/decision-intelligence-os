import { ExternalLink, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toaster";
import {
  type AssumptionFieldsFragment,
  type EvidenceFieldsFragment,
  type EvidenceSourceType,
  useRemoveEvidenceMutation,
} from "@/graphql/generated/graphql";

import { ConfirmDialog } from "./ConfirmDialog";
import { EvidenceDialog } from "./EvidenceDialog";

const SOURCE_TYPE_LABELS: Record<EvidenceSourceType, string> = {
  URL: "URL",
  DOCUMENT: "Document",
  NOTE: "Note",
  DATASET: "Dataset",
};

interface EvidenceListProps {
  assumptionId: string;
  evidence: AssumptionFieldsFragment["evidence"];
  onChanged(): void;
}

/** Lists, attaches, edits, and removes evidence for a single assumption. */
export function EvidenceList({
  assumptionId,
  evidence,
  onChanged,
}: EvidenceListProps) {
  const toast = useToast();
  const [attachOpen, setAttachOpen] = useState(false);
  const [editing, setEditing] = useState<EvidenceFieldsFragment | null>(null);
  const [removing, setRemoving] = useState<EvidenceFieldsFragment | null>(null);

  const [removeEvidence, { loading: removeLoading }] =
    useRemoveEvidenceMutation();

  async function handleRemove() {
    if (!removing) {
      return;
    }
    try {
      const { data } = await removeEvidence({ variables: { id: removing.id } });
      const payload = data?.removeEvidence;
      if (payload && payload.userErrors.length > 0) {
        toast.error(payload.userErrors[0].message);
        return;
      }
      toast.success("Evidence removed");
      setRemoving(null);
      onChanged();
    } catch {
      toast.error("Something went wrong. Please try again.");
    }
  }

  return (
    <div className="mt-3 flex flex-col gap-2 border-l-2 border-border pl-4">
      {evidence.length === 0 ? (
        <p className="text-xs text-muted-foreground">No evidence attached yet.</p>
      ) : (
        <ul className="flex flex-col gap-2">
          {evidence.map((item) => (
            <li
              key={item.id}
              className="flex items-start justify-between gap-3 rounded-md bg-muted/40 px-3 py-2"
            >
              <div className="flex min-w-0 flex-col gap-1">
                <div className="flex items-center gap-2">
                  <Badge variant="outline">
                    {SOURCE_TYPE_LABELS[item.sourceType]}
                  </Badge>
                  {item.sourceUrl ? (
                    <a
                      href={item.sourceUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                    >
                      Open source
                      <ExternalLink className="size-3" />
                    </a>
                  ) : null}
                </div>
                <p className="whitespace-pre-wrap break-words text-sm text-foreground">
                  {item.content}
                </p>
              </div>
              <div className="flex shrink-0 items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-8"
                  aria-label="Edit evidence"
                  onClick={() => setEditing(item)}
                >
                  <Pencil className="size-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-8 text-destructive"
                  aria-label="Remove evidence"
                  onClick={() => setRemoving(item)}
                >
                  <Trash2 className="size-4" />
                </Button>
              </div>
            </li>
          ))}
        </ul>
      )}

      <div>
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setAttachOpen(true)}
          className="text-muted-foreground"
        >
          <Plus className="size-4" />
          Attach evidence
        </Button>
      </div>

      <EvidenceDialog
        assumptionId={assumptionId}
        open={attachOpen}
        onOpenChange={setAttachOpen}
        onSaved={onChanged}
      />

      {editing ? (
        <EvidenceDialog
          assumptionId={assumptionId}
          evidence={editing}
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
        title="Remove evidence?"
        description="This evidence will be permanently removed from the assumption."
        confirmLabel="Remove"
        destructive
        loading={removeLoading}
        onConfirm={() => void handleRemove()}
      />
    </div>
  );
}
