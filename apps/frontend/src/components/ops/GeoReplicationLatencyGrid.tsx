import React from 'react';
import { Database, CheckCircle2, AlertTriangle, ShieldCheck } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { cn } from '@/lib/utils';

export interface ReplicationNode {
  region: string;
  dbRole: 'primary' | 'standby';
  replicationLagMs: number;
  syncStatus: 'in_sync' | 'lagging';
  slaMet: boolean;
}

export interface GeoReplicationLatencyGridProps {
  nodes: ReplicationNode[];
  className?: string;
}

export function GeoReplicationLatencyGrid({
  nodes,
  className,
}: GeoReplicationLatencyGridProps) {
  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl border border-zinc-200 bg-white shadow-xs dark:border-zinc-800 dark:bg-zinc-950',
        className
      )}
    >
      <div className="border-b border-zinc-100 p-5 dark:border-zinc-900">
        <div className="flex items-center gap-2">
          <Database className="h-4 w-4 text-zinc-900 dark:text-zinc-100" />
          <h3 className="text-sm font-semibold tracking-tight text-zinc-900 dark:text-zinc-100">
            Multi-Region Database Replication
          </h3>
        </div>
        <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
          PostgreSQL read replica cluster synchronization with &lt;500ms global SLA guarantee.
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-xs text-zinc-700 dark:text-zinc-300">
          <thead className="border-b border-zinc-200 bg-zinc-50/75 text-[11px] font-semibold uppercase tracking-wider text-zinc-500 dark:border-zinc-800 dark:bg-zinc-900/50 dark:text-zinc-400">
            <tr>
              <th className="px-4 py-3">Cluster Region</th>
              <th className="px-4 py-3">Database Role</th>
              <th className="px-4 py-3 text-right">Replication Lag</th>
              <th className="px-4 py-3">Sync Status</th>
              <th className="px-4 py-3 text-right">SLA Guarantee</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-100 dark:divide-zinc-900">
            {nodes.map((node) => (
              <tr
                key={node.region}
                className="hover:bg-zinc-50/60 transition-colors dark:hover:bg-zinc-900/40"
              >
                <td className="px-4 py-3 font-semibold text-zinc-900 dark:text-zinc-100 font-mono">
                  {node.region}
                </td>

                <td className="px-4 py-3">
                  <Badge
                    variant={node.dbRole === 'primary' ? 'zinc' : 'blue'}
                    size="sm"
                    className="font-mono text-[10px] uppercase"
                  >
                    {node.dbRole}
                  </Badge>
                </td>

                <td className="px-4 py-3 text-right font-mono font-bold text-zinc-900 dark:text-zinc-100">
                  {node.replicationLagMs === 0 ? '0 ms (Leader)' : `${node.replicationLagMs} ms`}
                </td>

                <td className="px-4 py-3">
                  <Badge
                    variant={node.syncStatus === 'in_sync' ? 'emerald' : 'amber'}
                    size="sm"
                    dot
                  >
                    {node.syncStatus === 'in_sync' ? 'In Sync' : 'Lagging'}
                  </Badge>
                </td>

                <td className="px-4 py-3 text-right">
                  <Badge
                    variant={node.slaMet ? 'emerald' : 'rose'}
                    size="sm"
                  >
                    {node.slaMet ? 'SLA Compliant' : 'Breached'}
                  </Badge>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default GeoReplicationLatencyGrid;
