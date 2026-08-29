# Changelog

## 2026-08-29
**Type:** DOCUMENTATION
**Task:** TASK-000 (Initial Audit)
**Change:** Created project control system (`PROJECT_STATE.md`, `ROADMAP.md`, `FEATURE_REGISTRY.md`, etc.).
**Reason:** To establish a source of truth for the actual state of the repository vs. the claimed state in README.

## 2026-08-29 (Part 2)
**Type:** FEATURE
**Task:** TASK-005 (Implement Authentication)
**Change:** Built native PostgreSQL `users` table, native bcrypt and JWT generation. Created `/auth/signup` and `/auth/login` APIs. Rewrote frontend `AuthContext` to submit credentials securely.
**Reason:** To replace the dangerous frontend mock-auth bypass with a real production-grade authentication flow.

### Phase 8: Clerk Authentication & Workspace Migration
* **Date:** 2026-08-29
* **Feature:** FEAT-004
* **Change:** Stripped out native authentication (bcrypt, local JWT, `/auth` endpoints). Configured Clerk as the single source of truth for Users and Organizations (Workspaces).
* **Change:** Backend now relies on Clerk SDK to verify JWTs, and uses a Just-In-Time (JIT) mapping in `ClerkJWTMiddleware` to map `clerk_user_id` and `clerk_org_id` into native Postgres UUIDs (`users` and `workspaces` tables).
