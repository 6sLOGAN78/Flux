# Subsystem Specification: Time Series Analytics

> **Source**: Extracted from `prompt/2/01.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 01.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 1 — ADVANCED ANALYTICS DASHBOARD & TIME-SERIES INFRASTRUCTURE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 1 — Advanced Analytics Dashboard & Time-Series Infrastructure** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build an enterprise-grade analytics engine for link tracking capable of ingesting millions of redirect events and rendering real-time dashboard analytics.

Key Features:
- Real-time time-series click volume charts (hourly, daily, weekly, monthly, custom ranges).
- Live streaming click feed using Server-Sent Events (SSE) / WebSockets.
- Top dimensions breakdown: Top Links, Top Referrers, Top Countries (ISO-3166), Top Browsers, Top Device Types, Top Operating Systems.
- Click heatmaps (time-of-day x day-of-week visualization).
- Scheduled & on-demand report exports (CSV, JSON, PDF formats).

2. Technical Architecture

                       Click Event Ingestion
                               │
                       PostgreSQL / ClickHouse
                               │
               ┌───────────────┴───────────────┐
               ▼                               ▼
       Time-Series Engine             Real-Time SSE Engine
    (ClickHouse / Timescale)            (Redis Pub/Sub)
               │                               │
               └───────────────┬───────────────┘
                               ▼
                        Analytics API
                               │
                     Next.js Dashboard UI

3. Database Schema Design (ClickHouse / PostgreSQL Analytical Schema)

```sql
-- Click Events Fact Table
CREATE TABLE IF NOT EXISTS click_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL,
    domain_id UUID,
    user_id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    country_code VARCHAR(2),
    region VARCHAR(100),
    city VARCHAR(100),
    latitude NUMERIC(9,6),
    longitude NUMERIC(9,6),
    user_agent TEXT,
    browser VARCHAR(50),
    browser_version VARCHAR(20),
    os VARCHAR(50),
    os_version VARCHAR(20),
    device_type VARCHAR(20), -- 'desktop', 'mobile', 'tablet', 'bot'
    referrer_domain VARCHAR(255),
    referrer_url TEXT,
    utm_source VARCHAR(100),
    utm_medium VARCHAR(100),
    utm_campaign VARCHAR(100),
    utm_term VARCHAR(100),
    utm_content VARCHAR(100),
    qr_code_scan BOOLEAN DEFAULT FALSE,
    response_time_ms INTEGER
);

CREATE INDEX idx_click_events_link_time ON click_events (link_id, timestamp DESC);
CREATE INDEX idx_click_events_user_time ON click_events (user_id, timestamp DESC);
CREATE INDEX idx_click_events_country ON click_events (country_code);
```

4. REST API Endpoint Contracts

- `GET /api/v1/analytics/time-series?link_id=:id&from=:start&to=:end&granularity=hour|day|month`
  Response:
  ```json
  {
    "status": "success",
    "data": [
      { "timestamp": "2026-08-06T00:00:00Z", "clicks": 1420, "unique_visitors": 1150 },
      { "timestamp": "2026-08-06T01:00:00Z", "clicks": 1890, "unique_visitors": 1430 }
    ]
  }
  ```

- `GET /api/v1/analytics/breakdown?link_id=:id&type=referrers|countries|browsers|devices&limit=10`
  Response:
  ```json
  {
    "status": "success",
    "dimension": "countries",
    "data": [
      { "code": "US", "name": "United States", "count": 4520, "percentage": 42.5 },
      { "code": "GB", "name": "United Kingdom", "count": 1820, "percentage": 17.1 }
    ]
  }
  ```

- `POST /api/v1/analytics/export`
  Request: `{ "link_ids": ["uuid..."], "format": "pdf", "range": "30d" }`
  Response: `{ "job_id": "job_12345", "status": "queued", "download_url": "/api/v1/analytics/exports/job_12345.pdf" }`

5. Implementation Step-by-Step

Step 1: Implement analytical aggregation queries with Redis caching layer (5-minute TTL for historical ranges).
Step 2: Build SSE (Server-Sent Events) endpoint `/api/v1/analytics/live` connected to Redis Pub/Sub topic `analytics:clicks`.
Step 3: Implement export worker using PDFKit / Puppeteer for PDF reports and stream CSV builder for data downloads.
Step 4: Integrate click heatmap aggregation query grouping by `EXTRACT(DOW FROM timestamp)` and `EXTRACT(HOUR FROM timestamp)`.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/core/qr_service.md](file:///home/logan78/Desktop/plan/docs/core/qr_service.md) |
| **Next** | [docs/growth/campaign_utm_builder.md](file:///home/logan78/Desktop/plan/docs/growth/campaign_utm_builder.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
