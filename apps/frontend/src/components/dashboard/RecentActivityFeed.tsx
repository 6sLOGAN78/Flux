import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { Copy, Check, ExternalLink, ArrowUpRight, BarChart2 } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { getShortDomain } from '@/config/env';

export interface LinkActivityItem {
  id: string;
  shortCode: string;
  originalUrl: string;
  clicks: number;
  createdAt: string;
  domain?: string;
}

export interface RecentActivityFeedProps {
  activities?: LinkActivityItem[];
  isLoading?: boolean;
}

const DEFAULT_ACTIVITIES: LinkActivityItem[] = [
  {
    id: 'act_1',
    shortCode: 'v2-launch',
    originalUrl: 'https://flux.to/blog/high-performance-edge-router-v2',
    clicks: 1420,
    createdAt: '2 mins ago',
  },
  {
    id: 'act_2',
    shortCode: 'docs-api',
    originalUrl: 'https://flux.to/docs/reference/openapi-v1',
    clicks: 890,
    createdAt: '15 mins ago',
  },
  {
    id: 'act_3',
    shortCode: 'summer-sale',
    originalUrl: 'https://store.acme.com/collections/summer-2026?utm_source=twitter',
    clicks: 3410,
    createdAt: '1 hour ago',
  },
  {
    id: 'act_4',
    shortCode: 'discord-invite',
    originalUrl: 'https://discord.gg/flux-dev-mesh-community',
    clicks: 560,
    createdAt: '3 hours ago',
  },
];

export function RecentActivityFeed({
  activities = DEFAULT_ACTIVITIES,
  isLoading = false,
}: RecentActivityFeedProps) {
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = (id: string, shortCode: string, domain = getShortDomain()) => {
    navigator.clipboard?.writeText(`https://${domain}/${shortCode}`);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div className="flex items-center justify-between border-b border-zinc-200 bg-zinc-50/75 px-5 py-3.5 dark:border-zinc-800 dark:bg-zinc-900/50">
        <div>
          <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-600 dark:text-zinc-300">
            Recent Link Activity
          </h3>
          <p className="text-[11px] text-zinc-400">
            Latest shortened links and live click counts across your workspace.
          </p>
        </div>
        <Link to="/links">
          <Button variant="ghost" size="sm" className="text-xs">
            <span>View All Links</span>
            <ArrowUpRight className="ml-1 h-3.5 w-3.5" />
          </Button>
        </Link>
      </div>

      <div className="divide-y divide-zinc-100 dark:divide-zinc-900">
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="flex items-center justify-between p-4 animate-pulse">
              <div className="space-y-2 w-2/3">
                <div className="h-4 w-32 rounded bg-zinc-200 dark:bg-zinc-800" />
                <div className="h-3 w-48 rounded bg-zinc-100 dark:bg-zinc-900" />
              </div>
              <div className="h-6 w-16 rounded bg-zinc-200 dark:bg-zinc-800" />
            </div>
          ))
        ) : activities.length === 0 ? (
          <div className="p-8 text-center text-xs text-zinc-400">
            No link activity recorded yet.
          </div>
        ) : (
          activities.map((item) => {
            const domain = item.domain || getShortDomain();
            const shortUrl = `${domain}/${item.shortCode}`;
            const isCopied = copiedId === item.id;

            return (
              <div
                key={item.id}
                className="flex flex-col justify-between gap-3 p-4 transition-colors hover:bg-zinc-50/60 dark:hover:bg-zinc-900/40 sm:flex-row sm:items-center"
              >
                <div className="min-w-0 space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                      {shortUrl}
                    </span>
                    <button
                      type="button"
                      onClick={() => handleCopy(item.id, item.shortCode, domain)}
                      className="rounded p-1 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700 dark:text-zinc-500 dark:hover:bg-zinc-900 dark:hover:text-zinc-300"
                      title="Copy short link"
                    >
                      {isCopied ? (
                        <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                    </button>
                  </div>
                  <p className="truncate font-mono text-[11px] text-zinc-400 dark:text-zinc-500">
                    {item.originalUrl}
                  </p>
                </div>

                <div className="flex items-center gap-4 shrink-0">
                  <div className="text-right">
                    <div className="flex items-center gap-1 font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                      <BarChart2 className="h-3.5 w-3.5 text-zinc-400" />
                      <span>{item.clicks.toLocaleString()}</span>
                    </div>
                    <div className="text-[10px] text-zinc-400">
                      {item.createdAt}
                    </div>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

export default RecentActivityFeed;
