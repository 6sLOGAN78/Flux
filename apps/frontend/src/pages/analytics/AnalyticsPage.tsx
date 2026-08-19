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

const MOCK_TIME_SERIES: TimeSeriesPoint[] = [
  { timestamp: '00:00', clicks: 120, uniqueVisitors: 90 },
  { timestamp: '04:00', clicks: 240, uniqueVisitors: 190 },
  { timestamp: '08:00', clicks: 890, uniqueVisitors: 720 },
  { timestamp: '12:00', clicks: 1420, uniqueVisitors: 1100 },
  { timestamp: '16:00', clicks: 1180, uniqueVisitors: 940 },
  { timestamp: '20:00', clicks: 650, uniqueVisitors: 510 },
  { timestamp: '23:59', clicks: 420, uniqueVisitors: 330 },
];

const MOCK_COUNTRIES: CountryStat[] = [
  { countryCode: 'US', countryName: 'United States', clicks: 14200, percentage: 48.2 },
  { countryCode: 'GB', countryName: 'United Kingdom', clicks: 6400, percentage: 21.7 },
  { countryCode: 'DE', countryName: 'Germany', clicks: 3100, percentage: 10.5 },
  { countryCode: 'JP', countryName: 'Japan', clicks: 2400, percentage: 8.1 },
  { countryCode: 'CA', countryName: 'Canada', clicks: 1800, percentage: 6.1 },
];

const MOCK_REFERRERS: ReferrerItem[] = [
  { domain: 'twitter.com', name: 'Twitter / X', clicks: 8200, percentage: 42.1 },
  { domain: 'google.com', name: 'Google Search', clicks: 5900, percentage: 30.3 },
  { domain: 'direct', name: 'Direct / Email', clicks: 3100, percentage: 15.9 },
  { domain: 'linkedin.com', name: 'LinkedIn', clicks: 2300, percentage: 11.7 },
];

const MOCK_DEVICES: DeviceStat[] = [
  { label: 'Mobile (iOS/Android)', value: 16400, percentage: 58.4, color: '#09090b' },
  { label: 'Desktop (macOS/Windows)', value: 10500, percentage: 37.4, color: '#2563eb' },
  { label: 'Tablet & Others', value: 1200, percentage: 4.2, color: '#10b981' },
];

export function AnalyticsPage() {
  const { data: analyticsData } = useAnalyticsSummary();

  const totalClicks = analyticsData?.totalClicks ?? 248920;
  const uniqueVisitors = 142300;

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
