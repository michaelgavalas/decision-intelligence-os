import { Flag } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toaster";
import {
  type DecisionStatus,
  type OutcomeFieldsFragment,
  useRecordOutcomeMutation,
} from "@/graphql/generated/graphql";
import { cn } from "@/lib/cn";
import { formatDate } from "@/lib/format";

import { dateInputToISO } from "./dates";
import { userErrorsToMap } from "./user-errors";

const SUMMARY_MAX = 5000;

interface OutcomeSectionProps {
  decisionId: string;
  status: DecisionStatus;
  outcome: OutcomeFieldsFragment | null | undefined;
  onChanged(): void;
}

/**
 * Shows a recorded outcome, or a form to record one. Recording an outcome moves
 * the decision to DECIDED. A DRAFT decision cannot be decided yet.
 */
export function OutcomeSection({
  decisionId,
  status,
  outcome,
  onChanged,
}: OutcomeSectionProps) {
  const toast = useToast();

  const today = new Date().toISOString().slice(0, 10);
  const [summary, setSummary] = useState("");
  const [success, setSuccess] = useState<boolean | null>(null);
  const [resolvedAt, setResolvedAt] = useState(today);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [recordOutcome, { loading }] = useRecordOutcomeMutation();

  const isDraft = status === "DRAFT";

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const errors: Record<string, string> = {};
    const trimmed = summary.trim();
    if (!trimmed) {
      errors.summary = "Summary is required.";
    } else if (trimmed.length > SUMMARY_MAX) {
      errors.summary = `Summary must be ${SUMMARY_MAX} characters or fewer.`;
    }
    if (success === null) {
      errors.success = "Choose whether the decision succeeded.";
    }
    const resolvedISO = dateInputToISO(resolvedAt);
    if (!resolvedISO) {
      errors.resolvedAt = "A resolution date is required.";
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }
    setFieldErrors({});

    try {
      const { data } = await recordOutcome({
        variables: {
          input: {
            decisionId,
            summary: trimmed,
            success: success as boolean,
            resolvedAt: resolvedISO as string,
          },
        },
      });
      const payload = data?.recordOutcome;
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
      toast.success("Outcome recorded");
      onChanged();
    } catch {
      setFormError("Something went wrong. Please try again.");
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Outcome</CardTitle>
      </CardHeader>
      <CardContent>
        {outcome ? (
          <div className="flex flex-col gap-3">
            <div className="flex items-center gap-3">
              <Badge variant={outcome.success ? "success" : "destructive"}>
                {outcome.success ? "Succeeded" : "Did not succeed"}
              </Badge>
              <span className="text-xs tabular-nums text-muted-foreground">
                Resolved {formatDate(outcome.resolvedAt)}
              </span>
            </div>
            <p className="whitespace-pre-wrap break-words text-sm text-foreground">
              {outcome.summary}
            </p>
          </div>
        ) : (
          <form
            className="flex flex-col gap-4"
            onSubmit={handleSubmit}
            noValidate
          >
            {isDraft ? (
              <div className="rounded-md border border-border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
                Move this decision out of draft before recording its outcome.
              </div>
            ) : null}

            {formError ? (
              <div
                role="alert"
                aria-live="assertive"
                className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive"
              >
                {formError}
              </div>
            ) : null}

            <FormField label="Summary" error={fieldErrors.summary} required>
              {(props) => (
                <Textarea
                  {...props}
                  value={summary}
                  maxLength={SUMMARY_MAX}
                  disabled={isDraft}
                  placeholder="What actually happened, and what did we learn?"
                  onChange={(event) => setSummary(event.target.value)}
                />
              )}
            </FormField>

            <FormField label="Result" error={fieldErrors.success} required>
              {() => (
                <div
                  role="radiogroup"
                  aria-label="Result"
                  className="inline-flex overflow-hidden rounded-md border border-border"
                >
                  <button
                    type="button"
                    role="radio"
                    aria-checked={success === true}
                    disabled={isDraft}
                    onClick={() => setSuccess(true)}
                    className={cn(
                      "cursor-pointer px-4 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50",
                      success === true
                        ? "bg-success/10 text-success"
                        : "text-muted-foreground hover:bg-muted",
                    )}
                  >
                    Succeeded
                  </button>
                  <button
                    type="button"
                    role="radio"
                    aria-checked={success === false}
                    disabled={isDraft}
                    onClick={() => setSuccess(false)}
                    className={cn(
                      "cursor-pointer border-l border-border px-4 py-2 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50",
                      success === false
                        ? "bg-destructive/10 text-destructive"
                        : "text-muted-foreground hover:bg-muted",
                    )}
                  >
                    Did not succeed
                  </button>
                </div>
              )}
            </FormField>

            <FormField
              label="Resolved on"
              error={fieldErrors.resolvedAt}
              required
            >
              {(props) => (
                <Input
                  {...props}
                  type="date"
                  className="w-48"
                  disabled={isDraft}
                  value={resolvedAt}
                  onChange={(event) => setResolvedAt(event.target.value)}
                />
              )}
            </FormField>

            <div className="flex justify-end">
              <Button
                type="submit"
                loading={loading}
                disabled={loading || isDraft}
              >
                <Flag className="size-4" />
                Record outcome
              </Button>
            </div>
          </form>
        )}
      </CardContent>
    </Card>
  );
}
