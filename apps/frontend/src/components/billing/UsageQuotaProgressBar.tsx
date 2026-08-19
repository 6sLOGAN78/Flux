import React from 'react';
import { Gauge, AlertTriangle } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface QuotaItem {
  label: string;
  current: number;
  max: number;
  unit: string;
}

export interface UsageQuotaProgressBarProps {
  quotas: QuotaItem[];
  className?: string;
}

export function UsageQuotaProgressBar({
  quotas,
  className,
}: UsageQuotaProgressBarProps) {
  return (
    <div
      className={cn(
        'space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <Gauge className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Usage &amp; Resource Quotas
          </h3>
        </div>
        <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
          Monthly allocation limits reset on your subscription billing cycle date.
        </p>
      </div>

      <div className="space-y-5">
        {quotas.map((q) => {
          const percent = Math.min(100, Math.round((q.current / q.max) * 100));
          const isWarning = percent >= 80;

          return (
            <div key={q.label} className="space-y-2">
              <div className="flex items-center justify-between text-xs">
                <span className="font-medium text-zinc-900 dark:text-zinc-100">
                  {q.label}
                </span>

                <div className="flex items-center gap-2 font-mono">
                  <span className="font-bold text-zinc-900 dark:text-zinc-100">
                    {q.current.toLocaleString()}
                  </span>
                  <span className="text-zinc-400">/</span>
                  <span className="text-zinc-500">
                    {`${q.max.toLocaleString()} ${q.unit}`}
                  </span>
                  <span className={cn(
                    'ml-1 text-[11px] font-semibold',
                    isWarning ? 'text-amber-600 dark:text-amber-400' : 'text-zinc-400'
                  )}>
                    {`(${percent}%)`}
                  </span>
                </div>
              </div>

              {/* Progress bar */}
              <div className="h-2 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-900">
                <div
                  style={{ width: `${percent}%` }}
                  className={cn(
                    'h-full rounded-full transition-all duration-500',
                    isWarning
                      ? 'bg-amber-500'
                      : 'bg-zinc-900 dark:bg-zinc-100'
                  )}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default UsageQuotaProgressBar;
