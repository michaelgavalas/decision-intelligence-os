import { useState } from "react";
import { UserPlus, Users } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Select } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useToast } from "@/components/ui/toaster";
import {
  type Role,
  type TeamDetailQuery,
  useChangeMemberRoleMutation,
  useRemoveTeamMemberMutation,
  useTeamDetailQuery,
} from "@/graphql/generated/graphql";

import { useAuth } from "@/features/auth/use-auth";
import { ConfirmDialog } from "@/features/decisions/detail/ConfirmDialog";
import { useCurrentTeam } from "@/features/teams/use-current-team";
import { formatDate } from "@/lib/format";

import { AddMemberDialog } from "./AddMemberDialog";
import { CreateTeamDialog } from "./CreateTeamDialog";
import { ROLE_BADGE_VARIANTS, ROLE_LABELS, ROLE_OPTIONS } from "./roles";

type TeamMember = NonNullable<TeamDetailQuery["team"]>["members"][number];

export function TeamsPage() {
  const { currentTeamId, setCurrentTeamId } = useCurrentTeam();
  const [createOpen, setCreateOpen] = useState(false);

  const { data, loading, error, refetch } = useTeamDetailQuery({
    variables: { id: currentTeamId ?? "" },
    skip: !currentTeamId,
  });

  const team = data?.team ?? null;

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Team</h1>
          {team ? (
            <p className="text-sm text-muted-foreground">{team.name}</p>
          ) : null}
        </div>
        <Button onClick={() => setCreateOpen(true)}>New team</Button>
      </header>

      {!currentTeamId ? (
        <EmptyState
          icon={Users}
          title="No team yet"
          description="Create a team to start collaborating on decisions."
          action={<Button onClick={() => setCreateOpen(true)}>New team</Button>}
        />
      ) : loading ? (
        <TeamSkeleton />
      ) : error ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-10 text-center">
            <p className="text-sm text-muted-foreground">
              We couldn&apos;t load this team.
            </p>
            <Button variant="outline" onClick={() => void refetch()}>
              Retry
            </Button>
          </CardContent>
        </Card>
      ) : !team ? (
        <EmptyState
          icon={Users}
          title="Team not found"
          description="This team may have been removed or you no longer have access."
        />
      ) : (
        <TeamMembers team={team} onChanged={() => void refetch()} />
      )}

      <CreateTeamDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        onCreated={(teamId) => setCurrentTeamId(teamId)}
      />
    </div>
  );
}

function TeamSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <Skeleton className="h-10 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
      <Skeleton className="h-12 w-full" />
    </div>
  );
}

interface TeamMembersProps {
  team: NonNullable<TeamDetailQuery["team"]>;
  onChanged(): void;
}

function TeamMembers({ team, onChanged }: TeamMembersProps) {
  const { user } = useAuth();
  const [addOpen, setAddOpen] = useState(false);

  const currentMembership = team.members.find(
    (member) => member.user.id === user?.id,
  );
  const isAdmin = currentMembership?.role === "ADMIN";

  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between gap-3 space-y-0">
        <CardTitle>Members</CardTitle>
        {isAdmin ? (
          <Button size="sm" variant="outline" onClick={() => setAddOpen(true)}>
            <UserPlus className="size-4" />
            Add member
          </Button>
        ) : null}
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Joined</TableHead>
              {isAdmin ? (
                <TableHead className="text-right">Actions</TableHead>
              ) : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {team.members.map((member) => (
              <MemberRow
                key={member.user.id}
                teamId={team.id}
                member={member}
                isAdmin={isAdmin}
                isSelf={member.user.id === user?.id}
                onChanged={onChanged}
              />
            ))}
          </TableBody>
        </Table>
      </CardContent>

      <AddMemberDialog
        teamId={team.id}
        open={addOpen}
        onOpenChange={setAddOpen}
        onAdded={onChanged}
      />
    </Card>
  );
}

interface MemberRowProps {
  teamId: string;
  member: TeamMember;
  isAdmin: boolean;
  isSelf: boolean;
  onChanged(): void;
}

function MemberRow({
  teamId,
  member,
  isAdmin,
  isSelf,
  onChanged,
}: MemberRowProps) {
  const toast = useToast();
  const [confirmRemove, setConfirmRemove] = useState(false);

  const [changeMemberRole, { loading: changing }] = useChangeMemberRoleMutation();
  const [removeTeamMember, { loading: removing }] =
    useRemoveTeamMemberMutation();

  async function handleRoleChange(role: Role) {
    if (role === member.role) {
      return;
    }
    try {
      const { data } = await changeMemberRole({
        variables: { input: { teamId, userId: member.user.id, role } },
      });

      const payload = data?.changeMemberRole;
      if (!payload) {
        toast.error("Something went wrong. Please try again.");
        return;
      }
      if (payload.userErrors.length > 0) {
        toast.error(payload.userErrors[0].message);
        return;
      }

      toast.success("Role updated");
      onChanged();
    } catch {
      toast.error("Something went wrong. Please try again.");
    }
  }

  async function handleRemove() {
    try {
      const { data } = await removeTeamMember({
        variables: { input: { teamId, userId: member.user.id } },
      });

      const payload = data?.removeTeamMember;
      if (!payload) {
        toast.error("Something went wrong. Please try again.");
        return;
      }
      if (payload.userErrors.length > 0) {
        toast.error(payload.userErrors[0].message);
        return;
      }

      toast.success("Member removed");
      setConfirmRemove(false);
      onChanged();
    } catch {
      toast.error("Something went wrong. Please try again.");
    }
  }

  return (
    <TableRow>
      <TableCell className="font-medium">
        {member.user.name}
        {isSelf ? (
          <span className="ml-2 text-xs text-muted-foreground">(you)</span>
        ) : null}
      </TableCell>
      <TableCell className="text-muted-foreground">
        {member.user.email}
      </TableCell>
      <TableCell>
        {isAdmin ? (
          <Select
            aria-label={`Role for ${member.user.name}`}
            className="h-9 w-32"
            value={member.role}
            disabled={changing}
            onChange={(event) =>
              void handleRoleChange(event.target.value as Role)
            }
          >
            {ROLE_OPTIONS.map((option) => (
              <option key={option} value={option}>
                {ROLE_LABELS[option]}
              </option>
            ))}
          </Select>
        ) : (
          <Badge variant={ROLE_BADGE_VARIANTS[member.role]}>
            {ROLE_LABELS[member.role]}
          </Badge>
        )}
      </TableCell>
      <TableCell className="text-muted-foreground">
        {formatDate(member.createdAt)}
      </TableCell>
      {isAdmin ? (
        <TableCell className="text-right">
          <Button
            size="sm"
            variant="ghost"
            className="text-destructive hover:bg-destructive/10"
            loading={removing}
            disabled={removing}
            onClick={() => setConfirmRemove(true)}
          >
            Remove
          </Button>
          <ConfirmDialog
            open={confirmRemove}
            onOpenChange={setConfirmRemove}
            title="Remove member"
            description={`Remove ${member.user.name} from this team? They will lose access to its decisions.`}
            confirmLabel="Remove"
            destructive
            loading={removing}
            onConfirm={() => void handleRemove()}
          />
        </TableCell>
      ) : null}
    </TableRow>
  );
}
