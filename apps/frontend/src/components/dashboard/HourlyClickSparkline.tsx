import React from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { Badge } from '@/components/ui/Badge';
import { Activity } from 'lucide-react';

export interface HourlyDataPoint {
  hour: string;
  clicks: number;
}

export interface HourlyClickSparklineProps {
  data?: HourlyDataPoint[];
  title?: string;
  totalClicks?: number;
  isLoading?: boolean;
}

const DEFAULT_HOURLY_DATA: HourlyDataPoint[] = [
  { hour: '00:00', clicks: 0 },
  { hour: '02:00', clicks: 0 },
  { hour: '04:00', clicks: 0 },
  { hour: '06:00', clicks: 0 },
  { hour: '08:00', clicks: 0 },
  { hour: '10:00', clicks: 0 },
  { hour: '12:00', clicks: 0 },
  { hour: '14:00', clicks: 0 },
  { hour: '16:00', clicks: 0 },
  { hour: '18:00', clicks: 0 },
  { hour: '20:00', clicks: 0 },
  { hour: '22:00', clicks: 0 },
];

export function HourlyClickSparkline({
  data = DEFAULT_HOURLY_DATA,
  title = '24h Click Volume',
  totalClicks = 0,
  isLoading = false,
}: HourlyClickSparklineProps) {
  return (
    <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div className="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
              {title}
            </h3>
            <Badge variant="emerald" size="sm" dot>
              Live ClickHouse Ingestion
            </Badge>
          </div>
          <div className="mt-1 flex items-baseline gap-2">
            <span className="font-mono text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              {totalClicks.toLocaleString()}
            </span>
            <span className="text-xs text-zinc-500">redirects recorded</span>
          </div>
        </div>

        <div className="flex items-center gap-1.5 font-mono text-[11px] text-zinc-400">
          <Activity className="h-3.5 w-3.5 text-emerald-500" />
          <span>Real-Time Stream Active</span>
        </div>
      </div>

      <div className="mt-6 h-56 w-full">
        {isLoading ? (
          <div className="flex h-full w-full items-center justify-center">
            <div className="h-32 w-full animate-pulse rounded bg-zinc-100 dark:bg-zinc-900" />
          </div>
        ) : (
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart
              data={data}
              margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
            >
              <defs>
                <linearGradient id="clickGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#18181b" stopOpacity={0.15} />
                  <stop offset="95%" stopColor="#18181b" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="clickGradientDark" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#10b981" stopOpacity={0.25} />
                  <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis
                dataKey="hour"
                stroke="#a1a1aa"
                fontSize={10}
                tickLine={false}
                axisLine={false}
              />
              <YAxis
                stroke="#a1a1aa"
                fontSize={10}
                tickLine={false}
                axisLine={false}
                tickFormatter={(val) => (val >= 1000 ? `${(val / 1000).toFixed(1)}k` : `${val}`)}
              />
              <Tooltip
                content={({ active, payload, label }) => {
                  if (active && payload && payload.length) {
                    return (
                      <div className="rounded-lg border border-zinc-200 bg-white/95 p-2 shadow-md backdrop-blur-xs dark:border-zinc-800 dark:bg-zinc-900/95">
                        <div className="font-mono text-[10px] text-zinc-400">
                          {label}
                        </div>
                        <div className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                          {payload[0].value?.toLocaleString()} clicks
                        </div>
                      </div>
                    );
                  }
                  return null;
                }}
              />
              <Area
                type="monotone"
                dataKey="clicks"
                stroke="#18181b"
                strokeWidth={2}
                fillOpacity={1}
                fill="url(#clickGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  );
}

export default HourlyClickSparkline;
