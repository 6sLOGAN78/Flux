# Project State

## Project Identity
* **Name**: Flux
* **Purpose**: URL shortening, dynamic link routing, analytics, SaaS multi-tenancy, attribution platform.
* **Problem**: Claims to solve enterprise link management, attribution, and analytics at scale.
* **Target Users**: Marketers, developers, enterprises.
* **Current Objective**: Move from a "mocked/vaporware" state into an actual functional MVP.

## Current State
* **What works**: Basic CRUD for links (creation, listing, updating, deleting) backed by PostgreSQL. Simple HTTP 302/301 redirects using database lookups.
* **What partially works**: The frontend UI is highly polished but operates almost entirely on mock data, local state, and unimplemented APIs.
* **What is broken**: Analytics endpoints return 500 (provider is `nil`). Redis cache is passed as `nil`.
* **What is missing**: 90% of the advertised features (Campaigns, Billing, Workspaces, Webhooks, Domains, AI, ClickHouse event ingestion, Advanced Routing) are not implemented in the backend/database.
* **What is experimental**: The entire `apps/backend/internal/modules` directory contains Go logic that is disconnected from the HTTP server and database.
* **What is deprecated**: N/A

## Technology Stack
* **Frontend**: React 19, Vite, TypeScript, Tailwind CSS.
* **Backend**: Go 1.25, Echo v4.
* **Database**: PostgreSQL (pgx/v5). ClickHouse (schema only). Redis (not wired).

## Current Phase
* **Phase**: Architecture Reality Check & Backend Wiring
* **Status**: IN PROGRESS
* **Why**: The repository presents a massive surface area of features but only implements basic link shortening. The priority is to wire the existing UI to actual database/backend implementations.

## Current Priorities
1. Implement Redis caching for redirects.
2. Implement ClickHouse event ingestion on redirect.
3. Hook up the backend Analytics endpoint to ClickHouse.
4. Expand the PostgreSQL schema to support the mock features in the UI.

## Blockers
* Massive discrepancy between UI state and backend capability.

## Stability
* **Stable**: PostgreSQL link CRUD, basic frontend shell.
* **Experimental**: Disconnected business logic modules.
* **Risky**: Highly decoupled frontend that expects complex API contracts that don't exist.
