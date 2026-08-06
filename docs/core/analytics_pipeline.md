# Subsystem Specification: Analytics Ingestion & Processing Pipeline

> **Source**: Consolidated from `prompt/1/08.md`, `prompt/1/10.md`, `prompt/2/01.md`, `prompt/4/02.md`, `prompt/5/04.md`.

## Part I: Event Collection & Core Ingestion Pipeline
> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Core Platform Monolith)**.
> You MUST implement **Part 1 (Core Platform Monolith) — File 08.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART 8 — EVENT COLLECTION & ANALYTICS INGESTION PIPELINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux v1** — a production-ready, modular monolith URL shortening platform.
> Your task in this step is to implement **Part 8 — Event Collection & Analytics Ingestion Pipeline** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

Good catch. This is exactly where most tutorials make a mistake.

They build:

Redirect
↓

Dashboard

But the dashboard has nothing to display.

Real companies build:

Redirect
↓

Event Collection

↓

Storage

↓

Dashboard

So let's do it the production way.

FLUX PART I — Chapter 8
Analytics Event Pipeline

This is NOT the analytics dashboard.

This chapter is about collecting click events.

Without this module,
there is no analytics.

1. Why Analytics is Separate

A beginner writes

Redirect

↓

Database

↓

Increment Click Count

↓

Redirect User

Looks fine.

It's terrible.

Why?

Imagine

10 Million Redirects

↓

10 Million UPDATE Queries

Your redirect latency becomes terrible.

Users should never wait because analytics is slow.

2. Analytics Architecture
               Browser

                   │

            GET /openai

                   │

         Redirect Service

                   │

          301 Redirect User

                   │

        Fire Click Event

                   │

      Analytics Collector

                   │

           PostgreSQL

Notice

Redirect happens FIRST.

Analytics SECOND.

3. Redirect Path

Critical path

Browser

↓

Redis

↓

Redirect

↓

Done

Analytics is not on the critical path.

4. Analytics Module

Create a new module

analytics/

handler.go

service.go

repository.go

event.go

aggregator.go

dto.go

This module owns all analytics.

5. What is a Click?

Not just

LinkID

A click contains much more.

Click

↓

Link

↓

Timestamp

↓

IP

↓

Country

↓

City

↓

Device

↓

Browser

↓

Operating System

↓

Referrer

↓

User Agent

Already

this is a separate domain.

6. Database

New table

clicks

Fields

ID

LinkID

Timestamp

IPAddress

Country

City

Browser

OS

Device

Referrer

UserAgent

Future

UTM

Session

Campaign

Conversion
7. Click Flow
Browser

↓

GET /abc123

↓

Redirect

↓

Send Analytics Event

↓

Analytics Service

↓

Insert Click

Notice

User already left.

8. Why Separate Module?

Because later

Analytics becomes

ClickHouse

Kafka

Redis

Workers

Redirect shouldn't know.

9. Analytics Service

Responsibilities

RecordClick()

Aggregate()

Count()

DailyStats()

No redirect logic.

10. Analytics Repository
Insert()

Count()

Find()

Daily()

TopCountries()

TopBrowsers()

Only persistence.

11. Request Information

Every request contains

Headers

↓

User-Agent

↓

Accept-Language

↓

IP

↓

Referer

Analytics extracts

everything.

12. User Agent Parsing

Input

Mozilla/5.0...

Output

Browser

Chrome

Version

138

OS

Windows

Device

Desktop

Store parsed values.

Don't parse every dashboard request.

13. Geo Location

Input

IP

Output

Country

India

City

Patna

Initially

Can use

MaxMind.

Later

Cloudflare.

14. Referrer

Store

google.com

twitter.com

github.com

linkedin.com

Very useful.

15. Event Model

Think

Event

Not

Click Counter.

One click

=

One event.

Future

Click

↓

Signup

↓

Purchase

↓

Conversion

Everything becomes events.

16. Why Not Store Only Count?

Many beginners

Links

Clicks

42

Bad.

Because later

Need

Countries

Devices

Browsers

Referrers

Heatmaps

Daily Charts

Impossible.

Store events.

Aggregate later.

17. Aggregation

Future dashboard

Needs

Yesterday

↓

100 clicks

↓

Today

↓

240 clicks

Don't compute

every request.

Later

Background workers

will aggregate.

For Part I

Simple SQL.

18. Performance

Insert

One row

per click.

No UPDATE.

INSERTs scale much better than repeatedly updating the same counter row under heavy concurrency.

19. Failure Strategy

Analytics fails?

Should redirect fail?

No.

Redirect

↓

Analytics Error

↓

Ignore

↓

User Redirected

Analytics is important.

Redirect is critical.

Critical path always wins.

20. Future Event Pipeline

Part I

Redirect

↓

Analytics

↓

PostgreSQL

Part V

Redirect

↓

Kafka

↓

Workers

↓

ClickHouse

Architecture already supports it.

21. Analytics APIs

Eventually

GET /analytics

GET /analytics/countries

GET /analytics/devices

GET /analytics/browsers

GET /analytics/referrers

These power the dashboard later.

22. Folder Structure
analytics/

handler.go

service.go

repository.go

event.go

mapper.go

validator.go

dto.go

routes.go
23. Tests

Test

✓ Insert

✓ IP parsing

✓ Browser parsing

✓ Referrer parsing

✓ Device parsing

✓ Country parsing

✓ Performance

24. Future

Later

Analytics grows into

Real Time

↓

Kafka

↓

Workers

↓

ClickHouse

↓

Dashboard

↓

AI Reports

No rewrite needed.

End of Chapter 8

Flux now contains

Authentication

Link Management

Redirect Engine

Analytics Pipeline

At this point, you are no longer building a "URL shortener." You're building a link platform.

What Comes Next?

The natural order now is:

Chapter 9 — Dashboard Backend
Search
Pagination
Filters
Statistics
Link management UI APIs
Chapter 10 — QR Service
PNG/SVG generation
Download endpoints
Storage strategy
Chapter 11 — Production Hardening
Rate limiting
Security headers
Observability
Health/readiness checks
Backups
Release checklist
Chapter 12 — Deploy Flux v1
Docker
Docker Compose
Nginx
CI/CD
First production deployment
One architectural note

If I were designing this from scratch today, I'd make one refinement to Chapter 8:

Instead of naming the module Analytics, I'd call it Events internally.

Why?

Because analytics is a consumer of events.

Tomorrow you'll also use events for:

notifications,
audit logs,
webhooks,
billing,
conversion tracking,
AI insights.

So the long-term architecture becomes:

Redirect
      │
      ▼
Event Producer
      │
      ▼
Event Store
      ├── Analytics
      ├── Audit Logs
      ├── Webhooks
      ├── Notifications
      └── Billing

That's a more scalable mental model and one that many event-driven systems adopt as they grow.


---

###

## Part II: GeoIP & User-Agent Enrichment
> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Core Platform Monolith)**.
> You MUST implement **Part 1 (Core Platform Monolith) — File 10.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART 10 — DASHBOARD ANALYTICS & ADVANCED FEATURES

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux v1** — a production-ready, modular monolith URL shortening platform.
> Your task in this step is to implement **Part 10 — Dashboard Analytics & Advanced Features** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

Excellent. This is how I'd write Chapter 9 if this were an engineering design document inside a real company.

FLUX PART I
Chapter 9 — Dashboard Backend

Goal: Build the backend that powers the entire web dashboard.

Important

This is NOT frontend.

We're building every API the frontend will consume.

Think of Next.js as just another client.

1. Problem Statement

A beginner thinks the dashboard is

SELECT * FROM links

That's wrong.

A dashboard is actually a Backend for Frontend (BFF).

It aggregates multiple domains into one optimized response.

Instead of the frontend making

Links

↓

Clicks

↓

User

↓

QR

↓

Statistics

through 10 different API calls,

the backend should aggregate everything.

2. Dashboard Responsibilities

The dashboard owns

Search

Pagination

Sorting

Filtering

Statistics

Dashboard Summary

Recent Activity

Bulk Operations

It does NOT own

Authentication
Redirects
Analytics
QR generation

Those belong to other modules.

3. Dashboard Architecture
                Browser

                    │

             Dashboard API

                    │

          Dashboard Service

      ┌────────┬─────────┐

      │        │         │

 Link Service Analytics User

      │        │         │

      └────────┴─────────┘

             PostgreSQL

Notice

Dashboard is mostly an orchestrator.

4. Dashboard Home

The first page should load everything needed in one request.

GET /dashboard

Returns

Current User

Workspace Summary

Recent Links

Quick Statistics

Recent Activity

Not

10 different API calls.

5. Dashboard Layout

Think like a product.

Dashboard

├── Summary Cards
├── Recent Links
├── Search
├── Filters
├── Statistics
├── Activity Feed
└── Quick Actions

Backend powers every section.

6. Summary Cards

Cards

Total Links

↓

Active Links

↓

Disabled Links

↓

Expired Links

↓

Total Clicks

↓

Today's Clicks

One SQL query should provide most of these.

Avoid separate queries for each metric.

7. Link Listing API
GET /links

Supports

Page

Page Size

Search

Sort

Filters

Everything should be optional.

8. Pagination

Bad

SELECT *

Good

LIMIT

OFFSET

Response

Items

Current Page

Page Size

Total Count

Has Next

Has Previous

Frontend should never calculate pagination.

9. Sorting

Support

Newest

Oldest

Alphabetical

Recently Updated

Most Clicked

Backend validates allowed values.

Never trust arbitrary sort fields from clients.

10. Searching

Search should work on

Slug

Destination

Title

Not

ID

Password

Internal fields

Future

OpenSearch

Current

PostgreSQL indexes.

11. Filtering

Support

Status

↓

Active

Disabled

Expired

Future

Domain

Tags

Folder

Campaign

Design filters so new ones can be added without changing API structure.

12. Link Detail API
GET /links/:id

Returns

Destination

Slug

QR URL

Status

Expiration

Created

Updated

Recent Clicks

Notice

The frontend doesn't need to call Analytics separately for recent click summary.

13. Dashboard Statistics API

Separate endpoint

GET /dashboard/stats

Returns

Links

Clicks

Daily Growth

Recent Activity

Top Links

Small

Fast

Frequently refreshed.

14. Bulk Operations

Users shouldn't delete

100 links

one at a time.

Support

Bulk Delete

Bulk Disable

Bulk Enable

API

POST /links/bulk

Payload

Action

IDs
15. Search Architecture

Flow

Browser

↓

Dashboard

↓

Link Service

↓

Repository

↓

SQL

↓

Response

Search logic belongs inside the service.

Database executes optimized query.

16. Dashboard Repository

Repository should expose

ListLinks()

CountLinks()

Summary()

Recent()

Search()

BulkUpdate()

BulkDelete()

No business logic.

17. Dashboard Service

Service responsibilities

BuildSummary()

ValidateFilters()

ValidatePagination()

Search()

BulkDelete()

BulkDisable()

RecentLinks()

Everything user-facing belongs here.

18. Authorization

Every query must include ownership.

Never

SELECT * FROM links;

Always

WHERE owner_id = ?

Future

workspace_id = ?

This is one of the most common sources of data leaks in SaaS products.

19. Performance

Avoid N+1 queries.

Wrong

Links

↓

Loop

↓

Get Click Count

↓

Loop

↓

Get QR

Correct

One Query

↓

Everything

Or a small number of well-designed queries.

20. Response Shape

Every list endpoint should look identical.

{
  "data": [],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 325,
    "totalPages": 17
  }
}

Consistency simplifies frontend development.

21. Error Handling

Possible errors

Invalid Page

Invalid Sort

Invalid Filter

Unauthorized

Link Not Found

Every error should map to a consistent error code and HTTP status.

22. Caching Strategy

Should the dashboard be cached?

Generally:

Individual dashboard pages: No (user-specific, changes often)
Summary statistics: Yes (short TTL if expensive)
Static metadata (e.g., available filter values): Yes

Don't cache entire user dashboards.

23. Future Dashboard

The API should already be designed to support future widgets.

Dashboard

├── Links
├── Analytics
├── Domains
├── Teams
├── Billing
├── API Usage
├── Notifications
├── AI Insights

If each widget has a clear backend contract now, adding new ones later won't require redesigning existing APIs.

24. Testing

Test cases should include:

Pagination boundaries
Empty search results
Partial search matches
Sorting correctness
Filter combinations
Bulk delete success/failure
Unauthorized access
Ownership enforcement
Large datasets
Performance under load
25. Definition of Done

This chapter is complete when:

Users can view their dashboard.
Link lists support pagination, sorting, filtering, and search.
Summary statistics are available.
Bulk operations work.
APIs are documented and tested.
Query performance is acceptable for the expected dataset.
What We've Built So Far
Core Infrastructure
        │
Authentication
        │
Link Management
        │
Redirect Engine
        │
Event Collection
        │
Dashboard Backend   ✅

At this point, the backend is capable of powering a real web interface for managing links.

Before Chapter 10

One architectural improvement I'd make before implementing the QR Service is to introduce a Storage Abstraction.

Instead of having the QR module write directly to the local filesystem, define a storage interface from the beginning:

Storage
├── Local Disk
├── Amazon S3
├── Cloudflare R2
└── MinIO

Then the QR Service depends on the storage interface, not on a specific storage implementation. In Part I you can use local disk, and in Part V switch to cloud object storage with almost no changes to the QR business logic. This is a common pattern in production systems and keeps infrastructure concerns separated from domain logic.


---

###

## Part III: Time-Series Analytics & ClickHouse Storage
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

## Part IV: Enterprise Conversion Tracking & Funnels
> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Enterprise)**.
> You MUST implement **Part 4 (Flux Enterprise) — File 02.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART IV — FLUX ENTERPRISE
## CHAPTER 2 — ENTERPRISE CONVERSION TRACKING & FUNNEL ANALYTICS

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART IV — Flux Enterprise: Chapter 2 — Enterprise Conversion Tracking & Funnel Analytics** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Track conversion events (Signups, Purchases, Custom Lead Events, Revenue) and build multi-stage Conversion Funnels to visualize customer drop-off.

Key Features:
- Conversion Tracking JS Snippet & REST API (`POST /api/v1/track/conversion`).
- Revenue association per conversion event.
- Funnel Builder Engine (Visualizing conversion rates step-by-step).

2. Database Schema Design

```sql
CREATE TABLE IF NOT EXISTS conversion_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    visitor_token VARCHAR(64) NOT NULL,
    event_name VARCHAR(100) NOT NULL, -- e.g. "purchase", "signup"
    revenue NUMERIC(12,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS funnels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    steps JSONB NOT NULL, -- Array of step definition objects
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

3. Implementation Step-by-Step

Step 1: Build robust Conversion Collector REST Endpoint with CORS validation.
Step 2: Implement Funnel Analysis Aggregator SQL query counting unique visitor progression through configured steps.
Step 3: Build visual funnel chart rendering API endpoint.


---

###

## Part V: Global Real-Time Analytics Pipeline
> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Flux Global Scale)**.
> You MUST implement **Part 5 (Flux Global Scale) — File 04.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART V — FLUX GLOBAL SCALE
## CHAPTER 4 — GLOBAL REAL-TIME ANALYTICS INGESTION PIPELINE

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux (Growth, SaaS, Enterprise & Global Scale)** — a production-ready link management & analytics platform.
> Your task in this step is to implement **PART V — Flux Global Scale: Chapter 4 — Global Real-Time Analytics Ingestion Pipeline** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

1. Product Requirement & Goal

Deploy a global real-time analytics ingestion pipeline capable of processing 1,000,000 click events per second using Apache Kafka / Redpanda, Vector log collectors, and ClickHouse cluster tables.

Key Features:
- Vector log collection agents running on edge proxies.
- Apache Kafka / Redpanda distributed log streaming backbone.
- ClickHouse Distributed Tables with MergeTree engine.

2. Architecture & ClickHouse Schema

              Edge Vector Log Agents
                        │
             Apache Kafka / Redpanda
            (Topic: `flux.click.events`)
                        │
             ClickHouse Kafka Engine
                        │
             ClickHouse Distributed Table

```sql
-- ClickHouse Distributed Table Definition
CREATE TABLE IF NOT EXISTS ClickHouse_events_local ON CLUSTER flux_cluster (
    id UUID,
    link_id UUID,
    timestamp DateTime64(3, 'UTC'),
    country_code LowCardinality(String),
    user_agent String,
    device_type LowCardinality(String),
    referrer_domain String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (link_id, timestamp);

CREATE TABLE IF NOT EXISTS ClickHouse_events ON CLUSTER flux_cluster AS ClickHouse_events_local
ENGINE = Distributed(flux_cluster, default, ClickHouse_events_local, rand());
```

3. Implementation Step-by-Step

Step 1: Set up Apache Kafka / Redpanda cluster with 16 partitions per topic.
Step 2: Configure ClickHouse Kafka Engine table consuming directly from Kafka topics into MergeTree storage.
Step 3: Verify real-time query aggregation throughput on 10,000,000 sample records.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [docs/core/redirect_engine.md](file:///home/logan78/Desktop/plan/docs/core/redirect_engine.md) |
| **Next** | [docs/core/qr_service.md](file:///home/logan78/Desktop/plan/docs/core/qr_service.md) |
| **Children** | [task_100_bootstrap_backend.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_100_bootstrap_backend.md), [task_101_auth_service.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_101_auth_service.md), [task_102_base62_encoder.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_102_base62_encoder.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
