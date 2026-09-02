# Roadmap & Phases

**Current Phase:** Phase 14 — Billing

## Phase 0 — Repository Stabilization
**Status**: COMPLETE
**Objective**: Audit the repository, document the actual state versus the theoretical state, and set up tracking.

## Phase 1 — Core Architecture Wiring
- TASK-005: Implement Authentication (FEAT-015, BUG-005)
**Status**: IN_PROGRESS
**Objective**: Connect the existing but decoupled backend components (Redis, ClickHouse, Base62).
**Tasks**:
- TASK-001: Wire Redis to `RedirectService`.
- TASK-002: Implement click event publishing to ClickHouse on redirect.
- TASK-003: Connect `AnalyticsProvider` to ClickHouse to serve `/api/v1/analytics/summary`.
- TASK-004: Replace random shortcode generator with `pkg/base62`.

## Phase 2 — Database Expansion
**Status**: NOT_STARTED
**Objective**: Create PostgreSQL schemas for the missing core entities.
**Tasks**:
- Create users and tenants schema.
- Create custom_domains schema.
- Create campaigns schema.

## Phase 3 — Backend API Implementation
**Status**: NOT_STARTED
**Objective**: Implement the missing OpenAPI endpoints in `router/v1/v1.go` using the disconnected modules in `internal/modules`.
**Tasks**:
- Wire Workspaces API.
- Wire Domains API.
- Wire Campaigns API.

## Phase 4 — Frontend Integration
**Status**: NOT_STARTED
**Objective**: Replace mocked React state with real API calls.
**Tasks**:
- Connect CampaignsPage to real API.
- Connect Analytics pages to real API.

## Phase 5 — Advanced Features
**Status**: NOT_STARTED
**Objective**: Implement AB Testing, Smart Routing, and AI Insights.
