# Feature Research

**Domain:** Production-grade multi-tenant link attribution and marketing analytics SaaS  
**Researched:** 2026-10-05  
**Confidence:** HIGH for category baseline and project constraints; MEDIUM for prioritization until validated with target users

## Feature Landscape

Flux should launch as an attribution product with link management, rather than as a broad URL-shortening suite. The minimum credible promise is: a workspace member creates a link, a visitor is redirected without waiting on analytics, the click becomes a durable event, the customer's application sends a conversion, Flux attributes it under an explicit model and window, and the marketer can inspect clicks, conversions, and revenue. Any v1 feature that does not make this loop usable, trustworthy, secure, or explainable should be deferred.

### Table Stakes (Users Expect These)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Authentication and workspace isolation | A multi-tenant SaaS cannot expose one customer's links, events, or revenue to another | HIGH | Authenticate all product routes; put every owned record under a workspace; enforce authorization in service and persistence boundaries. This precedes every product feature. |
| Basic team membership | Link programs are operated by teams, not only individuals | MEDIUM | v1 needs owner/admin/member/viewer semantics and invitations. Advanced custom roles, SSO, and SCIM are later. |
| Create a short link | It is the entry point to every other product capability | MEDIUM | Require a validated HTTP(S) destination; generate a collision-safe key; optionally accept a custom key; return a ready-to-use platform-domain URL. |
| Link edit, enable/disable, archive, and restore | Campaign destinations change, bad links must be stopped, and records should remain recoverable | MEDIUM | Preserve key uniqueness and history. Treat permanent deletion as a controlled later operation because analytics and attribution may reference the link. |
| Link list, search, filter, and status | A dashboard becomes unusable once a workspace has more than a handful of links | MEDIUM | v1 search by key, title, and destination; filter by active/disabled/archived; cursor pagination. Tags, folders, saved views, and bulk actions can follow. |
| Reliable redirect behavior | The public short URL is the product's most visible contract | HIGH | Resolve host + key, honor active/disabled/expired state, redirect quickly, and serve a safe missing/disabled response. Redirect availability must not depend on analytics persistence. |
| Asynchronous durable click capture | Users expect click counts without slower or less reliable redirects | HIGH | Emit an immutable click event after deciding the redirect. Queue failure must be observable and recoverable without putting analytics on the response critical path. |
| Basic bot classification and unique-visitor semantics | Raw request counts are misleading; users need to know what “click” and “unique” mean | HIGH | Store bot classification separately and define whether default dashboards exclude known bots. Use a first-party anonymous ID where possible; avoid fingerprinting. Document the uniqueness window. |
| Privacy-conscious enrichment | Location, device, referrer, and UTM breakdowns are normal link-analytics expectations | MEDIUM | Derive coarse geography and user-agent attributes asynchronously. Do not retain raw IP indefinitely. Clearly surface unknown values rather than inventing precision. |
| One production-grade conversion ingestion path | Click counts alone do not validate the product's core value | HIGH | v1 should provide an authenticated server Events API for `lead` and `purchase`, plus a minimal documented way to carry the click identifier from redirect to the customer's backend. Require an external event ID/idempotency key. |
| Revenue-safe purchase events | Marketing users need revenue, not only conversion counts | MEDIUM | Accept integer minor units and ISO currency; never accept floating-point money. Preserve original event payload/version for auditability. |
| Anonymous-to-known identity link | A click often happens before signup or purchase | HIGH | Support an anonymous visitor/click identifier and later association to a stable customer ID. Identity merge rules must be deterministic and testable. |
| Explainable first- and last-click attribution | Attribution must answer why a link received credit, not only return a total | HIGH | For each conversion, persist eligible touchpoints, selected click, model, window, link, timestamp, and reason. A workspace chooses first- or last-click. Use one documented default lookback in v1; per-link and per-event windows are later. |
| Idempotent conversion processing | Provider retries and customer retry logic are normal; duplicate purchases destroy trust | HIGH | Deduplicate within workspace + source using an external event ID or idempotency key. Return the prior result for safe retries and retain a visible duplicate/rejection reason. |
| Core analytics dashboard | Users expect an immediate answer to “what worked?” | HIGH | v1 KPIs: clicks, unique visitors, leads, purchases, conversion rate, revenue, and average order value. Include a time series and top links. Show data freshness because analytics are eventually consistent. |
| Essential breakdowns and filters | Aggregate totals alone cannot guide a campaign decision | MEDIUM | v1 filters: date range and link. Breakdowns: link, referrer domain, country, device class, and UTM source/medium/campaign. Arbitrary multi-dimensional report builders are later. |
| Event and attribution drill-down | Operators need to debug missing or surprising credit | HIGH | Provide a recent event stream or conversion detail showing click, identity match, selected model/window, attributed link, revenue, and processing state without exposing sensitive raw network data. |
| Clear empty, delayed, and error states | Eventual consistency otherwise looks like data loss | LOW | Explain that a link is live immediately, clicks may take a short time to appear, and conversions can be rejected or remain unattributed with a reason. |
| Basic abuse controls | Public redirect infrastructure attracts phishing, spam, brute force, and crawler floods | HIGH | Validate destinations, rate-limit creation and ingestion, support link suspension, and keep an operator-visible abuse state. External threat-intelligence automation can follow. |

The link-management baseline is corroborated by Bitly's official description of branded/customized links, real-time click data, referrer/location reporting, integrations, and QR extensions, and by Rebrandly's official create-link flow, generated/customizable link keys, workspace scoping, tags, and domain resources. Bitly's current analytics documentation also treats time, location, referrer, device, and item-level performance as normal dimensions. Dub and Branch's official materials show that attribution products go beyond clicks by connecting touchpoints to leads, purchases, revenue, and an explicit attribution window.

### Differentiators (Competitive Advantage)

Flux should differentiate on trust in the answer and in the redirect, not on having the longest link-builder form.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Explainable attribution ledger | A marketer or developer can see exactly which click was credited, under which model/window, and why | HIGH | Make this visible in conversion detail and queryable in the API. Preserve inputs and decisions so recalculation never silently rewrites history. This is the strongest v1 differentiator. |
| Redirect reliability isolated from analytics | Marketing links keep working during analytics, webhook, email, or dashboard failures | HIGH | This is a user-facing reliability property: redirect health and analytics freshness should be monitored separately. |
| Server-authoritative conversion and revenue ingestion | Idempotent backend events resist ad blockers and accidental/spoofed duplicate revenue better than browser-only tracking | HIGH | Start with a narrow server API. A browser SDK can later handle identity handoff and non-monetary events while purchases remain verifiable server-side. |
| Privacy-conscious identity by design | Useful attribution without invasive fingerprinting or indefinite raw-IP retention | HIGH | Explain the identity limits. Prefer explicit click IDs, first-party anonymous IDs, and later customer identification over probabilistic identity. |
| Fast path from first link to first attributed conversion | A guided verification flow proves setup before users invest in complex campaigns | MEDIUM | Provide a test link, click inspector, sample event request, idempotency result, and attribution confirmation. Treat setup diagnostics as product features. |
| Data freshness and processing transparency | Users can distinguish “still processing,” “unattributed,” “duplicate,” and “rejected” from missing data | MEDIUM | Surface ingestion time, processing state, freshness watermark, and structured rejection reasons. |
| Clean workspace-level isolation and auditability | Agencies and teams can trust that links and revenue never cross tenant boundaries | HIGH | Tenant tests and scoped identifiers are essential. Detailed audit logs and cross-workspace agency consoles are later. |

### Anti-Features (Commonly Requested, Often Problematic)

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Machine-learning or multi-touch attribution in v1 | Sounds more sophisticated than first/last click | It is hard to explain, impossible to validate without substantial clean history, and expands identity/privacy risk | Ship deterministic first- and last-click with a visible window and evidence trail; add other models only after customers demonstrate a decision they improve. |
| Synchronous analytics writes during redirect | Appears to guarantee immediate dashboards | Analytics or database latency would become redirect latency and outages would break public links | Publish an immutable event asynchronously and show a freshness watermark. |
| Broad browser fingerprinting | Seems to improve cross-session and cross-device matching | It is privacy-invasive, brittle under browser changes, and conflicts with the stated product constraints | Use explicit click IDs, first-party anonymous IDs, customer identification, and honest “unattributed” states. |
| Indefinite raw-IP and user-agent retention | Makes later reprocessing seem easier | It increases privacy and breach impact without improving most reporting after enrichment | Store derived/coarse geography and device attributes; apply short, explicit retention to any transient source data. |
| Partner programs and commissions in v1 | Affiliate use cases are commercially attractive | Programs, partner identities, commission rules, payouts, tax/compliance, and partner permissions form a second product and can hide flaws in core attribution | First prove partner-like referral links as ordinary links and conversions; build programs and commissions in v2 on the validated attribution ledger. |
| SaaS billing and plan enforcement in v1 | Every SaaS eventually needs monetization | Subscriptions, entitlements, usage metering, invoices, and failure states add cross-cutting complexity before value is proven | Record usage counters and keep entitlement seams; add real billing after activation and usage patterns are known. |
| Dozens of native integrations | Reduces setup work in theory | Each integration adds schema mapping, authentication, retries, version drift, and support load | Stabilize the generic Events API first; add one Stripe conversion integration after the core loop is reliable, then prioritize by demand. |
| Custom domains before the core loop | Branded links are category-standard and improve trust | DNS verification, domain conflict handling, TLS automation, routing, and cache invalidation delay attribution validation | Launch with a managed platform domain; add infrastructure-managed custom domains in v1.x immediately after end-to-end attribution works. |
| Advanced geo/device/OS routing, A/B routing, and cloaking in v1 | Makes links feel powerful | Rule precedence, preview crawlers, compliance, cache invalidation, and test combinations dramatically expand redirect risk | Support one destination and simple lifecycle states first; add routing rules after the redirect and event pipeline are proven. |
| QR design studio, landing pages, and link-in-bio | Competitors bundle adjacent acquisition surfaces | These are separate editors and content products; they do not prove conversion attribution | Add simple QR export in v1.x because it reuses links; defer hosted pages and page builders to a separate product decision. |
| Fully customizable analytics/report builder | Enterprise buyers ask for every dimension and formula | Query cost, permissions, UX, exports, and semantic consistency grow faster than the core product | Ship opinionated KPIs and breakdowns; later add saved views/export, then a bounded report builder based on observed questions. |
| AI-generated insights | Feels current and can demo well | It can confidently summarize incomplete or semantically inconsistent data before the metrics are trusted | Make metric definitions, drill-downs, and freshness trustworthy first; add narrative insights only after users validate the underlying numbers. |
| Bespoke certificate infrastructure | Gives theoretical control over domains | It creates a high-risk operational and security burden unrelated to attribution differentiation | Use reverse-proxy/cloud-provider certificate automation. |
| Elasticsearch, Kafka, and many microservices at launch | Signals “enterprise scale” | They create operational work before traffic demonstrates need and make correctness harder to trace | Use PostgreSQL, Redis, the existing job foundation, and a small number of independently deployable binaries; preserve replaceable boundaries. |

## Feature Dependencies

```text
[Stable foundation: config, migrations, lifecycle, health, CI, tests]
    └──requires-before──> [Authentication + workspace + membership + authorization]
                              └──requires-before──> [Links + lifecycle + search]
                                                        └──requires-before──> [Fast redirect resolution]
                                                                                  └──emits──> [Durable click event]
                                                                                                   ├──feeds──> [Click analytics]
                                                                                                   └──provides touchpoint──> [Attribution]

[Click identifier handoff] ──requires──> [Fast redirect resolution]
[Anonymous visitor identity] ──requires──> [Durable click event]
[Customer identification] ──links──> [Anonymous visitor identity]

[Authenticated, idempotent conversion Events API]
    ├──requires──> [Workspace/API authentication]
    ├──requires──> [Money + event schema]
    └──feeds──> [Attribution]

[Attribution]
    ├──requires──> [Durable click event]
    ├──requires──> [Conversion Events API]
    ├──requires──> [Identity/click match]
    ├──requires──> [Defined model + window]
    └──feeds──> [Conversion + revenue analytics and explanation]

[Custom domains] ──enhances──> [Links + redirects]
[Campaigns/tags/folders/UTMs/QR/bulk actions] ──enhance──> [Link productivity + analytics grouping]
[API keys/OpenAPI/webhooks/SDKs] ──scale access to──> [Stable links, events, and analytics contracts]
[Stripe and other integrations] ──adapt external events into──> [Stable Events API]
[Partner programs + commissions] ──require──> [Stable attribution + money semantics + separate partner authorization]
[Billing] ──requires──> [Usage metering + centralized entitlements]
[Enterprise identity and audit controls] ──extend──> [Proven workspace authorization model]
```

### Dependency Notes

- **Foundation precedes product schema:** The current repository has initialization scaffolding but no product DDL, repositories, or executable product tests. Health, lifecycle, migration, validation, and process-separation concerns should be corrected before redirect and event behavior depends on them.
- **Workspace authorization precedes all owned resources:** Links, clicks, conversions, analytics queries, and credentials must be scoped from their first migration. Adding tenant boundaries later is a rewrite and a security risk.
- **A working redirect precedes click analytics:** A click event is a consequence of resolving a valid link. The redirect decision, event envelope, and idempotent event identifier should be defined together.
- **Click identifier handoff precedes deterministic conversion attribution:** Flux needs a documented way for a customer application to receive and return a click identifier. Otherwise v1 can report clicks and conversions but cannot reliably connect them.
- **Identity rules precede first-/last-click models:** The model can only choose among touchpoints that Flux can associate to the same anonymous or known customer. Unknown identity must yield a visible unattributed result rather than a guess.
- **Idempotency precedes revenue dashboards:** A duplicate purchase event must not double revenue. Deduplication identity and money semantics are part of the event contract, not a reporting cleanup.
- **Attribution explanation precedes optimization features:** A campaign dashboard built on opaque or mutable credit will accelerate wrong decisions. Store attribution facts before adding complex reports.
- **Custom domains follow the managed-domain loop:** They increase trust and are a market expectation, but they do not establish whether click-to-revenue attribution works. Add them immediately after the core loop, using managed certificate automation.
- **Productivity features follow reliable entities:** Tags, folders, campaigns, QR codes, bulk operations, and import/export should operate on stable link and analytics models rather than defining those models indirectly.
- **Developer-platform breadth follows stable internal contracts:** The v1 Events API can use a narrowly scoped workspace credential. General API keys, fine-grained scopes, public CRUD APIs, webhooks, SDKs, and compatibility promises belong in v1.x after resource semantics settle.
- **Integrations adapt into core APIs:** Stripe or Shopify code should translate provider events into the same idempotent conversion model. It should not create a parallel attribution path.
- **Partner and billing products come after attribution validation:** Both depend on trustworthy conversion identity, integer money, lifecycle states, tenant authorization, and audit history. They should not shape v1 UI or phase ordering beyond clean extension points.

## MVP Definition

### Launch With (v1)

This is the smallest production-credible vertical slice that proves the complete value proposition.

- [ ] **Stabilized service foundation** — migrations, separate liveness/readiness, graceful lifecycle, tests, CI, and observable API/redirect/worker processes are prerequisites for reliable product behavior.
- [ ] **Authentication, workspace, membership, invitation, and fixed roles** — establishes tenant isolation and a credible team SaaS surface.
- [ ] **Platform-domain link creation and lifecycle** — create using a generated or custom key; edit destination/title; enable/disable; archive/restore; search/filter/list with pagination.
- [ ] **Independent redirect path with cache and safe fallback behavior** — resolves active links quickly and does not wait on analytics.
- [ ] **Durable click events and minimal enrichment** — event ID, time, workspace, link, click/anonymous visitor ID, referrer domain, UTM values, country, device class, bot classification, and processing status.
- [ ] **Documented click-ID handoff** — propagate a namespaced click ID to an opted-in destination and provide a tiny, original integration helper or example that preserves it for the customer's server.
- [ ] **Authenticated idempotent server Events API** — accept `lead` and `purchase` with external event ID, click/customer identity, timestamp, integer amount, currency, and metadata; expose structured validation and duplicate results.
- [ ] **Anonymous-to-customer association** — allow a signup or server event to bind the pre-conversion click identity to the customer's stable identifier.
- [ ] **Explainable first- and last-click attribution** — workspace-selected model, one documented default lookback, immutable decision record, visible unattributed reason, and no probabilistic fingerprint match.
- [ ] **Focused workspace analytics** — clicks, unique visitors, leads, purchases, conversion rate, revenue, average order value, time series, top links, and simple referrer/country/device/UTM breakdowns with date/link filters.
- [ ] **Event and conversion detail** — inspect the source click, customer/anonymous identity state, attribution decision, event processing state, and revenue.
- [ ] **Abuse and privacy baseline** — destination validation, creation/ingestion rate limits, suspension state, bot separation, coarse enrichment, and explicit raw-IP retention policy.
- [ ] **Setup verification journey** — create a test link, click it, send a sample idempotent conversion, and see the attributed result and data-freshness status.

The v1 acceptance journey should be executable end to end:

```text
workspace member creates managed-domain link
    → visitor receives an immediate redirect and click ID
    → click becomes durable asynchronously
    → customer backend sends lead or purchase with an idempotency key
    → Flux associates identity and applies first- or last-click within the documented window
    → marketer sees click, conversion, revenue, and the exact attribution explanation
```

### Add After Validation (v1.x)

Add these once the v1 loop is reliable under retries, delayed workers, cache invalidation, and multi-tenant authorization tests.

- [ ] **Custom domains with DNS verification and managed TLS** — category-standard branding; add after redirect and cache semantics are proven.
- [ ] **Campaigns, tags, folders, UTM builder/templates** — improve organization and grouped reporting after links and attribution dimensions are stable.
- [ ] **QR generation (PNG/SVG)** — low conceptual risk once a short URL is stable; QR scans should flow through the same redirect and click pipeline.
- [ ] **Bulk create/update/archive/tag and CSV import/export** — justified when real workspaces exceed manual management scale.
- [ ] **Minimal browser SDK** — capture and persist first-party click/anonymous IDs, identify users, and send trusted non-monetary events; keep purchase events server-authoritative by default.
- [ ] **First public Links and Analytics APIs** — reuse settled internal contracts with workspace-scoped API keys, pagination, consistent errors, and OpenAPI.
- [ ] **Scoped API keys, rotation, and rate-limit metadata** — required before customers automate link/event workflows broadly.
- [ ] **Signed webhooks with retries and delivery logs** — add when customers need attributed conversion output in their own systems.
- [ ] **One Stripe conversion adapter** — validate the integration boundary by translating provider events into the same idempotent Events API model.
- [ ] **Saved analytics views and CSV export** — add when repeated reporting workflows are observed; retain opinionated metric definitions.
- [ ] **Simple expiration, password protection, and destination fallback** — useful link controls after redirect rule precedence and cache invalidation are reliable.
- [ ] **Basic audit log for security-relevant changes** — API keys, domain verification, membership/role changes, and webhook configuration.

### Future Consideration (v2+)

- [ ] **Partner/affiliate programs, partner portal, referral links, commissions, and payouts** — a distinct product surface built on proven attribution and money semantics.
- [ ] **SaaS billing, subscriptions, usage, quotas, plans, invoices, and centralized entitlements** — add when packaging and usage drivers are known.
- [ ] **Integration catalog** — Shopify, HubSpot, Segment/CDPs, Zapier, Slack, and others prioritized by customer demand after the generic event boundary succeeds.
- [ ] **Advanced link routing** — geo/device/OS/language rules, A/B destinations, deep links, preview crawlers, and controlled cloaking require a separate redirect-rules phase.
- [ ] **Additional attribution models and configurable per-link/event windows** — linear or multi-touch only after sufficient clean data and a concrete decision use case exist.
- [ ] **Cross-device or mobile-app attribution** — requires dedicated SDKs, platform privacy handling, stronger identity design, and separate validation.
- [ ] **Customer LTV, retention, cohort, funnel, and journey analytics** — useful after conversion identity is trustworthy and recurring revenue events are normalized.
- [ ] **Enterprise identity and governance** — SAML/OIDC SSO, SCIM, advanced/custom RBAC, retention controls, data residency, and expanded audit/export controls.
- [ ] **Agency cross-workspace console** — requires explicit delegated access and aggregate-query boundaries; never bypass ordinary tenant checks.
- [ ] **Advanced abuse/fraud automation** — threat-intelligence providers, reputation scoring, automated review, and fraud heuristics after baseline abuse operations exist.
- [ ] **Hosted pages/link-in-bio and content tools** — separate product decision; do not assume they belong in link attribution.
- [ ] **Custom dashboards, natural-language analytics, and AI insights** — only after metric semantics, freshness, and query controls are trusted.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Foundation reliability and tests | HIGH | HIGH | P1 |
| Workspace identity, membership, and isolation | HIGH | HIGH | P1 |
| Create/manage/search links on a managed domain | HIGH | MEDIUM | P1 |
| Reliable cached redirect | HIGH | HIGH | P1 |
| Durable asynchronous click pipeline | HIGH | HIGH | P1 |
| Click-ID handoff and identity association | HIGH | HIGH | P1 |
| Idempotent lead/purchase Events API | HIGH | HIGH | P1 |
| Explainable first-/last-click attribution | HIGH | HIGH | P1 |
| Core conversion and revenue analytics | HIGH | HIGH | P1 |
| Event/attribution drill-down and setup diagnostics | HIGH | MEDIUM | P1 |
| Destination validation, rate limits, suspension, privacy baseline | HIGH | HIGH | P1 |
| Custom domains with managed TLS | HIGH | HIGH | P2 |
| Campaigns/tags/folders/UTM tools | MEDIUM | MEDIUM | P2 |
| QR generation | MEDIUM | LOW | P2 |
| Bulk actions and import/export | MEDIUM | MEDIUM | P2 |
| Browser SDK | HIGH | HIGH | P2 |
| Public link/analytics APIs, scoped keys, and OpenAPI | HIGH | HIGH | P2 |
| Webhooks with retries and delivery logs | MEDIUM | HIGH | P2 |
| First Stripe conversion integration | MEDIUM | HIGH | P2 |
| Partner programs and commissions | HIGH for affiliate segment | HIGH | P3 |
| SaaS billing and entitlements | HIGH for commercialization | HIGH | P3 |
| Advanced routing and deep links | MEDIUM | HIGH | P3 |
| Multi-touch/ML attribution | LOW until data proves need | HIGH | P3 |
| Enterprise SSO/SCIM/custom RBAC | HIGH for enterprise segment | HIGH | P3 |
| AI analytics and hosted pages | LOW for core proof | HIGH | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have after the core loop is validated
- P3: Later product/segment expansion

## Focused Competitor Signal

This is a category-baseline check, not an implementation blueprint. Flux must use original product design, naming, code, and assets.

| Market Signal | Current Official Evidence | Implication for Flux |
|---------------|---------------------------|----------------------|
| Link creation is expected to be immediate and manageable | Rebrandly's official API guide creates a working link from a destination and generated key; Bitly describes shortening, customization, management, and analysis as the core link product | v1 needs a polished create/list/edit/lifecycle workflow on a managed domain, not a raw redirect endpoint |
| Workspace scoping is normal for team link assets | Rebrandly documents links, tags, scripts, and domains inside workspaces shared with teammates | Tenant ownership and team authorization are foundational, not enterprise polish |
| Analytics normally include time, location, referrer, device, and per-link performance | Bitly's current analytics documentation lists engagement over time, location, referrer, device, and item-level views | v1 should cover these focused dimensions; a general report builder can wait |
| Attribution must link touchpoints to downstream outcomes under a time window | Branch documents click-to-conversion windows; Dub describes click → lead → sale and revenue analytics | v1 must include conversion ingestion, identity/click matching, a documented window, and revenue—not only click counts |
| Anonymous and identified identities need explicit linking | PostHog's official identity documentation distinguishes anonymous events from identified profiles and describes linking earlier anonymous activity on identification | Flux needs deterministic anonymous-to-customer association and visible limits; fingerprinting is unnecessary |
| QR, custom domains, bulk tooling, and integrations are mature-category extensions | Bitly and Rebrandly expose these as product/API capabilities; their presence is broad but does not prove attribution correctness | Plan them for v1.x behind the end-to-end attribution proof |

## Product Validation Questions

These questions should be answered with design partners before expanding v1.x:

1. Can a developer integrate the click-ID handoff and send the first attributed server conversion in under one hour using only docs and the setup verifier?
2. Do marketers trust the difference between total clicks, human clicks, unique visitors, leads, and purchases?
3. When a conversion is unattributed, duplicated, delayed, or rejected, can the user determine why without support?
4. Does first- versus last-click attribution change a real budget or campaign decision for the target segment?
5. Which next feature removes the largest activation barrier: custom domains, browser SDK, Stripe ingestion, or team workflow?
6. At what link count do tags/folders, bulk operations, and imports become necessary?
7. Which retention and privacy controls are required by the first customers, rather than assumed from enterprise checklists?

## Sources

### Project sources

- [Flux project definition and constraints](../PROJECT.md) — authoritative project scope and active requirements.
- [Product specification](../../spec.md) — authoritative long-term intent, success loop, and non-goals.
- [Codebase concerns](../codebase/CONCERNS.md) — source-inspected brownfield baseline and implementation gaps.

### Current official/public product documentation

- [Bitly: What is Bitly?](https://support.bitly.com/hc/en-us/articles/230895688-What-is-Bitly) — updated 2026-08-10; official baseline for link management, customization, click analytics, QR codes, and integrations. **Confidence: HIGH.**
- [Bitly: What metrics are available?](https://support.bitly.com/hc/en-us/articles/20370474672141-What-metrics-are-available-in-Bitly) — updated 2026-07-23; official analytics dimensions and per-item reporting. **Confidence: HIGH.**
- [Rebrandly: Get started](https://developers.rebrandly.com/docs/get-started) — current official API flow for creating a branded short link and selecting a domain. **Confidence: HIGH.**
- [Rebrandly: Workspaces](https://developers.rebrandly.com/docs/workspaces) — updated 2026; official workspace/resource scoping and team-sharing behavior. **Confidence: HIGH.**
- [Rebrandly: Tags](https://developers.rebrandly.com/docs/tags) — current official link-organization capability. **Confidence: HIGH.**
- [Rebrandly: Traffic routing](https://developers.rebrandly.com/docs/advanced-link-options) — current official evidence that conditional routing is an advanced, separately scoped capability. **Confidence: HIGH.**
- [Branch: Attribution](https://help.branch.io/docs/attribution-page-new) — updated 2026-08-18; official definitions for click-to-conversion attribution windows and downstream events. **Confidence: HIGH.**
- [Dub: Introducing Dub Conversions](https://dub.co/blog/introducing-dub-conversions) — official description of click → lead → sale conversion analytics and revenue attribution. Used only as public category evidence, not as implementation or design material. **Confidence: HIGH for stated product behavior.**
- [PostHog: People and identity](https://posthog.com/docs/data/persons) — current official documentation for anonymous events, identify calls, and linking prior anonymous activity to an identified person. **Confidence: HIGH.**

## Confidence and Gaps

- **HIGH confidence:** The category expects competent link management, team/workspace scoping, click analytics, common breakdowns, custom domains, and automation surfaces. These are consistent across current official sources.
- **HIGH confidence:** The core attribution product requires conversion ingestion, identity or click matching, a time window, deduplication, revenue semantics, and an explainable result. These follow directly from the project intent and current attribution-product documentation.
- **MEDIUM confidence:** Custom domains should follow rather than join v1. They are a market expectation, but the ordering is a roadmap judgment optimized for proving Flux's core value quickly.
- **MEDIUM confidence:** Both first- and last-click should ship in v1. The product specification explicitly names them as initial models, but design-partner evidence may justify launching with one model and adding the second immediately afterward.
- **MEDIUM confidence:** Fixed roles and invitations belong in v1. They are appropriate for a production-grade multi-tenant product, but an internal technical preview could temporarily use a single-member workspace without changing the underlying authorization model.
- **Open gap:** No direct target-user interviews, pricing research, or usage telemetry were available. Prioritization should be rechecked after the first five to ten integrated workspaces complete the setup journey.
- **Open gap:** The exact click-ID propagation mechanism needs phase-specific product and privacy design because destination query parameters, cookies, cross-domain behavior, and customer backend frameworks impose different constraints.

---
*Feature research for: Flux multi-tenant link attribution and marketing analytics SaaS*  
*Researched: 2026-10-05*
