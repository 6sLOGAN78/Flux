# Subsystem Specification: Multi Tenant Rbac

> **Source**: Extracted from `prompt/3/01.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux SaaS)**.
> You MUST implement **Part 3 (Flux SaaS) — File 01.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART III — FLUX SAAS
## CHAPTER 1 — MULTI-TENANT ARCHITECTURE, ORGANIZATIONS, WORKSPACES & RBAC

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART III — Flux SaaS: Chapter 1 — Multi-Tenant Architecture, Organizations, Workspaces & RBAC** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Transform Flux into a multi-tenant SaaS architecture supporting Organizations, Workspaces, Teams, Member Invitations, and granular Role-Based Access Control (RBAC).

Roles & Permissions:
- Owner: Full access, billing management, workspace deletion.
- Admin: Manage team members, custom domains, security settings.
- Editor: Create, edit, delete, and manage short links and campaigns.
- Viewer: Read-only access to dashboard and analytics.
- Custom Permissions: Fine-grained toggle per workspace.

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) UNIQUE NOT NULL,
    billing_email VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, slug)
);

CREATE TABLE IF NOT EXISTS workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(30) NOT NULL DEFAULT 'editor', -- 'owner', 'admin', 'editor', 'viewer'
    custom_permissions JSONB DEFAULT '{}',
    invited_by UUID REFERENCES users(id),
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(30) NOT NULL DEFAULT 'editor',
    token VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. API Contracts

- `POST /api/v1/workspaces/:id/invitations`
  Request: `{ "email": "engineer@acme.com", "role": "admin" }`
  Response: `{ "invitation_id": "uuid", "invite_url": "https://flux.dev/accept-invite?token=abc..." }`

- `GET /api/v1/workspaces/:id/members`
  Response:
  ```json
  [
    { "user_id": "uuid1", "email": "alice@acme.com", "role": "owner" },
    { "user_id": "uuid2", "email": "bob@acme.com", "role": "editor" }
  ]
  ```

4. Implementation Step-by-Step

Step 1: Write RBAC Middleware evaluating user workspace member permissions on every API endpoint.
Step 2: Build Workspace Context Switcher (`X-Workspace-ID` header or URL scope).
Step 3: Build transactional invitation email workflow with token verification.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/async_queue.md](file:///home/logan78/Desktop/flux/docs/growth/async_queue.md) |
| **Next** | [docs/saas/billing_stripe.md](file:///home/logan78/Desktop/flux/docs/saas/billing_stripe.md) |
| **Children** | [task_301_tenant_rbac.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_301_tenant_rbac.md), [task_302_stripe_billing.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_302_stripe_billing.md), [task_303_public_api.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_303_public_api.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
