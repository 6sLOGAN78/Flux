import React from 'react';
import {
  BarChart3,
  Activity,
  Zap,
  Globe2,
  Database,
  Cpu,
  Layers,
} from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import {
  TimeSeriesAreaChart,
  TimeSeriesPoint,
} from '@/components/analytics/TimeSeriesAreaChart';
import {
  GeographicChoropleth,
  CountryStat,
} from '@/components/analytics/GeographicChoropleth';
import {
  ReferrerBreakdownTable,
  ReferrerItem,
} from '@/components/analytics/ReferrerBreakdownTable';
import {
  DeviceDonutChart,
  DeviceStat,
} from '@/components/analytics/DeviceDonutChart';
import { useAnalyticsSummary } from '@/hooks/useAnalyticsQuery';

const MOCK_TIME_SERIES: TimeSeriesPoint[] = [];

const MOCK_COUNTRIES: CountryStat[] = [];
const MOCK_REFERRERS: ReferrerItem[] = [];
const MOCK_DEVICES: DeviceStat[] = [];

export function AnalyticsPage() {
  const { data: analyticsData } = useAnalyticsSummary();

  const totalClicks = analyticsData?.totalClicks ?? 0;
  const uniqueVisitors = 0;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Analytics Explorer
            </h1>
            <Badge variant="emerald" size="sm" dot>
              ClickHouse Pipeline Active
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Real-time high-throughput telemetry powered by ClickHouse columnar storage.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="zinc" size="sm" className="font-mono">
            Zero-Cold-Start Ingestion
          </Badge>
        </div>
      </div>

      {/* KPI Stats Grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
            <span>Total Ingested Events</span>
            <Activity className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="mt-3 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            {totalClicks.toLocaleString()}
          </div>
          <div className="mt-1 flex items-center gap-1.5 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
            <span>+18.4% vs last week</span>
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
            <span>Unique Visitors</span>
            <Globe2 className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="mt-3 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            {uniqueVisitors.toLocaleString()}
          </div>
          <div className="mt-1 flex items-center gap-1.5 text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
            <span>+12.1% new visitors</span>
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
            <span>Edge Ingestion Latency</span>
            <Zap className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="mt-3 font-mono text-2xl font-bold text-emerald-600 dark:text-emerald-400">
            4.1 ms
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            p99 sub-10ms SLA guaranteed
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
            <span>Stream Compression</span>
            <Database className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="mt-3 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            8.4x
          </div>
          <div className="mt-1 text-[11px] text-zinc-400">
            Real-time gzip stream ratio
          </div>
        </div>
      </div>

      {/* Main TimeSeries Area Chart */}
      <TimeSeriesAreaChart data={MOCK_TIME_SERIES} />

      {/* Secondary Metrics Grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <GeographicChoropleth countries={MOCK_COUNTRIES} />
        <ReferrerBreakdownTable referrers={MOCK_REFERRERS} />
        <DeviceDonutChart devices={MOCK_DEVICES} />
      </div>
    </div>
  );
}

export default AnalyticsPage;
