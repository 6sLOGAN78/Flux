# DevOps Specification: Terraform Infrastructure

> You are acting as the Lead Staff Backend & Systems Engineer building **Flux (Core Platform Monolith)**.
> You MUST implement **Part 1 (Core Platform Monolith) — File 13.md** by strictly following the **25-Step Engineering Lifecycle** (from `step_by_step.md`) and integrating all topics from `whatshouldbeineverything.md`.
> 
> **MANDATORY EXECUTION SEQUENCE (25 STEPS):**
> 1. PRD ➔ 2. Feature Requirements ➔ 3. HLD ➔ 4. LLD ➔ 5. Database Design ➔ 6. API Design ➔ 7. Folder Structure ➔ 8. Domain Models ➔ 9. DTOs ➔ 10. Validation Rules ➔ 11. Business Logic ➔ 12. Repository Layer ➔ 13. Service Layer ➔ 14. Handlers ➔ 15. Middleware ➔ 16. Background Jobs ➔ 17. Caching ➔ 18. Security ➔ 19. Testing ➔ 20. Documentation ➔ 21. Deployment ➔ 22. Monitoring ➔ 23. Performance Review ➔ 24. Refactoring ➔ 25. Release

---


# CODING AGENT IMPLEMENTATION PROMPT — PART 13 — PACKAGING, CONTAINERIZATION & PRODUCTION DEPLOYMENT

> **ROLE & INSTRUCTION FOR CODING AGENT / AI DEVELOPER:**
> You are acting as the Lead Backend & Full-Stack Engineer building **Flux v1** — a production-ready, modular monolith URL shortening platform.
> Your task in this step is to implement **Part 13 — Packaging, Containerization & Production Deployment** exactly as specified below.
> 
> **EXECUTION REQUIREMENTS:**
> 1. Read and analyze every section, architecture layout, database schema, API contract, and edge case in this document.
> 2. Write clean, production-grade code that satisfies all requirements of this phase without taking shortcuts.
> 3. Ensure full compatibility with preceding and subsequent modules of the Flux platform.

---

Excellent. This is the last chapter of Part I.

This is where your project stops being "code on GitHub" and becomes software that can actually be deployed.

FLUX PART I
Chapter 12 — Deploy Flux v1

Goal: Package, deploy, release, and operate Flux Core v1.

This chapter isn't just "how to run Docker."

It's about the software delivery lifecycle.

1. Deployment Philosophy

Many beginners think deployment is

git push

↓

server

Professional deployment is

Developer

↓

Git

↓

CI

↓

Tests

↓

Build

↓

Container

↓

Registry

↓

Deploy

↓

Monitor

Deployment starts when you commit code.

2. Production Architecture

For Flux v1

                Internet
                     │
             HTTPS (443)
                     │
                 Nginx
                     │
          ┌──────────┴──────────┐
          │                     │
     Static Frontend      Flux Backend
                                │
                       ┌────────┴─────────┐
                       │                  │
                  PostgreSQL          Redis

Simple.

Reliable.

Easy to maintain.

3. Docker Strategy

Every component gets its own container.

frontend

backend

postgres

redis

nginx

Never install PostgreSQL directly on the server.

Everything runs in containers.

4. Backend Dockerfile

Use a multi-stage build.

Why?

Wrong

Go Compiler

↓

Application

↓

Production

Huge image.

Correct

Compile

↓

Copy Binary

↓

Tiny Runtime Image

Benefits

Smaller image
Faster deployment
Better security
5. Frontend

Frontend is independent.

Next.js

↓

Production Build

↓

Nginx

No Node.js in production if you're exporting static assets.

If you use SSR, Node.js stays.

6. Docker Compose

Development architecture

Docker Compose

↓

Backend

↓

Frontend

↓

Redis

↓

PostgreSQL

↓

Nginx

One command starts everything.

7. Environment Variables

Never hardcode

JWT Secret

Database URL

Redis URL

API Keys

Use

.env

↓

Container

↓

Application

Different environments

Development

Staging

Production

Different configs.

8. Reverse Proxy

Nginx responsibilities

HTTPS

↓

Compression

↓

Caching

↓

Routing

↓

Security Headers

Backend shouldn't terminate TLS directly.

9. HTTPS

Never deploy

http://

Always

https://

Certificate

↓

Let's Encrypt

or

Cloudflare

10. CI/CD Pipeline

Every push

GitHub

↓

Run Tests

↓

Lint

↓

Build Docker Image

↓

Push Registry

↓

Deploy

Never deploy manually.

11. Build Pipeline
Commit

↓

GitHub Actions

↓

Go Tests

↓

Frontend Tests

↓

Docker Build

↓

Image Registry

↓

Deploy

Every build should be reproducible.

12. Image Registry

Images live in

GitHub Container Registry

or

Docker Hub

Production servers pull immutable images.

13. Deployment Strategy

For Flux v1

Rolling restart is sufficient.

Old Container

↓

Start New

↓

Health Check

↓

Stop Old

Avoid downtime.

14. Startup Order

Containers should start in this order.

PostgreSQL

↓

Redis

↓

Backend

↓

Frontend

↓

Nginx

Backend shouldn't start before PostgreSQL is ready.

15. Health Checks

Docker

↓

Health endpoint

↓

Healthy?

↓

Continue

If unhealthy

↓

Restart container

16. Logs

Don't log to files inside containers.

Instead

stdout

stderr

↓

Docker

↓

Log Aggregator

Containers should be disposable.

17. Database Migrations

Deployment

New Version

↓

Run Migrations

↓

Start Application

Never manually execute SQL on production.

18. Rollback Strategy

Bad deployment?

Version 1.0.1

↓

Problem

↓

Deploy Version 1.0.0

Rollback should take minutes, not hours.

19. Backup Strategy

Before deployment

Backup Database

↓

Deploy

↓

Success?

↓

Keep Backup

Never deploy without a recovery plan.

20. Monitoring

Minimum production dashboard

CPU

Memory

Requests

Latency

Errors

Database

Redis

Know when something breaks.

21. Domain Setup

Example

flux.dev

↓

DNS

↓

Nginx

↓

Backend

Future

go.company.com

Part II.

22. Release Checklist

Before every release:

Code
All tests pass
Lint passes
API documentation updated
Database
Migrations reviewed
Backup completed
Security
Secrets configured
HTTPS enabled
JWT expiration verified
Infrastructure
Health checks passing
Monitoring active
Logs visible
Product
Smoke tests completed
Redirects verified
Dashboard verified
23. Versioning

Use semantic versioning.

v1.0.0

Patch

v1.0.1

Minor

v1.1.0

Major

v2.0.0

Every release should be tagged in Git.

24. Definition of Done

Flux Core v1 is complete when:

Backend runs in Docker.
Frontend runs in Docker.
PostgreSQL and Redis are containerized.
HTTPS is configured.
CI/CD builds successfully.
Health checks pass.
Monitoring basics are in place.
Deployment is repeatable.
Rollback is documented.
Release is tagged.
25. Flux Core v1 Architecture
                     Internet
                          │
                     HTTPS (443)
                          │
                    Nginx Reverse Proxy
                          │
          ┌───────────────┴───────────────┐
          │                               │
     Next.js Frontend              Go Backend
                                          │
                         ┌────────────────┴──────────────┐
                         │                               │
                    PostgreSQL                      Redis
                         │
                  Local Storage (QR)

This is a production-ready modular monolith.

It is simple enough for a small team yet structured enough to evolve.

🎉 Part I Complete

You have built:

Product requirements and architecture
Clean backend foundation
Authentication
Link management
Redirect engine
Event collection
Dashboard backend
QR service
Production hardening
Deployment pipeline

This is already a substantial portfolio project.

Before Part II: Engineering Retrospective

A real team would pause here and review the release.

What went well?
Clear module boundaries.
Feature-based architecture.
Clean separation of concerns.
Ready for extension.
Technical debt accepted intentionally
Offset pagination instead of cursor pagination.
Analytics stored in PostgreSQL instead of a specialized analytics database.
Local file storage instead of cloud object storage.
Single backend process (modular monolith).

These are conscious trade-offs, not mistakes.

Extension points prepared

Part I was designed so Part II can add features without major rewrites:

Custom domains
Rich analytics dashboards
Smart redirects
A/B testing
Deep links
Campaigns
Advanced QR customization
🚀 Part II Preview — Flux Growth

Part II is not just "more features."

It's where Flux transforms from a URL shortener into a link management and marketing platform, similar in spirit to products like Dub.

We'll redesign and extend the system with modules such as:

Analytics Dashboard
↓

Time-Series Analytics
↓

Campaigns & UTM Builder
↓

Custom Domains
↓

DNS Verification
↓

SSL Management
↓

Smart Redirect Rules
↓

Geo & Device Targeting
↓

A/B Testing
↓

Deep Links
↓

Advanced QR Customization
↓

Background Workers
↓

Caching Improvements

The architecture will also evolve, introducing asynchronous jobs, richer analytics, and more sophisticated domain models—while still keeping the modular monolith intact before the transition to distributed systems in later parts.


---

###

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [ai/DEPLOYMENT.md](file:///home/logan78/Desktop/flux/ai/DEPLOYMENT.md) |
| **Previous** | [ops/kubernetes.md](file:///home/logan78/Desktop/flux/ops/kubernetes.md) |
| **Next** | [ops/cicd.md](file:///home/logan78/Desktop/flux/ops/cicd.md) |
| **Children** | None |
| **Dependencies** | [ai/DEPLOYMENT.md](file:///home/logan78/Desktop/flux/ai/DEPLOYMENT.md), [docs/ARCHITECTURE.md](file:///home/logan78/Desktop/flux/docs/ARCHITECTURE.md) |
| **Related Documents** | [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->
