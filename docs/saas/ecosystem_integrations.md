# Subsystem Specification: Ecosystem Integrations

> **Source**: Extracted from `prompt/3/06.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux SaaS)**.
> You MUST implement **Part 3 (Flux SaaS) — File 06.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART III — FLUX SAAS
## CHAPTER 6 — THIRD-PARTY ECOSYSTEM INTEGRATIONS (ZAPIER, GA4, SHOPIFY, META/TIKTOK)

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART III — Flux SaaS: Chapter 6 — Third-Party Ecosystem Integrations (Zapier, GA4, Shopify, Meta/TikTok)** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build native integrations with popular SaaS tools: Zapier, Google Analytics 4, Shopify, Meta Pixel, and TikTok Pixel.

Key Features:
- Zapier Integration App (Trigger on link created/clicked, Action to shorten link).
- Google Analytics 4 Measurement Protocol integration (sending offline server-side click events).
- Shopify App embed (auto-shortening product links & tracking conversions).
- Meta & TikTok Pixel server-side conversion API dispatcher.

2. Implementation Step-by-Step

Step 1: Implement GA4 Measurement Protocol HTTP payload formatter (`v=2&tid=G-XXX&cid=...`).
Step 2: Implement Meta Conversions API (`/v18.0/{pixel_id}/events`) sending hashed user metadata (IP, UA).
Step 3: Create Zapier REST hook endpoints for trigger subscriptions.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/saas/notifications.md](file:///home/logan78/Desktop/plan/docs/saas/notifications.md) |
| **Next** | [docs/saas/admin_audit_flags.md](file:///home/logan78/Desktop/plan/docs/saas/admin_audit_flags.md) |
| **Children** | [task_301_tenant_rbac.md](file:///home/logan78/Desktop/plan/tasks/03_saas/task_301_tenant_rbac.md), [task_302_stripe_billing.md](file:///home/logan78/Desktop/plan/tasks/03_saas/task_302_stripe_billing.md), [task_303_public_api.md](file:///home/logan78/Desktop/plan/tasks/03_saas/task_303_public_api.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
