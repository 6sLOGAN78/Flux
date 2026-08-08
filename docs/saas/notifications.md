# Subsystem Specification: Notifications

> **Source**: Extracted from `prompt/3/05.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux SaaS)**.
> You MUST implement **Part 3 (Flux SaaS) — File 05.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART III — FLUX SAAS
## CHAPTER 5 — MULTI-CHANNEL NOTIFICATION ENGINE (EMAIL, SLACK, DISCORD, IN-APP)

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART III — Flux SaaS: Chapter 5 — Multi-Channel Notification Engine (Email, Slack, Discord, In-App)** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build a multi-channel notification dispatcher supporting Email, Slack Webhooks, Discord Webhooks, and WebSockets in-app notifications.

Key Features:
- Transactional & Marketing Email Engine (Resend / SendGrid / Postmark).
- Slack App Webhook integration (rich Block Kit messages).
- Discord Webhook integration (embed cards).
- In-App Notification Center backed by WebSockets / SSE.

2. Database Schema & API Contract

```sql
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'info', -- 'info', 'warning', 'alert'
    link_url TEXT,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

- `GET /api/v1/notifications` -> Returns list of unread user notifications.
- `POST /api/v1/notifications/mark-read` -> Marks notifications as read.

3. Implementation Step-by-Step

Step 1: Abstract Notification Delivery Provider interface (`SendEmail`, `SendSlack`, `SendDiscord`, `SendInApp`).
Step 2: Construct Slack Block Kit template builder for link threshold alerts.
Step 3: Connect WebSocket broker (`/ws/notifications`) sending instant alerts to active user dashboard sessions.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/saas/webhooks_event_bus.md](file:///home/logan78/Desktop/flux/docs/saas/webhooks_event_bus.md) |
| **Next** | [docs/saas/ecosystem_integrations.md](file:///home/logan78/Desktop/flux/docs/saas/ecosystem_integrations.md) |
| **Children** | [task_301_tenant_rbac.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_301_tenant_rbac.md), [task_302_stripe_billing.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_302_stripe_billing.md), [task_303_public_api.md](file:///home/logan78/Desktop/flux/tasks/03_saas/task_303_public_api.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
