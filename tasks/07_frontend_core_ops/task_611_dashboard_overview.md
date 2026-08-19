---
id: TASK-611
title: Minimalist Overview Dashboard with Dub-Style Sparklines & Metrics
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_601_api_client_query.md
  - tasks/06_frontend_foundation/task_603_app_shell_navigation.md
references:
  - api/openapi_v1_core.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - dashboard
  - overview
  - sparklines
  - kpi
  - dub-style
---

# TASK-611: Minimalist Overview Dashboard with Dub-Style Sparklines & Metrics

## 1. Goal
Implement the main overview dashboard (`/dashboard`) adhering to the **Dub.co & Notion minimalist aesthetic**: clean flat KPI metric cards (Total Links, 24h Clicks, CTR, Active Domains), subtle monochrome hourly click sparklines, and compact link activity feed.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/dashboard/OverviewPage.tsx`
- `apps/frontend/src/components/dashboard/MetricCardsGrid.tsx`
- `apps/frontend/src/components/dashboard/HourlyClickSparkline.tsx`
- `apps/frontend/src/components/dashboard/RecentActivityFeed.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`getAnalyticsSummary` endpoint)
- `recharts` for time-series sparklines

## 4. Referenced Architecture & Product Specs
- [docs/core/analytics_pipeline.md](file:///home/logan78/Desktop/flux/docs/core/analytics_pipeline.md)
- [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md)

## 5. Acceptance Criteria
- [x] 4 KPI cards styled with clean borders (`border-zinc-200 dark:border-zinc-800`), crisp numbers, and subtle percentage badges (+12% in emerald/zinc).
- [x] Dub-style clean hourly click sparkline with subtle fill opacity and clean tooltip without clutter.
- [x] Inline Quick Link shortening bar with solid black/white action button.

## 6. Target Deliverables
- `apps/frontend/src/pages/dashboard/OverviewPage.tsx`
- `apps/frontend/src/components/dashboard/MetricCardsGrid.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write Overview page unit test with mocked API client.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement Overview components.
- [x] Step 4: Confirm tests pass with `bun test`.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/dashboard/OverviewPage.test.tsx`
