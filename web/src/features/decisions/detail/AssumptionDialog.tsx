import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FormField } from "@/components/ui/form-field";
import { PercentSlider } from "@/components/ui/percent-slider";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toaster";
import {
  useAddAssumptionMutation,
  useUpdateAssumptionMutation,
} from "@/graphql/generated/graphql";

import { fromPercent, toPercentValue } from "./percent";
import { userErrorsToMap } from "./user-errors";

const STATEMENT_MAX = 2000;

interface AssumptionDialogProps {
  decisionId: string;
  /** When provided, the dialog edits this assumption instead of adding one. */
  assumption?: { id: string; statement: string; confidence: number };
  open: boolean;
  onOpenChange(open: boolean): void;
  onSaved(): void;
}

/** Add or edit an assumption, capturing confidence as a 0-100 percent. */
export function AssumptionDialog({
  decisionId,
  assumption,
  open,
  onOpenChange,
  onSaved,
}: AssumptionDialogProps) {
  const toast = useToast();
  const statementRef = useRef<HTMLTextAreaElement>(null);
  const isEdit = Boolean(assumption);

  const [statement, setStatement] = useState("");
  const [confidence, setConfidence] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [addAssumption, addState] = useAddAssumptionMutation();
  const [updateAssumption, updateState] = useUpdateAssumptionMutation();
  const loading = addState.loading || updateState.loading;

  useEffect(() => {
    if (open) {
      setStatement(assumption?.statement ?? "");
      setConfidence(
        assumption ? String(toPercentValue(assumption.confidence)) : "50",
      );
      setFieldErrors({});
      setFormError(null);
    }
  }, [open, assumption]);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const errors: Record<string, string> = {};
    const trimmed = statement.trim();
    if (!trimmed) {
      errors.statement = "Statement is required.";
    } else if (trimmed.length > STATEMENT_MAX) {
      errors.statement = `Statement must be ${STATEMENT_MAX} characters or fewer.`;
    }

    const parsed = fromPercent(confidence);
    if (parsed.error) {
      errors.confidence = parsed.error;
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      if (errors.statement) {
        statementRef.current?.focus();
      }
      return;
    }
    setFieldErrors({});

    const fraction = parsed.fraction as number;

    try {
      if (isEdit && assumption) {
        const { data } = await updateAssumption({
          variables: {
            input: { id: assumption.id, statement: trimmed, confidence: fraction },
          },
        });
        const payload = data?.updateAssumption;
        if (!payload) {
          setFormError("Something went wrong. Please try again.");
          return;
        }
        if (payload.userErrors.length > 0) {
          const mapped = userErrorsToMap(payload.userErrors);
          setFieldErrors(mapped.fields);
          setFormError(mapped.formError);
          return;
        }
        toast.success("Assumption updated");
      } else {
        const { data } = await addAssumption({
          variables: {
            input: { decisionId, statement: trimmed, confidence: fraction },
          },
        });
        const payload = data?.addAssumption;
        if (!payload) {
          setFormError("Something went wrong. Please try again.");
          return;
        }
        if (payload.userErrors.length > 0) {
          const mapped = userErrorsToMap(payload.userErrors);
          setFieldErrors(mapped.fields);
          setFormError(mapped.formError);
          return;
        }
        toast.success("Assumption added");
      }

      onSaved();
      onOpenChange(false);
    } catch {
      setFormError("Something went wrong. Please try again.");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit assumption" : "Add assumption"}</DialogTitle>
          <DialogDescription>
            State what you believe must hold, and how confident you are.
          </DialogDescription>
        </DialogHeader>

        <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
          {formError ? (
            <div
              role="alert"
              aria-live="assertive"
              className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
            >
              {formError}
            </div>
          ) : null}

          <FormField label="Statement" error={fieldErrors.statement} required>
            {(props) => (
              <Textarea
                {...props}
                ref={statementRef}
                value={statement}
                maxLength={STATEMENT_MAX}
                placeholder="We assume demand will hold through Q4."
                onChange={(event) => setStatement(event.target.value)}
              />
            )}
          </FormField>

          <FormField
            label="Confidence"
            error={fieldErrors.confidence}
            helperText="Drag to set how likely this assumption holds."
            required
          >
            {(props) => (
              <PercentSlider
                id={props.id}
                ariaDescribedby={props["aria-describedby"]}
                label="Confidence"
                value={Number(confidence) || 0}
                onChange={(v) => setConfidence(String(v))}
              />
            )}
          </FormField>

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" disabled={loading}>
                Cancel
              </Button>
            </DialogClose>
            <Button type="submit" loading={loading} disabled={loading}>
              {isEdit ? "Save assumption" : "Add assumption"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
