# Subsystem Specification: Deep Linking

> **Source**: Extracted from `prompt/2/05.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 05.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 5 — DEEP LINKING & MOBILE APP INTEGRATION

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 5 — Deep Linking & Mobile App Integration** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Support seamless mobile deep linking for iOS and Android apps, including deferred deep linking for non-installed users.

Key Features:
- iOS Universal Links support (`/.well-known/apple-app-site-association`).
- Android App Links support (`/.well-known/assetlinks.json`).
- Mobile Deep Link Fallback (App Store / Google Play redirection when app is not installed).
- Deferred Deep Link Engine (passing parameters through installation flow).

2. Technical Configuration & Schema

```json
// apple-app-site-association template
{
  "applinks": {
    "apps": [],
    "details": [
      {
        "appID": "TEAMID.com.acme.fluxapp",
        "paths": [ "/*" ]
      }
    ]
  }
}
```

```sql
CREATE TABLE IF NOT EXISTS deep_link_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ios_app_store_id VARCHAR(100),
    ios_bundle_id VARCHAR(100),
    ios_custom_scheme VARCHAR(100),
    android_package_name VARCHAR(100),
    android_sha256_fingerprint TEXT,
    android_custom_scheme VARCHAR(100),
    fallback_url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. API Contracts

- `GET /.well-known/apple-app-site-association` -> Serves dynamic AASA JSON.
- `GET /.well-known/assetlinks.json` -> Serves dynamic Digital Asset Links JSON.

4. Implementation Step-by-Step

Step 1: Build dynamic `/.well-known` endpoints powered by user's deep link configurations.
Step 2: Construct JavaScript client-side intent wrapper HTML page for browser fallback when scheme fails to open.
Step 3: Store deferred deep link token in Redis with 24-hour TTL for SDK retrieval upon first launch.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/custom_domains_ssl.md](file:///home/logan78/Desktop/plan/docs/growth/custom_domains_ssl.md) |
| **Next** | [docs/growth/dynamic_og_metadata.md](file:///home/logan78/Desktop/plan/docs/growth/dynamic_og_metadata.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/plan/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
