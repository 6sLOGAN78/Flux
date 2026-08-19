---
id: TASK-650
title: Predictive AI CTR Forecasting & Anomaly Detection Center
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/09_frontend_analytics_attribution/task_630_timeseries_analytics_explorer.md
references:
  - api/openapi_v4_enterprise.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - ai-insights
  - anomaly-detection
  - ctr-forecast
  - enterprise
---

# TASK-650: Predictive AI CTR Forecasting & Anomaly Detection Center

## 1. Goal
Implement the Enterprise AI Insights page (`/enterprise/ai-insights`) featuring predictive CTR trend charts, real-time traffic anomaly stream (traffic spikes, sudden drops, bot surges with Z-scores), and automated link optimization suggestions.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/enterprise/AIInsightsPage.tsx`
- `apps/frontend/src/components/ai/CTRPredictionChart.tsx`
- `apps/frontend/src/components/ai/AnomalyEventStream.tsx`
- `apps/frontend/src/components/ai/OptimizationTipsCard.tsx`

## 3. Dependencies & Prerequisites
- `@flux/zod` (`ZAnomalyDetectionResult`, `ZCTRPredictionResult`, `ZAnomalyLog`)

## 4. Referenced Architecture & Product Specs
- [docs/enterprise/ai_engine.md](file:///home/logan78/Desktop/flux/docs/enterprise/ai_engine.md)
- [tasks/04_enterprise/task_404_ai_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_404_ai_engine.md)

## 5. Acceptance Criteria
- [ ] CTR prediction chart plots historical vs predicted trajectory with confidence bands.
- [ ] Anomaly feed displays real-time badges for `traffic_spike` (green), `traffic_drop` (rose), and `bot_surge` (amber) with statistical Z-scores.
- [ ] Optimization tips present actionable recommendation chips (e.g. "Optimal posting time: 14:00 UTC").

## 6. Target Deliverables
- `apps/frontend/src/pages/enterprise/AIInsightsPage.tsx`
- `apps/frontend/src/components/ai/AnomalyEventStream.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write anomaly event rendering unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement AI Insights page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/enterprise/AIInsightsPage.test.tsx`
