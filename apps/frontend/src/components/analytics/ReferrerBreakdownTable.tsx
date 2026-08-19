import React from 'react';
import { Share2, ExternalLink, Globe2 } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface ReferrerItem {
  domain: string;
  name: string;
  clicks: number;
  percentage: number;
}

export interface ReferrerBreakdownTableProps {
  referrers: ReferrerItem[];
  className?: string;
}

export function ReferrerBreakdownTable({
  referrers,
  className,
}: ReferrerBreakdownTableProps) {
  return (
    <div
      className={cn(
        'space-y-4 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div>
          <div className="flex items-center gap-2">
            <Share2 className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Traffic Referrers
            </h2>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Top channels driving click-throughs to your links.
          </p>
        </div>
      </div>

      <div className="space-y-3">
        {referrers.map((ref) => (
          <div key={ref.domain} className="space-y-1.5">
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {ref.name}
                </span>
                <span className="font-mono text-[11px] text-zinc-400">
                  ({ref.domain})
                </span>
              </div>

              <div className="flex items-center gap-2 font-mono">
                <span className="font-bold text-zinc-900 dark:text-zinc-100">
                  {ref.clicks.toLocaleString()}
                </span>
                <span className="text-zinc-400">({ref.percentage}%)</span>
              </div>
            </div>

            <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-900">
              <div
                style={{ width: `${ref.percentage}%` }}
                className="h-full rounded-full bg-blue-600 transition-all duration-500"
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default ReferrerBreakdownTable;
