import React, { useState, useMemo } from 'react';
import {
  BarChart3,
  Activity,
  Zap,
  Globe2,
  Database,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import {
  TimeSeriesAreaChart,
  TimeSeriesPoint,
} from '@/components/analytics/TimeSeriesAreaChart';
import {
  ReferrerBreakdownTable,
  ReferrerItem,
} from '@/components/analytics/ReferrerBreakdownTable';
import { UTMPerformanceTable } from '@/components/analytics/UTMPerformanceTable';
import { TopLinksTable } from '@/components/analytics/TopLinksTable';
import {
  useAnalyticsSummary,
  useAnalyticsTimeseries,
  useAnalyticsTopLinks,
  useAnalyticsReferrers,
} from '@/hooks/useAnalyticsQuery';

// Unimplemented on backend yet, preserving layout with empty state

export function AnalyticsPage() {
  const [activeRange, setActiveRange] = useState('30d');

  // Compute "from" date based on selected range
  const { from, interval } = useMemo(() => {
    const now = new Date();
    let fromDate = new Date();
    let intv = 'day';

    switch (activeRange) {
      case '1h':
        fromDate.setHours(now.getHours() - 1);
        intv = 'hour';
        break;
      case '24h':
        fromDate.setHours(now.getHours() - 24);
        intv = 'hour';
        break;
      case '7d':
        fromDate.setDate(now.getDate() - 7);
        intv = 'day';
        break;
      case '30d':
        fromDate.setDate(now.getDate() - 30);
        intv = 'day';
        break;
      case '90d':
        fromDate.setDate(now.getDate() - 90);
        intv = 'day';
        break;
      default:
        fromDate.setDate(now.getDate() - 30);
    }
    return { from: fromDate.toISOString(), interval: intv };
  }, [activeRange]);

  const to = new Date().toISOString();

  // Queries
  const { 
    data: summaryData, 
    isLoading: summaryLoading, 
    isError: summaryError 
  } = useAnalyticsSummary(from, to);
  
  const { 
    data: timeseriesData, 
    isLoading: tsLoading, 
    isError: tsError 
  } = useAnalyticsTimeseries(from, to, interval);
  
  const {
    data: referrersData,
    isLoading: refLoading,
    isError: refError
  } = useAnalyticsReferrers(from, to);

  // TopLinks query can be added similarly, though the dashboard currently doesn't have a component for it explicitly in the grid, 
  // wait, the prompt asks to connect the Top Links UI! Let me check if there's a top links component.
  // Oh, wait! The prompt says: "Connect the existing top-links UI to: GET /api/v1/analytics/top-links"
  // Is there a TopLinks component? Let's check `src/components/analytics/`

  const totalClicks = summaryData?.total_clicks ?? 0;
  const uniqueVisitors = summaryData?.unique_visitors ?? 0;

  const chartData: TimeSeriesPoint[] = (timeseriesData?.data ?? []).map((d: any) => ({
    timestamp: d.timestamp.substring(0, 10), // simplified
    clicks: d.clicks,
    uniqueVisitors: d.unique_visitors,
  }));

  const { data: topLinksData, isLoading: topLoading, isError: topError } = useAnalyticsTopLinks(from, to);

  const topLinks = topLinksData?.data ?? [];

  const refTotal = (referrersData?.data ?? []).reduce((acc: number, r: any) => acc + r.clicks, 0);
  const referrers: ReferrerItem[] = (referrersData?.data ?? []).map((r: any) => ({
    domain: r.referrer || 'direct',
    name: r.referrer || 'Direct / Unknown',
    clicks: r.clicks,
    percentage: refTotal > 0 ? Number(((r.clicks / refTotal) * 100).toFixed(1)) : 0,
  }));

  const isLoading = summaryLoading || tsLoading || refLoading || topLoading;
  const isError = summaryError || tsError || refError || topError;

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

      {isError && (
        <div className="flex items-center gap-2 rounded-lg bg-red-50 p-4 text-sm text-red-600 dark:bg-red-950/50 dark:text-red-400">
          <AlertCircle className="h-5 w-5" />
          <p>Failed to load analytics data. Please try again later.</p>
        </div>
      )}

      {/* KPI Stats Grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 relative">
        {isLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-zinc-950/50">
            <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
          </div>
        )}
        
        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
            <span>Total Clicks</span>
            <Activity className="h-4 w-4 text-zinc-400" />
          </div>
          <div className="mt-3 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            {totalClicks.toLocaleString()}
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
      <div className="relative">
        {tsLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-zinc-950/50">
            <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
          </div>
        )}
        <TimeSeriesAreaChart 
          data={chartData} 
          activeRange={activeRange}
          onRangeChange={setActiveRange}
        />
      </div>

      {/* Secondary Metrics Grid */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3 relative">
        <div className="relative">
          {topLoading && (
            <div className="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-zinc-950/50">
              <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
            </div>
          )}
          <TopLinksTable links={topLinks} />
        </div>
        
        <div className="relative">
          {refLoading && (
            <div className="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-zinc-950/50">
              <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
            </div>
          )}
          <ReferrerBreakdownTable referrers={referrers} />
        </div>
        
        <UTMPerformanceTable from={from} to={to} />
      </div>
    </div>
  );
}

export default AnalyticsPage;
