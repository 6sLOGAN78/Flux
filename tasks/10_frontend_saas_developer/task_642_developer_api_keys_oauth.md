---
id: TASK-642
title: Scoped API Key Generator & OAuth 2.0 Client Apps Manager
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_604_ui_component_primitives.md
references:
  - api/openapi_v3_saas.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - api-keys
  - oauth
  - developer
  - credentials
---

# TASK-642: Scoped API Key Generator & OAuth 2.0 Client Apps Manager

## 1. Goal
Build the Developer API Keys & OAuth 2.0 page (`/developer/api-keys`) supporting granular scope selection (`links:read`, `links:write`, `analytics:read`), secure one-time token reveal dialogs, and OAuth 2.0 client application registration.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/developer/ApiKeysPage.tsx`
- `apps/frontend/src/components/developer/ApiKeyTable.tsx`
- `apps/frontend/src/components/developer/CreateApiKeyModal.tsx`
- `apps/frontend/src/components/developer/OAuthClientsCard.tsx`

## 3. Dependencies & Prerequisites
- `@flux/zod` (`ZAPIKey`, `ZOAuthTokenResponse`)

## 4. Referenced Architecture & Product Specs
- [docs/saas/public_api_oauth.md](file:///home/logan78/Desktop/flux/docs/saas/public_api_oauth.md)
- [tasks/03_saas/task_303_public_api.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_303_public_api.md)

## 5. Acceptance Criteria
- [x] Creating an API key displays the full raw secret token in a modal with warning: "Save this key now; it will not be shown again."
- [x] API Key list masks secret and shows `tokenPrefix` (e.g. `flx_live_a1b2...`) and active scopes.
- [x] OAuth Clients card allows generating `client_id` and rotating `client_secret`.

## 6. Target Deliverables
- `apps/frontend/src/pages/developer/ApiKeysPage.tsx`
- `apps/frontend/src/components/developer/CreateApiKeyModal.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write API key masking and clipboard unit tests.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement API Keys page.
- [x] Step 4: Confirm tests pass.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/developer/ApiKeysPage.test.tsx`
