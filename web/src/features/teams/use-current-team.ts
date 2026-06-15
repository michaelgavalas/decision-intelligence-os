import { useContext } from "react";

import {
  CurrentTeamContext,
  type CurrentTeamContextValue,
} from "./current-team";

/** Access the current-team context. Must be used within a CurrentTeamProvider. */
export function useCurrentTeam(): CurrentTeamContextValue {
  const context = useContext(CurrentTeamContext);
  if (!context) {
    throw new Error(
      "useCurrentTeam must be used within a CurrentTeamProvider",
    );
  }
  return context;
}
