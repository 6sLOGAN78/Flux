import React from 'react';
import { Globe2, Radio, Server, Activity, Zap } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface EdgePoP {
  id: string;
  code: string;
  city: string;
  region: string;
  latencyMs: number;
  status: 'healthy' | 'degraded' | 'withdrawn';
  bgpRoutes: number;
}

export interface PoPWorldMapProps {
  pops: EdgePoP[];
  className?: string;
}

export function PoPWorldMap({
  pops,
  className,
}: PoPWorldMapProps) {
  const getStatusBadge = (status: EdgePoP['status']) => {
    switch (status) {
      case 'healthy':
        return <Badge variant="emerald" size="sm" dot>Healthy</Badge>;
      case 'degraded':
        return <Badge variant="amber" size="sm" dot>Degraded</Badge>;
      case 'withdrawn':
      default:
        return <Badge variant="zinc" size="sm">Withdrawn</Badge>;
    }
  };

  return (
    <div
      className={cn(
        'rounded-2xl border border-zinc-200 bg-white p-6 shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="flex flex-col justify-between gap-4 border-b border-zinc-100 pb-4 dark:border-zinc-900 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <Globe2 className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
            <h2 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
              Global Anycast BGP Edge Network
            </h2>
            <Badge variant="emerald" size="sm" dot>
              AS13335 Announced
            </Badge>
          </div>
          <p className="mt-1 text-xs text-zinc-500 dark:text-zinc-400">
            Real-time Point of Presence (PoP) health, BGP route telemetry, and sub-10ms edge latency.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="zinc" size="sm" className="font-mono">
            Anycast IP: 198.51.100.1
          </Badge>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {pops.map((pop) => (
          <div
            key={pop.id}
            className="flex flex-col justify-between rounded-xl border border-zinc-200 bg-zinc-50/50 p-4 transition-all hover:border-zinc-300 dark:border-zinc-800 dark:bg-zinc-900/40 dark:hover:border-zinc-700"
          >
            <div>
              <div className="flex items-center justify-between">
                <span className="font-mono text-sm font-bold text-zinc-900 dark:text-zinc-100">
                  {pop.code}
                </span>
                {getStatusBadge(pop.status)}
              </div>

              <div className="mt-1 text-xs font-medium text-zinc-700 dark:text-zinc-300">
                {pop.city}
              </div>
              <div className="text-[11px] text-zinc-400">
                {pop.region}
              </div>
            </div>

            <div className="mt-4 flex items-center justify-between border-t border-zinc-100 pt-3 dark:border-zinc-800/60">
              <div className="flex items-center gap-1 text-[11px] text-zinc-500 font-mono">
                <Radio className="h-3 w-3 text-zinc-400" />
                <span>{`${pop.bgpRoutes} BGP peers`}</span>
              </div>

              <div className="font-mono text-xs font-bold text-emerald-600 dark:text-emerald-400">
                {`${pop.latencyMs.toFixed(1)} ms`}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default PoPWorldMap;
