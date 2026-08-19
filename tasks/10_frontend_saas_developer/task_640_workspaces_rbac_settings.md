---
id: TASK-640
title: Organization Workspaces & Team RBAC Management
layer: Level 5 (Executable Task Unit)
status: Ready
owner: Frontend Engineer
depends_on:
  - tasks/06_frontend_foundation/task_603_app_shell_navigation.md
references:
  - api/openapi_v3_saas.yaml
agent_mode: TDD-Execution
token_budget_est: ~1.8KB
tags:
  - workspaces
  - rbac
  - team
  - permissions
---

# TASK-640: Organization Workspaces & Team RBAC Management

## 1. Goal
Implement the Workspaces & Team Settings page (`/settings/workspaces`) with member list, role indicators (`Owner`, `Admin`, `Member`, `Viewer`), email invitation modal, and workspace creation.

## 2. Scope & Target Boundaries
Target source files:
- `apps/frontend/src/pages/settings/WorkspacesPage.tsx`
- `apps/frontend/src/components/settings/TeamMemberTable.tsx`
- `apps/frontend/src/components/settings/InviteMemberModal.tsx`
- `apps/frontend/src/components/settings/CreateWorkspaceModal.tsx`

## 3. Dependencies & Prerequisites
- `@flux/openapi` (`getWorkspaces`, `ZWorkspace`, `ZWorkspaceMember`)

## 4. Referenced Architecture & Product Specs
- [docs/saas/multi_tenant_rbac.md](file:///home/logan78/Desktop/flux/docs/saas/multi_tenant_rbac.md)
- [tasks/03_saas/task_301_tenant_rbac.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_301_tenant_rbac.md)

## 5. Acceptance Criteria
- [ ] Lists active team members with role badges and avatar chips.
- [ ] Invite Member modal allows selecting role level and entering recipient emails.
- [ ] Non-owner users cannot invite owners or alter workspace billing settings.

## 6. Target Deliverables
- `apps/frontend/src/pages/settings/WorkspacesPage.tsx`
- `apps/frontend/src/components/settings/TeamMemberTable.tsx`

## 7. Definition of Done (DoD)
- [ ] Step 1: Write member role permission unit tests.
- [ ] Step 2: Confirm test failure.
- [ ] Step 3: Implement Workspaces page.
- [ ] Step 4: Confirm tests pass.
- [ ] Step 5: Verify zero linter warnings.

## 8. Testing Strategy
- **Verification Command**: `cd apps/frontend && bun test src/pages/settings/WorkspacesPage.test.tsx`
