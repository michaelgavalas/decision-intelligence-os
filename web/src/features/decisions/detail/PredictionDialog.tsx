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
import { DatePicker } from "@/components/ui/date-picker";
import { FormField } from "@/components/ui/form-field";
import { Slider } from "@/components/ui/slider";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toaster";
import {
  useCreatePredictionMutation,
  useUpdatePredictionMutation,
} from "@/graphql/generated/graphql";

import { dateInputValue, dateInputToISO } from "./dates";
import { fromPercent, toPercentValue } from "./percent";
import { userErrorsToMap } from "./user-errors";

const STATEMENT_MAX = 2000;

interface PredictionDialogProps {
  decisionId: string;
  /** When provided, the dialog edits this prediction instead of creating one. */
  prediction?: {
    id: string;
    statement: string;
    probability: number;
    resolvesAt?: string | null;
  };
  open: boolean;
  onOpenChange(open: boolean): void;
  onSaved(): void;
}

/** Create or edit a forecast, with probability captured as a 0-100 percent. */
export function PredictionDialog({
  decisionId,
  prediction,
  open,
  onOpenChange,
  onSaved,
}: PredictionDialogProps) {
  const toast = useToast();
  const statementRef = useRef<HTMLTextAreaElement>(null);
  const isEdit = Boolean(prediction);

  const [statement, setStatement] = useState("");
  const [probability, setProbability] = useState("");
  const [resolvesAt, setResolvesAt] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [createPrediction, createState] = useCreatePredictionMutation();
  const [updatePrediction, updateState] = useUpdatePredictionMutation();
  const loading = createState.loading || updateState.loading;

  useEffect(() => {
    if (open) {
      setStatement(prediction?.statement ?? "");
      setProbability(
        prediction ? String(toPercentValue(prediction.probability)) : "50",
      );
      setResolvesAt(dateInputValue(prediction?.resolvesAt));
      setFieldErrors({});
      setFormError(null);
    }
  }, [open, prediction]);

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

    const parsed = fromPercent(probability);
    if (parsed.error) {
      errors.probability = parsed.error;
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
    const resolvesAtISO = dateInputToISO(resolvesAt);

    try {
      if (isEdit && prediction) {
        const { data } = await updatePrediction({
          variables: {
            input: {
              id: prediction.id,
              statement: trimmed,
              probability: fraction,
              resolvesAt: resolvesAtISO,
            },
          },
        });
        const payload = data?.updatePrediction;
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
        toast.success("Prediction updated");
      } else {
        const { data } = await createPrediction({
          variables: {
            input: {
              decisionId,
              statement: trimmed,
              probability: fraction,
              resolvesAt: resolvesAtISO,
            },
          },
        });
        const payload = data?.createPrediction;
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
        toast.success("Prediction created");
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
          <DialogTitle>{isEdit ? "Edit prediction" : "Add prediction"}</DialogTitle>
          <DialogDescription>
            Forecast an outcome and how likely you think it is.
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
                placeholder="We will hit 1,000 paying customers by year end."
                onChange={(event) => setStatement(event.target.value)}
              />
            )}
          </FormField>

          <FormField
            label="Probability"
            error={fieldErrors.probability}
            helperText="Drag to set your estimated likelihood."
            required
          >
            {(props) => (
              <div className="flex items-center gap-4 pt-1">
                <Slider
                  id={props.id}
                  aria-describedby={props["aria-describedby"]}
                  thumbLabel="Probability percent"
                  min={0}
                  max={100}
                  step={1}
                  value={[Number(probability) || 0]}
                  onValueChange={([v]) => setProbability(String(v))}
                  className="flex-1"
                />
                <span className="w-12 text-right text-base font-semibold tabular-nums text-foreground">
                  {Number(probability) || 0}%
                </span>
              </div>
            )}
          </FormField>

          <FormField
            label="Resolves on"
            error={fieldErrors.resolvesAt}
            helperText="Optional date when this forecast can be settled."
          >
            {(props) => (
              <DatePicker
                id={props.id}
                ariaDescribedby={props["aria-describedby"]}
                value={resolvesAt}
                onChange={setResolvesAt}
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
              {isEdit ? "Save prediction" : "Add prediction"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
