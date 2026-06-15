import type { MockedResponse } from "@apollo/client/testing";
import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { DecisionsListPage } from "@/features/decisions/DecisionsListPage";
import { DecisionsDocument } from "@/graphql/generated/graphql";

import { renderWithProviders, TEST_TEAM } from "./test-utils";

const PAGE_SIZE = 20;

function decisionNode(overrides: Record<string, unknown> = {}) {
  return {
    __typename: "Decision",
    id: "d1",
    title: "Expand to the EU market",
    description: "Evaluate a European launch",
    status: "ACTIVE",
    decidedAt: null,
    createdAt: "2026-06-01T00:00:00Z",
    updatedAt: "2026-06-01T00:00:00Z",
    owner: { __typename: "User", id: "u1", name: "Ada" },
    ...overrides,
  };
}

function decisionsMock(
  nodes: ReturnType<typeof decisionNode>[],
): MockedResponse {
  return {
    request: {
      query: DecisionsDocument,
      variables: { teamId: TEST_TEAM.id, first: PAGE_SIZE },
    },
    result: {
      data: {
        decisions: {
          __typename: "DecisionConnection",
          totalCount: nodes.length,
          edges: nodes.map((node) => ({
            __typename: "DecisionEdge",
            cursor: `cursor-${node.id}`,
            node,
          })),
          pageInfo: {
            __typename: "PageInfo",
            hasNextPage: false,
            endCursor: nodes.length ? `cursor-${nodes[nodes.length - 1].id}` : null,
          },
        },
      },
    },
  };
}

describe("DecisionsListPage", () => {
  it("renders decisions with status badges", async () => {
    renderWithProviders(<DecisionsListPage />, {
      mocks: [
        decisionsMock([
          decisionNode(),
          decisionNode({ id: "d2", title: "Hire a head of sales", status: "DRAFT" }),
        ]),
      ],
    });

    expect(
      await screen.findByRole("link", { name: "Expand to the EU market" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Hire a head of sales" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("Draft")).toBeInTheDocument();
    expect(screen.getByText("2 decisions")).toBeInTheDocument();
  });

  it("shows an empty state when there are no decisions", async () => {
    renderWithProviders(<DecisionsListPage />, {
      mocks: [decisionsMock([])],
    });

    expect(await screen.findByText("No decisions yet")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /create your first decision/i }),
    ).toBeInTheDocument();
  });
});
