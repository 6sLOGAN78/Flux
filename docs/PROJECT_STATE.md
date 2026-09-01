# Project State

## Project Identity
* **Name**: Flux
* **Purpose**: URL shortening, dynamic link routing, analytics, SaaS multi-tenancy, attribution platform.
* **Problem**: Claims to solve enterprise link management, attribution, and analytics at scale.
* **Target Users**: Marketers, developers, enterprises.
* **Current Objective**: Establish a trustworthy, production-hardened baseline.

## Current State
* **What works**: Link CRUD is fully functional and isolated by tenant. Real-time Analytics ingestion (Redis Streams -> ClickHouse) is deployed and powering the frontend dashboard. Authentication uses Clerk exclusively with secure tenant mapping. Campaigns and UTM tracking are fully functional end-to-end with immutable attribution.
* **What partially works**: Custom Domains, Webhooks, Billing, and AI are currently just mock UI shells pending backend implementation.
* **What is broken**: Nothing in the critical path. Previous production blockers (auth spoofing, graceful shutdown, short-code PRNG collisions) have been remediated.
* **What is missing**: Advanced feature modules (Custom Domains, Billing, Webhooks, AI).
* **What is deprecated**: N/A

## Technology Stack
* **Frontend**: React 19, Vite, TypeScript, Tailwind CSS, Clerk, React Query.
* **Backend**: Go 1.25, Echo v4, Clerk SDK.
* **Database**: PostgreSQL (pgx/v5) for relational state. ClickHouse for analytics. Redis for async stream buffering and redirect caching.

## Current Phase
* **Phase**: PHASE 13 - Multi-Touch Attribution (Task 13D Complete)
* **Status**: IN PROGRESS
* **Why**: The Phase 13 backend infrastructure is fully operational. Conversion events are tracked, buffered through Redis, ingested into ClickHouse, and successfully mapped to historical touchpoints via `GET /api/v1/analytics/attribution`. Next is 13E (Frontend Attribution UI).

## Current Priorities
1. Expand features: Campaigns, UTM, Custom Domains.
2. Expand features: Campaigns, UTM, Custom Domains.

## Blockers
* None.

## Stability
* **Stable**: Auth, Multi-tenancy, Link CRUD, Redirects, Analytics Pipeline, API.
* **Experimental**: None currently.
* **Risky**: None.
