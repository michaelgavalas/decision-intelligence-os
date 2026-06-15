import {
  Activity,
  BarChart3,
  CheckCircle2,
  ListChecks,
  Target,
} from "lucide-react";

import { BucketBarChart } from "@/components/charts/BucketBarChart";
import {
  CalibrationChart,
  type CalibrationPoint,
} from "@/components/charts/CalibrationChart";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { brierQuality } from "@/features/analytics/brier-quality";
import { StatCard } from "@/features/analytics/StatCard";
import { useCurrentTeam } from "@/features/teams/use-current-team";
import {
  type CalibrationQuery,
  type TeamMetricsQuery,
  useCalibrationQuery,
  useTeamMetricsQuery,
} from "@/graphql/generated/graphql";

/** Human label for a decile bucket index (1..10), e.g. 3 -> "20-30%". */
function bucketLabel(bucket: number): string {
  const lower = (bucket - 1) * 10;
  const upper = bucket * 10;
  return `${lower}-${upper}%`;
}

interface DashboardProps {
  metrics: TeamMetricsQuery["teamMetrics"];
  bins: CalibrationQuery["calibration"]["bins"];
}

function StatCardsSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {Array.from({ length: 4 }).map((_, index) => (
        <Card key={index} className="flex flex-col gap-3 p-5">
          <Skeleton className="h-4 w-24" />
          <Skeleton className="h-8 w-20" />
          <Skeleton className="h-3 w-28" />
        </Card>
      ))}
    </div>
  );
}

function ChartCardSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-3 w-64" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-72 w-full" />
      </CardContent>
    </Card>
  );
}

function Dashboard({ metrics, bins }: DashboardProps) {
  const quality = brierQuality(metrics.brierScore);

  const calibrationPoints: CalibrationPoint[] = bins.map((bin) => ({
    predicted: bin.meanPredicted,
    observed: bin.observedFrequency,
    sampleSize: bin.sampleSize,
  }));

  const distribution = bins.map((bin) => ({
    label: bucketLabel(bin.bucket),
    value: bin.sampleSize,
  }));

  const calibrationSummary = `Calibration across ${bins.length} probability ${
    bins.length === 1 ? "bin" : "bins"
  }. Points on the diagonal are well-calibrated.`;

  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          label="Brier score"
          icon={Target}
          value={metrics.brierScore.toFixed(3)}
          hint="Lower is better (0-1)"
          badge={
            metrics.forecastCount > 0 ? (
              <Badge variant={quality.variant}>{quality.label}</Badge>
            ) : undefined
          }
        />
        <StatCard
          label="Decision success rate"
          icon={CheckCircle2}
          value={`${Math.round(metrics.decisionSuccessRate * 100)}%`}
          hint="of resolved decisions"
        />
        <StatCard
          label="Forecasts scored"
          icon={Activity}
          value={metrics.forecastCount}
          hint="resolved forecasts"
        />
        <StatCard
          label="Decisions resolved"
          icon={ListChecks}
          value={metrics.resolvedDecisionCount}
          hint="with a recorded outcome"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Calibration</CardTitle>
          <CardDescription>
            Points on the diagonal are well-calibrated; above the line means
            under-confident, below means over-confident.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <CalibrationChart
            data={calibrationPoints}
            ariaLabel={calibrationSummary}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Forecast distribution</CardTitle>
          <CardDescription>
            How many resolved forecasts fall into each predicted-probability
            range.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <BucketBarChart
            data={distribution}
            ariaLabel="Number of forecasts per predicted-probability range"
            valueFormatter={(value) => String(value)}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export function AnalyticsDashboard() {
  const { currentTeam, currentTeamId, loading: teamLoading } = useCurrentTeam();

  const metricsQuery = useTeamMetricsQuery({
    variables: { teamId: currentTeamId ?? "" },
    skip: !currentTeamId,
  });
  const calibrationQuery = useCalibrationQuery({
    variables: { teamId: currentTeamId ?? "" },
    skip: !currentTeamId,
  });

  const loading =
    (teamLoading && !currentTeamId) ||
    metricsQuery.loading ||
    calibrationQuery.loading;
  const error = metricsQuery.error ?? calibrationQuery.error;

  const metrics = metricsQuery.data?.teamMetrics;
  const bins = calibrationQuery.data?.calibration.bins ?? [];

  function retry() {
    void metricsQuery.refetch();
    void calibrationQuery.refetch();
  }

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold tracking-tight text-foreground">
          Analytics
        </h1>
        <p className="text-sm text-muted-foreground">
          {currentTeam
            ? `Forecast quality for ${currentTeam.name}`
            : "Forecast quality for your team"}
        </p>
      </header>

      {!currentTeamId && !teamLoading ? (
        <EmptyState
          icon={BarChart3}
          title="No team selected"
          description="Select a team to see its forecast quality and calibration."
        />
      ) : loading ? (
        <div className="flex flex-col gap-6">
          <StatCardsSkeleton />
          <ChartCardSkeleton />
          <ChartCardSkeleton />
        </div>
      ) : error ? (
        <Card className="flex flex-col items-center gap-3 p-8 text-center">
          <p className="text-sm font-medium text-foreground">
            We couldn&apos;t load analytics.
          </p>
          <p className="text-sm text-muted-foreground">
            Something went wrong while computing this team&apos;s metrics.
          </p>
          <Button variant="outline" onClick={retry}>
            Try again
          </Button>
        </Card>
      ) : !metrics || metrics.forecastCount === 0 || bins.length === 0 ? (
        <EmptyState
          icon={BarChart3}
          title="No forecast data yet"
          description="Add predictions to your decisions and record outcomes to measure calibration."
        />
      ) : (
        <Dashboard metrics={metrics} bins={bins} />
      )}
    </div>
  );
}
