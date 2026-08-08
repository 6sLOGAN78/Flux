# Subsystem Specification: Custom Domains Ssl

> **Source**: Extracted from `prompt/2/03.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 03.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 3 — CUSTOM DOMAINS, DNS VERIFICATION & AUTOMATED SSL PROVISIONING

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 3 — Custom Domains, DNS Verification & Automated SSL Provisioning** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Allow users and organizations to use custom branded domains (e.g. `link.acme.com` or `ac.me`) instead of the default `flux.dev` domain.

Key Features:
- Custom domain registration & DNS verification (`CNAME` / `A` record validation).
- Automated TLS/SSL Certificate Provisioning via Let's Encrypt / ZeroSSL ACME challenge using Caddy Proxy / ACME client.
- Wildcard domain support & Custom Root/404 fallback handling.
- Background Domain Health Checker (periodic DNS, SSL expiry, and uptime checks).

2. Technical Architecture

     User Adds Domain "link.acme.com"
                     │
         DNS CNAME Verification
                     │
      Caddy / ACME Gateway Dynamic SSL
                     │
      Route `link.acme.com/:slug` -> Flux Core

3. Database Schema

```sql
CREATE TABLE IF NOT EXISTS custom_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain VARCHAR(255) UNIQUE NOT NULL,
    verification_token VARCHAR(64) NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    ssl_status VARCHAR(30) DEFAULT 'pending', -- 'pending', 'active', 'failed', 'expired'
    ssl_expires_at TIMESTAMPTZ,
    custom_root_redirect TEXT,
    custom_404_redirect TEXT,
    is_wildcard BOOLEAN DEFAULT FALSE,
    last_health_check TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_custom_domains_domain ON custom_domains(domain);
```

4. API Contracts

- `POST /api/v1/domains`
  Request: `{ "domain": "link.acme.com", "custom_root_redirect": "https://acme.com" }`
  Response: `{ "id": "uuid", "domain": "link.acme.com", "cname_target": "cname.flux.dev", "verification_token": "flux-verify=abc123xyz" }`

- `POST /api/v1/domains/:id/verify`
  Response: `{ "status": "verified", "ssl_status": "provisioning" }`

5. Implementation Step-by-Step

Step 1: Write DNS lookup verifier in Go/Node.js using `net.LookupCNAME` and `net.LookupTXT`.
Step 2: Configure Caddy Server dynamic SSL endpoint (`on_demand_tls` asking `/api/v1/internal/check-domain`).
Step 3: Build cron health-check worker checking SSL expiration (<14 days remaining triggers auto-renewal).


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/growth/campaign_utm_builder.md](file:///home/logan78/Desktop/flux/docs/growth/campaign_utm_builder.md) |
| **Next** | [docs/growth/deep_linking.md](file:///home/logan78/Desktop/flux/docs/growth/deep_linking.md) |
| **Children** | [task_201_clickhouse_pipeline.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_201_clickhouse_pipeline.md), [task_202_utm_builder.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_202_utm_builder.md), [task_203_custom_domains.md](file:///home/logan78/Desktop/flux/tasks/02_growth/task_203_custom_domains.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
