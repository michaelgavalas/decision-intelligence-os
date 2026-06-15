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
import { useUpdateDecisionMutation } from "@/graphql/generated/graphql";

import { userErrorsToMap } from "./user-errors";

const TITLE_MAX = 200;

interface EditDecisionDialogProps {
  decisionId: string;
  initialTitle: string;
  initialDescription: string;
  open: boolean;
  onOpenChange(open: boolean): void;
  onSaved(): void;
}

/** Modal form for editing a decision's title and description. */
export function EditDecisionDialog({
  decisionId,
  initialTitle,
  initialDescription,
  open,
  onOpenChange,
  onSaved,
}: EditDecisionDialogProps) {
  const toast = useToast();
  const titleRef = useRef<HTMLInputElement>(null);

  const [title, setTitle] = useState(initialTitle);
  const [description, setDescription] = useState(initialDescription);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [updateDecision, { loading }] = useUpdateDecisionMutation();

  // Re-seed the form from the latest decision each time the dialog opens.
  useEffect(() => {
    if (open) {
      setTitle(initialTitle);
      setDescription(initialDescription);
      setFieldErrors({});
      setFormError(null);
    }
  }, [open, initialTitle, initialDescription]);

  function validate(): Record<string, string> {
    const errors: Record<string, string> = {};
    const trimmed = title.trim();
    if (!trimmed) {
      errors.title = "Title is required.";
    } else if (trimmed.length > TITLE_MAX) {
      errors.title = `Title must be ${TITLE_MAX} characters or fewer.`;
    }
    return errors;
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
      const { data } = await updateDecision({
        variables: {
          input: {
            id: decisionId,
            title: title.trim(),
            description: description.trim(),
          },
        },
      });

      const payload = data?.updateDecision;
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

      toast.success("Decision updated");
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
          <DialogTitle>Edit decision</DialogTitle>
          <DialogDescription>
            Update the title and description for this decision.
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
                onChange={(event) => setTitle(event.target.value)}
              />
            )}
          </FormField>

          <FormField label="Description" error={fieldErrors.description}>
            {(props) => (
              <Textarea
                {...props}
                value={description}
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
              Save changes
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
