---
id: TASK-620
title: Marketing Campaigns & Visual Multi-Channel UTM Builder
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/07_frontend_core_ops/task_612_links_management_hub.md
references:
  - api/openapi_v2_growth.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - campaigns
  - utm-builder
  - marketing
  - growth
---

# TASK-620: Marketing Campaigns & Visual Multi-Channel UTM Builder

## 1. Goal
Implement the Campaigns Management page (`/campaigns`) with a visual UTM Parameter Builder featuring channel presets (Google Ads, Meta, Twitter, LinkedIn, Newsletter) and batch campaign short-link generation.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/growth/CampaignsPage.tsx`
- `apps/frontend/src/components/campaigns/UTMBuilderStudio.tsx`
- `apps/frontend/src/components/campaigns/CampaignListTable.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`createCampaign`, `ZCreateCampaignInput`)

## 4. Referenced Architecture & Product Specs
- [docs/growth/campaign_utm_builder.md](file:///home/logan78/Desktop/flux/docs/growth/campaign_utm_builder.md)
- [tasks/02_growth/task_202_utm_builder.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_202_utm_builder.md)

## 5. Acceptance Criteria
- [x] Presets automatically populate `utm_source`, `utm_medium`, and `utm_campaign` based on selected channel.
- [x] Real-time output preview shows the full sanitized URL and generated short code.
- [x] Campaign table lists active campaigns, aggregate click counts, and conversion metrics.

## 6. Target Deliverables
- `apps/frontend/src/pages/growth/CampaignsPage.tsx`
- `apps/frontend/src/components/campaigns/UTMBuilderStudio.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write UTM builder parameter serialization unit tests.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement UTM Studio and Campaigns page.
- [x] Step 4: Confirm tests pass.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/components/campaigns/UTMBuilderStudio.test.tsx`
