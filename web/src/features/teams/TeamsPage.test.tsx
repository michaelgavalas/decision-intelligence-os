import { MockedProvider, type MockedResponse } from "@apollo/client/testing";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import {
  ChangeMemberRoleDocument,
  CreateTeamDocument,
  MyTeamsDocument,
  TeamDetailDocument,
} from "@/graphql/generated/graphql";

const successToast = vi.fn();
const errorToast = vi.fn();

vi.mock("@/components/ui/toaster", () => ({
  useToast: () => ({ success: successToast, error: errorToast }),
}));

const setCurrentTeamId = vi.fn();

vi.mock("@/features/teams/use-current-team", () => ({
  useCurrentTeam: () => ({
    teams: [{ id: "team-1", name: "Acme" }],
    currentTeamId: "team-1",
    currentTeam: { id: "team-1", name: "Acme" },
    setCurrentTeamId,
    loading: false,
  }),
}));

vi.mock("@/features/auth/use-auth", () => ({
  useAuth: () => ({
    user: { id: "u1", name: "Ada", email: "ada@x.io" },
    status: "authenticated",
  }),
}));

import { TeamsPage } from "./TeamsPage";

const TEAM_ID = "team-1";

function teamData(adminId = "u1") {
  return {
    team: {
      __typename: "Team",
      id: TEAM_ID,
      name: "Acme",
      createdAt: "2026-06-01T00:00:00Z",
      members: [
        {
          __typename: "Membership",
          user: {
            __typename: "User",
            id: adminId,
            name: "Ada",
            email: "ada@x.io",
          },
          role: "ADMIN",
          createdAt: "2026-06-01T00:00:00Z",
        },
        {
          __typename: "Membership",
          user: {
            __typename: "User",
            id: "u2",
            name: "Bo",
            email: "bo@x.io",
          },
          role: "MEMBER",
          createdAt: "2026-06-02T00:00:00Z",
        },
      ],
    },
  };
}

function teamMock(adminId = "u1"): MockedResponse {
  return {
    request: { query: TeamDetailDocument, variables: { id: TEAM_ID } },
    result: { data: teamData(adminId) },
  };
}

function renderPage(mocks: MockedResponse[]) {
  return render(<MockedProvider mocks={mocks}>{<TeamsPage />}</MockedProvider>);
}

describe("TeamsPage", () => {
  it("renders the members table from the query", async () => {
    renderPage([teamMock()]);

    expect(await screen.findByText("Bo")).toBeInTheDocument();
    expect(screen.getByText("bo@x.io")).toBeInTheDocument();
    // The current user (Ada) is an admin, so roles render as selects.
    expect(
      screen.getByRole("combobox", { name: /role for ada/i }),
    ).toHaveValue("ADMIN");
    expect(
      screen.getByRole("combobox", { name: /role for bo/i }),
    ).toHaveValue("MEMBER");
  });

  it("changes a member role and toasts on success", async () => {
    successToast.mockClear();
    const user = userEvent.setup();

    const changeMock: MockedResponse = {
      request: {
        query: ChangeMemberRoleDocument,
        variables: { input: { teamId: TEAM_ID, userId: "u2", role: "ADMIN" } },
      },
      result: {
        data: {
          changeMemberRole: {
            __typename: "ChangeMemberRolePayload",
            membership: {
              __typename: "Membership",
              user: {
                __typename: "User",
                id: "u2",
                name: "Bo",
                email: "bo@x.io",
              },
              role: "ADMIN",
            },
            userErrors: [],
          },
        },
      },
    };

    renderPage([teamMock(), changeMock, teamMock()]);

    await screen.findByText("Bo");

    await user.selectOptions(
      screen.getByRole("combobox", { name: /role for bo/i }),
      "ADMIN",
    );

    await waitFor(() => {
      expect(successToast).toHaveBeenCalledWith("Role updated");
    });
  });

  it("surfaces a LAST_ADMIN user error when changing a role", async () => {
    successToast.mockClear();
    errorToast.mockClear();
    const user = userEvent.setup();

    const changeMock: MockedResponse = {
      request: {
        query: ChangeMemberRoleDocument,
        variables: {
          input: { teamId: TEAM_ID, userId: "u1", role: "MEMBER" },
        },
      },
      result: {
        data: {
          changeMemberRole: {
            __typename: "ChangeMemberRolePayload",
            membership: null,
            userErrors: [
              {
                __typename: "UserError",
                field: null,
                message: "A team must keep at least one admin.",
                code: "LAST_ADMIN",
              },
            ],
          },
        },
      },
    };

    renderPage([teamMock(), changeMock]);

    await screen.findByText("Bo");

    await user.selectOptions(
      screen.getByRole("combobox", { name: /role for ada/i }),
      "MEMBER",
    );

    await waitFor(() => {
      expect(errorToast).toHaveBeenCalledWith(
        "A team must keep at least one admin.",
      );
    });
    expect(successToast).not.toHaveBeenCalled();
  });

  it("creates a team and switches to it on success", async () => {
    successToast.mockClear();
    setCurrentTeamId.mockClear();
    const user = userEvent.setup();

    const createMock: MockedResponse = {
      request: {
        query: CreateTeamDocument,
        variables: { input: { name: "Beta" } },
      },
      result: {
        data: {
          createTeam: {
            __typename: "CreateTeamPayload",
            team: { __typename: "Team", id: "team-2", name: "Beta" },
            userErrors: [],
          },
        },
      },
    };

    const myTeamsMock: MockedResponse = {
      request: { query: MyTeamsDocument },
      result: {
        data: {
          myTeams: [
            {
              __typename: "Team",
              id: "team-1",
              name: "Acme",
              createdAt: "2026-06-01T00:00:00Z",
            },
            {
              __typename: "Team",
              id: "team-2",
              name: "Beta",
              createdAt: "2026-06-10T00:00:00Z",
            },
          ],
        },
      },
    };

    renderPage([teamMock(), createMock, myTeamsMock]);

    await screen.findByText("Bo");

    await user.click(screen.getByRole("button", { name: /new team/i }));

    const dialog = await screen.findByRole("dialog");
    await user.type(within(dialog).getByLabelText(/name/i), "Beta");
    await user.click(
      within(dialog).getByRole("button", { name: /create team/i }),
    );

    await waitFor(() => {
      expect(setCurrentTeamId).toHaveBeenCalledWith("team-2");
    });
    expect(successToast).toHaveBeenCalledWith("Team created");
  });
});
