# Subsystem Specification: Attribution Engine

> **Source**: Extracted from `prompt/4/01.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 01.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 1 — MULTI-TOUCH ATTRIBUTION ENGINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 1 — Multi-Touch Attribution Engine** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build an enterprise attribution engine capable of evaluating customer touchpoints across multiple attribution models.

Supported Attribution Models:
- First-Touch Attribution: 100% credit assigned to first link clicked.
- Last-Touch Attribution: 100% credit assigned to final link clicked before conversion.
- Linear Attribution: Equal credit distributed across all touchpoints in session history.
- Time-Decay Attribution: Exponentially higher weight given to recent touchpoints.
- Position-Based (U-Shaped) Attribution: 40% to first, 40% to last, 20% split among middle touchpoints.

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS visitor_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visitor_token VARCHAR(64) NOT NULL, -- Cookie / Fingerprint hash
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS touchpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visitor_session_id UUID NOT NULL REFERENCES visitor_sessions(id) ON DELETE CASCADE,
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    campaign_id UUID REFERENCES campaigns(id),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    referrer_domain VARCHAR(255),
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(100)
);

CREATE INDEX idx_touchpoints_session ON touchpoints(visitor_session_id, timestamp ASC);
```

3. API Contracts

- `GET /api/v1/enterprise/attribution?model=multi-touch|first-touch|last-touch&from=:start&to=:end`
  Response:
  ```json
  {
    "model": "position_based",
    "total_conversions": 142,
    "total_attributed_revenue": 45200.00,
    "campaigns": [
      { "campaign_name": "Summer Launch", "attributed_conversions": 56.8, "attributed_revenue": 18080.00 }
    ]
  }
  ```

4. Implementation Step-by-Step

Step 1: Write client-side tracking script (`flux-pixel.js`) maintaining visitor session token in `localStorage`/cookie.
Step 2: Build Touchpoint Collector Endpoint storing journey history in `touchpoints` table.
Step 3: Build Attribution Math Calculator Engine implementing all 5 attribution algorithms.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/saas/admin_audit_flags.md](file:///home/logan78/Desktop/flux/docs/saas/admin_audit_flags.md) |
| **Next** | [docs/enterprise/revenue_analytics.md](file:///home/logan78/Desktop/flux/docs/enterprise/revenue_analytics.md) |
| **Children** | [task_401_attribution_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_401_attribution_engine.md), [task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md), [task_403_revenue_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_403_revenue_analytics.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
