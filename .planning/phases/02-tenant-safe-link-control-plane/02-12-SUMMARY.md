---
phase: 02-tenant-safe-link-control-plane
plan: "12"
subsystem: links
tags: [go, postgres, idempotency, custom-keys, nextjs, tdd]
requires:
  - phase: 02-11
    provides: Network-free destination policy and consumed scoped form/detail components
provides:
  - ASCII-normalized custom keys with global collision protection and safe feedback
  - Canonical custom-key retries with fresh authorization and atomic link/ledger effects
  - Actual browser normalization, collision, changed-submission and uncertain-retry proof
affects: [02-13, 02-14, 02-27, 02-30, 02-33, 02-35]
tech-stack:
  added: []
  patterns: [closed canonical error-code projection, forward key constraint migration, distinct-actor collision proof]
key-files:
  created: [apps/backend/internal/database/migrations/006_custom_link_keys.sql]
  modified:
    - apps/backend/internal/service/link.go
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/handler/product.go
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/components/link-form.tsx
    - apps/frontend/lib/api.ts
    - apps/frontend/tests/control-plane.spec.ts
    - packages/zod/src/identity.ts
    - packages/openapi/src/contracts/identity.ts
    - packages/openapi/openapi.json
    - apps/backend/static/openapi.json
    - apps/backend/internal/transport/health.gen.go
key-decisions:
  - Validate ASCII before case folding; normalize custom keys before hashing and enforce global host/key uniqueness including deleted rows.
  - Omit empty customKey from canonical hashes to preserve existing generated-key retry compatibility.
  - Project only three recognized canonical error codes with matching status through a bounded browser error reader; never render response messages.
requirements-completed: []
requirements-addressed: [LINK-01, LINK-02, LINK-03, LINK-07, LINK-08, TEN-05, TEN-07, TEN-08, SAFE-03, SAFE-04]
duration: 15min
completed: 2026-10-09
---

# Phase 2 Plan 12: Custom Keys and Safe Retries Summary

**Managed-domain creation accepts normalized custom keys, protects global uniqueness and replays one committed outcome after fresh authorization, with actual browser collision and uncertain-retry feedback.**

## Performance

- Duration: approximately 15 minutes, starting around 11:52 UTC and completing implementation at 12:06 UTC on 2026-10-09.
- Tasks: 2/2.
- Implementation/test paths: 13, including one forward migration and all three generated contract outputs. Summary and tracking metadata are additional documentation paths.

## Accomplishments

- Optional `customKey` accepts only 3–64 ASCII letters/digits/hyphens/underscores, starting with a letter or digit. ASCII validation precedes lowercasing; slash, percent escapes, spaces, Unicode lookalikes, short/long keys and all eleven reserved system paths deny. Browser ASCII uppercase visibly becomes lowercase while invalid characters remain visible for correction. Hostname remains server-selected.
- The existing workspace→actor→resource transaction freshly requires write authorization before replay. Normalized custom key participates in SHA-256 canonical request hashing. Same key and canonical hash returns the exact committed snapshot; changed content returns safe 409 `REQUEST_REUSE_CONFLICT`. Empty custom key is omitted from the hash to preserve the earlier generated-key canonical payload. Link and ledger remain one transaction with at least 24-hour retention.
- Global managed-host/key uniqueness remains authoritative across workspaces and deleted records. Custom-key insertion never retries a collision: only the exact named unique violation maps to safe 409 `KEY_UNAVAILABLE`, with no owner hint. Generated entropy uses the unchanged secure source and bounded five-attempt named-constraint-only savepoint retry.
- Registered real-PG actual HTTP tests prove normalized same-hash replay, changed-hash rejection, nineteen invalid/reserved examples, independent durable actors concurrently competing across two workspaces with exactly one winner, four concurrent identical requests with one link and one ledger record, deferred commit failure with neither effect nor ledger, and revoked-actor replay denial. Existing generated entropy/collision, unrelated-constraint, rollback, destination-egress, foreign detail, viewer and provenance tests remain intact and passed.
- Existing scoped form and detail routes remain consumed. Identical uncertain retry retains the precise pending payload and idempotency key; deliberate changed submission rotates the key. Success renders only after committed response/replay. Fixed feedback comes from recognized canonical constraint codes, never raw response messages; the error reader is bounded to 8 KiB within the existing request deadline and cancellation boundary.

## Task Commits

1. Task 1 — Specify custom key normalization and safe retries: `d1b8f8b` (`test`).
2. Task 2 — Deliver normalized custom keys and safe committed retries: `eb7751a` (`feat`).

## Verification

- Meaningful Go RED: `TestProductActualHTTP/custom-key` received 400 instead of expected committed 201 for the previously absent `customKey` field.
- Actual browser RED: initial root gate failed before UI implementation. A sanitized diagnostic repeated the RED-commit form revision `d1b8f8b` by temporarily preserving/restoring only that owned file; **1 failed, 0 skipped, 0 flaky, 0 infrastructure errors**, with `getByLabel("Custom short key")` failing because the field was absent. This diagnostic is explicitly baseline-form evidence, not a claim about the finished form.
- `bun run test:unit`: **92 discovered race-enabled Go top-level unit tests**, all four workspace suites, script suites and tool self-tests passed.
- `bun run test:integration`: **17 registered top-level tests across six groups** passed. Final run includes distinct-actor global uniqueness and custom rollback/revocation proof. The existing telemetry transient did not recur and is not claimed repaired.
- Targeted actual browser: `bun run test:e2e -- --project=local --grep custom-key` — **2 completed**, zero skipped/flaky/infrastructure errors.
- Full actual browser regression: `bun run test:e2e -- --project=local` — **25 completed**, zero skipped/flaky/infrastructure errors. Invalid characters remain visible; actual server collision renders exact approved copy; changed submissions use fresh keys; lost committed response retries identical payload/key and displays exact committed short URL.
- Root format, lint and typecheck passed. `bun run generate:check` passed after jointly regenerating authored/served OpenAPI and Go transport outputs. `bun run migrate:check` passed exact latest version 6 and existing prefix upgrades.
- All role/workspace production builds passed after browser execution. Existing exported `cleanFrontendBuild` validated and removed only the production frontend output before `CI=true bun run scan`. Bun dependencies, all-role/test imported Go package vulnerability checks, complete worktree secrets and full-history secrets passed. Unused OpenPGP inventory advisory `GO-2026-5932` remains visible; no scanner exceptions or gate changes.

## Decisions Made

Implement the actual plan12 custom-key objective; the prior summary's reference to guarded transport ownership was stale. No destination fetching was introduced. Preserve the PostgreSQL global uniqueness invariant and safe bounded transaction behavior. The forward CHECK amendment follows PostgreSQL [constraint guidance](https://www.postgresql.org/docs/17/ddl-constraints.html); its down migration validates existing data and fails rather than deleting incompatible custom-key links.

## Deviations from Plan

1. **[Rule 3 - Blocking] Forward custom-key constraint migration.** Existing migration005 restricted `short_key` to generated 20-character base32, causing custom insert 503. Added `apps/backend/internal/database/migrations/006_custom_link_keys.sql`, broadening the existing column CHECK to approved lowercase ASCII3–64 and denying reserved paths. Existing uniqueness, provenance and deleted-row reservation remain. Down migration fails safely when custom rows prevent the old CHECK. Existing migration assertions derive the exact latest count and needed no changes. Verified real-PG HTTP and migration gates. Commit: `eb7751a`.
2. **[Rule 3 - Blocking] Recognized frontend constraint projection.** Existing `apps/frontend/lib/api.ts` discarded all non-success response codes. Extended that otherwise undeclared path to parse bounded canonical error envelopes and expose only matching-status `KEY_UNAVAILABLE`, `REQUEST_REUSE_CONFLICT` and `INVALID_CUSTOM_KEY`. Existing token, timeout, abort, fixed-origin, no-cookie and no-raw-message behavior remains. Verified frontend suites, typecheck and actual browser collision. Commit: `eb7751a`.

Exactly 13 implementation/test paths changed: eleven declared paths plus those two necessary wiring paths. Declared `repository/mutation.go`, `components/link-detail.tsx` and `lib/links.ts` required no change. No undeclared aggregate, runtime pin, gate or handwritten transport was introduced.

## Issues Encountered

New lint findings (transaction complexity, test shadows and reserved-path spelling counting against an unrelated role literal) were resolved inside owned code without modifying gates or role policy. No remaining blocker or authentication gate.

## Known Stubs and Scope Limits

No goal-blocking stub exists in changed paths. Redirect/analytics availability notice remains truthful. Later library/team/lifecycle and live configured-provider acceptance remain pending their owning plans, including plan35. No whole requirement is marked complete by this bounded slice. No editing, custom domains, redirect implementation, billing, analytics or destination fetch was added.

## Self-Check: PASSED

Created forward migration exists; both task commits resolve in history. Exactly 13 implementation/test paths are present in the task commits. All required verification passed before the GREEN commit. No tracked files were deleted, and no generated/runtime file remains untracked.
