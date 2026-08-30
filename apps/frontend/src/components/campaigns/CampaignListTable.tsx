import React from 'react';
import { Sparkles, BarChart2, TrendingUp, Calendar } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';
import type { CampaignPerformance } from '@flux/zod';

export interface CampaignListTableProps {
  campaigns: any[]; // The merged campaign data
  isLoading?: boolean;
}

export function CampaignListTable({
  campaigns,
  isLoading = false,
}: CampaignListTableProps) {
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

  if (campaigns.length === 0) {
    return (
      <div className="rounded-xl border border-zinc-200 bg-white p-12 text-center text-xs text-zinc-400 dark:border-zinc-800 dark:bg-zinc-950">
        No active marketing campaigns found. Use the UTM Builder to launch one.
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Campaign Name</th>
              <th className="px-4 py-3">UTM Source</th>
              <th className="px-4 py-3">UTM Medium</th>
              <th className="px-4 py-3">UTM Campaign</th>
              <th className="px-4 py-3 text-right">Clicks</th>
              <th className="px-4 py-3 text-right">Unique Visitors</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {campaigns.map((camp) => {
              return (
                <tr
                  key={camp.id}
                  className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
                >
                  <td className="px-4 py-3 font-semibold text-zinc-900 dark:text-zinc-100">
                    {camp.name}
                  </td>
                  <td className="px-4 py-3 font-mono text-zinc-500">
                    {camp.utm_source || '-'}
                  </td>
                  <td className="px-4 py-3 font-mono text-zinc-500">
                    {camp.utm_medium || '-'}
                  </td>
                  <td className="px-4 py-3 font-mono text-zinc-500">
                    {camp.utm_campaign || '-'}
                  </td>
                  <td className="px-4 py-3 text-right font-mono font-bold text-zinc-900 dark:text-zinc-100">
                    {(camp.clicks || 0).toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-emerald-600 dark:text-emerald-400">
                    {(camp.unique_visitors || 0).toLocaleString()}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default CampaignListTable;
