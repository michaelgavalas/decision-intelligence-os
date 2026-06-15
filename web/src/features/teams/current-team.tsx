import { createContext, useCallback, useEffect, useMemo, useState } from "react";

import { useMyTeamsQuery } from "@/graphql/generated/graphql";
import { useAuth } from "@/features/auth/use-auth";

export interface CurrentTeam {
  id: string;
  name: string;
}

export interface CurrentTeamContextValue {
  teams: CurrentTeam[];
  currentTeamId: string | null;
  currentTeam: CurrentTeam | null;
  setCurrentTeamId(id: string): void;
  loading: boolean;
}

export const CurrentTeamContext = createContext<CurrentTeamContextValue | null>(
  null,
);

/** localStorage key persisting the operator's most recently selected team. */
const STORAGE_KEY = "dios-current-team";

function readStoredTeamId(): string | null {
  try {
    return window.localStorage.getItem(STORAGE_KEY);
  } catch {
    return null;
  }
}

function storeTeamId(id: string): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, id);
  } catch {
    // Persistence is best-effort; ignore storage failures (e.g. private mode).
  }
}

export function CurrentTeamProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const { status } = useAuth();
  const isAuthenticated = status === "authenticated";

  const { data, loading } = useMyTeamsQuery({ skip: !isAuthenticated });

  const teams = useMemo<CurrentTeam[]>(
    () =>
      (data?.myTeams ?? []).map((team) => ({ id: team.id, name: team.name })),
    [data?.myTeams],
  );

  const [selectedId, setSelectedId] = useState<string | null>(() =>
    readStoredTeamId(),
  );

  // Resolve the active team: prefer the user's selection when it still exists
  // in the list, otherwise fall back to the first team.
  const currentTeamId = useMemo<string | null>(() => {
    if (teams.length === 0) {
      return null;
    }
    if (selectedId && teams.some((team) => team.id === selectedId)) {
      return selectedId;
    }
    return teams[0].id;
  }, [teams, selectedId]);

  // Keep state and storage aligned once a valid team is resolved.
  useEffect(() => {
    if (currentTeamId && currentTeamId !== selectedId) {
      setSelectedId(currentTeamId);
      storeTeamId(currentTeamId);
    }
  }, [currentTeamId, selectedId]);

  const setCurrentTeamId = useCallback((id: string) => {
    setSelectedId(id);
    storeTeamId(id);
  }, []);

  const currentTeam = useMemo<CurrentTeam | null>(
    () => teams.find((team) => team.id === currentTeamId) ?? null,
    [teams, currentTeamId],
  );

  const value = useMemo<CurrentTeamContextValue>(
    () => ({
      teams,
      currentTeamId,
      currentTeam,
      setCurrentTeamId,
      loading: isAuthenticated && loading,
    }),
    [teams, currentTeamId, currentTeam, setCurrentTeamId, isAuthenticated, loading],
  );

  return (
    <CurrentTeamContext.Provider value={value}>
      {children}
    </CurrentTeamContext.Provider>
  );
}
