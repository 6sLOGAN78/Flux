# Subsystem Specification: Redirect Engine (Core, Smart Routing & Edge)

> **Source**: Consolidated from `prompt/1/07.md`, `prompt/2/04.md`, and `prompt/5/01.md`.

## Part I: Core High-Performance Redirect Handler (Go Backend)
> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Core Platform Monolith)**.
> You MUST implement **Part 1 (Core Platform Monolith) — File 07.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART 7 — HIGH-PERFORMANCE REDIRECT ENGINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux v1** — a production-ready, modular monolith URL shortening platform.
> Your task in this step is to implement **Part 7 — High-Performance Redirect Engine** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

xcellent. This chapter is the most performance-critical part of the entire platform.

Every feature in Flux exists to support one endpoint:

GET /:slug

This endpoint will receive 95–99% of all traffic.

Creating links might happen a few hundred times per day.

Redirects might happen millions of times per day.

Because of that, we design this module very differently from every other module.

FLUX PART I — Chapter 7
Redirect Engine
1. Product Requirement

When someone visits

https://flux.dev/openai

they should be redirected to

https://openai.com

within a few milliseconds.

Sounds simple.

It isn't.

2. Why Redirect is a Separate Module

A beginner would do

Browser

↓

Link Handler

↓

Database

↓

Redirect

Wrong.

Redirect is not part of CRUD.

It deserves its own module.

3. Redirect Architecture
                Browser
                    │
          GET /openai
                    │
             Redirect Handler
                    │
             Redirect Service
             ┌───────────────┐
             │               │
         Redis Cache     PostgreSQL
             │               │
             └───────┬───────┘
                     │
                 Destination URL
                     │
               HTTP 301 Redirect

Notice

The redirect service never talks to authentication.

It doesn't care who owns the link.

4. Redirect Lifecycle
Incoming Request

↓

Extract Slug

↓

Validate Slug

↓

Redis Lookup

↓

Cache Hit?

↓

YES → Redirect

↓

NO

↓

Database Lookup

↓

Validate Link

↓

Store in Redis

↓

Redirect

This is called the read path.

5. Why Redis Comes First

Suppose

/openai

gets

2 million requests/day

Without Redis

2 million PostgreSQL queries

With Redis

Database

↓

Only first request

↓

Everything else

↓

Redis

Database load becomes tiny.

6. Redis Cache Design

Key

link:{slug}

Example

link:openai

Value

Destination URL

Status

Expiration

Password Flag

Only information needed for redirect.

Don't cache unnecessary fields.

7. Redirect Decision Tree

Every redirect follows this order.

Slug Exists?

↓

No

↓

404


↓

Yes

↓

Deleted?

↓

410


↓

Active?

↓

No

↓

403


↓

Expired?

↓

Yes

↓

410


↓

Password Protected?

↓

Yes

↓

Password Page


↓

Redirect

This order matters.

8. Redirect Status Codes

Choose the correct HTTP response.

Permanent
301 Moved Permanently

Browser may cache.

Temporary
302 Found

Used when destination may change.

Flux should support both.

Deleted
410 Gone

Better than

404

because

The link existed

but was removed.

Missing
404 Not Found

Slug never existed.

9. Cache Strategy

We use

Cache Aside

Flow

Redis

↓

Hit

↓

Return


↓

Miss

↓

Database

↓

Update Cache

↓

Return

Simple.

Reliable.

Industry standard.

10. Cache Expiration

Never cache forever.

Example

TTL

5 minutes

or

30 minutes

Why?

Because links can change.

11. Updating Links

Suppose

/openai

changes destination.

Database

↓

Updated

Redis

↓

Still old

Problem.

Solution

Whenever link updates

Update Database

↓

Delete Redis Key

Next redirect

↓

Cache rebuilt.

This is called Cache Invalidation.

One of the hardest problems in distributed systems.

12. Redirect Sequence Diagram
Browser

↓

GET /openai

↓

Router

↓

Redirect Handler

↓

Redirect Service

↓

Redis

↓

Miss

↓

Repository

↓

Database

↓

Redis Update

↓

301 Redirect
13. Repository Responsibilities

Repository should expose

FindBySlug()

IncrementClickLater()


Nothing else.

Notice

No redirect logic.

14. Service Responsibilities

Redirect Service

Resolve()

Validate()

Cache()

Redirect()

Business decisions live here.

15. Click Tracking

Should redirect wait?

No.

Wrong

Redirect

↓

Save Click

↓

Analytics

↓

Redirect User

Correct

Redirect User

↓

Background Click Event

The user should never wait for analytics.

Even in Part I, separate redirect latency from analytics.

16. Reserved Routes

Never treat these as slugs.

/login

/signup

/api

/docs

/health

/dashboard

/favicon.ico

/robots.txt

Router should match them before redirect.

17. Security Checks

Redirects are a common attack vector.

Validate:

Valid destination URL
Allowed schemes (http, https)
Reject javascript:
Reject malformed URLs
Optionally block known malicious domains (future feature)
18. Failure Scenarios

Redis down?

↓

Use PostgreSQL.

Database down?

↓

Return

503 Service Unavailable

Never panic.

19. Performance Goals

Target

Redirect latency

Cache Hit

<5 ms

Cache Miss

<30 ms

Database queries

One query maximum.

20. Rate Limiting

Should redirect endpoint be rate-limited?

Generally:

API endpoints (/links, /auth) → Yes.
Public redirects (/:slug) → Usually no, or use a much higher threshold plus bot detection.

Otherwise, popular links could be throttled unnecessarily.

21. Future Compatibility

The Redirect Engine should already be designed so it can later support:

↓

Geo Redirect

↓

Device Redirect

↓

Language Redirect

↓

A/B Testing

↓

Deep Links

Even if those features aren't implemented yet, don't make assumptions that prevent them.

22. Redirect Module Structure
redirect/

handler.go

service.go

repository.go

cache.go

validator.go

dto.go

errors.go

routes.go

Notice

The redirect module owns cache logic.

Not the link module.

23. Tests

This module deserves extensive tests because it's the busiest path in the system.

Test cases include:

Existing slug
Missing slug
Deleted link
Disabled link
Expired link
Password-protected link
Redis hit
Redis miss
Cache rebuild
Cache invalidation after update
Invalid destination
301 vs 302 behavior
Chapter 7 Complete

Flux now has:

✅ Authentication
✅ Link Management
✅ High-performance Redirect Engine

At this point, Flux is already a working URL shortener.

Before moving on

Here's one improvement I'd make to the roadmap.

Instead of building Dashboard next, I'd build Infrastructure for Analytics first.

Reason:

The dashboard only displays data.

The analytics pipeline creates that data.

A better order is:

Redirect Engine ✅
Analytics Event Pipeline (capture clicks)
Dashboard (visualize clicks)
QR Service
Release v1

This mirrors how real products evolve: collect reliable data first, then build dashboards on top of it.


---

###

## Part II: Smart Redirect Engine (Geo, Device & Conditional Routing)
> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Growth)**.
> You MUST implement **Part 2 (Flux Growth) — File 04.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART II — FLUX GROWTH
## CHAPTER 4 — SMART REDIRECT ENGINE: GEO, DEVICE & CONDITIONAL ROUTING

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART II — Flux Growth: Chapter 4 — Smart Redirect Engine: Geo, Device & Conditional Routing** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Transform simple 1-to-1 short links into intelligent dynamic routing engines based on real-time visitor metadata.

Routing Types:
- Geo Routing: Redirect US visitors to `/us`, UK visitors to `/uk`, default to `/global`.
- Device Routing: Redirect iOS users to App Store, Android to Google Play, Desktop to Web App.
- Browser & Language Routing: Redirect based on `Accept-Language` header and browser type.
- Time-Based Routing: Change destination URL automatically before/after a specified date or operating hours.
- Conditional Fallback & Default Target.

2. Architecture & Rule Evaluation Flow

                Incoming Visitor GET /:slug
                            │
               Extract IP, User-Agent, Headers
                            │
               MaxMind GeoIP2 + UserAgent Parser
                            │
              Evaluate Rule Matrix (Priority 1..N)
                            │
           ┌────────────────┼────────────────┐
           ▼                ▼                ▼
     Geo Match?       Device Match?    Time Match?
           │                │                │
           └────────────────┼────────────────┘
                            ▼
               Execute Matching Redirect 302

3. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS redirect_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    rule_type VARCHAR(30) NOT NULL, -- 'geo', 'device', 'language', 'time'
    priority INTEGER DEFAULT 1,
    condition_key VARCHAR(100) NOT NULL, -- e.g. 'US', 'iOS', 'en-US'
    target_url TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_redirect_rules_link ON redirect_rules(link_id, priority ASC);
```

4. API Contracts

- `POST /api/v1/links/:id/rules`
  Request:
  ```json
  {
    "rules": [
      { "rule_type": "device", "condition_key": "iOS", "target_url": "https://apps.apple.com/app/id12345", "priority": 1 },
      { "rule_type": "geo", "condition_key": "DE", "target_url": "https://acme.de/deal", "priority": 2 }
    ]
  }
  ```
  Response: `{ "status": "success", "rules_count": 2 }`

5. Implementation Step-by-Step

Step 1: Integrate MaxMind GeoIP2 Reader / IP-API client into high-speed Redirect Engine memory.
Step 2: Build fast User-Agent parser (UAParser) caching parsed rules in Redis under `rules:link_id`.
Step 3: Write rule matching evaluation loop in execution flow of `GET /:slug`.


---

###

## Part III: Multi-Region Edge Redirect Engine (Cloudflare Workers / Fly.io)
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
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Previous** | [docs/core/link_management.md](file:///home/logan78/Desktop/flux/docs/core/link_management.md) |
| **Next** | [docs/core/analytics_pipeline.md](file:///home/logan78/Desktop/flux/docs/core/analytics_pipeline.md) |
| **Children** | [task_100_bootstrap_backend.md](file:///home/logan78/Desktop/flux/tasks/01_core/task_100_bootstrap_backend.md), [task_101_auth_service.md](file:///home/logan78/Desktop/flux/tasks/01_core/task_101_auth_service.md), [task_102_base62_encoder.md](file:///home/logan78/Desktop/flux/tasks/01_core/task_102_base62_encoder.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
