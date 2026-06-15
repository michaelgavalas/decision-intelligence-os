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
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/toaster";
import {
  type EvidenceSourceType,
  useAttachEvidenceMutation,
  useUpdateEvidenceMutation,
} from "@/graphql/generated/graphql";

import { userErrorsToMap } from "./user-errors";

const CONTENT_MAX = 5000;

const SOURCE_TYPE_OPTIONS: { value: EvidenceSourceType; label: string }[] = [
  { value: "URL", label: "URL" },
  { value: "DOCUMENT", label: "Document" },
  { value: "NOTE", label: "Note" },
  { value: "DATASET", label: "Dataset" },
];

interface EvidenceDialogProps {
  assumptionId: string;
  /** When provided, the dialog edits this evidence instead of attaching new. */
  evidence?: {
    id: string;
    sourceType: EvidenceSourceType;
    sourceUrl?: string | null;
    content: string;
  };
  open: boolean;
  onOpenChange(open: boolean): void;
  onSaved(): void;
}

/** Attach or edit a piece of evidence on an assumption. */
export function EvidenceDialog({
  assumptionId,
  evidence,
  open,
  onOpenChange,
  onSaved,
}: EvidenceDialogProps) {
  const toast = useToast();
  const contentRef = useRef<HTMLTextAreaElement>(null);
  const isEdit = Boolean(evidence);

  const [sourceType, setSourceType] = useState<EvidenceSourceType>("NOTE");
  const [sourceUrl, setSourceUrl] = useState("");
  const [content, setContent] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [attachEvidence, attachState] = useAttachEvidenceMutation();
  const [updateEvidence, updateState] = useUpdateEvidenceMutation();
  const loading = attachState.loading || updateState.loading;

  useEffect(() => {
    if (open) {
      setSourceType(evidence?.sourceType ?? "NOTE");
      setSourceUrl(evidence?.sourceUrl ?? "");
      setContent(evidence?.content ?? "");
      setFieldErrors({});
      setFormError(null);
    }
  }, [open, evidence]);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const errors: Record<string, string> = {};
    const trimmedContent = content.trim();
    const trimmedUrl = sourceUrl.trim();

    if (!trimmedContent) {
      errors.content = "Content is required.";
    } else if (trimmedContent.length > CONTENT_MAX) {
      errors.content = `Content must be ${CONTENT_MAX} characters or fewer.`;
    }

    if (sourceType === "URL" && !trimmedUrl) {
      errors.sourceUrl = "A URL is required for URL evidence.";
    } else if (trimmedUrl && !/^https?:\/\/.+/i.test(trimmedUrl)) {
      errors.sourceUrl = "Enter a valid http(s) URL.";
    }

    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors);
      return;
    }
    setFieldErrors({});

    const sourceUrlValue = trimmedUrl ? trimmedUrl : null;

    try {
      if (isEdit && evidence) {
        const { data } = await updateEvidence({
          variables: {
            input: {
              id: evidence.id,
              sourceType,
              sourceUrl: sourceUrlValue,
              content: trimmedContent,
            },
          },
        });
        const payload = data?.updateEvidence;
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
        toast.success("Evidence updated");
      } else {
        const { data } = await attachEvidence({
          variables: {
            input: {
              assumptionId,
              sourceType,
              sourceUrl: sourceUrlValue,
              content: trimmedContent,
            },
          },
        });
        const payload = data?.attachEvidence;
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
        toast.success("Evidence attached");
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
          <DialogTitle>{isEdit ? "Edit evidence" : "Attach evidence"}</DialogTitle>
          <DialogDescription>
            Record a source that supports or challenges this assumption.
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

          <FormField label="Source type" error={fieldErrors.sourceType} required>
            {(props) => (
              <Select
                {...props}
                value={sourceType}
                onChange={(event) =>
                  setSourceType(event.target.value as EvidenceSourceType)
                }
              >
                {SOURCE_TYPE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Select>
            )}
          </FormField>

          <FormField
            label="URL"
            error={fieldErrors.sourceUrl}
            helperText={
              sourceType === "URL"
                ? "Required for URL evidence."
                : "Optional link to the source."
            }
            required={sourceType === "URL"}
          >
            {(props) => (
              <Input
                {...props}
                type="url"
                value={sourceUrl}
                placeholder="https://example.com/report"
                onChange={(event) => setSourceUrl(event.target.value)}
              />
            )}
          </FormField>

          <FormField label="Content" error={fieldErrors.content} required>
            {(props) => (
              <Textarea
                {...props}
                ref={contentRef}
                value={content}
                maxLength={CONTENT_MAX}
                placeholder="Summarize what this evidence shows."
                onChange={(event) => setContent(event.target.value)}
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
              {isEdit ? "Save evidence" : "Attach evidence"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
