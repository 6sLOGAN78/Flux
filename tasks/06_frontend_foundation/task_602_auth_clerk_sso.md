---
id: TASK-602
title: Clerk Authentication Provider, Protected Routes & SSO Handler
layer: Level 5 (Executable Task Unit)
status: Completed
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_601_api_client_query.md
references:
  - ai/SECURITY.md
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - auth
  - clerk
  - rbac
  - sso
---

# TASK-602: Clerk Authentication Provider, Protected Routes & SSO Handler

## 1. Goal
Integrate `@clerk/clerk-react` Provider, build `<ProtectedRoute>` and `<PublicRoute>` guards in React Router v7, and implement enterprise SSO redirect handler.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/components/auth/ProtectedRoute.tsx`
- `apps/frontend/src/components/auth/PublicRoute.tsx`
- `apps/frontend/src/pages/auth/SignInPage.tsx`
- `apps/frontend/src/pages/auth/SignUpPage.tsx`
- `apps/frontend/src/pages/auth/SSOPage.tsx`

## 3. Dependencies & Prerequisites
- `@clerk/clerk-react` v5
- `react-router-dom` v7

## 4. Referenced Architecture & Product Specs
- Master Agent Operating Protocol: [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md)
- [docs/saas/saml_scim_sso.md](file:///home/logan78/Desktop/flux/docs/enterprise/saml_scim_sso.md)

## 5. Acceptance Criteria
- [x] Unauthenticated requests to protected routes redirect cleanly to `/sign-in` with return URL state.
- [x] Authenticated users visiting `/sign-in` or `/sign-up` redirect to `/dashboard`.
- [x] Enterprise SSO page `/auth/sso` captures company domain and triggers SAML 2.0 IdP initiation.

## 6. Target Deliverables
- `apps/frontend/src/components/auth/ProtectedRoute.tsx`
- `apps/frontend/src/pages/auth/SignInPage.tsx`
- `apps/frontend/src/pages/auth/SignUpPage.tsx`
- `apps/frontend/src/pages/auth/SSOPage.tsx`

## 7. Definition of Done (DoD)
- [x] Step 1: Write routing unit tests simulating authenticated and unauthenticated states.
- [x] Step 2: Confirm test failure.
- [x] Step 3: Implement auth wrappers and pages.
- [x] Step 4: Confirm tests pass.
- [x] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/components/auth/ProtectedRoute.test.tsx`
