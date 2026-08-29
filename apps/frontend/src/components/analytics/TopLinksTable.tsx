import React from 'react';
import { Link2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { getShortDomain } from '@/config/env';

export interface TopLinkItem {
  link_id: string;
  short_code: string;
  clicks: number;
}

export interface TopLinksTableProps {
  links: TopLinkItem[];
  className?: string;
}

export function TopLinksTable({ links, className }: TopLinksTableProps) {
  const domain = getShortDomain();
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
            <Link2 className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Top Links
            </h2>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Highest performing short links by click volume.
          </p>
        </div>
      </div>

      <div className="space-y-3">
        {links.length === 0 ? (
          <div className="text-center text-xs text-zinc-500 p-4">No top links found.</div>
        ) : (
          links.map((link) => (
            <div key={link.link_id} className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {domain}/{link.short_code}
                </span>
              </div>
              <div className="font-mono font-bold text-zinc-900 dark:text-zinc-100">
                {link.clicks.toLocaleString()}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

export default TopLinksTable;
