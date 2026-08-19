import { describe, expect, it } from 'bun:test';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { GlobalOpsPage } from './GlobalOpsPage';
import { PoPWorldMap, EdgePoP } from '@/components/ops/PoPWorldMap';
import {
  GeoReplicationLatencyGrid,
  ReplicationNode,
} from '@/components/ops/GeoReplicationLatencyGrid';
import { FailoverStatusCard } from '@/components/ops/FailoverStatusCard';

describe('Global Edge PoP Health & Disaster Recovery Monitor', () => {
  const testQueryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  const mockPoPs: EdgePoP[] = [
    {
      id: 'pop_iad',
      code: 'IAD-01',
      city: 'Washington DC (Ashburn)',
      region: 'North America',
      latencyMs: 3.2,
      status: 'healthy',
      bgpRoutes: 14,
    },
    {
      id: 'pop_lhr',
      code: 'LHR-01',
      city: 'London Heathrow',
      region: 'Europe',
      latencyMs: 5.8,
      status: 'healthy',
      bgpRoutes: 12,
    },
    {
      id: 'pop_nrt',
      code: 'NRT-01',
      city: 'Tokyo Narita',
      region: 'Asia Pacific',
      latencyMs: 8.4,
      status: 'healthy',
      bgpRoutes: 10,
    },
  ];

  const mockReplication: ReplicationNode[] = [
    {
      region: 'us-east-1 (Primary)',
      dbRole: 'primary',
      replicationLagMs: 0,
      syncStatus: 'in_sync',
      slaMet: true,
    },
    {
      region: 'eu-central-1 (Frankfurt)',
      dbRole: 'standby',
      replicationLagMs: 38,
      syncStatus: 'in_sync',
      slaMet: true,
    },
  ];

  it('renders PoPWorldMap with edge locations, latency, and BGP status', () => {
    const html = renderToString(
      <PoPWorldMap pops={mockPoPs} />
    );

    expect(html).toContain('Global Anycast BGP Edge Network');
    expect(html).toContain('IAD-01');
    expect(html).toContain('3.2 ms');
    expect(html).toContain('LHR-01');
    expect(html).toContain('Healthy');
  });

  it('renders GeoReplicationLatencyGrid with replication lag and SLA compliance', () => {
    const html = renderToString(
      <GeoReplicationLatencyGrid nodes={mockReplication} />
    );

    expect(html).toContain('Multi-Region Database Replication');
    expect(html).toContain('us-east-1 (Primary)');
    expect(html).toContain('eu-central-1 (Frankfurt)');
    expect(html).toContain('38 ms');
    expect(html).toContain('SLA Compliant');
  });

  it('renders FailoverStatusCard with primary region and circuit breaker state', () => {
    const html = renderToString(
      <FailoverStatusCard
        primaryRegion="us-east-1"
        standbyRegion="eu-central-1"
        circuitBreaker="closed"
        onTriggerTestFailover={() => {}}
      />
    );

    expect(html).toContain('Disaster Recovery &amp; Failover Control');
    expect(html).toContain('us-east-1');
    expect(html).toContain('Circuit Breaker Normal');
    expect(html).toContain('Trigger Test Failover');
  });

  it('renders full GlobalOpsPage with PoP map, replication grid, and failover control', () => {
    const html = renderToString(
      <QueryClientProvider client={testQueryClient}>
        <MemoryRouter>
          <GlobalOpsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(html).toContain('Global Edge Operations &amp; HA Health');
    expect(html).toContain('Global Anycast BGP Edge Network');
    expect(html).toContain('Multi-Region Database Replication');
  });
});
