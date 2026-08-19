import React from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Legend,
} from 'recharts';
import { Sparkles, TrendingUp, Cpu } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface CTRPredictionPoint {
  hour: string;
  actual: number | null;
  predicted: number;
}

export interface CTRPredictionChartProps {
  data: CTRPredictionPoint[];
  className?: string;
}

export function CTRPredictionChart({
  data,
  className,
}: CTRPredictionChartProps) {
  return (
    <div
      className={cn(
        'rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-4 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              CTR Forecasting &amp; Trend Trajectory
            </h2>
            <Badge variant="blue" size="sm">
              AI Predictive Model
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Transformer-based Click-Through Rate forecasting model with 95% confidence intervals.
          </p>
        </div>

        <div className="flex items-center gap-4 text-xs font-mono">
          <div className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-zinc-900 dark:bg-zinc-100" />
            <span className="text-zinc-600 dark:text-zinc-400">Observed CTR</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-blue-500" />
            <span className="text-zinc-600 dark:text-zinc-400">AI Forecast</span>
          </div>
        </div>
      </div>

      <div className="mt-6 h-72 w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart
            data={data}
            margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
          >
            <defs>
              <linearGradient id="colorActual" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#09090b" stopOpacity={0.4} />
                <stop offset="95%" stopColor="#09090b" stopOpacity={0.0} />
              </linearGradient>
              <linearGradient id="colorPredicted" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.0} />
              </linearGradient>
            </defs>
            <CartesianGrid
              strokeDasharray="3 3"
              vertical={false}
              className="stroke-zinc-100 dark:stroke-zinc-900"
            />
            <XAxis
              dataKey="hour"
              tickLine={false}
              axisLine={false}
              tick={{ fontSize: 11, fill: '#71717a' }}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickFormatter={(v) => `${v}%`}
              tick={{ fontSize: 11, fill: '#71717a' }}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: '#09090b',
                borderColor: '#27272a',
                borderRadius: '12px',
                color: '#fff',
                fontSize: '12px',
              }}
              formatter={(value: any, name: any) => [
                value !== null ? `${Number(value).toFixed(1)}%` : 'Projected',
                name === 'actual' ? 'Observed CTR' : 'Predicted Forecast',
              ]}
            />
            <Area
              type="monotone"
              dataKey="actual"
              stroke="#09090b"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#colorActual)"
              name="actual"
            />
            <Area
              type="monotone"
              dataKey="predicted"
              stroke="#3b82f6"
              strokeWidth={2}
              strokeDasharray="4 4"
              fillOpacity={1}
              fill="url(#colorPredicted)"
              name="predicted"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

export default CTRPredictionChart;
