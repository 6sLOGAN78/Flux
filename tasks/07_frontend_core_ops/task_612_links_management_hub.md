---
id: TASK-612
title: Dub-Style Links Hub with Search, Compact Row Cards & Bulk Actions
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_601_api_client_query.md
  - tasks/06_frontend_foundation/task_604_ui_component_primitives.md
references:
  - api/openapi_v1_core.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - links
  - dub-style
  - bulk-actions
  - filtering
  - minimalist
---

# TASK-612: Dub-Style Links Hub with Search, Compact Row Cards & Bulk Actions

## 1. Goal
Build the central Links Management page (`/links`) with Dub-style compact link row cards, instant search, subtle tag pills, one-click copy with toast feedback, and slide-over link creation drawer.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/links/LinksListPage.tsx`
- `apps/frontend/src/components/links/LinksTable.tsx`
- `apps/frontend/src/components/links/CreateLinkDrawer.tsx`
- `apps/frontend/src/components/links/BulkCategorizeModal.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`createLink`, `bulkCategorizeLinks`, `getLink`)
- `@tanstack/react-table`

## 4. Referenced Architecture & Product Specs
- [docs/core/link_management.md](file:///home/logan78/Desktop/flux/docs/core/link_management.md)
- [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md)

## 5. Acceptance Criteria
- [x] Dub-style link card rows displaying favicon, short link slug, destination URL preview, click count badge, and quick copy button.
- [x] Minimalist search bar with category filters and keyboard shortcut indicators.
- [x] Slide-over link creation drawer with clean inputs, custom slug input, and single solid black/white "Create link" CTA.

## 6. Target Deliverables
- `apps/frontend/src/pages/links/LinksListPage.tsx`
- `apps/frontend/src/components/links/LinksTable.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write unit tests for table filtering and drawer validation.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement LinksListPage and components.
- [x] Step 4: Confirm tests pass with `bun test`.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/links/LinksListPage.test.tsx`
