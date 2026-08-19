import React from 'react';
import { Filter, ArrowDown, Users, TrendingDown } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface FunnelStepItem {
  id: string;
  name: string;
  visitors: number;
  dropoffPercentage: number;
  conversionRateFromStart: number;
}

export interface FunnelVisualizerProps {
  steps: FunnelStepItem[];
  funnelName?: string;
  className?: string;
}

export function FunnelVisualizer({
  steps,
  funnelName = 'Conversion Funnel',
  className,
}: FunnelVisualizerProps) {
  const topStepVisitors = steps[0]?.visitors || 1;
  const finalStep = steps[steps.length - 1];
  const overallConversion = finalStep
    ? ((finalStep.visitors / topStepVisitors) * 100).toFixed(2)
    : '0.00';

  return (
    <div
      className={cn(
        'space-y-6 rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-2 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              {funnelName}
            </h2>
            <Badge variant="emerald" size="sm">
              {`${overallConversion}% overall conversion`}
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Step-by-step visitor progression and conversion drop-off analysis.
          </p>
        </div>
      </div>

      {/* Sequential Funnel Steps */}
      <div className="space-y-3">
        {steps.map((step, idx) => {
          const widthPercent = Math.max(
            15,
            Math.round((step.visitors / topStepVisitors) * 100)
          );

          return (
            <div key={step.id} className="space-y-2">
              <div className="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 transition-all dark:border-zinc-800 dark:bg-zinc-900/40">
                <div className="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
                  <div className="flex items-center gap-3">
                    <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-zinc-900 font-mono text-xs font-bold text-white dark:bg-zinc-100 dark:text-zinc-900">
                      {idx + 1}
                    </span>
                    <div>
                      <h4 className="text-xs font-bold text-zinc-900 dark:text-zinc-100">
                        {step.name}
                      </h4>
                      <div className="flex items-center gap-2 text-[11px] text-zinc-400 font-mono">
                        <span>{`${step.conversionRateFromStart}% of top of funnel`}</span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-4">
                    {step.dropoffPercentage > 0 && (
                      <span className="flex items-center gap-1 text-[11px] font-medium text-rose-600 dark:text-rose-400">
                        <TrendingDown className="h-3 w-3" />
                        {`-${step.dropoffPercentage}% dropoff`}
                      </span>
                    )}

                    <div className="text-right font-mono">
                      <span className="text-sm font-bold text-zinc-900 dark:text-zinc-100">
                        {step.visitors.toLocaleString()}
                      </span>
                      <span className="ml-1 text-[11px] text-zinc-400">users</span>
                    </div>
                  </div>
                </div>

                {/* Relative Width Bar */}
                <div className="mt-3 h-2 w-full overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-800">
                  <div
                    style={{ width: `${widthPercent}%` }}
                    className="h-full rounded-full bg-zinc-900 transition-all duration-500 dark:bg-zinc-100"
                  />
                </div>
              </div>

              {idx < steps.length - 1 && (
                <div className="flex justify-center py-0.5">
                  <ArrowDown className="h-3.5 w-3.5 text-zinc-300 dark:text-zinc-700" />
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default FunnelVisualizer;
