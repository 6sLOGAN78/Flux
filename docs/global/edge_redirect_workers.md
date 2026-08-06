# Subsystem Specification: Edge Redirect Workers

> **Source**: Extracted from `prompt/5/01.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Global Scale)**.
> You MUST implement **Part 5 (Flux Global Scale) — File 01.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART V — FLUX GLOBAL SCALE
## CHAPTER 1 — MULTI-REGION EDGE REDIRECT ENGINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART V — Flux Global Scale: Chapter 1 — Multi-Region Edge Redirect Engine** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Build a sub-10ms global edge redirect engine deployed across 250+ Edge Locations worldwide (Cloudflare Workers, Fly.io Edge, AWS CloudFront Functions).

Key Requirements:
- Global Edge Execution (evaluating redirects at edge closest to visitor).
- Edge KV / Memory Storage for instant link destination lookup.
- Asynchronous click event streaming back to core ingestion pipeline.

2. Architecture Diagram

                         Global End Users
                                │
          ┌─────────────────────┼─────────────────────┐
          ▼                     ▼                     ▼
   Edge Node: US-East     Edge Node: EU-West     Edge Node: AP-East
   (Cloudflare Worker)   (Cloudflare Worker)    (Cloudflare Worker)
          │                     │                     │
   Edge KV Lookup         Edge KV Lookup        Edge KV Lookup
          │                     │                     │
   HTTP 302 Redirect      HTTP 302 Redirect     HTTP 302 Redirect
          │                     │                     │
          └─────────────────────┼─────────────────────┘
                                ▼
                   Async Analytics Event Queue
                   (Apache Kafka / Vector Ingest)

3. Implementation Code Contract (TypeScript / Cloudflare Worker Engine)

```typescript
export interface Env {
  LINK_KV: KVNamespace;
  ANALYTICS_QUEUE: Queue;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const slug = url.pathname.slice(1);

    if (!slug) {
      return Response.redirect('https://flux.dev', 302);
    }

    // 1. Edge KV Lookup (<2ms)
    const targetUrl = await env.LINK_KV.get(`slug:${slug}`);
    if (!targetUrl) {
      return new Response('Link Not Found', { status: 404 });
    }

    # 2. Async Non-Blocking Click Event Queue (<1ms)
    ctx.waitUntil(
      env.ANALYTICS_QUEUE.send({
        slug,
        timestamp: new Date().toISOString(),
        ip: request.headers.get('cf-connecting-ip'),
        country: request.headers.get('cf-ipcountry'),
        userAgent: request.headers.get('user-agent'),
        referrer: request.headers.get('referer')
      })
    );

    // 3. Instant 302 Redirect
    return Response.redirect(targetUrl, 302);
  }
};
```

4. Implementation Step-by-Step

Step 1: Write Cloudflare Worker / Fly.io Edge runtime script with non-blocking event queue dispatcher.
Step 2: Implement KV Invalidation Webhook from Core Flux API syncing slug updates to global edge KV in <500ms.
Step 3: Benchmark global response latency verifying sub-10ms redirect performance worldwide.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/enterprise/abuse_detection.md](file:///home/logan78/Desktop/plan/docs/enterprise/abuse_detection.md) |
| **Next** | [docs/global/geo_db_replication.md](file:///home/logan78/Desktop/plan/docs/global/geo_db_replication.md) |
| **Children** | [task_501_edge_redirects.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_501_edge_redirects.md), [task_502_geo_replication.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_502_geo_replication.md), [task_503_anycast_dns.md](file:///home/logan78/Desktop/plan/tasks/05_global/task_503_anycast_dns.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
