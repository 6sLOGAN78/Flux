import React from 'react';
import { DollarSign, BarChart2, TrendingUp } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface CampaignAttributionItem {
  channel: string;
  campaign: string;
  touchpoints: number;
  conversions: number;
  revenue: number;
  sharePercentage: number;
}

export interface AttributionComparisonTableProps {
  data: CampaignAttributionItem[];
  currency?: string;
  isLoading?: boolean;
  className?: string;
}

export function AttributionComparisonTable({
  data,
  currency = '$',
  isLoading = false,
  className,
}: AttributionComparisonTableProps) {
  if (isLoading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <div
            key={i}
            className="h-16 rounded-xl border border-zinc-200 bg-white p-4 animate-pulse dark:border-zinc-800 dark:bg-zinc-950"
          />
        ))}
      </div>
    );
  }

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 p-5 dark:border-zinc-900">
        <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
          Campaign Attribution Performance
        </h3>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          Calculated conversions and pipeline value attributed under the active algorithm.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Marketing Channel</th>
              <th className="px-4 py-3">Campaign Tag</th>
              <th className="px-4 py-3 text-right">Touchpoints</th>
              <th className="px-4 py-3 text-right">Attributed Conv.</th>
              <th className="px-4 py-3 text-right">Attributed Revenue</th>
              <th className="px-4 py-3 w-40">Revenue Share</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {data.map((row) => (
              <tr
                key={row.campaign}
                className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
              >
                <td className="px-4 py-3 font-semibold text-zinc-900 dark:text-zinc-100">
                  {row.channel}
                </td>
                <td className="px-4 py-3 font-mono text-zinc-500">
                  {row.campaign}
                </td>
                <td className="px-4 py-3 text-right font-mono text-zinc-600 dark:text-zinc-400">
                  {row.touchpoints.toLocaleString()}
                </td>
                <td className="px-4 py-3 text-right font-mono font-bold text-zinc-900 dark:text-zinc-100">
                  {row.conversions.toFixed(1)}
                </td>
                <td className="px-4 py-3 text-right font-mono font-bold text-emerald-600 dark:text-emerald-400">
                  {`${currency}${row.revenue.toLocaleString()}`}
                </td>
                <td className="px-4 py-3">
                  <div className="space-y-1">
                    <div className="flex justify-between text-[11px] font-mono">
                      <span>{`${row.sharePercentage}%`}</span>
                    </div>
                    <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-900">
                      <div
                        style={{ width: `${row.sharePercentage}%` }}
                        className="h-full rounded-full bg-zinc-900 dark:bg-zinc-100"
                      />
                    </div>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default AttributionComparisonTable;
