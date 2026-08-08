# Subsystem Specification: Ai Engine

> **Source**: Extracted from `prompt/4/04.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 04.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 4 — ENTERPRISE AI ENGINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 4 — Enterprise AI Engine** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Integrate Enterprise AI Features powered by Large Language Model (LLM) APIs (Gemini / OpenAI API).

Key Features:
- AI Custom Slug Generator (Generating catchy, relevant short slugs from destination page content).
- AI Campaign Performance Summaries (Generating plain English executive briefs).
- AI Anomaly Detection (Detecting sudden spikes or drops in traffic and identifying potential spam/bot surges).

2. Architecture & API Contract

```sql
CREATE TABLE IF NOT EXISTS ai_anomaly_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    anomaly_type VARCHAR(50) NOT NULL, -- 'traffic_spike', 'traffic_drop', 'bot_surge'
    confidence_score NUMERIC(4,3),
    summary TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

- `POST /api/v1/ai/generate-slugs`
  Request: `{ "url": "https://acme.com/products/quantum-processor-v2" }`
  Response: `{ "suggestions": ["quantum-v2", "next-gen-cpu", "quantum-proc"] }`

- `POST /api/v1/ai/summarize-campaign`
  Request: `{ "campaign_id": "uuid..." }`
  Response: `{ "summary": "Campaign 'Q3 Launch' performed 45% above average driven by Twitter mobile traffic." }`

3. Implementation Step-by-Step

Step 1: Build LLM Service Wrapper with structured JSON response parsing and fallback retry logic.
Step 2: Implement web scraper extracting title/meta-description from destination URL for slug generation.
Step 3: Build statistical anomaly detector evaluating rolling z-score of hourly click volumes (`z = (x - μ) / σ`).


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/enterprise/revenue_analytics.md](file:///home/logan78/Desktop/flux/docs/enterprise/revenue_analytics.md) |
| **Next** | [docs/enterprise/saml_scim_sso.md](file:///home/logan78/Desktop/flux/docs/enterprise/saml_scim_sso.md) |
| **Children** | [task_401_attribution_engine.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_401_attribution_engine.md), [task_402_funnel_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_402_funnel_analytics.md), [task_403_revenue_analytics.md](file:///home/logan78/Desktop/flux/tasks/04_enterprise/task_403_revenue_analytics.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
