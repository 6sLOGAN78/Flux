import React from 'react';
import { Sparkles, BarChart2, TrendingUp, Calendar } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface CampaignItem {
  id: string;
  name: string;
  channel: string;
  utmCampaign: string;
  totalClicks: number;
  conversions: number;
  status: 'active' | 'paused' | 'completed';
  createdAt: string;
}

export interface CampaignListTableProps {
  campaigns: CampaignItem[];
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
              <th className="px-4 py-3">Channel</th>
              <th className="px-4 py-3">UTM Campaign</th>
              <th className="px-4 py-3 text-right">Clicks</th>
              <th className="px-4 py-3 text-right">Conversions</th>
              <th className="px-4 py-3 text-right">Conv. Rate</th>
              <th className="px-4 py-3 text-center">Status</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {campaigns.map((camp) => {
              const convRate =
                camp.totalClicks > 0
                  ? ((camp.conversions / camp.totalClicks) * 100).toFixed(1)
                  : '0.0';

              return (
                <tr
                  key={camp.id}
                  className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
                >
                  <td className="px-4 py-3 font-semibold text-zinc-900 dark:text-zinc-100">
                    {camp.name}
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant="zinc" size="sm">
                      {camp.channel}
                    </Badge>
                  </td>
                  <td className="px-4 py-3 font-mono text-zinc-500">
                    {camp.utmCampaign}
                  </td>
                  <td className="px-4 py-3 text-right font-mono font-bold text-zinc-900 dark:text-zinc-100">
                    {camp.totalClicks.toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-right font-mono text-emerald-600 dark:text-emerald-400">
                    {camp.conversions.toLocaleString()}
                  </td>
                  <td className="px-4 py-3 text-right font-mono font-medium text-zinc-700 dark:text-zinc-300">
                    {convRate}%
                  </td>
                  <td className="px-4 py-3 text-center">
                    <Badge
                      variant={camp.status === 'active' ? 'emerald' : 'zinc'}
                      size="sm"
                      dot
                    >
                      {camp.status}
                    </Badge>
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
