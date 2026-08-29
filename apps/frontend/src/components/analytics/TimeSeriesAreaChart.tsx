import React, { useState } from 'react';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';
import { BarChart3, Calendar } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface TimeSeriesPoint {
  timestamp: string;
  clicks: number;
  uniqueVisitors: number;
}

export interface TimeSeriesAreaChartProps {
  data: TimeSeriesPoint[];
  title?: string;
  className?: string;
  activeRange?: string;
  onRangeChange?: (range: string) => void;
}

const RANGES = ['1h', '24h', '7d', '30d', '90d'];

export function TimeSeriesAreaChart({
  data,
  title = 'Click Volume Over Time',
  className,
  activeRange = '24h',
  onRangeChange,
}: TimeSeriesAreaChartProps) {
  // If parent controls state, use it, else fallback (though parent should always control it now)
  const [internalRange, setInternalRange] = useState('24h');
  const currentRange = onRangeChange ? activeRange : internalRange;
  const handleRangeChange = (r: string) => {
    if (onRangeChange) onRangeChange(r);
    else setInternalRange(r);
  };

  const totalClicks = data.reduce((sum, p) => sum + p.clicks, 0);
  const totalUniques = data.reduce((sum, p) => sum + p.uniqueVisitors, 0);

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      {/* Header & Controls */}
      <div className="flex flex-col justify-between gap-4 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              {title}
            </h2>
            <Badge variant="zinc" size="sm">
              ClickHouse Materialized Views
            </Badge>
          </div>
          <div className="mt-1 flex items-center gap-4 text-xs">
            <span className="font-mono font-bold text-zinc-900 dark:text-zinc-100">
              {totalClicks.toLocaleString()} total clicks
            </span>
            <span className="text-zinc-400">·</span>
            <span className="font-mono text-zinc-500">
              {totalUniques.toLocaleString()} unique visitors
            </span>
          </div>
        </div>

        {/* Range Selector */}
        <div className="flex items-center rounded-lg border border-zinc-200 bg-zinc-50 p-0.5 dark:border-zinc-800 dark:bg-zinc-900">
          {RANGES.map((r) => (
            <button
              key={r}
              type="button"
              onClick={() => handleRangeChange(r)}
              className={cn(
                'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                currentRange === r
                  ? 'bg-white text-zinc-900 shadow-xs dark:bg-zinc-800 dark:text-zinc-100'
                  : 'text-zinc-500 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-200'
              )}
            >
              {r}
            </button>
          ))}
        </div>
      </div>

      {/* Chart */}
      <div className="mt-6 h-64 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart
            data={data}
            margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
          >
            <defs>
              <linearGradient id="clickGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#09090b" stopOpacity={0.2} />
                <stop offset="95%" stopColor="#09090b" stopOpacity={0.0} />
              </linearGradient>
              <linearGradient id="uniqueGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#10b981" stopOpacity={0.2} />
                <stop offset="95%" stopColor="#10b981" stopOpacity={0.0} />
              </linearGradient>
            </defs>
            <CartesianGrid
              strokeDasharray="3 3"
              stroke="#e4e4e7"
              className="dark:stroke-zinc-800"
              vertical={false}
            />
            <XAxis
              dataKey="timestamp"
              stroke="#71717a"
              fontSize={11}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              stroke="#71717a"
              fontSize={11}
              tickLine={false}
              axisLine={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'rgba(9, 9, 11, 0.95)',
                borderRadius: '8px',
                border: 'none',
                color: '#fff',
                fontSize: '11px',
                boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
              }}
            />
            <Area
              type="monotone"
              dataKey="clicks"
              name="Total Clicks"
              stroke="#09090b"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#clickGradient)"
            />
            <Area
              type="monotone"
              dataKey="uniqueVisitors"
              name="Unique Visitors"
              stroke="#10b981"
              strokeWidth={1.5}
              strokeDasharray="4 4"
              fillOpacity={1}
              fill="url(#uniqueGradient)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

export default TimeSeriesAreaChart;
