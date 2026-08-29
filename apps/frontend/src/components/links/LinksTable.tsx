import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Copy,
  Check,
  ExternalLink,
  BarChart2,
  QrCode,
  MoreVertical,
  Globe2,
  Trash2,
  Tag,
} from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils';
import { getShortDomain } from '@/config/env';

export interface LinkItem {
  id: string;
  shortCode: string;
  destinationUrl: string;
  title?: string;
  clicks: number;
  createdAt: string;
  category?: string;
  domain?: string;
}

export interface LinksTableProps {
  links: LinkItem[];
  selectedLinkIds: string[];
  onToggleSelect: (id: string) => void;
  onSelectAll: () => void;
  onDeleteLink?: (id: string) => void;
  isLoading?: boolean;
}

export function LinksTable({
  links,
  selectedLinkIds,
  onToggleSelect,
  onSelectAll,
  onDeleteLink,
  isLoading = false,
}: LinksTableProps) {
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = (id: string, shortUrl: string) => {
    navigator.clipboard?.writeText(`http${window.location.hostname === 'localhost' ? '' : 's'}://${shortUrl}`);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const isAllSelected =
    links.length > 0 && selectedLinkIds.length === links.length;

  return (
    <div className="overflow-hidden rounded-xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      {/* Table Header / Selection Bar */}
      <div className="flex items-center justify-between border-b border-zinc-200 bg-zinc-50/75 px-4 py-2.5 text-xs text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            checked={isAllSelected}
            onChange={onSelectAll}
            className="h-4 w-4 rounded border-zinc-300 text-zinc-900 focus:ring-zinc-900/20 dark:border-zinc-700 dark:bg-zinc-900 dark:checked:bg-zinc-100"
            aria-label="Select all links"
          />
          <span className="font-medium text-zinc-700 dark:text-zinc-300">
            {selectedLinkIds.length > 0
              ? `${selectedLinkIds.length} of ${links.length} selected`
              : `All Links (${links.length})`}
          </span>
        </div>
        <div className="hidden items-center gap-8 sm:flex">
          <span className="w-24 text-right">Analytics</span>
          <span className="w-28 text-right">Created</span>
        </div>
      </div>

      {/* Rows */}
      <div className="divide-y divide-zinc-100 dark:divide-zinc-900">
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="flex items-center gap-4 p-4 animate-pulse">
              <div className="h-4 w-4 rounded bg-zinc-200 dark:bg-zinc-800" />
              <div className="h-8 w-8 rounded-full bg-zinc-200 dark:bg-zinc-800" />
              <div className="flex-1 space-y-2">
                <div className="h-4 w-40 rounded bg-zinc-200 dark:bg-zinc-800" />
                <div className="h-3 w-64 rounded bg-zinc-100 dark:bg-zinc-900" />
              </div>
              <div className="h-6 w-16 rounded bg-zinc-200 dark:bg-zinc-800" />
            </div>
          ))
        ) : links.length === 0 ? (
          <div className="p-12 text-center text-xs text-zinc-400">
            No links found matching your search.
          </div>
        ) : (
          links.map((link) => {
            const domain = link.domain || getShortDomain();
            const shortUrl = `${domain}/${link.shortCode}`;
            const isSelected = selectedLinkIds.includes(link.id);
            const isCopied = copiedId === link.id;

            return (
              <div
                key={link.id}
                className={cn(
                  'group flex flex-col justify-between gap-3 p-4 transition-colors hover:bg-zinc-50/60 dark:hover:bg-zinc-900/40 sm:flex-row sm:items-center',
                  isSelected && 'bg-zinc-50/80 dark:bg-zinc-900/60'
                )}
              >
                <div className="flex items-start gap-3 min-w-0">
                  <div className="pt-0.5">
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => onToggleSelect(link.id)}
                      className="h-4 w-4 rounded border-zinc-300 text-zinc-900 focus:ring-zinc-900/20 dark:border-zinc-700 dark:bg-zinc-900"
                    />
                  </div>

                  {/* Favicon or fallback icon */}
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-900">
                    <Globe2 className="h-4 w-4 text-zinc-500" />
                  </div>

                  {/* Link Details */}
                  <div className="min-w-0 space-y-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-mono text-xs font-semibold text-zinc-900 dark:text-zinc-100">
                        {shortUrl}
                      </span>

                      <button
                        type="button"
                        onClick={() => handleCopy(link.id, shortUrl)}
                        className="rounded p-1 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-900 dark:hover:text-zinc-300"
                        title="Copy short URL"
                      >
                        {isCopied ? (
                          <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
                        ) : (
                          <Copy className="h-3 w-3" />
                        )}
                      </button>

                      {link.category && (
                        <Badge variant="zinc" size="sm">
                          {link.category}
                        </Badge>
                      )}
                    </div>

                    {link.title && (
                      <p className="text-xs font-medium text-zinc-700 dark:text-zinc-300">
                        {link.title}
                      </p>
                    )}

                    <div className="flex items-center gap-1.5 font-mono text-[11px] text-zinc-400 dark:text-zinc-500">
                      <span className="truncate max-w-sm sm:max-w-md">
                        {link.destinationUrl}
                      </span>
                      <a
                        href={link.destinationUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="opacity-0 transition-opacity group-hover:opacity-100 hover:text-zinc-700 dark:hover:text-zinc-300"
                      >
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    </div>
                  </div>
                </div>

                {/* Right side stats and quick actions */}
                <div className="flex items-center justify-between gap-6 sm:justify-end shrink-0 pt-2 sm:pt-0">
                  <div className="flex items-center gap-1 font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                    <BarChart2 className="h-3.5 w-3.5 text-zinc-400" />
                    <span>{link.clicks.toLocaleString()}</span>
                  </div>

                  <span className="text-[11px] text-zinc-400 font-mono">
                    {new Date(link.createdAt).toLocaleDateString(undefined, {
                      month: 'short',
                      day: 'numeric',
                    })}
                  </span>

                  <div className="flex items-center gap-1">
                    <Link to={`/qr-studio?url=https://${shortUrl}`}>
                      <button
                        type="button"
                        className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-900 dark:hover:text-zinc-300"
                        title="Generate QR code"
                      >
                        <QrCode className="h-3.5 w-3.5" />
                      </button>
                    </Link>

                    {onDeleteLink && (
                      <button
                        type="button"
                        onClick={() => onDeleteLink(link.id)}
                        className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-950/50 dark:hover:text-red-400"
                        title="Delete link"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    )}
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

export default LinksTable;
