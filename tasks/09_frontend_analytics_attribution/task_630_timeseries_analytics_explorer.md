---
id: TASK-630
title: ClickHouse Time-Series Analytics Explorer & Geographic Heatmap
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/07_frontend_core_ops/task_611_dashboard_overview.md
references:
  - database/clickhouse_analytics_schema.sql
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - analytics
  - clickhouse
  - timeseries
  - geo-heatmap
  - recharts
---

# TASK-630: ClickHouse Time-Series Analytics Explorer & Geographic Heatmap

## 1. Goal
Implement the full Analytics Explorer page (`/analytics`) featuring interactive time-range filters (1h, 24h, 7d, 30d, custom), ClickHouse time-series click charts, geographic country intensity heatmap, referrer breakdown, and device/OS donut charts.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/analytics/AnalyticsPage.tsx`
- `apps/frontend/src/components/analytics/TimeSeriesAreaChart.tsx`
- `apps/frontend/src/components/analytics/GeographicChoropleth.tsx`
- `apps/frontend/src/components/analytics/ReferrerBreakdownTable.tsx`
- `apps/frontend/src/components/analytics/DeviceDonutChart.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`getAnalyticsSummary`, `getLinkMetrics`, `getAnalyticsStreamMetrics`)
- `recharts`

## 4. Referenced Architecture & Product Specs
- [docs/growth/time_series_analytics.md](file:///home/logan78/Desktop/flux/docs/growth/time_series_analytics.md)
- [tasks/02_growth/task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_201_clickhouse_pipeline.md)

## 5. Acceptance Criteria
- [ ] Time-series area chart allows zooming, panning, and hovering with tooltips showing date and click volume.
- [ ] Geographic distribution ranks top countries with flag icons, click counts, and percentage of total traffic.
- [ ] Live Stream meter displays total ingested events and real-time gzip stream compression ratio factor.

## 6. Target Deliverables
- `apps/frontend/src/pages/analytics/AnalyticsPage.tsx`
- `apps/frontend/src/components/analytics/TimeSeriesAreaChart.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write chart data transformation unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Analytics Explorer components.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/analytics/AnalyticsPage.test.tsx`
