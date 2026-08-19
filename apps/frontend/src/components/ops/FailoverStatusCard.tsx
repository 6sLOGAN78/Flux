import React from 'react';
import { ShieldCheck, RefreshCw, AlertTriangle, ArrowRight, HeartPulse } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface FailoverStatusCardProps {
  primaryRegion: string;
  standbyRegion: string;
  circuitBreaker: 'closed' | 'open' | 'half_open';
  onTriggerTestFailover: () => void;
  isLoading?: boolean;
  className?: string;
}

export function FailoverStatusCard({
  primaryRegion,
  standbyRegion,
  circuitBreaker,
  onTriggerTestFailover,
  isLoading = false,
  className,
}: FailoverStatusCardProps) {
  const isCircuitHealthy = circuitBreaker === 'closed';

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-4 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Disaster Recovery &amp; Failover Control
            </h3>
            <Badge
              variant={isCircuitHealthy ? 'emerald' : 'rose'}
              size="sm"
              dot
            >
              {isCircuitHealthy ? 'Circuit Breaker Normal' : 'Circuit Breaker Tripped'}
            </Badge>
          </div>
          <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
            Automated health probe polling every 2,000ms with sub-30s DNS cutover failover.
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={onTriggerTestFailover}
          isLoading={isLoading}
          leftIcon={<RefreshCw className="h-3.5 w-3.5" />}
        >
          Trigger Test Failover
        </Button>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div className="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 dark:border-zinc-800 dark:bg-zinc-900/40">
          <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
            Primary Active Region
          </div>
          <div className="mt-1 flex items-center justify-between">
            <span className="font-mono text-sm font-bold text-zinc-900 dark:text-zinc-100">
              {primaryRegion}
            </span>
            <Badge variant="emerald" size="sm" dot>
              100% Traffic
            </Badge>
          </div>
          <div className="mt-3 flex items-center gap-1 text-[11px] text-zinc-400 font-mono">
            <HeartPulse className="h-3 w-3 text-emerald-500" />
            <span>Health Check: 200 OK (2ms probe)</span>
          </div>
        </div>

        <div className="rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 dark:border-zinc-800 dark:bg-zinc-900/40">
          <div className="text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
            Hot Standby Region
          </div>
          <div className="mt-1 flex items-center justify-between">
            <span className="font-mono text-sm font-bold text-zinc-900 dark:text-zinc-100">
              {standbyRegion}
            </span>
            <Badge variant="blue" size="sm">
              Standby Ready
            </Badge>
          </div>
          <div className="mt-3 flex items-center gap-1 text-[11px] text-zinc-400 font-mono">
            <ShieldCheck className="h-3 w-3 text-blue-500" />
            <span>Replication Lag: 38ms (RTO &lt; 30s)</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default FailoverStatusCard;
