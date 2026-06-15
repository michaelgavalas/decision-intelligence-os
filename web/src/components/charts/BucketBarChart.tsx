import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { ChartEmpty, ChartLoading } from "@/components/charts/chart-states";
import { CHART_TOKENS, resolveColor } from "@/components/charts/chart-colors";

export interface BucketDatum {
  label: string;
  value: number;
}

interface BucketBarChartProps {
  data: BucketDatum[];
  loading?: boolean;
  height?: number;
  ariaLabel?: string;
  valueFormatter?: (value: number) => string;
}

export function BucketBarChart({
  data,
  loading,
  height = 280,
  ariaLabel = "Distribution by bucket",
  valueFormatter,
}: BucketBarChartProps) {
  if (loading) {
    return <ChartLoading height={height} />;
  }
  if (data.length === 0) {
    return <ChartEmpty height={height} />;
  }

  const secondary = resolveColor(CHART_TOKENS.secondary);
  const grid = resolveColor(CHART_TOKENS.border);
  const axis = resolveColor(CHART_TOKENS.mutedForeground);
  const surface = resolveColor(CHART_TOKENS.surface, "#ffffff");

  return (
    <div role="img" aria-label={ariaLabel} style={{ width: "100%", height }}>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data} margin={{ top: 8, right: 8, bottom: 8, left: 0 }}>
          <CartesianGrid stroke={grid} strokeDasharray="3 3" vertical={false} />
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
            allowDecimals={false}
          />
          <Tooltip
            cursor={{ fill: grid, opacity: 0.3 }}
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
          <Bar dataKey="value" fill={secondary} radius={[4, 4, 0, 0]} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
