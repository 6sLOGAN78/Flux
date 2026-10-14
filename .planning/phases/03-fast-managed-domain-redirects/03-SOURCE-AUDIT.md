# Phase 3 source coverage

Planning audit2026-10-14; coverage describes executable plans, not shipped behavior. There is no Phase3 CONTEXT and no invented user decision IDs. ROADMAP/REQUIREMENTS are scope authority; broader spec capabilities remain assigned to later phases.

| Requirement | Delivering plans | Required observable evidence |
|---|---|---|
| LINK-04 | 03-02,03-03,03-05,03-08,03-11 | Rendered destination/title edit, current membership, canonical version, immutable retry, actual updated public redirect and transactional invalidation. |
| REDIR-01 | 03-01,03-02,03-11 | Actual role GET/HEAD302 with stored target, no-store, query/fragment policy and no auto-follow in client. |
| REDIR-02 | 03-04,03-08,03-10 | Real Redis candidate hit, mandatory current-primary guard, SQL fallback, normalized identity and cache-state classification. |
| REDIR-03 | 03-05,03-06,03-08 | Every mutation atomically emits; real relay/fences, rollback/replay, multiple claimers, crash/reclaim and reversed messages. |
| REDIR-04 | 03-01,03-04,03-05,03-08 | Every unavailable/unsafe state warm/cold, poisoned cache, stale fill and old cache restoration; no prior target after current authority denies it. |
| REDIR-05 | 03-01,03-04,03-06,03-07,03-08,03-09 | Real redirect with absent email/analytics/worker/webhook/exporter; bounded optional Redis outage; single-deadline cleanup. Primary outage explicitly503. |
| REDIR-06 | 03-10,03-11 | Actual role, pinned real dependencies, documented load/concurrency/environment; >=1000 valid samples/profile and conservative p95<30/<100ms with guard enabled. |
| REDIR-07 | 03-09,03-10 | All required fixed-label metrics survive app+collector filters, no planted private data, actual server latency buckets and honest not_configured handoff. |
| REDIR-08 | 03-01,03-07,03-11 | Explicit direct extractor, exact trusted-CIDR XFF policy, spoof/oversize/mapped-IP tests and deployed proxy Host instructions. |

Roadmap criterion1 ->03-01/04/08/11; criterion2 ->03-02/04/05/06/08; criterion3 ->03-06/07/08/09; criterion4 ->03-09/10/11.

Research resolution: strict primary status/version guard resolves the conflict between TTL-only bounded caching and literal REDIR-04. A cache hit still performs a narrow primary read. If that fails the performance target, measure/optimize without deleting the correctness check; any architecture change requires a revised plan and explicit requirement reconciliation. No five-second permission to serve a disabled destination is implied.

UI contract coverage: initial form and save03-02; unknown results/conflict/access loss/focus/mobile03-03; availability helper and complete journey03-11. No navigation to analytics or click-delivery claims. No custom domains, advanced routing, click facts/loss SLO, visitor identity, billing or conversion attribution is silently implemented here; those are outside the phase's nine requirements.

Execution prerequisites: Phase2 remains34/35 with unfinished real provider acceptance. No03-SUMMARY or VERIFICATION passed record exists. Local fixtures do not close this dependency. Phase3 plans are available for review while execution waits.
