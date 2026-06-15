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
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toaster";
import {
  DecisionsDocument,
  useCreateDecisionMutation,
} from "@/graphql/generated/graphql";

const TITLE_MAX = 200;
const DECISIONS_PAGE_SIZE = 20;

interface FieldErrors {
  title?: string;
  description?: string;
}

interface CreateDecisionDialogProps {
  teamId: string;
  open: boolean;
  onOpenChange(open: boolean): void;
}

/**
 * Modal form for creating a decision within the active team. Validates the
 * title client-side, surfaces server-side `userErrors` inline, and refetches
 * the decisions list so the new row appears on success.
 */
export function CreateDecisionDialog({
  teamId,
  open,
  onOpenChange,
}: CreateDecisionDialogProps) {
  const toast = useToast();
  const titleRef = useRef<HTMLInputElement>(null);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [createDecision, { loading }] = useCreateDecisionMutation();

  // Reset the form whenever the dialog closes so it opens clean next time.
  useEffect(() => {
    if (!open) {
      setTitle("");
      setDescription("");
      setFieldErrors({});
      setFormError(null);
    }
  }, [open]);

  function validate(): FieldErrors {
    const errors: FieldErrors = {};
    const trimmed = title.trim();
    if (!trimmed) {
      errors.title = "Title is required.";
    } else if (trimmed.length > TITLE_MAX) {
      errors.title = `Title must be ${TITLE_MAX} characters or fewer.`;
    }
    return errors;
  }

  function mapUserErrors(
    errors: ReadonlyArray<{ field?: string | null; message: string }>,
  ) {
    const next: FieldErrors = {};
    const formMessages: string[] = [];
    for (const error of errors) {
      if (error.field === "title") {
        next.title = error.message;
      } else if (error.field === "description") {
        next.description = error.message;
      } else {
        formMessages.push(error.message);
      }
    }
    setFieldErrors(next);
    setFormError(formMessages[0] ?? null);
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setFieldErrors(validationErrors);
      titleRef.current?.focus();
      return;
    }
    setFieldErrors({});

    try {
      const { data } = await createDecision({
        variables: {
          input: {
            teamId,
            title: title.trim(),
            description: description.trim(),
          },
        },
        refetchQueries: [
          {
            query: DecisionsDocument,
            variables: { teamId, first: DECISIONS_PAGE_SIZE },
          },
        ],
      });

      const payload = data?.createDecision;
      if (!payload) {
        setFormError("Something went wrong. Please try again.");
        return;
      }

      if (payload.userErrors.length > 0) {
        mapUserErrors(payload.userErrors);
        return;
      }

      toast.success("Decision created");
      onOpenChange(false);
    } catch {
      setFormError("Something went wrong. Please try again.");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New decision</DialogTitle>
          <DialogDescription>
            Capture a decision to track through its lifecycle.
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

          <FormField label="Title" error={fieldErrors.title} required>
            {(props) => (
              <Input
                {...props}
                ref={titleRef}
                value={title}
                maxLength={TITLE_MAX}
                placeholder="Should we expand to the EU market?"
                onChange={(event) => setTitle(event.target.value)}
              />
            )}
          </FormField>

          <FormField
            label="Description"
            error={fieldErrors.description}
            helperText="Optional context for this decision."
          >
            {(props) => (
              <Textarea
                {...props}
                value={description}
                placeholder="What's the situation, and why now?"
                onChange={(event) => setDescription(event.target.value)}
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
              Create decision
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
