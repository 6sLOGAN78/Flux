---
id: TASK-652
title: Global Edge PoP Health, Anycast BGP & Disaster Recovery Monitor
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_603_app_shell_navigation.md
references:
  - api/asyncapi_events.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - global-ops
  - anycast
  - edge-health
  - failover
---

# TASK-652: Global Edge PoP Health, Anycast BGP & Disaster Recovery Monitor

## 1. Goal
Implement the Global Operations & System Health dashboard (`/ops/global-health`) providing an interactive world map of Anycast BGP Point-of-Presence (PoP) edge nodes, database replication sync latency monitors (<500ms SLA), and automated disaster recovery failover status.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/enterprise/GlobalOpsPage.tsx`
- `apps/frontend/src/components/ops/PoPWorldMap.tsx`
- `apps/frontend/src/components/ops/GeoReplicationLatencyGrid.tsx`
- `apps/frontend/src/components/ops/FailoverStatusCard.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`getGeoClusterHealth`, `getAnycastStatus`, `getFailoverStatus`)
- `@flux/zod` (`ZAnycastStatusResponse`, `ZGeoClusterHealthResponse`, `ZClusterFailoverStatus`)

## 4. Referenced Architecture & Product Specs
- [docs/global/anycast_dns_tls.md](file:///home/logan78/Desktop/flux/docs/global/anycast_dns_tls.md)
- [docs/global/geo_db_replication.md](file:///home/logan78/Desktop/flux/docs/global/geo_db_replication.md)
- [docs/global/ha_disaster_recovery.md](file:///home/logan78/Desktop/flux/docs/global/ha_disaster_recovery.md)

## 5. Acceptance Criteria
- [ ] PoP nodes display operational health (`healthy` green pulse, `withdrawn` grey, `degraded` amber) and latency ms.
- [ ] Geo-replication table tracks regional ping latencies with SLA compliance badge.
- [ ] Failover card shows active primary region and backup health with live circuit breaker trip status.

## 6. Target Deliverables
- `apps/frontend/src/pages/enterprise/GlobalOpsPage.tsx`
- `apps/frontend/src/components/ops/PoPWorldMap.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write PoP status calculation unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Global Ops page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/enterprise/GlobalOpsPage.test.tsx`
