import type { MockedResponse } from "@apollo/client/testing";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import {
  CreateDecisionDocument,
  DecisionsDocument,
} from "@/graphql/generated/graphql";

const successToast = vi.fn();

vi.mock("@/components/ui/toaster", () => ({
  useToast: () => ({ success: successToast }),
}));

import { CreateDecisionDialog } from "@/features/decisions/CreateDecisionDialog";

import { renderWithProviders, TEST_TEAM } from "./test-utils";

const PAGE_SIZE = 20;

const emptyDecisionsRefetch: MockedResponse = {
  request: {
    query: DecisionsDocument,
    variables: { teamId: TEST_TEAM.id, first: PAGE_SIZE },
  },
  result: {
    data: {
      decisions: {
        __typename: "DecisionConnection",
        totalCount: 0,
        edges: [],
        pageInfo: {
          __typename: "PageInfo",
          hasNextPage: false,
          endCursor: null,
        },
      },
    },
  },
};

function createMock(): MockedResponse {
  return {
    request: {
      query: CreateDecisionDocument,
      variables: {
        input: {
          teamId: TEST_TEAM.id,
          title: "Adopt OKRs",
          description: "Trial quarterly OKRs",
        },
      },
    },
    result: {
      data: {
        createDecision: {
          __typename: "DecisionPayload",
          decision: {
            __typename: "Decision",
            id: "new-1",
            title: "Adopt OKRs",
            description: "Trial quarterly OKRs",
            status: "DRAFT",
            decidedAt: null,
            createdAt: "2026-06-14T00:00:00Z",
            updatedAt: "2026-06-14T00:00:00Z",
            owner: { __typename: "User", id: "u1", name: "Ada" },
          },
          userErrors: [],
        },
      },
    },
  };
}

describe("CreateDecisionDialog", () => {
  it("creates a decision, toasts, and closes on success", async () => {
    successToast.mockClear();
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(
      <CreateDecisionDialog
        teamId={TEST_TEAM.id}
        open
        onOpenChange={onOpenChange}
      />,
      { mocks: [createMock(), emptyDecisionsRefetch] },
    );

    await user.type(screen.getByLabelText(/title/i), "Adopt OKRs");
    await user.type(screen.getByLabelText(/description/i), "Trial quarterly OKRs");
    await user.click(
      screen.getByRole("button", { name: /create decision/i }),
    );

    await waitFor(() => {
      expect(successToast).toHaveBeenCalledWith("Decision created");
    });
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("blocks submission when the title is empty", async () => {
    successToast.mockClear();
    const onOpenChange = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(
      <CreateDecisionDialog
        teamId={TEST_TEAM.id}
        open
        onOpenChange={onOpenChange}
      />,
      { mocks: [] },
    );

    await user.click(
      screen.getByRole("button", { name: /create decision/i }),
    );

    expect(await screen.findByText("Title is required.")).toBeInTheDocument();
    expect(successToast).not.toHaveBeenCalled();
    expect(onOpenChange).not.toHaveBeenCalled();
  });
});
