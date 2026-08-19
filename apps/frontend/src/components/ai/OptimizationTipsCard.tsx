import React from 'react';
import { Lightbulb, ArrowRight, Zap, Target } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface OptimizationTip {
  id: string;
  title: string;
  description: string;
  impact: 'high' | 'medium' | 'low';
  actionLabel?: string;
}

export interface OptimizationTipsCardProps {
  tips: OptimizationTip[];
  onActionClick?: (id: string) => void;
  className?: string;
}

export function OptimizationTipsCard({
  tips,
  onActionClick,
  className,
}: OptimizationTipsCardProps) {
  const getImpactBadge = (impact: OptimizationTip['impact']) => {
    switch (impact) {
      case 'high':
        return <Badge variant="emerald" size="sm">High Impact</Badge>;
      case 'medium':
        return <Badge variant="blue" size="sm">Medium Impact</Badge>;
      case 'low':
      default:
        return <Badge variant="zinc" size="sm">Low Impact</Badge>;
    }
  };

  return (
    <div
      className={cn(
        'rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex items-center justify-between border-b border-zinc-100 pb-4 dark:border-zinc-900">
        <div>
          <div className="flex items-center gap-2">
            <Lightbulb className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              AI Optimization Recommendations
            </h3>
          </div>
          <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
            Algorithmic tips to maximize click-through rates and prevent bot drain.
          </p>
        </div>
      </div>

      <div className="mt-4 space-y-3">
        {tips.map((tip) => (
          <div
            key={tip.id}
            className="flex flex-col justify-between gap-3 rounded-xl border border-zinc-100 bg-zinc-50/50 p-4 transition-all hover:bg-zinc-100/50 dark:border-zinc-900 dark:bg-zinc-900/30 dark:hover:bg-zinc-900/60 sm:flex-row sm:items-center"
          >
            <div>
              <div className="flex items-center gap-2">
                <h4 className="text-xs font-bold text-zinc-900 dark:text-zinc-100">
                  {tip.title}
                </h4>
                {getImpactBadge(tip.impact)}
              </div>
              <p className="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
                {tip.description}
              </p>
            </div>

            {tip.actionLabel && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => onActionClick?.(tip.id)}
                rightIcon={<ArrowRight className="h-3 w-3" />}
                className="shrink-0"
              >
                {tip.actionLabel}
              </Button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export default OptimizationTipsCard;
