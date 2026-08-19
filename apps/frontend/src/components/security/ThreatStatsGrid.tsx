import React from 'react';
import { ShieldCheck, ShieldAlert, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface ThreatStats {
  totalScanned: number;
  threatsBlocked: number;
  reputationScore: number;
  quarantineCount: number;
}

export interface ThreatStatsGridProps {
  stats: ThreatStats;
  className?: string;
}

export function ThreatStatsGrid({
  stats,
  className,
}: ThreatStatsGridProps) {
  return (
    <div className={cn('grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4', className)}>
      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Scanned URLs</span>
          <ShieldCheck className="h-4 w-4 text-zinc-400" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          {stats.totalScanned.toLocaleString()}
        </div>
        <div className="mt-1 text-[11px] text-zinc-400">
          Continuous background scanning
        </div>
      </div>

      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Threats Neutralized</span>
          <ShieldAlert className="h-4 w-4 text-rose-500" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-rose-600 dark:text-rose-400">
          {stats.threatsBlocked.toLocaleString()}
        </div>
        <div className="mt-1 text-[11px] text-zinc-400">
          Phishing &amp; malware stopped
        </div>
      </div>

      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Domain Reputation Score</span>
          <CheckCircle2 className="h-4 w-4 text-emerald-500" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-emerald-600 dark:text-emerald-400">
          {`${stats.reputationScore.toFixed(2)}%`}
        </div>
        <div className="mt-1 text-[11px] text-zinc-400">
          Safe Browsing verified clean
        </div>
      </div>

      <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
        <div className="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
          <span>Quarantine Queue</span>
          <AlertTriangle className="h-4 w-4 text-amber-500" />
        </div>
        <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          {stats.quarantineCount}
        </div>
        <div className="mt-1 text-[11px] text-zinc-400">
          Awaiting moderator review
        </div>
      </div>
    </div>
  );
}

export default ThreatStatsGrid;
