import { MockedProvider, type MockedResponse } from "@apollo/client/testing";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import {
  CurrentTeamContext,
  type CurrentTeamContextValue,
} from "@/features/teams/current-team";
import {
  CalibrationDocument,
  TeamMetricsDocument,
} from "@/graphql/generated/graphql";

import { AnalyticsDashboard } from "./AnalyticsDashboard";

// Recharts' ResponsiveContainer renders nothing without a measured size in
// jsdom, so give it a deterministic box for the chart elements to mount into.
vi.mock("recharts", async () => {
  const actual = await vi.importActual<typeof import("recharts")>("recharts");
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: ReactNode }) => (
      <div style={{ width: 800, height: 320 }}>{children}</div>
    ),
  };
});

const TEST_TEAM = { id: "team-1", name: "Acme" };

function stubTeam(
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

function metricsMock(
  metrics: {
    brierScore: number;
    forecastCount: number;
    decisionSuccessRate: number;
    resolvedDecisionCount: number;
  },
): MockedResponse {
  return {
    request: {
      query: TeamMetricsDocument,
      variables: { teamId: TEST_TEAM.id },
    },
    result: {
      data: {
        teamMetrics: { __typename: "TeamMetrics", ...metrics },
      },
    },
  };
}

function calibrationMock(
  bins: {
    bucket: number;
    meanPredicted: number;
    observedFrequency: number;
    sampleSize: number;
  }[],
): MockedResponse {
  return {
    request: {
      query: CalibrationDocument,
      variables: { teamId: TEST_TEAM.id },
    },
    result: {
      data: {
        calibration: {
          __typename: "CalibrationReport",
          bins: bins.map((bin) => ({ __typename: "CalibrationBin", ...bin })),
        },
      },
    },
  };
}

function renderDashboard(
  mocks: MockedResponse[],
  team: CurrentTeamContextValue = stubTeam(),
) {
  return render(
    <MockedProvider mocks={mocks}>
      <CurrentTeamContext.Provider value={team}>
        <AnalyticsDashboard />
      </CurrentTeamContext.Provider>
    </MockedProvider>,
  );
}

describe("AnalyticsDashboard", () => {
  it("renders KPI values and the Brier quality badge", async () => {
    renderDashboard([
      metricsMock({
        brierScore: 0.04,
        forecastCount: 12,
        decisionSuccessRate: 0.6667,
        resolvedDecisionCount: 9,
      }),
      calibrationMock([
        {
          bucket: 7,
          meanPredicted: 0.68,
          observedFrequency: 0.7,
          sampleSize: 12,
        },
      ]),
    ]);

    expect(await screen.findByText("0.040")).toBeInTheDocument();
    expect(screen.getByText("Excellent")).toBeInTheDocument();
    expect(screen.getByText("67%")).toBeInTheDocument();
    expect(screen.getByText("12")).toBeInTheDocument();
    expect(screen.getByText("9")).toBeInTheDocument();
  });

  it("renders the calibration and distribution cards when bins exist", async () => {
    renderDashboard([
      metricsMock({
        brierScore: 0.18,
        forecastCount: 20,
        decisionSuccessRate: 0.5,
        resolvedDecisionCount: 15,
      }),
      calibrationMock([
        {
          bucket: 6,
          meanPredicted: 0.55,
          observedFrequency: 0.5,
          sampleSize: 20,
        },
      ]),
    ]);

    expect(await screen.findByText("Calibration")).toBeInTheDocument();
    expect(screen.getByText("Forecast distribution")).toBeInTheDocument();
    expect(screen.getByText("Good")).toBeInTheDocument();
    expect(
      screen.getByRole("img", {
        name: /Calibration across 1 probability bin/i,
      }),
    ).toBeInTheDocument();
  });

  it("shows an empty state when there is no forecast data", async () => {
    renderDashboard([
      metricsMock({
        brierScore: 0,
        forecastCount: 0,
        decisionSuccessRate: 0,
        resolvedDecisionCount: 0,
      }),
      calibrationMock([]),
    ]);

    expect(await screen.findByText("No forecast data yet")).toBeInTheDocument();
  });
});
