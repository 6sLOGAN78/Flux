# Subsystem Specification: Billing Stripe

> **Source**: Extracted from `prompt/3/02.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux SaaS)**.
> You MUST implement **Part 3 (Flux SaaS) — File 02.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART III — FLUX SAAS
## CHAPTER 2 — SUBSCRIPTIONS, STRIPE MONETIZATION & USAGE LIMITS

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART III — Flux SaaS: Chapter 2 — Subscriptions, Stripe Monetization & Usage Limits** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Integrate Stripe for recurring SaaS subscriptions, tier usage limits, trials, discount coupons, and automated invoice management.

Subscription Tiers:
- Free: 1,000 links/mo, 10,000 clicks/mo, 1 workspace.
- Pro ($29/mo): 50,000 links/mo, 500,000 clicks/mo, custom domains, 5 seats.
- Business ($99/mo): 500,000 links/mo, 5,000,000 clicks/mo, unlimited seats, SLA.

2. Architecture & Webhook Handling

                   Stripe Checkout / Customer Portal
                                  │
                          Stripe Webhooks
                                  │
                 `/api/v1/webhooks/stripe` Endpoint
                                  │
                  Verify Signature (HMAC-SHA256)
                                  │
                 Update Organization Subscription State

3. Database Schema

```sql
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(100) UNIQUE NOT NULL,
    stripe_subscription_id VARCHAR(100) UNIQUE,
    plan_tier VARCHAR(50) NOT NULL DEFAULT 'free', -- 'free', 'pro', 'business'
    status VARCHAR(30) NOT NULL DEFAULT 'active', -- 'active', 'past_due', 'canceled', 'trialing'
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

4. API Contracts

- `POST /api/v1/billing/checkout`
  Request: `{ "plan": "pro", "interval": "monthly" }`
  Response: `{ "checkout_url": "https://checkout.stripe.com/c/pay/cs_live_..." }`

- `POST /api/v1/billing/portal`
  Response: `{ "portal_url": "https://billing.stripe.com/p/session/test_..." }`

5. Implementation Step-by-Step

Step 1: Implement Stripe Billing SDK wrapper handling customer creation, checkout sessions, and customer portal URLs.
Step 2: Build robust Stripe Webhook Handler with idempotency processing (`evt_...` event IDs stored in Redis).
Step 3: Implement Usage Metering Middleware blocking link creation when tier limits are exceeded.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/saas/multi_tenant_rbac.md](file:///home/logan78/Desktop/flux/docs/saas/multi_tenant_rbac.md) |
| **Next** | [docs/saas/public_api_oauth.md](file:///home/logan78/Desktop/flux/docs/saas/public_api_oauth.md) |
| **Children** | [task_301_tenant_rbac.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_301_tenant_rbac.md), [task_302_stripe_billing.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_302_stripe_billing.md), [task_303_public_api.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_303_public_api.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
