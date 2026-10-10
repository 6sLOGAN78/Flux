# Phase 2 revised coverage and bounded execution

All four mandatory artifact types and approved UI are audited. Every row is COVERED; exclusions below are not gaps.

| SOURCE | ID | Complete feature/protocol | Plans | Status |
|---|---|---|---|---|
| GOAL | Phase2 | Authenticated users manage teams/managed-domain links inside enforceable workspace boundaries | 01–35 | COVERED |
| GOAL | SC1 | Sign-in/atomic owner/authorized switch and foreign-record denial | 01,02,03,04,05,06,07,08,09,17,18,30,31,32,33,34 | COVERED |
| GOAL | SC2 | Invite/accept/list roles/removal/owner preservation | 16,17,18,19,20,21,22,23,24,25,26,30,31,32,33,34 | COVERED |
| GOAL | SC3 | Generated/custom keys/unsafe URL/collision/retry/concurrency | 10,11,12,27,28,30,31,32,33 | COVERED |
| GOAL | SC4 | Inspect/search/filter/paginate/lifecycle/suspend/audit | 10,11,12,13,14,15,27,28,29,30,31,32,33 | COVERED |
| GOAL | SC5 | Session/CSRF/SQL/transactions/egress/two-workspace proof | 01–35 | COVERED |
| REQ | TEN-01 | Full exact REQUIREMENTS.md capability | 01,02,03,04,05,30,31,32,33,34,35 | COVERED |
| REQ | TEN-02 | Full exact REQUIREMENTS.md capability | 06,07,30,31,32,33,34,35 | COVERED |
| REQ | TEN-03 | Full exact REQUIREMENTS.md capability | 19,20,21,22,23,24,25,26,30,31,32,33,34,35 | COVERED |
| REQ | TEN-04 | Full exact REQUIREMENTS.md capability | 16,17,18,24,25,26,30,31,32,33 | COVERED |
| REQ | TEN-05 | Full exact REQUIREMENTS.md capability | 06,07,10,11,12,16,17,18,19,20,24,25,26,27,28,29,30,31,32,33 | COVERED |
| REQ | TEN-06 | Full exact REQUIREMENTS.md capability | 08,09,13,14,15,17,18,24,25,26,30,31,32,33,34,35 | COVERED |
| REQ | TEN-07 | Full exact REQUIREMENTS.md capability | 06,07,08,09,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33 | COVERED |
| REQ | TEN-08 | Full exact REQUIREMENTS.md capability | 06,07,08,09,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33 | COVERED |
| REQ | LINK-01 | Full exact REQUIREMENTS.md capability | 10,11,12,30,31,32,33 | COVERED |
| REQ | LINK-02 | Full exact REQUIREMENTS.md capability | 10,11,12,30,31,32,33 | COVERED |
| REQ | LINK-03 | Full exact REQUIREMENTS.md capability | 10,11,12,13,14,15,30,31,32,33 | COVERED |
| REQ | LINK-05 | Full exact REQUIREMENTS.md capability | 27,28,29,30,31,32,33 | COVERED |
| REQ | LINK-06 | Full exact REQUIREMENTS.md capability | 13,14,15,30,31,32,33 | COVERED |
| REQ | LINK-07 | Full exact REQUIREMENTS.md capability | 10,11,12,27,28,29,30,31,32,33 | COVERED |
| REQ | LINK-08 | Full exact REQUIREMENTS.md capability | 10,11,12,30,31,32,33 | COVERED |
| REQ | LINK-09 | Full exact REQUIREMENTS.md capability | 29,30,31,32,33 | COVERED |
| REQ | SAFE-02 | Full exact REQUIREMENTS.md capability | 01,02,03,04,05,08,09,24,25,26,30,31,32,33,34,35 | COVERED |
| REQ | SAFE-03 | Full exact REQUIREMENTS.md capability | 01,02,03,04,05,06,07,08,09,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33 | COVERED |
| REQ | SAFE-04 | Full exact REQUIREMENTS.md capability | 10,11,12,21,22,23,30,31,32,33,34,35 | COVERED |
| CONTEXT | D-01 | Password+verification/recovery+Google/GitHub actual provider | 01–05,34–35 | COVERED |
| CONTEXT | D-02 | Explicit named atomic owner; invite acceptance no unrelated workspace | 06–07,24,35 | COVERED |
| CONTEXT | D-03 | Current authorized preference/chooser/switch/tab cancellation | 08–09,18,30,35 | COVERED |
| CONTEXT | D-04 | First Links/Create your first link/invites optional/no redirect claims | 06,10,35 | COVERED |
| RESEARCH | AUTH | SDK exact issuer/azp/expiry/session/JWKS; unique durable user; no globals/org claims | 02–05,35 | COVERED |
| RESEARCH | CSRF | Explicitbearer/JSON/Origin/cross-site/preflight/no-store | 04,30,35 | COVERED |
| RESEARCH | SESSION | Seven-day max/refresh/recovery/signout/rotation/cookies/quota evidence | 04,34–35 | COVERED |
| RESEARCH | SQL | ParameterizedScope/compositeFK/transaction/shared vs exclusive lock order | 06–30 | COVERED |
| RESEARCH | OWNER | Four roles/admin target restrictions/locked final owner race | 07,16–18 | COVERED |
| RESEARCH | CREATOR | Durable users FK; membership deletion retains active/deleted links/audit provenance | 06,10,18–19 | COVERED |
| RESEARCH | KEY | 12random bytes/base32/ASCII3–64/systemreserved/globalhost-key/softdelete reservation | 10–12,27–28 | COVERED |
| RESEARCH | LEDGER | Bootstrap identity/tenant actor operation SHA256/24h/fresh replay authority | 06–07,10,12,17–19,24–29 | COVERED |
| RESEARCH | PAGINATION | 25/max100/timestampUUID seek/tenant-filter signed cursor/query escaping | 13–15 | COVERED |
| RESEARCH | URL | URL corpus/IDNA/publicHTTP(S)/metadata/private/mappedIP/blocklist/no fetch | 10–11 | COVERED |
| RESEARCH | EGRESS | No userURL fetch; fixed provider transports; guard any introduced processing | 04,11,21–22,35 | COVERED |
| RESEARCH | INVITE | Seven-day/digest256bit/verified recipient/explicitaccept/repeat role preservation | 19–20,24–26 | COVERED |
| RESEARCH | ENCRYPT | Standard AES256GCMrandomnonce/AAD/keyID/rotation/nonce bound/no plaintext | 19–23,35 | COVERED |
| RESEARCH | DELIVERY | DurablePG/realRedis scoped reread/leasegeneration/ack/idempotency/crash/outage/replay | 21–23,25,30,35 | COVERED |
| RESEARCH | ROLES | Independent worker ownsPG/readiness/drain/cleanup/producer independence | 21–23 | COVERED |
| RESEARCH | LIFECYCLE | All edges/version428/409/softdelete/restoreDisabled/independentrestriction/audit | 27–29 | COVERED |
| RESEARCH | PACKAGES | Exact approveddeps/preserved React/pins/frozen install/scans/no unaudited scaffolder | 01,33 | COVERED |
| RESEARCH | GENERATION | Authored canonical types/errors/headers/bearer and deterministic three outputs/template | 02–03,05–06,08,10,12–13,15–19,21,24–25,27,29,33 | COVERED |
| RESEARCH | GATES | RunnerbeforeRED/nonzeroBun+explicitPlaywrightJSON/classifiedGo/race/oldgates | 01,03,05,21,23,30,33–35 | COVERED |
| RESEARCH | MIGRATION | Forward001→latest and preserved role/outage latestversion assertions | 03,05–07 | COVERED |
| RESEARCH | PRIVACY | Protected audit/creator provenance; no names/email/url/reasons/tokens telemetry | 04,10–12,19–30,35 | COVERED |
| RESEARCH | PROVIDER | Independent implementation first; named config/evidence commands; no absentconfig skip | 34–35 | COVERED |
| UI | NATIVE | Original native CSS/tokens/fonts/type/colorcontrast/no registries/new dependencies | 01,04,31–32 | COVERED |
| UI | AUTH | Provider-native factors and exact safe wrapper headings/copy | 01,04,34–35 | COVERED |
| UI | ONBOARD | Required name/idempotency/Links empty/invited alternative | 06–07,24 | COVERED |
| UI | SWITCH | Chooser/authorizedselection/unsaveddialog/late-response/tab/focus behavior | 08–09 | COVERED |
| UI | LINK | Create/detail/copy/timezone/fullaccessible values/immutablehostname/no edits | 10–12 | COVERED |
| UI | LIBRARY | Responsive tablecards/search/closedfilter/NextPrevious/emptyloadingerrors | 13–15,32 | COVERED |
| UI | TEAM | Role options/finalownerhelper/invitation states/secret-free explicit acceptance | 16–26 | COVERED |
| UI | STATE | Native destructiveconfirmation/conflictreview/suspensionprecedence/cleartoDisabled | 27–29 | COVERED |
| UI | A11Y | Keyboard/forms/dialogfocus/announcements/loading/reducedmotion/forcedcolors/320/zoom | 31–32 | COVERED |
| UI | HONESTY | No analytics/redirect testing/compulsory team/customdomain/billing/enterprise paths | 06–35 | COVERED |

## Execution outline and ownership
| Wave/plan | Reachable core or refinement | Unique files | Autonomous |
|---|---|---|---|
| 01 | Run real provider sign-in with a working browser gate | 14 | yes |
| 02 | Verify bearer sessions through the authenticated Go boundary | 14 | yes |
| 03 | Persist unique provider identities in PostgreSQL | 14 | yes |
| 04 | Protect browser mutations and recover safely from session failures | 14 | yes |
| 05 | Exercise the production identity path with deterministic generation and real fixtures | 14 | yes |
| 06 | Create the named workspace and owner atomically | 14 | yes |
| 07 | Validate workspace retries, tenant constraints and role capabilities | 14 | yes |
| 08 | Restore only a current member workspace | 14 | yes |
| 09 | Switch workspaces without stale data across tabs | 9 | yes |
| 10 | Create a safe generated-key managed-domain link | 14 | yes |
| 11 | Configure managed domains and reject adversarial destinations | 14 | yes |
| 12 | Request normalized custom keys and retry uncertain creates safely | 14 | yes |
| 13 | Show the authorized newest-first Links library | 14 | yes |
| 14 | Navigate signed tenant-bound cursor pages | 12 | yes |
| 15 | Search and filter links with scoped parameterized queries | 12 | yes |
| 16 | Inspect Team through server capabilities | 13 | yes |
| 17 | Promote and demote roles while preserving the final owner | 14 | yes |
| 18 | Remove members without losing creator provenance or access isolation | 14 | yes |
| 19 | Queue a tenant invitation with encrypted durable delivery intent | 14 | yes |
| 20 | Expose invitation status and strengthen encryption/retry boundaries | 11 | yes |
| 21 | Deliver the durable invitation through the independent worker | 14 | yes |
| 22 | Reject forged worker envelopes and recover delivery across failures | 10 | yes |
| 23 | Keep invitation worker readiness and shutdown honest | 9 | yes |
| 24 | Inspect and explicitly accept a verified-recipient invitation | 13 | yes |
| 25 | Resend or revoke invitations without resurrecting old credentials | 13 | yes |
| 26 | Protect invitation secrets through browser authentication and recovery | 7 | yes |
| 27 | Apply versioned link lifecycle transitions | 14 | yes |
| 28 | Review stale lifecycle conflicts and safely retry uncertain mutations | 8 | yes |
| 29 | Suspend and clear abuse restrictions with immutable provenance | 13 | yes |
| 30 | Verify two-workspace denial across every delivered workflow | 8 | yes |
| 31 | Make forms and dialogs keyboard accessible with truthful feedback | 11 | yes |
| 32 | Reflow the control plane on small screens and under zoom | 8 | yes |
| 33 | Run complete frozen builds, generation and security gates | 11 | yes |
| 34 | Prepare executable private provider prerequisite and evidence validators | 11 | yes |
|35|Actual configured provider factors/session/delivery acceptance|11|no: unavailable prerequisites/provider-controlled evidence|

All plans have fewer than15 unique concrete files, including generated outputs/tests/manifests; no glob or directory aliases conceal mutations. Each task owns a subset of its plan. Shared canonical authority, generated outputs, handler adapters and browser specifications force explicit serial dependency/waves; concurrent mutations of those paths are prohibited.

First genuine capability is provider sign-in through real UI, with a working root runner BEFORE RED tests. Then Go bearer verification and durable identity→workspace→link/team slices extend the reachable path. Major logic/first schema use meaningful actualPG HTTP tests. Later validation/security/recovery/accessibility refinements extend already delivered capabilities, never defer an entire API/UI layer or weaken source scope. No intermediate deployment/phase completion occurs with required security/refinements pending.

Authored control-plane schemas/routes consolidate incrementally in packages/zod/src/identity.ts and packages/openapi/src/contracts/identity.ts, already exported by their aggregates. This is explicit file ownership under delegated discretion, not a hidden path; later feature plans own both authored modules and all three generated outputs. Existing runtime/product scaffolds are extended without rewriting functioning foundation code. Thin ProductHandler adapts Echo and narrow service/repository constructors; protected persistence tests reuse registered TestProductActualHTTP subtests. New independent top-level container suites are registered in their owning plans.

## Checker findings resolved
1. Root package.json test:e2e and actual runner/parsing/forwarding exist in01Task1 BEFORE first RED. Root provider:check/provider:acceptance:check/test:e2e:provider are declared in34 and retained in35.
2. Link creator, audit actor and invitation provenance reference durable users, never deletable workspace_memberships.18 proves removal with active AND deleted links succeeds, creator/key/links persist and removed actor loses all access/replay.
3.35 bounded plans replace the14 oversized prompts; maximum14 actual unique paths; complete security/validation/recovery/UI/refinement coverage retained.
4.35 checkpoints verify with runnable bun run provider:check and bun run test:e2e:provider && bun run provider:acceptance:check. No shell command contains a prose sentence after semicolon.

## External acceptance and exclusions
Local browser fixtures use production Go middleware/services/SQL with test-only injected SDK provider transports and real pinnedPG/Redis. No production auth bypass/test-login route/runtime alternate key. Fixtures cannot count actual verification/recovery/social/cookie/rotation evidence;34 prepares validators/specs before35 configuration checkpoint. 02-02 bootstraps/explicitly registers signed production-router TestBearerBrowserFixture before consumption;02-03 registers TestProductActualHTTP and extends real migrations before identity cases.02-05 only extracts/refines existing fixtures.02-21 updates roles_test.go worker PostgreSQL expectations at ownership change, before verification;02-23 refines readiness/drain/cleanup.02-34 adds provider:automated:check, permitting only enumerated human-only evidence pending while rejecting failed/missing automated rows;02-35Task2 uses it to reach the human checkpoint. The final provider:acceptance:check runs only afterTask3 evidence and rejects every pending row. Missing provider prerequisites fail nonzero rather than skip/pass; real authorized recipient/account only, no secret paste or unsolicited mail.

SAFE-04 adds no destination fetch. Creation proves zero userURL egress; fixed provider endpoints remain trusted. Any introduced userURL transport requires validatedDNS/addressdial/TLS/redirect/proxy/time/bytes defenses before use. TEN-08 proves delivered SQL/HTTP/browser/worker boundaries and absent analytics/cache404, not imagined analytics execution.

LINK-04 edits/redirect invalidation, actual redirects, analytics/clicks/conversions, custom domains/productivity/billing/enterprise/previewfetch/harddelete are scoped elsewhere and excluded. CONTEXT has no additional deferred feature. All nineteen requirements, D01–D04 and five roadmap outcomes remain required.

## 02-19 execution scope corrections

Rule 3: preserve existing 007_link_listing.sql and add the invitation schema as 008_invitations.sql; align 02-19-PLAN.md artifact references. Reachable production wiring additionally requires configuration and its tests/sample, role and browser fixture configuration, router/service registration, protected audit insertion, role constants, closed frontend error mapping, and generated-route inventory verification. These concrete paths are recorded in 02-19-SUMMARY.md.

Two native browser diagnostics proved repeated public pinned Clerk CDN asset requests failing during test completion. The test-only apps/frontend/tests/provider-assets.ts cache and apps/frontend/lib/provider-assets.test.ts controls retain exact URL provenance for Clerk JS 6.38.1/UI 1.39.1, successful JavaScript status/content-type and byte identity, deduplicate in-flight public fetches, reject failed/private/unpinned responses, and bound entries/asset/total bytes. Requests contain only a public Accept header; no provider API/session response or credential header is cached. The existing 15-second asset timeout and per-page cancellation/drain remain. The real production SDK still executes. This fixture repair is required to complete the unchanged browser gate, and does not establish live-provider acceptance.
