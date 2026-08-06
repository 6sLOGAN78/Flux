# Subsystem Specification: Async Queue

> **Source**: Extracted from `prompt/2/09.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 09.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 9 — ASYNC BACKGROUND JOB PROCESSING & QUEUE INFRASTRUCTURE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 9 — Async Background Job Processing & Queue Infrastructure** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Decouple intensive or non-blocking tasks from HTTP request/response loops using a distributed background job queue with retry policies, rate limits, and scheduling.

Job Types:
- Asynchronous Analytics Aggregation Rollups (hourly/daily table updates).
- Email notifications (welcome, password resets, report exports).
- QR code batch generation.
- Custom domain SSL renewal checks.
- Dead link checking (periodically checking for 404 destination URLs).

2. Queue Architecture & Job Schema

                     HTTP API / Redirect Engine
                                │
                       Push Job to Redis Queue
                       (Asynq / BullMQ Queue)
                                │
                 ┌──────────────┴──────────────┐
                 ▼                             ▼
          Worker Process 1              Worker Process 2
       (Analytics Aggregation)       (Email / SSL Worker)
                 │                             │
                 └──────────────┬──────────────┘
                                ▼
                       PostgreSQL / S3

3. Job Definition Structure

```json
{
  "job_id": "job_99812",
  "queue": "emails",
  "task": "send_campaign_report",
  "payload": { "user_id": "uuid...", "campaign_id": "uuid...", "format": "pdf" },
  "max_retries": 5,
  "retry_delay_seconds": 30,
  "created_at": "2026-08-06T04:00:00Z"
}
```

4. Implementation Step-by-Step

Step 1: Set up Redis-backed distributed queue framework (Asynq in Go or BullMQ in Node.js).
Step 2: Build worker concurrency manager handling job registration, graceful shutdown signals (SIGTERM), and dead-letter queues.
Step 3: Implement exponential backoff retry handler (`delay = 2^retry_count * base_seconds`).


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/ab_testing.md](file:///home/logan78/Desktop/plan/docs/growth/ab_testing.md) |
| **Next** | [docs/saas/multi_tenant_rbac.md](file:///home/logan78/Desktop/plan/docs/saas/multi_tenant_rbac.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
