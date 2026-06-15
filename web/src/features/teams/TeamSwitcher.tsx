import { Building2, Check, ChevronDown } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentTeam } from "@/features/teams/use-current-team";
import { cn } from "@/lib/cn";

/** Topbar control for switching the active team scope. */
export function TeamSwitcher() {
  const { teams, currentTeam, currentTeamId, setCurrentTeamId, loading } =
    useCurrentTeam();

  if (loading && teams.length === 0) {
    return <Skeleton className="h-9 w-36" />;
  }

  if (teams.length === 0) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="max-w-44 gap-2"
          aria-label="Switch team"
        >
          <Building2 className="size-4 shrink-0 text-muted-foreground" />
          <span className="truncate">{currentTeam?.name ?? "Team"}</span>
          <ChevronDown className="size-4 shrink-0 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-52">
        <DropdownMenuLabel>Teams</DropdownMenuLabel>
        <DropdownMenuSeparator />
        {teams.map((team) => (
          <DropdownMenuItem
            key={team.id}
            onSelect={() => setCurrentTeamId(team.id)}
          >
            <Check
              className={cn(
                "size-4 shrink-0",
                team.id === currentTeamId ? "opacity-100" : "opacity-0",
              )}
            />
            <span className="truncate">{team.name}</span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
