# Subsystem Specification: Link Management

> **Source**: Extracted from `prompt/1/06.md`.

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Core Platform Monolith)**.
> You MUST implement **Part 1 (Core Platform Monolith) — File 06.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART 6 — LINK SHORTENING & MANAGEMENT SYSTEM

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux v1** — a production-ready, modular monolith URL shortening platform.
> Your task in this step is to implement **Part 6 — Link Shortening & Management System** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

Perfect. This is where Flux actually becomes a URL shortener.

Until now, we have built the platform. Now we build the core business.

This chapter is the heart of the entire project.

FLUX PART I — Chapter 6
Link Module

Goal: Design and implement the complete Link Management System.

When this chapter is finished, Flux can:

Create short URLs
Edit them
Delete them
Disable them
Expire them
Password protect them
Search them
List them
Generate QR
Prepare them for analytics

This is the most important module in the project.

Step 1 — Business Problem

A beginner thinks:

"A link is just a URL."

A backend engineer thinks:

"A link is an aggregate with business rules."

A Flux link has:

Destination

Slug

Owner

Status

Visibility

Password

Expiration

Metadata

QR

Analytics

Already that's much more than a URL.

Step 2 — Link Domain

The Link Aggregate owns:

Link
│
├── Destination
├── Slug
├── QR
├── Status
├── Expiration
├── Password
└── Metadata

Everything about links belongs here.

Not in Redirect.

Not in Dashboard.

Step 3 — Database

The first instinct is:

links

id

url

Wrong.

Production table:

links

id

user_id

slug

destination

title

description

favicon

status

visibility

password_hash

expires_at

created_at

updated_at

Future versions will add

workspace_id

domain_id

utm_template

redirect_rules

tags

folder

Design for growth.

Step 4 — Link States

A link is never simply "exists."

It has a lifecycle.

Draft

↓

Active

↓

Disabled

↓

Expired

↓

Deleted

Redirect only works if

status == ACTIVE
Step 5 — Business Rules

Creating a link isn't just inserting a row.

Rules:

Destination

Must exist
Must be valid URL
Must use HTTP/HTTPS

Slug

Unique
Reserved words forbidden
Length limit
Allowed characters

Password

Optional
Stored as hash

Expiration

Optional
Must be future
Step 6 — Slug Strategy

This decision affects everything.

Possible options

Option 1

Database ID

1

2

3

Terrible.

Predictable.

Option 2

Base62(Database ID)

a

b

c

d

Still sequential.

Leaks traffic.

Option 3

UUID

8b57a6f3-...

Too long.

Option 4

NanoID ⭐

Ab9KxP

K82aMw

MzQ1Lp

Best choice.

Fast.

Tiny.

Collision probability extremely low.

Flux should use NanoID.

Step 7 — Reserved Slugs

Never allow

login

signup

dashboard

admin

docs

api

auth

health

favicon.ico

robots.txt

assets

Otherwise

flux.dev/login

could redirect somewhere else.

Step 8 — Link Creation Flow
POST /links

↓

Validate Request

↓

Validate URL

↓

Generate Slug

↓

Check Collision

↓

Insert Database

↓

Generate QR

↓

Return Link

Notice

QR generation happens after persistence.

Never before.

Step 9 — Collision Handling

NanoID collisions are rare.

Still.

Always

Generate

↓

Exists?

↓

No

↓

Save

If yes

Generate another.

Never trust probability.

Step 10 — Editing Links

Editable

Destination

Expiration

Status

Password

Not editable

ID

Owner

CreatedAt

Slug?

Configurable.

Some platforms allow changing it.

Some don't.

Flux v1:

Allow changing slug.

Step 11 — Delete Strategy

Never

DELETE FROM links

Instead

Status

↓

Deleted

Soft delete.

Benefits

Restore
Analytics preserved
Audit history
Safer operations
Step 12 — Search

Support

Slug

Destination

Title

Eventually

Full-text search.

For v1

SQL LIKE + indexes.

Step 13 — Pagination

Never

SELECT * FROM links;

Instead

Limit

Offset

Sort

Filter

Example

Newest

Oldest

Active

Expired
Step 14 — Link Status
Active

Disabled

Expired

Deleted

Redirect checks this first.

Step 15 — Password Protected Links

Flow

Visit Link

↓

Password?

↓

No

↓

Redirect

↓

Yes

↓

Password Page

↓

Verify

↓

Redirect

Never expose password hash.

Step 16 — Expiration

Every redirect

Checks

Current Time

>

Expiration?

If expired

410 Gone

or

Custom page.

Step 17 — Link Metadata

Store

Title

Description

Image

Favicon

Later

Social preview.

Step 18 — Repository

Repository should only know persistence.

Create()

Update()

Delete()

FindBySlug()

FindByID()

FindByUser()

Search()

Exists()

Count()

No validation.

No slug generation.

Step 19 — Service

Business logic.

CreateLink()

GenerateSlug()

Validate()

Update()

Disable()

Delete()

GenerateQR()

Search()

Every business rule lives here.

Step 20 — Handler

Only HTTP.

POST

GET

PATCH

DELETE

Never hash passwords.

Never generate slugs.

Never execute SQL.

Step 21 — API Design
POST /links

Create.

GET /links

List.

GET /links/:id

Detail.

PATCH /links/:id

Update.

DELETE /links/:id

Delete.

Step 22 — DTOs

Request

Destination

Slug

Password

Expiration

Response

ID

ShortURL

Slug

QR

Status
Step 23 — Sequence Diagram

Create Link

Browser

↓

POST

↓

Handler

↓

Validate

↓

Service

↓

Generate NanoID

↓

Repository

↓

Database

↓

QR

↓

Response
Step 24 — Error Cases
Slug Exists

↓

409

Invalid URL

↓

400

Expired Date

↓

400

Not Found

↓

404

Deleted Link

↓

410

Step 25 — Unit Tests

Must test

Slug generation
Slug collisions
URL validation
Password hashing
Expiration logic
Soft delete
Search
Pagination
Reserved slugs
Duplicate slug rejection
Step 26 — Future Compatibility

When designing today, think about tomorrow.

Don't hard-code assumptions that make later features painful.

For example:

A link currently belongs to a user, but later it may belong to a workspace.
Today it has one destination, later it may support A/B testing or geo/device redirects.
Today it uses the default domain, later it may use custom domains.

Design the module so these additions require extending the model rather than rewriting it.

What Part I Still Needs

After the Link Module, the remaining major chapters are:

Chapter 7 — Redirect Engine (the highest-performance part of the system)
Chapter 8 — Dashboard & Search
Chapter 9 — QR Service
Chapter 10 — Testing & Production Hardening
Chapter 11 — Docker, Deployment & Release v1

Only after those are complete is Flux Core v1 truly finished.

One suggestion

From this point on, I would stop writing purely descriptive chapters and switch to the style of a real engineering design document.

Each remaining chapter would include:

Requirements
Architecture
Database changes
API contract
Sequence diagrams
Failure scenarios
Performance considerations
Security considerations
Implementation order
Testing checklist

That format is much closer to what you'd see in a software design review at a startup or large tech company, and it will make the project documentation much more valuable as your portfolio grows.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/plan/docs/ARCHITECTURE.md) |
| **Previous** | [CHANGELOG.md](file:///home/logan78/Desktop/plan/CHANGELOG.md) |
| **Next** | [docs/core/redirect_engine.md](file:///home/logan78/Desktop/plan/docs/core/redirect_engine.md) |
| **Children** | [task_100_bootstrap_backend.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_100_bootstrap_backend.md), [task_101_auth_service.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_101_auth_service.md), [task_102_base62_encoder.md](file:///home/logan78/Desktop/plan/tasks/01_core/task_102_base62_encoder.md) |
| **Dependencies** | [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/plan/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/plan/database/postgres_master_schema.sql) |
| **Related Documents** | [ai/PERFORMANCE.md](file:///home/logan78/Desktop/plan/ai/PERFORMANCE.md), [ai/SECURITY.md](file:///home/logan78/Desktop/plan/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/plan/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/plan/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
