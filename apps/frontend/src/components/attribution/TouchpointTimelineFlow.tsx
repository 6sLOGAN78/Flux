import React from 'react';
import { GitCommit, ArrowRight, CheckCircle, Sparkles } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { AttributionModelType } from './ModelSelectorBar';
import { cn } from '@/lib/utils';

export interface TouchpointNode {
  channel: string;
  type: 'first_touch' | 'middle_touch' | 'last_touch';
  timestamp: string;
  weightPercentage: number;
}

export interface TouchpointTimelineFlowProps {
  touchpoints: TouchpointNode[];
  model?: AttributionModelType;
  className?: string;
}

export function TouchpointTimelineFlow({
  touchpoints,
  model = 'u_shaped',
  className,
}: TouchpointTimelineFlowProps) {
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
            <GitCommit className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Customer Journey Attribution Path
            </h2>
            <Badge variant="zinc" size="sm">
              Example Lead Journey
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Visual sequence illustrating how conversion credit is split across sequential touchpoints.
          </p>
        </div>
      </div>

      <div className="flex flex-col items-center justify-between gap-3 overflow-x-auto py-2 sm:flex-row">
        {touchpoints.map((tp, idx) => (
          <React.Fragment key={idx}>
            <div className="flex w-full flex-col items-center rounded-xl border border-zinc-200 bg-zinc-50/75 p-4 text-center dark:border-zinc-800 dark:bg-zinc-900/40 sm:w-44 shrink-0">
              <span className="text-[10px] font-semibold uppercase tracking-wider text-zinc-400">
                {tp.timestamp}
              </span>
              <h4 className="mt-1 text-xs font-bold text-zinc-900 dark:text-zinc-100">
                {tp.channel}
              </h4>
              <div className="mt-3 flex items-center justify-center gap-1.5">
                <Badge
                  variant={
                    tp.type === 'first_touch'
                      ? 'blue'
                      : tp.type === 'last_touch'
                      ? 'emerald'
                      : 'zinc'
                  }
                  size="sm"
                >
                  {tp.type === 'first_touch'
                    ? 'First Touch'
                    : tp.type === 'last_touch'
                    ? 'Conversion'
                    : 'Assist'}
                </Badge>
              </div>

              <div className="mt-2 font-mono text-sm font-bold text-zinc-900 dark:text-zinc-100">
                {`${tp.weightPercentage}%`}
              </div>
              <span className="text-[10px] text-zinc-400">attribution credit</span>
            </div>

            {idx < touchpoints.length - 1 && (
              <ArrowRight className="hidden h-4 w-4 text-zinc-400 sm:block shrink-0" />
            )}
          </React.Fragment>
        ))}
      </div>
    </div>
  );
}

export default TouchpointTimelineFlow;
