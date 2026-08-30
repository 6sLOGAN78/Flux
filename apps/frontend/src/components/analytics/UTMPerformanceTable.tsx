import React, { useState } from 'react';
import { Tag } from 'lucide-react';
import { useAnalyticsUTM } from '@/hooks/useAnalyticsQuery';
import { Loader2 } from 'lucide-react';

export function UTMPerformanceTable({ from, to }: { from?: string, to?: string }) {
  const [dimension, setDimension] = useState('utm_source');
  
  const { data, isLoading, isError } = useAnalyticsUTM(dimension, from, to);

  const utmData = data?.data || [];
  
  return (
    <div className="flex flex-col h-full rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
      <div className="flex items-center justify-between border-b border-zinc-100 p-4 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600 dark:bg-indigo-500/10 dark:text-indigo-400">
            <Tag className="h-4 w-4" />
          </div>
          <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">UTM Performance</h2>
        </div>
        <select 
          className="text-xs border rounded px-2 py-1 bg-white dark:bg-zinc-900 dark:border-zinc-800"
          value={dimension}
          onChange={(e) => setDimension(e.target.value)}
        >
          <option value="utm_source">Source</option>
          <option value="utm_medium">Medium</option>
          <option value="utm_campaign">Campaign</option>
          <option value="utm_term">Term</option>
          <option value="utm_content">Content</option>
        </select>
      </div>

      <div className="relative flex-1 p-4">
        {isLoading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-white/50 backdrop-blur-sm dark:bg-zinc-950/50">
            <Loader2 className="h-6 w-6 animate-spin text-zinc-400" />
          </div>
        )}
        {isError && (
          <div className="text-red-500 text-xs">Failed to load UTM data</div>
        )}
        {!isLoading && !isError && utmData.length === 0 && (
          <div className="flex h-full items-center justify-center text-xs text-zinc-400">
            No UTM data available for this range.
          </div>
        )}
        {!isLoading && !isError && utmData.length > 0 && (
          <div className="space-y-3">
            {utmData.map((item: any, i: number) => (
              <div key={i} className="flex items-center justify-between">
                <div className="flex items-center gap-3 truncate">
                  <span className="truncate text-sm font-medium text-zinc-900 dark:text-zinc-100">
                    {item.utm_value}
                  </span>
                </div>
                <div className="text-right">
                  <div className="text-sm font-semibold text-zinc-900 dark:text-zinc-100">
                    {item.clicks.toLocaleString()}
                  </div>
                  <div className="text-[10px] text-zinc-500">
                    {item.unique_visitors.toLocaleString()} uv
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
