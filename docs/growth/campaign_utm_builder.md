# Subsystem Specification: Campaign Utm Builder

> **Source**: Extracted from `prompt/2/02.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 02.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 2 — CAMPAIGN MANAGEMENT & UTM BUILDER ENGINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 2 — Campaign Management & UTM Builder Engine** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build an enterprise marketing campaign management module that enables marketers to tag, organize, template, and track performance of short URLs across multi-channel campaigns.

Key Features:
- UTM Builder with auto-completion and validation for `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, and `utm_content`.
- Campaign Templates & Presets (e.g. "Summer Sale 2026", "Newsletter Weekly", "Product Hunt Launch").
- Campaign Grouping & Multi-Link Association.
- Campaign Performance Comparisons & ROI tracking.

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL,
    description TEXT,
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(100),
    utm_term VARCHAR(100),
    utm_content VARCHAR(100),
    budget NUMERIC(12,2) DEFAULT 0.00,
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'archived', 'completed'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS campaign_links (
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (campaign_id, link_id)
);

CREATE TABLE IF NOT EXISTS utm_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. API Contracts

- `POST /api/v1/campaigns`
  Request: `{ "name": "Q3 Growth Campaign", "utm_source": "twitter", "utm_medium": "cpc", "utm_campaign": "q3_launch" }`
  Response: `{ "id": "uuid", "name": "Q3 Growth Campaign", "status": "active" }`

- `POST /api/v1/campaigns/:id/links`
  Request: `{ "destination_url": "https://example.com/pricing", "custom_slug": "q3-pricing" }`
  Response: `{ "short_url": "https://flux.dev/q3-pricing", "full_url": "https://example.com/pricing?utm_source=twitter&utm_medium=cpc&utm_campaign=q3_launch" }`

- `GET /api/v1/campaigns/:id/stats`
  Response: `{ "campaign_id": "uuid", "total_links": 12, "total_clicks": 45890, "top_link": "q3-pricing" }`

4. Implementation Step-by-Step

Step 1: Build URL parsing and query parameter injection engine preserving existing destination parameters.
Step 2: Create CRUD REST endpoints for Campaigns, Campaign Links, and UTM Templates.
Step 3: Implement aggregated campaign analytics endpoint summarizing performance across all campaign links.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/time_series_analytics.md](file:///home/logan78/Desktop/flux/docs/growth/time_series_analytics.md) |
| **Next** | [docs/growth/custom_domains_ssl.md](file:///home/logan78/Desktop/flux/docs/growth/custom_domains_ssl.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
