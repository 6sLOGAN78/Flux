# Subsystem Specification: Ab Testing

> **Source**: Extracted from `prompt/2/08.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 08.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 8 — A/B TESTING & DYNAMIC TRAFFIC SPLITTING

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 8 — A/B Testing & Dynamic Traffic Splitting** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Enable marketers to perform A/B split testing across multiple destination URLs from a single short link, tracking conversion efficiency and auto-selecting winners.

Key Features:
- Multi-variant destination support with weighted percentage distribution (e.g. Variant A: 50%, Variant B: 30%, Variant C: 20%).
- Automated conversion winner selection based on sample size and statistical confidence.
- Real-time side-by-side performance analytics (Click-Through Rate, Conversions, Bounce Rate).

2. Traffic Splitter Engine & Database Schema

```sql
CREATE TABLE IF NOT EXISTS ab_test_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL, -- e.g. "Landing Page V2"
    target_url TEXT NOT NULL,
    weight_percentage INTEGER NOT NULL CHECK (weight_percentage >= 0 AND weight_percentage <= 100),
    clicks_count BIGINT DEFAULT 0,
    conversions_count BIGINT DEFAULT 0,
    is_winner BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ab_variants_link ON ab_test_variants(link_id);
```

3. API Contracts

- `POST /api/v1/links/:id/ab-test`
  Request:
  ```json
  {
    "variants": [
      { "name": "Control Page A", "target_url": "https://acme.com/v1", "weight_percentage": 50 },
      { "name": "Challenger Page B", "target_url": "https://acme.com/v2", "weight_percentage": 50 }
    ]
  }
  ```
  Response: `{ "status": "active", "total_variants": 2 }`

4. Implementation Step-by-Step

Step 1: Implement weighted random selection algorithm in memory inside Redirect Engine (`random(1..100)` mapped to ranges).
Step 2: Store visitor assigned variant in cookie (`flux_ab_:link_id`) to ensure sticky consistent user experience.
Step 3: Build conversion reporting endpoint incrementing `conversions_count` for matching variant.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/dynamic_og_metadata.md](file:///home/logan78/Desktop/flux/docs/growth/dynamic_og_metadata.md) |
| **Next** | [docs/growth/async_queue.md](file:///home/logan78/Desktop/flux/docs/growth/async_queue.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
