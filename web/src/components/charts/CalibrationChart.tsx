import {
  CartesianGrid,
  ComposedChart,
  Line,
  ReferenceLine,
  ResponsiveContainer,
  Scatter,
  Tooltip,
  type TooltipProps,
  XAxis,
  YAxis,
  ZAxis,
} from "recharts";

import { ChartEmpty, ChartLoading } from "@/components/charts/chart-states";
import { CHART_TOKENS, resolveColor } from "@/components/charts/chart-colors";

export interface CalibrationPoint {
  /** Mean predicted probability for the bin, in [0, 1]. */
  predicted: number;
  /** Observed success frequency for the bin, in [0, 1]. */
  observed: number;
  /** Number of forecasts in the bin; encoded as dot size when present. */
  sampleSize?: number;
}

interface CalibrationChartProps {
  data: CalibrationPoint[];
  loading?: boolean;
  height?: number;
  ariaLabel?: string;
}

const percentTick = (value: number) => `${Math.round(value * 100)}%`;

function CalibrationTooltip({ active, payload }: TooltipProps<number, string>) {
  if (!active || !payload || payload.length === 0) {
    return null;
  }
  const point = payload[0]?.payload as CalibrationPoint | undefined;
  if (!point) {
    return null;
  }
  return (
    <div className="rounded-md border border-border bg-surface px-2.5 py-1.5 text-xs text-foreground shadow-md">
      <div className="tabular-nums">
        Predicted {Math.round(point.predicted * 100)}%
      </div>
      <div className="tabular-nums">
        Observed {Math.round(point.observed * 100)}%
      </div>
      {typeof point.sampleSize === "number" ? (
        <div className="tabular-nums text-muted-foreground">
          n = {point.sampleSize}
        </div>
      ) : null}
    </div>
  );
}

export function CalibrationChart({
  data,
  loading,
  height = 320,
  ariaLabel = "Calibration: predicted versus observed probability",
}: CalibrationChartProps) {
  if (loading) {
    return <ChartLoading height={height} />;
  }
  if (data.length === 0) {
    return <ChartEmpty height={height} />;
  }

  const primary = resolveColor(CHART_TOKENS.primary);
  const grid = resolveColor(CHART_TOKENS.border);
  const axis = resolveColor(CHART_TOKENS.mutedForeground);

  // Soft draw-in on load, disabled when the user prefers reduced motion so the
  // data is readable immediately (Recharts animation is JS-driven and not covered
  // by the global CSS reduced-motion guard).
  const animate =
    typeof window === "undefined" ||
    typeof window.matchMedia !== "function" ||
    !window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const sorted = [...data].sort((a, b) => a.predicted - b.predicted);
  const hasSampleSize = sorted.some(
    (point) => typeof point.sampleSize === "number",
  );

  return (
    <div role="img" aria-label={ariaLabel} style={{ width: "100%", height }}>
      <ResponsiveContainer width="100%" height="100%">
        <ComposedChart
          data={sorted}
          margin={{ top: 16, right: 18, bottom: 8, left: 0 }}
        >
          <CartesianGrid stroke={grid} strokeDasharray="3 3" />
          <XAxis
            type="number"
            dataKey="predicted"
            domain={[0, 1]}
            ticks={[0, 0.25, 0.5, 0.75, 1]}
            tickFormatter={percentTick}
            stroke={axis}
            tick={{ fill: axis, fontSize: 12 }}
            tickLine={false}
            padding={{ left: 6, right: 16 }}
            name="Predicted"
          />
          <YAxis
            type="number"
            dataKey="observed"
            domain={[0, 1]}
            ticks={[0, 0.25, 0.5, 0.75, 1]}
            tickFormatter={percentTick}
            stroke={axis}
            tick={{ fill: axis, fontSize: 12 }}
            tickLine={false}
            padding={{ top: 14, bottom: 6 }}
            name="Observed"
          />
          {hasSampleSize ? (
            <ZAxis
              type="number"
              dataKey="sampleSize"
              range={[40, 360]}
              name="Sample size"
            />
          ) : null}
          <ReferenceLine
            segment={[
              { x: 0, y: 0 },
              { x: 1, y: 1 },
            ]}
            stroke={axis}
            strokeDasharray="6 6"
            ifOverflow="extendDomain"
          />
          <Tooltip
            cursor={{ strokeDasharray: "3 3" }}
            content={<CalibrationTooltip />}
          />
          <Line
            type="monotone"
            dataKey="observed"
            stroke={primary}
            strokeWidth={2}
            dot={false}
            activeDot={false}
            isAnimationActive={animate}
            animationDuration={700}
            animationEasing="ease-out"
          />
          <Scatter
            dataKey="observed"
            fill={primary}
            isAnimationActive={animate}
            animationBegin={animate ? 350 : 0}
            animationDuration={400}
            animationEasing="ease-out"
          />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
