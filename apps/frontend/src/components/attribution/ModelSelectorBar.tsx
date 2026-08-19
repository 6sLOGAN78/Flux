import React from 'react';
import { Layers, Sparkles, PieChart, Clock, GitMerge } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export type AttributionModelType =
  | 'first_touch'
  | 'last_touch'
  | 'linear'
  | 'time_decay'
  | 'u_shaped';

export interface AttributionModelOption {
  id: AttributionModelType;
  name: string;
  description: string;
  formula: string;
}

export const ATTRIBUTION_MODELS: AttributionModelOption[] = [
  {
    id: 'first_touch',
    name: 'First-Touch',
    description: '100% credit to the first marketing touchpoint that introduced the user.',
    formula: '100% / 0% / 0%',
  },
  {
    id: 'last_touch',
    name: 'Last-Touch',
    description: '100% credit to the final touchpoint directly preceding the conversion.',
    formula: '0% / 0% / 100%',
  },
  {
    id: 'linear',
    name: 'Linear',
    description: 'Equal credit distributed evenly across every customer interaction.',
    formula: '1/N equal share',
  },
  {
    id: 'time_decay',
    name: 'Time-Decay',
    description: 'Exponential weighting increasing as touchpoints get closer to conversion.',
    formula: '7-day half-life',
  },
  {
    id: 'u_shaped',
    name: 'Position-Based (U-Shaped)',
    description: '40% to first touch, 40% to lead conversion, 20% split among middle touches.',
    formula: '40% / 20% / 40%',
  },
];

export interface ModelSelectorBarProps {
  selectedModel: AttributionModelType;
  onSelectModel: (model: AttributionModelType) => void;
  className?: string;
}

export function ModelSelectorBar({
  selectedModel,
  onSelectModel,
  className,
}: ModelSelectorBarProps) {
  return (
    <div
      className={cn(
        'rounded-2xl border border-zinc-200 bg-white p-4 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Layers className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <span className="text-xs font-semibold text-zinc-900 dark:text-zinc-100">
            Attribution Algorithm Models
          </span>
        </div>
        <span className="text-[11px] text-zinc-400">
          ClickHouse Multi-Touch Graph
        </span>
      </div>

      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-5">
        {ATTRIBUTION_MODELS.map((m) => {
          const isSelected = selectedModel === m.id;
          return (
            <button
              key={m.id}
              type="button"
              onClick={() => onSelectModel(m.id)}
              className={cn(
                'flex flex-col justify-between rounded-xl border p-3 text-left transition-all',
                isSelected
                  ? 'border-zinc-900 bg-zinc-900 text-white shadow-xs dark:border-zinc-100 dark:bg-zinc-100 dark:text-zinc-900'
                  : 'border-zinc-200 bg-zinc-50/50 text-zinc-700 hover:border-zinc-300 hover:bg-zinc-100/60 dark:border-zinc-800 dark:bg-zinc-900/40 dark:text-zinc-300 dark:hover:border-zinc-700'
              )}
            >
              <div>
                <div className="flex items-center justify-between">
                  <span className="text-xs font-bold">{m.name}</span>
                </div>
                <p
                  className={cn(
                    'mt-1.5 text-[11px] leading-relaxed',
                    isSelected
                      ? 'text-zinc-300 dark:text-zinc-600'
                      : 'text-zinc-500 dark:text-zinc-400'
                  )}
                >
                  {m.description}
                </p>
              </div>

              <div className="mt-3 font-mono text-[10px] font-semibold opacity-80">
                {m.formula}
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

export default ModelSelectorBar;
