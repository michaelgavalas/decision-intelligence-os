import { MockedProvider, type MockedResponse } from "@apollo/client/testing";
import { render } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import {
  CurrentTeamContext,
  type CurrentTeamContextValue,
} from "@/features/teams/current-team";

export const TEST_TEAM = { id: "team-1", name: "Acme" };

/** A stub current-team context value backed by a single test team. */
export function stubCurrentTeam(
  overrides: Partial<CurrentTeamContextValue> = {},
): CurrentTeamContextValue {
  return {
    teams: [TEST_TEAM],
    currentTeamId: TEST_TEAM.id,
    currentTeam: TEST_TEAM,
    setCurrentTeamId: () => {},
    loading: false,
    ...overrides,
  };
}

interface RenderOptions {
  mocks?: MockedResponse[];
  team?: CurrentTeamContextValue;
}

/**
 * Renders a decisions feature component wrapped in a MockedProvider, a router,
 * and a stubbed current-team context so tests need not mock MyTeams.
 */
export function renderWithProviders(
  ui: React.ReactNode,
  { mocks = [], team = stubCurrentTeam() }: RenderOptions = {},
) {
  return render(
    <MockedProvider mocks={mocks}>
      <CurrentTeamContext.Provider value={team}>
        <MemoryRouter initialEntries={["/decisions"]}>
          <Routes>
            <Route path="/decisions" element={ui} />
            <Route
              path="/decisions/:id"
              element={<div>Decision detail</div>}
            />
          </Routes>
        </MemoryRouter>
      </CurrentTeamContext.Provider>
    </MockedProvider>,
  );
}
