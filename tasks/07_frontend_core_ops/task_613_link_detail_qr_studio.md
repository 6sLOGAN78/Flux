---
id: TASK-613
title: Link Detail Editor, Interactive QR Code Studio & Categories Manager
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/07_frontend_core_ops/task_612_links_management_hub.md
references:
  - api/openapi_v1_core.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - link-detail
  - qr-code
  - categories
  - editor
---

# TASK-613: Link Detail Editor, Interactive QR Code Studio & Categories Manager

## 1. Goal
Implement the detailed Link settings page (`/links/:id`), interactive SVG/PNG QR Code Studio with custom color pickers and center logo branding, and the Categories & Tag Taxonomy page (`/categories`).

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/links/LinkDetailPage.tsx`
- `apps/frontend/src/pages/links/CategoriesPage.tsx`
- `apps/frontend/src/components/qr/QRStudioCanvas.tsx`
- `apps/frontend/src/components/categories/CategoryGrid.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`updateLink`, `createCategory`, `getLinkMetrics`)
- `@flux/zod` (`ZQRCustomization`, `ZCategory`)

## 4. Referenced Architecture & Product Specs
- [docs/core/qr_service.md](file:///home/logan78/Desktop/flux/docs/core/qr_service.md)
- [tasks/01_core/task_107_qr_generator.md](file:///home/logan78/Desktop/flux/tasks/01_core/task_107_qr_generator.md)

## 5. Acceptance Criteria
- [ ] Link Detail allows updating destination URL, title, description, and assigned category with live validation.
- [ ] QR Code Studio provides real-time canvas preview of foreground/background colors and logo uploads with PNG/SVG export buttons.
- [ ] Categories manager provides color-coded card grid, link count badges, and create/edit modal.

## 6. Target Deliverables
- `apps/frontend/src/pages/links/LinkDetailPage.tsx`
- `apps/frontend/src/pages/links/CategoriesPage.tsx`
- `apps/frontend/src/components/qr/QRStudioCanvas.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write unit tests for QR studio color sanitization and category mutations.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement detail page and QR studio.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/links/LinkDetailPage.test.tsx`
