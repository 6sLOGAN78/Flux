# Architecture Decisions

## DEC-001
**Date:** 2026-08-29
**Decision:** Focus on wiring existing foundation before expanding features.
**Context:** The repository presents a massive surface area, but core features (Redis caching, ClickHouse analytics, Postgres schemas for tenants/users) are missing or passed as `nil`.
**Chosen Approach:** Do not build new UI or modules. First, connect the existing backend modules (Redis cache, ClickHouse ingestion) to the live request lifecycle.
**Affected Components:** `server.go`, `redirect.go`, `links.go`

## DEC-002
**Date:** 2026-08-29
**Decision:** Implement Native JWT Authentication.
**Context:** The codebase has Clerk support but defaults to a fake local auth if the Clerk key is missing. The `AuthService` already implements JWT and bcrypt.
**Chosen Approach:** Build a native `users` table and `/auth/signup`, `/auth/login` endpoints. Use the existing JWT middleware to protect routes.
**Affected Components:** `server.go`, `router.go`, `v1.go`, `SignInPage.tsx`, `SignUpPage.tsx`, `AuthContext.tsx`.

## DEC-003: Migrate Auth & Workspaces to Clerk
* **Date:** 2026-08-29
* **Status:** Implemented
* **Decision:** Replaced native JWT authentication and RBAC with Clerk. Clerk is now the single source of truth for Users, Organizations, Roles, and Sessions.
* **Reasoning:** Reduces security overhead, offloads identity management, and simplifies multi-tenant organization handling.
* **Implementation:** Backend uses JIT (Just-In-Time) sync via `ClerkJWTMiddleware` to map `clerk_user_id` and `clerk_org_id` to local PostgreSQL UUIDs (`users` and `workspaces` tables) for foreign key relations.
