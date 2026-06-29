"use client";

import { AreaChart, Area, ResponsiveContainer } from "recharts";
import type { PricePoint } from "@/lib/types";

interface PriceHistoryChartProps {
  data: PricePoint[];
}

function LastDot({ cx, cy, index, total, color }: {
  cx?: number;
  cy?: number;
  index?: number;
  total: number;
  color: string;
}) {
  if (index !== total - 1 || cx === undefined || cy === undefined) return null;
  return <circle cx={cx} cy={cy} r={4} fill={color} stroke="#14141b" strokeWidth={2} />;
}

export function PriceHistoryChart({ data }: PriceHistoryChartProps) {
  if (data.length === 0) return null;

  const isGrowing = data[data.length - 1].price >= data[0].price;
  const color = isGrowing ? "#22c55e" : "#ef4444";
  const gradientId = `price-fill-${isGrowing ? "up" : "down"}`;

  return (
    <ResponsiveContainer width="100%" height={170}>
      <AreaChart data={data}>
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity={0.25} />
            <stop offset="100%" stopColor={color} stopOpacity={0.02} />
          </linearGradient>
        </defs>
        <Area
          type="monotone"
          dataKey="price"
          stroke={color}
          strokeWidth={2}
          fill={`url(#${gradientId})`}
          isAnimationActive={false}
          dot={(props) => (
            <LastDot
              key={props.index}
              cx={props.cx}
              cy={props.cy}
              index={props.index}
              total={data.length}
              color={color}
            />
          )}
          activeDot={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
