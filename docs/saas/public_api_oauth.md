# Subsystem Specification: Public Api Oauth

> **Source**: Extracted from `prompt/3/03.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux SaaS)**.
> You MUST implement **Part 3 (Flux SaaS) — File 03.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART III — FLUX SAAS
## CHAPTER 3 — PUBLIC REST API PLATFORM, OAUTH 2.0 & DEVELOPER SDKS

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART III — Flux SaaS: Chapter 3 — Public REST API Platform, OAuth 2.0 & Developer SDKs** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Expose a public developer REST API platform backed by API Key authentication, OAuth 2.0 Provider flow, token bucket rate limiting, and official Client SDKs in Go, TypeScript, Python, and CLI.

Key Features:
- API Key creation with scoped permissions (`links:read`, `links:write`, `analytics:read`).
- OAuth 2.0 Authorization Server (`/oauth/authorize`, `/oauth/token`).
- Token Bucket Rate Limiting per API key (e.g. 100 req/min for Free, 5,000 req/min for Pro).
- Auto-generated OpenAPI v3 (Swagger) docs.

2. API Key Database Schema

```sql
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(10) NOT NULL, -- e.g. "flx_live_"
    key_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash of secret
    scopes TEXT[] DEFAULT '{"links:read","links:write"}',
    rate_limit_per_min INTEGER DEFAULT 100,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. Implementation Step-by-Step

Step 1: Build API Key Authenticator middleware hashing incoming headers `Authorization: Bearer flx_live_secretkey` and checking Redis cache.
Step 2: Implement Redis sliding window rate limiter (`INCRBY` + `EXPIRE`).
Step 3: Generate OpenAPI / Swagger documentation definitions.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/saas/billing_stripe.md](file:///home/logan78/Desktop/flux/docs/saas/billing_stripe.md) |
| **Next** | [docs/saas/webhooks_event_bus.md](file:///home/logan78/Desktop/flux/docs/saas/webhooks_event_bus.md) |
| **Children** | [task_301_tenant_rbac.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_301_tenant_rbac.md), [task_302_stripe_billing.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_302_stripe_billing.md), [task_303_public_api.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_303_public_api.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
