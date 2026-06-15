import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { ChartEmpty, ChartLoading } from "@/components/charts/chart-states";
import { CHART_TOKENS, resolveColor } from "@/components/charts/chart-colors";

export interface TrendPoint {
  label: string;
  value: number;
}

interface TrendLineChartProps {
  data: TrendPoint[];
  loading?: boolean;
  height?: number;
  ariaLabel?: string;
  valueFormatter?: (value: number) => string;
}

export function TrendLineChart({
  data,
  loading,
  height = 280,
  ariaLabel = "Trend over time",
  valueFormatter,
}: TrendLineChartProps) {
  if (loading) {
    return <ChartLoading height={height} />;
  }
  if (data.length === 0) {
    return <ChartEmpty height={height} />;
  }

  const primary = resolveColor(CHART_TOKENS.primary);
  const grid = resolveColor(CHART_TOKENS.border);
  const axis = resolveColor(CHART_TOKENS.mutedForeground);
  const surface = resolveColor(CHART_TOKENS.surface, "#ffffff");

  return (
    <div role="img" aria-label={ariaLabel} style={{ width: "100%", height }}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart
          data={data}
          margin={{ top: 8, right: 8, bottom: 8, left: 0 }}
        >
          <CartesianGrid stroke={grid} strokeDasharray="3 3" />
          <XAxis
            dataKey="label"
            stroke={axis}
            tick={{ fill: axis, fontSize: 12 }}
            tickLine={false}
          />
          <YAxis
            stroke={axis}
            tick={{ fill: axis, fontSize: 12 }}
            tickLine={false}
            tickFormatter={valueFormatter}
          />
          <Tooltip
            contentStyle={{
              background: surface,
              border: `1px solid ${grid}`,
              borderRadius: 8,
              fontSize: 12,
            }}
            formatter={(value: number) =>
              valueFormatter ? valueFormatter(value) : value
            }
          />
          <Line
            type="monotone"
            dataKey="value"
            stroke={primary}
            strokeWidth={2}
            dot={false}
            activeDot={{ r: 4 }}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
