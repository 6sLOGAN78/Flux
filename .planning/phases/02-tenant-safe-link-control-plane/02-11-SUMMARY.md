---
phase: 02-tenant-safe-link-control-plane
plan: "11"
subsystem: links
tags: [go, url-safety, idna, postgres, nextjs, security, tdd]
requires:
  - phase: 02-10
    provides: Atomic scoped link creation, canonical contracts and real browser detail
provides:
  - Extracted destination policy with reserved IP and parser adversarial denials
  - API-only fail-closed operator host configuration and actual environment sample
  - Real-PG creation proof with live DNS/HTTP egress counters and injected entropy failures
  - Consumed link form/detail components with exact feedback and safe external destinations
affects: [02-12, 02-14, 02-27, 02-30, 02-35]
tech-stack:
  added: []
  patterns: [network-free destination validation, positive-control egress instrumentation, io.Reader entropy injection, scoped component extraction]
key-files:
  created:
    - apps/backend/internal/safety/url.go
    - apps/backend/internal/safety/url_test.go
    - apps/backend/internal/service/link_test.go
    - apps/frontend/components/link-form.tsx
    - apps/frontend/components/link-detail.tsx
    - apps/frontend/lib/links.ts
    - apps/frontend/lib/links.test.ts
  modified:
    - apps/backend/.env.sample
    - apps/backend/internal/config/config.go
    - apps/backend/internal/config/config_test.go
    - apps/backend/internal/service/link.go
    - apps/backend/internal/repository/link.go
    - apps/backend/internal/handler/product_test.go
    - apps/frontend/app/workspaces/[workspaceId]/links/new/page.tsx
    - apps/frontend/app/workspaces/[workspaceId]/links/[linkId]/page.tsx
    - apps/frontend/tests/control-plane.spec.ts
    - docs/development.md
key-decisions:
  - Preserve network-free creation; reject special IPv4 and IPv6 literals conservatively without resolving hostnames.
  - Preserve the existing repository constructor with crypto/rand.Reader and inject io.Reader only through explicit constructors for failure verification.
  - Wire extracted components into the existing scoped routes and retain canonical DTOs, dirty confirmation and cancellation boundaries.
requirements-completed: []
requirements-addressed: [LINK-01, LINK-02, LINK-03, LINK-07, LINK-08, TEN-05, TEN-07, TEN-08, SAFE-03, SAFE-04]
duration: 25min
completed: 2026-10-09
---

# Phase 2 Plan 11: Adversarial Destination Policy Summary

**Managed-domain creation rejects reserved destinations without DNS or HTTP egress, while real scoped form/detail components preserve retries, escaped content and safe destination links.**

## Performance

- **Duration:** Approximately 25 minutes, starting around 11:24 UTC and completing implementation at 11:48:42 UTC on 2026-10-09.
- **Tasks:** 2/2
- **Files changed:** 17 implementation/test/documentation paths, plus this summary and tracking metadata.

## Accomplishments

- Moved the existing creation policy into `internal/safety/url.go` and kept `LinkService.Create` wired to it. The 78-case adversarial corpus rejects credentials, opaque/relative/non-HTTP URLs, escaped host/control/backslash/parser ambiguities, invalid ports, numeric host tricks, private/local/metadata addresses, mapped addresses, multicast and special/reserved ranges. IPv6 must belong to the global-unicast allocation and avoid special ranges; deprecated site-local, relay, documentation and other special ranges deny. Canonical positive cases preserve encoded path/query/fragment meaning, IDNA hostname normalization, ports and public IPs.
- API startup continues to require an operator-managed hostname with no default. Local/internal operator hostnames now fail validation. Tests exercise the actual dot-nested environment loader, comma-separated blocked names, IDNA normalization, malformed entries and other roles' independence. The existing environment callback remains unchanged and authoritative; production startup normalizes once.
- The registered `TestProductActualHTTP/link-create` surrounds real signed-provider HTTP requests, PostgreSQL effects and the destination corpus with process resolver/default-HTTP observations. Positive controls first prove both instruments are live; accepted public DNS, IPv4 and IPv6 destinations then create links with **zero destination DNS and HTTP calls**. Exact local API/provider fixture traffic passes separately. Globals restore after the sequential subtest, and the race-enabled integration gate passed. Source inspection confirms no custom resolver, destination HTTP client or dialer in creation.
- An explicit standard `io.Reader` seam preserves the existing repository constructor and production `crypto/rand.Reader`. Actual HTTP tests prove deterministic collision retry, EOF/short-entropy failure, five-attempt exhaustion, safe 503 responses and no link or idempotency-ledger effects on failure. Existing named-constraint-only retries, transaction rollback, fresh membership before replay, foreign 404, viewer denial and durable creator projection tests passed unchanged.
- The existing new/detail pages now consume the declared components. They preserve canonical response validation, workspace generation/abort handling, dirty-draft confirmation and disposal. The server-owned hostname stays immutable; invalid and blocked destination messages match approved copy. Detail escapes values, exposes an HTTP(S)-only external destination with `noreferrer noopener`, retains copy fallback, local full timestamps and the management availability notice. Browser loss of a committed create response retries the identical key and payload and announces success only after confirmed replay.

## Task Commits

1. **Task 1: Specify adversarial destination and operator denials** — `3e24466` (`test`)
2. **Task 2: Deliver policy and consumed safe UI** — `acbae3b` (`feat`)

## Verification

- Meaningful Go RED: the existing policy accepted six newly specified reserved/site-local/IPv6 cases; operator local-domain validation also failed as expected before production changes.
- Actual browser RED: the named destination denial test failed its exact-feedback assertion; **1 failed, 0 skipped, 0 flaky, 0 infrastructure errors**. A bounded private capture retained no raw report in the repository.
- `bun run test:unit`: **92 discovered race-enabled Go top-level unit tests**, all four workspace suites, all four script test files and tool self-tests passed.
- `bun run test:integration`: **17 registered top-level tests across six groups** passed, including the real-PG egress, entropy, transaction, membership and collision proof.
- Final targeted `bun run test:e2e -- --project=local --grep destination`: **2 completed**, no skips, flaky cases or infrastructure errors.
- Final full browser suite: **24 completed**, no skips, flaky cases or infrastructure errors. Existing switch, multi-tab invalidation, draft disposal, creator/detail and provider-fixture behavior remained covered.
- Root format, lint, typecheck and generation checks passed. Generated canonical/served OpenAPI and Go transport bytes remain unchanged. All role/workspace production builds passed after browser execution.
- Validated production Next build cleanup preceded `CI=true bun run scan`. Bun dependencies, imported Go packages across all roles/tests, complete worktree secrets and full-history secrets scans passed. The existing unused `GO-2026-5932` OpenPGP inventory advisory remains visible; there are no scanner exceptions or weakened rules.

## Decisions Made

URL creation remains a pure parsing/validation operation. Syntactic hostname acceptance does not claim that DNS answers are public. Guarded outbound processing remains separately owned by plan 12. Reused the locked `golang.org/x/net` IDNA implementation without upgrades; checked its [v0.60.0 documentation](https://pkg.go.dev/golang.org/x/net@v0.60.0/idna) and the [IPv4](https://www.iana.org/assignments/iana-ipv4-special-registry/) and [IPv6](https://www.iana.org/assignments/iana-ipv6-special-registry/) special-purpose registries.

## Deviations from Plan

1. **[Rule 3 - Blocking] Consume extracted components through the actual routes.** Updated the two existing new/detail page wrappers outside the declared list; otherwise the requested extracted components would be unreachable and untested. Existing scope, dirty confirmation and denial behavior remain in place. Verification: full 24-case browser suite. Commit: `acbae3b`.
2. **[Rule 3 - Blocking] Use the actual backend environment sample.** The plan listed nonexistent `apps/backend/internal/.env.sample`; updated existing `apps/backend/.env.sample` and documented that real path instead. No duplicate sample remains. Verification: actual environment-loader tests and source inspection. Commit: `acbae3b`.
3. **[Rule 3 - Blocking] Add the explicit standard entropy seam.** `internal/repository/link.go` was outside the declared list, but injection was required for the plan's executable entropy/collision proof. Constructor compatibility, secure production entropy, lock order and bounded named-constraint behavior are preserved. Verification: actual HTTP real-PG tests and race-enabled integration. Commit: `acbae3b`.
4. **[Rule 1 - Bug] Preserve distinct destination text after adding its safe anchor.** Full browser diagnostic found only the previous exact destination-text visibility assertion failing (23 passed, 1 failed; zero infrastructure errors). Wrapped the original destination in a text element; did not alter the prior test. Verification: full rerun passed all 24. Commit: `acbae3b`.

## Issues Encountered

Initial local compile/type/lint findings in the new tests were corrected before final passing gates. No remaining blocker or authentication gate. The pre-existing telemetry transient in the phase deferred record did not recur in this plan's integration run and is not claimed repaired.

## Known Stubs and Scope Limits

No goal-blocking stub exists in the changed paths. The explicit redirects/analytics availability notice remains truthful. The library first-link CTA remains staged until plan 14; Team operations remain plan 16. Custom keys/domains, editing, redirects, analytics, billing and destination fetching are outside this plan. Broad requirements and live-provider factor acceptance remain pending subsequent plans and plan 35; no whole requirement is marked complete by this slice.

## Self-Check: PASSED

All seven created implementation/test files exist. Both task commits resolve in repository history. Exactly 17 implementation/test/documentation paths changed from the plan's baseline. Final verification passed before the GREEN commit; no tracked file was deleted.
