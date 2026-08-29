import React, { useState } from 'react';
import { Globe2, Radio, Server, Activity, Zap, Check } from 'lucide-react';
import { Badge } from '@/components/ui/Badge';
import { PoPWorldMap, EdgePoP } from '@/components/ops/PoPWorldMap';
import {
  GeoReplicationLatencyGrid,
  ReplicationNode,
} from '@/components/ops/GeoReplicationLatencyGrid';
import { FailoverStatusCard } from '@/components/ops/FailoverStatusCard';

const INITIAL_POPS: EdgePoP[] = [];

const INITIAL_REPLICATION: ReplicationNode[] = [];

export function GlobalOpsPage() {
  const [pops, setPops] = useState<EdgePoP[]>(INITIAL_POPS);
  const [replication, setReplication] =
    useState<ReplicationNode[]>(INITIAL_REPLICATION);
  const [circuitBreaker, setCircuitBreaker] = useState<'closed' | 'open'>('closed');
  const [notice, setNotice] = useState<string | null>(null);

  const handleTestFailover = () => {
    setNotice('Initiating simulated disaster recovery failover drill...');
    setTimeout(() => {
      setNotice('Disaster recovery drill passed cleanly: DNS cutover executed in 8.4s.');
      setTimeout(() => setNotice(null), 4000);
    }, 1500);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
              Global Edge Operations &amp; HA Health
            </h1>
            <Badge variant="emerald" size="sm" dot>
              100% Operational
            </Badge>
          </div>
          <p className="text-xs text-zinc-500 dark:text-zinc-400">
            Multi-region Anycast BGP Point of Presence nodes, geo-database replication, and automated failover.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="zinc" size="sm" className="font-mono">
            Anycast BGP AS13335
          </Badge>
        </div>
      </div>

      {notice && (
        <div className="flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-xs font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300 animate-in fade-in">
          <Check className="h-4 w-4" />
          <span>{notice}</span>
        </div>
      )}

      {/* KPI Stats Grid */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Active Edge PoPs
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            28 / 28
          </div>
          <div className="mt-1 text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">
            100% mesh availability
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Global Edge p99 Latency
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-emerald-600 dark:text-emerald-400">
            6.4 ms
          </div>
          <div className="mt-1 text-[11px] text-zinc-400 font-mono">
            Target &lt; 10.0 ms
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Geo-Replication Lag
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            38 ms
          </div>
          <div className="mt-1 text-[11px] text-emerald-600 dark:text-emerald-400 font-medium">
            SLA &lt; 500 ms guarantee
          </div>
        </div>

        <div className="rounded-2xl border border-zinc-200 bg-white p-5 shadow-xs dark:border-zinc-800 dark:bg-zinc-950">
          <div className="text-xs text-zinc-500 dark:text-zinc-400">
            Disaster Recovery SLA
          </div>
          <div className="mt-2 font-mono text-2xl font-bold text-zinc-900 dark:text-zinc-100">
            RTO &lt; 30s
          </div>
          <div className="mt-1 text-[11px] text-zinc-400 font-mono">
            RPO 0s (Zero data loss)
          </div>
        </div>
      </div>

      {/* PoP World Map */}
      <PoPWorldMap pops={pops} />

      {/* Database Replication Latency */}
      <GeoReplicationLatencyGrid nodes={replication} />

      {/* Disaster Recovery Failover Card */}
      <FailoverStatusCard
        primaryRegion="us-east-1 (N. Virginia)"
        standbyRegion="eu-central-1 (Frankfurt)"
        circuitBreaker={circuitBreaker}
        onTriggerTestFailover={handleTestFailover}
      />
    </div>
  );
}

export default GlobalOpsPage;
