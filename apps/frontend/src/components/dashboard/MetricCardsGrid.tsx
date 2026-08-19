import React from 'react';
import { Link2, MousePointerClick, TrendingUp, Globe2 } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface DashboardMetrics {
  totalLinks: number;
  totalClicks24h: number;
  avgCtr: number;
  activeDomains: number;
  trends?: {
    linksChange?: string;
    clicksChange?: string;
    ctrChange?: string;
    domainsChange?: string;
  };
}

export interface MetricCardsGridProps {
  metrics?: DashboardMetrics;
  isLoading?: boolean;
}

function formatNumber(num: number): string {
  if (num >= 1_000_000) {
    return `${(num / 1_000_000).toFixed(1)}M`;
  }
  if (num >= 1_000) {
    return `${(num / 1_000).toFixed(1)}K`;
  }
  return num.toLocaleString();
}

export function MetricCardsGrid({
  metrics = {
    totalLinks: 12480,
    totalClicks24h: 382400,
    avgCtr: 4.8,
    activeDomains: 6,
    trends: {
      linksChange: '+8.4%',
      clicksChange: '+14.2%',
      ctrChange: '+0.6%',
      domainsChange: '+2',
    },
  },
  isLoading = false,
}: MetricCardsGridProps) {
  const cards = [
    {
      title: 'Total Links',
      value: metrics.totalLinks.toLocaleString(),
      change: metrics.trends?.linksChange || '+8.4%',
      icon: <Link2 className="h-4 w-4 text-zinc-500" />,
      description: 'Active short redirects',
    },
    {
      title: '24h Clicks',
      value: formatNumber(metrics.totalClicks24h),
      change: metrics.trends?.clicksChange || '+14.2%',
      icon: <MousePointerClick className="h-4 w-4 text-zinc-500" />,
      description: 'Global edge traffic',
    },
    {
      title: 'Avg CTR',
      value: `${metrics.avgCtr.toFixed(1)}%`,
      change: metrics.trends?.ctrChange || '+0.6%',
      icon: <TrendingUp className="h-4 w-4 text-zinc-500" />,
      description: 'Click-through rate',
    },
    {
      title: 'Active Domains',
      value: metrics.activeDomains.toString(),
      change: metrics.trends?.domainsChange || '+2',
      icon: <Globe2 className="h-4 w-4 text-zinc-500" />,
      description: 'Verified DNS zones',
    },
  ];

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {cards.map((card, idx) => (
        <div
          key={idx}
          className="relative overflow-hidden rounded-xl border border-zinc-200 bg-white p-5 shadow-xs transition-all hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-950 dark:hover:border-zinc-700"
        >
          <div className="flex items-center justify-between">
            <span className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
              {card.title}
            </span>
            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-900">
              {card.icon}
            </div>
          </div>

          <div className="mt-3 flex items-baseline justify-between">
            {isLoading ? (
              <div className="h-7 w-24 animate-pulse rounded bg-zinc-200 dark:bg-zinc-800" />
            ) : (
              <div className="font-mono text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
                {card.value}
              </div>
            )}

            <Badge variant="emerald" size="sm" className="font-mono font-semibold">
              {card.change}
            </Badge>
          </div>

          <p className="mt-1 text-[11px] text-zinc-400 dark:text-zinc-500">
            {card.description}
          </p>
        </div>
      ))}
    </div>
  );
}

export default MetricCardsGrid;
