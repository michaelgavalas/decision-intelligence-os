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
import { useToast } from "@/components/ui/toaster";
import {
  type Role,
  useAddTeamMemberMutation,
} from "@/graphql/generated/graphql";

import { userErrorsToMap } from "@/features/decisions/detail/user-errors";

import { ROLE_LABELS, ROLE_OPTIONS } from "./roles";

interface AddMemberDialogProps {
  teamId: string;
  open: boolean;
  onOpenChange(open: boolean): void;
  /** Refetch the team after a member is added. */
  onAdded(): void;
}

/** Modal form for adding a member to a team by user id. */
export function AddMemberDialog({
  teamId,
  open,
  onOpenChange,
  onAdded,
}: AddMemberDialogProps) {
  const toast = useToast();
  const userIdRef = useRef<HTMLInputElement>(null);

  const [userId, setUserId] = useState("");
  const [role, setRole] = useState<Role>("MEMBER");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [formError, setFormError] = useState<string | null>(null);

  const [addTeamMember, { loading }] = useAddTeamMemberMutation();

  useEffect(() => {
    if (open) {
      setUserId("");
      setRole("MEMBER");
      setFieldErrors({});
      setFormError(null);
    }
  }, [open]);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setFormError(null);

    const trimmed = userId.trim();
    if (!trimmed) {
      setFieldErrors({ userId: "User ID is required." });
      userIdRef.current?.focus();
      return;
    }
    setFieldErrors({});

    try {
      const { data } = await addTeamMember({
        variables: { input: { teamId, userId: trimmed, role } },
      });

      const payload = data?.addTeamMember;
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

      toast.success("Member added");
      onAdded();
      onOpenChange(false);
    } catch {
      setFormError("Something went wrong. Please try again.");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add member</DialogTitle>
          <DialogDescription>
            Add a user to this team and choose the role they hold.
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

          <FormField
            label="User ID"
            error={fieldErrors.userId}
            helperText="Enter the user's ID."
            required
          >
            {(props) => (
              <Input
                {...props}
                ref={userIdRef}
                value={userId}
                onChange={(event) => setUserId(event.target.value)}
              />
            )}
          </FormField>

          <FormField label="Role" error={fieldErrors.role}>
            {(props) => (
              <Select
                {...props}
                value={role}
                onChange={(event) => setRole(event.target.value as Role)}
              >
                {ROLE_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {ROLE_LABELS[option]}
                  </option>
                ))}
              </Select>
            )}
          </FormField>

          <DialogFooter>
            <DialogClose asChild>
              <Button type="button" variant="outline" disabled={loading}>
                Cancel
              </Button>
            </DialogClose>
            <Button type="submit" loading={loading} disabled={loading}>
              Add member
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
