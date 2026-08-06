# Subsystem Specification: Dynamic Og Metadata

> **Source**: Extracted from `prompt/2/06.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 06.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 6 — DYNAMIC OPEN GRAPH METADATA & SOCIAL PREVIEWS

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 6 — Dynamic Open Graph Metadata & Social Previews** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Allow custom Open Graph metadata (title, description, social image, Twitter Card) for any short link, enabling rich previews across Twitter/X, iMessage, LinkedIn, Facebook, and Slack.

Key Features:
- Custom OG title, description, and thumbnail image override.
- Dynamic image generator service for automated social cards.
- Crawler/Scraper proxy handling (serving HTML with `<meta og:...>` tags to bots, while immediately redirecting real human visitors).

2. Bot Detection & Proxy Flow

            Visitor GET /:slug
                   │
         Is Social Media Bot?
       (Twitterbot, facebookexternalhit, Slackbot, etc.)
         ┌─────────┴─────────┐
         ▼                   ▼
       [YES]               [NO]
  Serve HTML with     HTTP 302 Redirect
  OG Meta Tags        to Destination
  (no redirect)

3. Database Schema & API Contract

```sql
CREATE TABLE IF NOT EXISTS link_og_meta (
    link_id UUID PRIMARY KEY REFERENCES links(id) ON DELETE CASCADE,
    title VARCHAR(255),
    description TEXT,
    image_url TEXT,
    twitter_card_type VARCHAR(50) DEFAULT 'summary_large_image',
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

- `POST /api/v1/links/:id/og`
  Request: `{ "title": "Exclusive Sale", "description": "Get 50% off today!", "image_url": "https://cdn.acme.com/banner.png" }`
  Response: `{ "status": "updated", "preview_url": "https://flux.dev/p/preview/:id" }`

4. Implementation Step-by-Step

Step 1: Implement bot detection middleware checking User-Agent regex (`bot|facebook|twitter|slack|linkedin|whatsapp|telegram`).
Step 2: Generate clean HTML markup response with OG tags for bots.
Step 3: Integrate headless thumbnail card generator service (Puppeteer / Canvas) rendering custom social cards on demand.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/deep_linking.md](file:///home/logan78/Desktop/plan/docs/growth/deep_linking.md) |
| **Next** | [docs/growth/ab_testing.md](file:///home/logan78/Desktop/plan/docs/growth/ab_testing.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
