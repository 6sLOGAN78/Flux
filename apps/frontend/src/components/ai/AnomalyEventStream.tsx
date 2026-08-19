import React from 'react';
import {
  Activity,
  TrendingUp,
  TrendingDown,
  ShieldAlert,
  CheckCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface AnomalyEvent {
  id: string;
  type: 'traffic_spike' | 'traffic_drop' | 'bot_surge';
  slug: string;
  zScore: number;
  description: string;
  timestamp: string;
}

export interface AnomalyEventStreamProps {
  anomalies: AnomalyEvent[];
  onResolveAnomaly: (id: string) => void;
  className?: string;
}

export function AnomalyEventStream({
  anomalies,
  onResolveAnomaly,
  className,
}: AnomalyEventStreamProps) {
  const getAnomalyTypeDetails = (type: AnomalyEvent['type']) => {
    switch (type) {
      case 'traffic_spike':
        return {
          icon: <TrendingUp className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />,
          variant: 'emerald' as const,
          label: 'traffic_spike',
        };
      case 'traffic_drop':
        return {
          icon: <TrendingDown className="h-4 w-4 text-rose-600 dark:text-rose-400" />,
          variant: 'rose' as const,
          label: 'traffic_drop',
        };
      case 'bot_surge':
      default:
        return {
          icon: <ShieldAlert className="h-4 w-4 text-amber-600 dark:text-amber-400" />,
          variant: 'amber' as const,
          label: 'bot_surge',
        };
    }
  };

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 p-5 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <Activity className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Real-Time Anomaly Stream
          </h3>
        </div>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          Continuous Z-score outlier detection powered by automated statistical telemetry.
        </p>
      </div>

      <div className="divide-y divide-zinc-100 dark:divide-zinc-900">
        {anomalies.map((a) => {
          const details = getAnomalyTypeDetails(a.type);
          const zPrefix = a.zScore >= 0 ? '+' : '';

          return (
            <div
              key={a.id}
              className="flex flex-col justify-between gap-3 p-4 transition-colors hover:bg-zinc-50/50 dark:hover:bg-zinc-900/30 sm:flex-row sm:items-center"
            >
              <div className="flex items-start gap-3">
                <div className="mt-0.5 rounded-lg border border-zinc-200 bg-zinc-50 p-2 dark:border-zinc-800 dark:bg-zinc-900">
                  {details.icon}
                </div>

                <div>
                  <div className="flex items-center gap-2">
                    <Badge variant={details.variant} size="sm">
                      {details.label}
                    </Badge>
                    <span className="font-mono text-xs font-bold text-zinc-900 dark:text-zinc-100">
                      {`/${a.slug}`}
                    </span>
                    <span className="font-mono text-[11px] font-bold text-zinc-500">
                      {`Z: ${zPrefix}${a.zScore.toFixed(2)}`}
                    </span>
                  </div>

                  <p className="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
                    {a.description}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3">
                <span className="font-mono text-[11px] text-zinc-400">
                  {new Date(a.timestamp).toLocaleTimeString([], {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onResolveAnomaly(a.id)}
                  leftIcon={<CheckCircle className="h-3 w-3" />}
                >
                  Resolve
                </Button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default AnomalyEventStream;
