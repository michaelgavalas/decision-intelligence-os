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
import { useToast } from "@/components/ui/toaster";
import {
  MyTeamsDocument,
  useCreateTeamMutation,
} from "@/graphql/generated/graphql";

import { userErrorsToMap } from "@/features/decisions/detail/user-errors";

const NAME_MAX = 200;

interface CreateTeamDialogProps {
  open: boolean;
  onOpenChange(open: boolean): void;
  /** Called with the new team's id after a successful creation. */
  onCreated(teamId: string): void;
}

/** Modal form for creating a new team. */
export function CreateTeamDialog({
  open,
  onOpenChange,
  onCreated,
}: CreateTeamDialogProps) {
  const toast = useToast();
  const nameRef = useRef<HTMLInputElement>(null);

  const [name, setName] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [createTeam, { loading }] = useCreateTeamMutation({
    refetchQueries: [{ query: MyTeamsDocument }],
    awaitRefetchQueries: true,
  });

  // Reset the form each time the dialog opens.
  useEffect(() => {
    if (open) {
      setName("");
      setFieldErrors({});
      setFormError(null);
    }
  }, [open]);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const trimmed = name.trim();
    if (!trimmed) {
      setFieldErrors({ name: "Name is required." });
      nameRef.current?.focus();
      return;
    }
    if (trimmed.length > NAME_MAX) {
      setFieldErrors({ name: `Name must be ${NAME_MAX} characters or fewer.` });
      nameRef.current?.focus();
      return;
    }
    setFieldErrors({});

    try {
      const { data } = await createTeam({
        variables: { input: { name: trimmed } },
      });

      const payload = data?.createTeam;
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

      if (!payload.team) {
        setFormError("Something went wrong. Please try again.");
        return;
      }

      toast.success("Team created");
      onCreated(payload.team.id);
      onOpenChange(false);
    } catch {
      setFormError("Something went wrong. Please try again.");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>New team</DialogTitle>
          <DialogDescription>
            Create a team to group the people you make decisions with.
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

          <FormField label="Name" error={fieldErrors.name} required>
            {(props) => (
              <Input
                {...props}
                ref={nameRef}
                value={name}
                maxLength={NAME_MAX}
                onChange={(event) => setName(event.target.value)}
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
              Create team
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
