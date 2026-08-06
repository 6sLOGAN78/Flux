# Subsystem Specification: Revenue Analytics

> **Source**: Extracted from `prompt/4/03.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 03.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 3 — REVENUE ANALYTICS & FINANCIAL MARKETING METRICS

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 3 — Revenue Analytics & Financial Marketing Metrics** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Calculate high-level financial marketing metrics: Customer Acquisition Cost (CAC), Lifetime Value (LTV), Return on Investment (ROI), and generate executive financial reports.

Key Features:
- Ad Spend Data Ingestion (Manual CSV import & API integrations).
- Automated CAC Calculation (`Total Spend / Total Customers Acquired`).
- ROI & ROAS metrics per campaign.

2. Database Schema & API Contract

```sql
CREATE TABLE IF NOT EXISTS ad_spend (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    amount_spent NUMERIC(12,2) NOT NULL,
    platform VARCHAR(50) NOT NULL, -- 'google', 'facebook', 'linkedin'
    UNIQUE(campaign_id, date, platform)
);
```

- `GET /api/v1/enterprise/metrics/roi?campaign_id=:id`
  Response: `{ "spend": 5000.00, "revenue": 18500.00, "roas": 3.7, "cac": 45.20 }`

3. Implementation Step-by-Step

Step 1: Implement Ad Spend ingestion endpoints & CSV uploader.
Step 2: Build Financial Math Engine aggregating spend vs conversion revenue.
Step 3: Build automated weekly PDF summary report generation job for marketing executives.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/enterprise/attribution_engine.md](file:///home/logan78/Desktop/plan/docs/enterprise/attribution_engine.md) |
| **Next** | [docs/enterprise/ai_engine.md](file:///home/logan78/Desktop/plan/docs/enterprise/ai_engine.md) |
| **Children** | [task_401_attribution_engine.md](file:///home/logan78/Desktop/plan/tasks/04_enterprise/task_401_attribution_engine.md), [task_402_funnel_analytics.md](file:///home/logan78/Desktop/plan/tasks/04_enterprise/task_402_funnel_analytics.md), [task_403_revenue_analytics.md](file:///home/logan78/Desktop/plan/tasks/04_enterprise/task_403_revenue_analytics.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
