import { MockedProvider, type MockedResponse } from "@apollo/client/testing";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import {
  AddAssumptionDocument,
  DecisionDocument,
  TransitionDecisionDocument,
} from "@/graphql/generated/graphql";

const successToast = vi.fn();
const errorToast = vi.fn();

vi.mock("@/components/ui/toaster", () => ({
  useToast: () => ({ success: successToast, error: errorToast }),
}));

import { DecisionDetailPage } from "./DecisionDetailPage";

const DECISION_ID = "dec-1";

function decisionData(overrides: Record<string, unknown> = {}) {
  return {
    decision: {
      __typename: "Decision",
      id: DECISION_ID,
      title: "Expand to the EU market",
      description: "Weigh the cost and upside of an EU launch.",
      status: "ACTIVE",
      decidedAt: null,
      createdAt: "2026-06-01T00:00:00Z",
      updatedAt: "2026-06-01T00:00:00Z",
      owner: { __typename: "User", id: "u1", name: "Ada", email: "ada@x.io" },
      team: { __typename: "Team", id: "team-1", name: "Acme" },
      assumptions: [
        {
          __typename: "Assumption",
          id: "asm-1",
          statement: "Demand will hold through Q4.",
          confidence: 0.7,
          createdAt: "2026-06-02T00:00:00Z",
          evidence: [
            {
              __typename: "Evidence",
              id: "ev-1",
              sourceType: "URL",
              sourceUrl: "https://example.com/report",
              content: "Industry report projects growth.",
              createdAt: "2026-06-03T00:00:00Z",
            },
          ],
        },
      ],
      predictions: [
        {
          __typename: "Prediction",
          id: "pred-1",
          statement: "We hit 1,000 customers by year end.",
          probability: 0.4,
          resolvesAt: "2026-12-31T00:00:00Z",
          createdAt: "2026-06-02T00:00:00Z",
        },
      ],
      outcome: null,
      ...overrides,
    },
  };
}

function decisionMock(
  overrides: Record<string, unknown> = {},
): MockedResponse {
  return {
    request: { query: DecisionDocument, variables: { id: DECISION_ID } },
    result: { data: decisionData(overrides) },
  };
}

function renderPage(mocks: MockedResponse[]) {
  return render(
    <MockedProvider mocks={mocks}>
      <MemoryRouter initialEntries={[`/decisions/${DECISION_ID}`]}>
        <Routes>
          <Route path="/decisions" element={<div>Decisions list</div>} />
          <Route path="/decisions/:id" element={<DecisionDetailPage />} />
        </Routes>
      </MemoryRouter>
    </MockedProvider>,
  );
}

describe("DecisionDetailPage", () => {
  it("renders nested decision data from the query", async () => {
    renderPage([decisionMock()]);

    expect(
      await screen.findByText("Expand to the EU market"),
    ).toBeInTheDocument();

    // Status badge.
    expect(screen.getByText("Active")).toBeInTheDocument();

    // Assumption statement, its confidence percent, and its evidence.
    expect(
      screen.getByText("Demand will hold through Q4."),
    ).toBeInTheDocument();
    expect(screen.getByText(/Confidence 70%/)).toBeInTheDocument();
    expect(
      screen.getByText("Industry report projects growth."),
    ).toBeInTheDocument();

    // Prediction probability percent.
    expect(screen.getByText(/Probability 40%/)).toBeInTheDocument();

    // Outcome record form is shown when no outcome exists.
    expect(
      screen.getByRole("button", { name: /record outcome/i }),
    ).toBeInTheDocument();
  });

  it("renders not-found when the decision is missing", async () => {
    const nullMock: MockedResponse = {
      request: { query: DecisionDocument, variables: { id: DECISION_ID } },
      result: { data: { decision: null } },
    };
    renderPage([nullMock]);

    expect(
      await screen.findByText("Decision not found"),
    ).toBeInTheDocument();
  });

  it("shows a recorded outcome instead of the form", async () => {
    renderPage([
      decisionMock({
        status: "DECIDED",
        decidedAt: "2026-06-10T00:00:00Z",
        outcome: {
          __typename: "Outcome",
          id: "out-1",
          summary: "Launched and met targets.",
          success: true,
          resolvedAt: "2026-06-10T00:00:00Z",
        },
      }),
    ]);

    expect(await screen.findByText("Succeeded")).toBeInTheDocument();
    expect(
      screen.getByText("Launched and met targets."),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: /record outcome/i }),
    ).not.toBeInTheDocument();
  });

  it("blocks adding an assumption with an empty statement", async () => {
    successToast.mockClear();
    const user = userEvent.setup();
    renderPage([decisionMock()]);

    await screen.findByText("Expand to the EU market");

    await user.click(
      screen.getByRole("button", { name: /add assumption/i }),
    );

    const dialog = await screen.findByRole("dialog");
    // Clear the prefilled confidence is fine; submit with empty statement.
    await user.click(
      within(dialog).getByRole("button", { name: /^add assumption$/i }),
    );

    expect(
      await within(dialog).findByText("Statement is required."),
    ).toBeInTheDocument();
    expect(successToast).not.toHaveBeenCalled();
  });

  it("adds an assumption and refetches on success", async () => {
    successToast.mockClear();
    const user = userEvent.setup();

    const addMock: MockedResponse = {
      request: {
        query: AddAssumptionDocument,
        variables: {
          input: {
            decisionId: DECISION_ID,
            statement: "Pricing stays competitive.",
            confidence: 0.5,
          },
        },
      },
      result: {
        data: {
          addAssumption: {
            __typename: "AssumptionPayload",
            assumption: {
              __typename: "Assumption",
              id: "asm-2",
              statement: "Pricing stays competitive.",
              confidence: 0.5,
              createdAt: "2026-06-05T00:00:00Z",
              evidence: [],
            },
            userErrors: [],
          },
        },
      },
    };

    renderPage([decisionMock(), addMock, decisionMock()]);

    await screen.findByText("Expand to the EU market");

    await user.click(
      screen.getByRole("button", { name: /add assumption/i }),
    );

    const dialog = await screen.findByRole("dialog");
    await user.type(
      within(dialog).getByLabelText(/statement/i),
      "Pricing stays competitive.",
    );
    // Confidence is a slider; the default (50% = 0.5) is submitted as-is.

    await user.click(
      within(dialog).getByRole("button", { name: /^add assumption$/i }),
    );

    await waitFor(() => {
      expect(successToast).toHaveBeenCalledWith("Assumption added");
    });
  });

  it("transitions the decision via the move-to menu", async () => {
    successToast.mockClear();
    const user = userEvent.setup();

    const transitionMock: MockedResponse = {
      request: {
        query: TransitionDecisionDocument,
        variables: { input: { id: DECISION_ID, status: "DECIDED" } },
      },
      result: {
        data: {
          transitionDecision: {
            __typename: "DecisionPayload",
            decision: {
              __typename: "Decision",
              id: DECISION_ID,
              status: "DECIDED",
              decidedAt: "2026-06-14T00:00:00Z",
            },
            userErrors: [],
          },
        },
      },
    };

    renderPage([decisionMock(), transitionMock, decisionMock()]);

    await screen.findByText("Expand to the EU market");

    await user.click(screen.getByRole("button", { name: /move to/i }));
    await user.click(
      await screen.findByRole("menuitem", { name: /decided/i }),
    );

    await waitFor(() => {
      expect(successToast).toHaveBeenCalledWith("Moved to Decided");
    });
  });
});
