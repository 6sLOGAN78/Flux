# Product Specification — Modern Link Attribution Platform

## 0. Project Intent

Build a production-grade, multi-tenant link attribution and marketing analytics platform inspired by the product category represented by Dub.

This project MUST NOT copy Dub source code, proprietary implementation details, UI assets, branding, naming, or internal architecture verbatim.

Dub may be studied as:

- a product reference,
- a feature reference,
- a UX reference,
- an architectural inspiration source,
- and a benchmark for expected product quality.

The implementation must be original and built primarily around a Go backend.

The existing repository already contains a master Go boilerplate. Preserve useful boilerplate conventions where they are sound rather than replacing everything blindly.

Before changing the repository:

1. inspect the entire existing codebase,
2. determine the current Go module/package layout,
3. identify reusable boilerplate,
4. identify technical debt or incompatible assumptions,
5. create a codebase map,
6. then adapt the architecture below to the repository.

Do not rewrite functioning boilerplate simply to match this document.

---

# 1. Product Vision

Create a modern link attribution platform that lets organizations:

- create branded short links,
- use custom domains,
- track clicks in real time,
- understand visitors,
- track conversions,
- attribute conversions to marketing links,
- manage campaigns,
- generate QR codes,
- integrate events from websites and servers,
- expose APIs and webhooks,
- collaborate through workspaces,
- create partner/affiliate programs,
- measure partner-driven conversions and revenue,
- and operate the entire system at high traffic volume.

The platform should feel like a serious SaaS product rather than a URL-shortener demo.

The long-term goal is:

> marketing links + attribution + conversions + analytics + partner programs + developer platform.

---

# 2. Core Principles

## Performance

Redirects are the most latency-sensitive operation in the system.

Target:

- cached redirect lookup: p95 < 30 ms server processing
- uncached redirect lookup: p95 < 100 ms server processing
- redirect path must not synchronously depend on analytics persistence
- analytics failures must never prevent redirects

## Reliability

A redirect must still work if:

- ClickHouse is unavailable,
- analytics workers are down,
- webhook delivery is failing,
- email infrastructure is unavailable,
- asynchronous enrichment is delayed.

## Scalability

Design the platform so redirect processing, API processing, analytics ingestion, and workers can scale independently.

Do not prematurely implement dozens of microservices.

Start as a modular system with independently deployable binaries where useful.

## Multi-tenancy

Workspace isolation must be part of the data model from day one.

Every workspace-owned entity must contain or resolve to a workspace identifier.

## Developer-first

The product must expose:

- REST API
- OpenAPI specification
- API keys
- webhooks
- SDK-friendly endpoints
- idempotency where appropriate
- predictable errors
- pagination
- rate-limit metadata

## Observability

All critical services must support:

- structured logs
- metrics
- traces
- request IDs
- health endpoints
- readiness endpoints

---

# 3. Primary Users

### Marketer

Creates links, campaigns and QR codes and studies their performance.

### Growth Team

Uses conversion attribution to understand acquisition channels.

### Product Team

Uses links and events for launches and experiments.

### Developer

Integrates link creation, tracking, attribution and webhooks using APIs.

### Agency

Manages multiple brands/domains/workspaces.

### Affiliate Manager

Creates partner programs, invites affiliates, configures commissions and tracks revenue.

### Affiliate / Partner

Receives referral links and sees clicks, leads, conversions and commissions.

---

# 4. Product Domains

The platform should eventually contain these major product domains:

1. Identity
2. Workspaces
3. Links
4. Redirect Engine
5. Domains
6. Tracking
7. Analytics
8. Events
9. Conversion Attribution
10. Campaigns
11. QR Codes
12. Tags / Organization
13. Webhooks
14. API / Developer Platform
15. Integrations
16. Billing
17. Partner / Affiliate Platform
18. Notifications
19. Administration
20. Fraud / Abuse Prevention

---

# 5. Link Management

A user must be able to create a short link with:

- destination URL
- generated short key
- custom short key
- selected domain
- title
- description
- tags
- folder/campaign
- UTM parameters
- expiration time
- disabled/enabled status
- password protection
- geographic targeting rules
- device targeting rules
- operating-system targeting rules
- link cloaking configuration where technically appropriate
- custom social preview metadata
- comments/notes
- external reference ID
- custom metadata

Example:

    https://go.example.com/launch

redirects to:

    https://example.com/products/new-launch

---

# 6. Link Operations

Support:

- create link
- retrieve link
- update link
- archive link
- soft delete
- restore
- permanent delete where allowed
- duplicate link
- bulk create
- bulk update
- bulk delete
- bulk tag
- bulk move between folders/campaigns
- import links
- export links
- filter links
- search links
- sort links
- pagination
- link status
- click totals
- conversion totals
- revenue totals

Prevent domain + key collisions.

---

# 7. Redirect Engine

Redirect handling should be isolated from the primary application API.

Recommended binary:

    cmd/redirector

Request:

    GET https://go.example.com/abc123

Flow:

    request
      ↓
    identify host
      ↓
    normalize host + path
      ↓
    rate/abuse checks
      ↓
    Redis lookup
      ↓ cache miss
    PostgreSQL lookup
      ↓
    populate cache
      ↓
    evaluate link rules
      ↓
    issue redirect
      ↓
    asynchronously publish click event

Never block the redirect waiting for analytics persistence.

Supported behavior:

- 301
- 302
- 307
- configurable redirect strategy
- expired links
- password protected links
- disabled links
- geo routing
- device routing
- fallback URL
- bot detection
- preview crawler handling

---

# 8. Custom Domains

Users must be able to connect domains such as:

    go.company.com

Domain lifecycle:

    pending
    verifying
    verified
    active
    failed
    disabled

Features:

- TXT/CNAME verification
- DNS instructions
- verification polling
- ownership protection
- domain conflict prevention
- HTTPS status
- default domain
- domain-level fallback URL
- domain-level branding configuration

Do not initially build your own certificate authority.

Use infrastructure or reverse-proxy/cloud-provider certificate automation.

---

# 9. Click Tracking

Every eligible redirect generates an immutable click event.

Minimum click fields:

- event_id
- timestamp
- link_id
- workspace_id
- domain
- key
- destination
- visitor_id
- session_id
- request_id
- referrer
- referrer_domain
- UTM source
- UTM medium
- UTM campaign
- UTM term
- UTM content
- country
- region
- city when available
- device type
- browser
- operating system
- language
- bot classification
- IP-derived metadata
- user-agent-derived metadata

Avoid retaining raw IP addresses indefinitely.

Prefer privacy-conscious derived/geographic information and configurable retention.

---

# 10. Visitor Identity

Support anonymous visitor identifiers.

Possible lifecycle:

    first click
       ↓
    anonymous visitor ID
       ↓
    session
       ↓
    website interaction
       ↓
    identify call
       ↓
    known customer/user
       ↓
    conversion
       ↓
    attribution

Never use invasive fingerprinting as the primary identity mechanism.

---

# 11. Analytics

Analytics dashboard should support:

- total clicks
- unique visitors
- conversions
- conversion rate
- leads
- sales
- revenue
- average order value
- top links
- top domains
- top destinations
- top countries
- top cities
- devices
- browsers
- operating systems
- referrers
- campaigns
- UTM parameters
- time-series charts

Filters:

- date range
- workspace
- link
- domain
- campaign
- tag
- country
- device
- browser
- OS
- referrer
- UTM values
- event type

Intervals:

- hourly
- daily
- weekly
- monthly

Analytics should support near-real-time updates.

---

# 12. Conversion Tracking

Support event classes including:

    click
    lead
    signup
    purchase
    custom_event

Conversion payload may include:

- event_name
- external event ID
- customer ID
- anonymous ID
- click ID
- link ID
- timestamp
- revenue
- currency
- metadata

Events should be accepted from:

- browser SDK
- backend/server API
- integrations
- webhooks
- supported payment providers

---

# 13. Attribution

The system needs an attribution engine.

Initial models:

- last click
- first click

Later:

- linear
- configurable attribution windows
- custom attribution rules

Example:

    LinkedIn campaign link
       ↓
    click c_123
       ↓
    visitor v_456
       ↓
    signup
       ↓
    purchase ₹5,000
       ↓
    conversion attributed to link/campaign

Persist enough information to explain:

- which click was credited
- which link was credited
- which campaign was credited
- attribution model
- attribution window
- resulting revenue

Attribution logic should live in a dedicated domain package instead of controllers.

---

# 14. Client Tracking SDK

Provide a lightweight browser SDK.

Example conceptual API:

    tracker.init(...)
    tracker.identify(...)
    tracker.track(...)
    tracker.lead(...)
    tracker.purchase(...)

Requirements:

- asynchronous
- small bundle
- retry strategy
- configurable endpoint
- privacy-conscious
- no UI dependency
- first-party cookie support where appropriate

SDK implementation can initially be TypeScript even though the backend is Go.

---

# 15. Server Events API

Example:

    POST /v1/events

Payload:

    {
      "event": "purchase",
      "customer_id": "cus_123",
      "amount": 4999,
      "currency": "USD",
      "metadata": {}
    }

Requirements:

- authentication
- request validation
- idempotency keys
- deduplication
- asynchronous processing
- clear errors
- event versioning

---

# 16. Campaigns

Users can create campaigns containing related links.

Fields:

- id
- workspace_id
- name
- slug
- description
- start date
- end date
- tags
- metadata

Campaign analytics should aggregate:

- clicks
- visitors
- leads
- sales
- revenue
- conversion rate

---

# 17. Tags

Links must support multiple tags.

Users can:

- create tags
- rename tags
- delete tags
- bulk tag links
- filter by tags
- view tag-level analytics

---

# 18. Folders

Optional organizational hierarchy:

    Workspace
      ├── Product Launch
      ├── Paid Ads
      ├── Influencers
      └── Newsletter

Folders are organizational.

Campaigns represent marketing initiatives.

Do not conflate the concepts.

---

# 19. QR Codes

Each link should support QR code generation.

Support:

- PNG
- SVG

Configuration:

- size
- error correction
- foreground/background
- optional logo
- margins

QR generation should reference the short URL rather than the destination URL so tracking remains intact.

---

# 20. Social Preview

Allow users to override:

- title
- description
- image

for supported link preview scenarios.

Metadata can be fetched from the destination URL asynchronously.

Never block link creation waiting for preview scraping.

---

# 21. Workspaces

Users belong to workspaces.

Suggested roles:

    owner
    admin
    member
    viewer

Later:

    billing
    developer
    analyst

Capabilities include:

- create workspace
- invite member
- accept invitation
- remove member
- update role
- transfer ownership
- workspace settings
- delete workspace

Authorization must be enforced server-side.

---

# 22. Authentication

Support initially:

- email/password OR magic-link
- GitHub OAuth
- Google OAuth

Architecture must allow:

- SAML SSO
- OIDC
- SCIM

in later enterprise milestones.

Use secure:

- HttpOnly cookies
- Secure flag
- CSRF protection where needed
- rotation
- session expiration
- token hashing

---

# 23. API Keys

API keys belong to a workspace.

Format should allow efficient identification, such as:

    lk_live_<prefix><secret>

Store only secure hashes of secrets.

Features:

- name
- created date
- last used
- scopes
- environment
- expiration
- revoke
- rotate

Possible scopes:

    links:read
    links:write
    analytics:read
    events:write
    domains:write
    webhooks:write

---

# 24. Public REST API

Base:

    /v1

Major resources:

    /links
    /domains
    /events
    /analytics
    /campaigns
    /tags
    /folders
    /customers
    /webhooks
    /partners
    /programs
    /commissions

API rules:

- versioned
- JSON
- consistent error envelope
- cursor pagination where appropriate
- idempotency support
- rate limits
- authentication
- OpenAPI specification
- generated documentation

Example error:

    {
      "error": {
        "code": "LINK_NOT_FOUND",
        "message": "The requested link does not exist.",
        "request_id": "req_..."
      }
    }

---

# 25. Webhooks

Workspace developers can create webhook endpoints.

Events may include:

    link.created
    link.updated
    link.deleted

    domain.verified

    click.created

    lead.created
    sale.created

    customer.created

    partner.created
    partner.approved

    commission.created
    commission.approved
    commission.paid

Webhook requirements:

- HMAC signing
- delivery IDs
- retries
- exponential backoff
- timeout
- delivery logs
- response status
- manual resend
- disable failing endpoints
- event filtering

Webhook delivery must run asynchronously.

---

# 26. Integrations

Architecture should allow integrations without polluting core domain code.

Potential integrations:

- Stripe
- Shopify
- HubSpot
- Segment
- Zapier
- Slack
- Google Analytics
- customer data platforms

Integrations should communicate through domain services/events instead of directly manipulating arbitrary database tables.

---

# 27. Stripe Integration

Stripe integration can support:

- SaaS billing
- purchase conversion events
- partner revenue attribution

Do not couple attribution core directly to Stripe-specific models.

Create an integration boundary.

---

# 28. Partner / Affiliate Platform

A later major milestone must support affiliate and referral programs.

Workspace can create programs.

Program fields:

- name
- description
- domain
- default commission
- cookie/attribution duration
- status
- application rules

---

# 29. Partners

Partner model:

- partner ID
- workspace
- program
- user/contact
- status
- referral code
- referral links
- metadata

States:

    invited
    applied
    pending
    approved
    rejected
    suspended

---

# 30. Commission Engine

Commission types:

- percentage
- flat amount

Potential future models:

- recurring
- tiered
- product-specific

Commission lifecycle:

    pending
    approved
    rejected
    payable
    paid

Commission must reference the underlying conversion.

Avoid floating-point arithmetic for money.

Store monetary values using integer minor units and currency.

---

# 31. Partner Portal

Partners should eventually access:

- referral links
- clicks
- leads
- sales
- conversion rate
- revenue generated
- commission earned
- payout history

Partner permissions must remain isolated from normal workspace administration.

---

# 32. Billing

Prepare SaaS billing architecture.

Entities:

- subscription
- plan
- usage
- invoice reference
- entitlements

Possible limits:

- links
- tracked events
- domains
- workspace members
- API calls
- retention
- partner programs

Do not scatter plan-name checks through application code.

Use centralized entitlements such as:

    CanCreateDomain()
    CanUseFeature()
    MaxWorkspaceMembers()
    AnalyticsRetentionDays()

---

# 33. Search

Users must be able to search across:

- link keys
- titles
- destination URLs
- tags
- campaigns

Start with PostgreSQL-supported search.

Do not add Elasticsearch until scale demonstrates a need.

---

# 34. Abuse Prevention

A public redirect platform attracts abuse.

Plan from the beginning for:

- malicious destinations
- phishing links
- spam
- brute force
- abusive APIs
- crawler floods

Implement:

- endpoint rate limiting
- workspace quotas
- link abuse states
- domain verification
- bot detection
- admin suspension
- audit logs

Future:

- threat-intelligence provider integration

---

# 35. Audit Logs

Security-relevant operations should create audit records.

Examples:

- member invited
- role changed
- domain added
- API key created
- API key revoked
- webhook changed
- billing changed
- partner payout changed

Fields:

- workspace
- actor
- action
- resource
- timestamp
- metadata

---

# 36. Technical Architecture

Use a Go-first modular architecture.

Initial deployable applications:

    cmd/api
    cmd/redirector
    cmd/worker
    cmd/migrate

Possible future binaries:

    cmd/ingester
    cmd/scheduler

Do NOT start by creating twenty microservices.

Use logical domain boundaries first.

---

# 37. Repository Architecture

Preferred structure:

    .
    ├── cmd/
    │   ├── api/
    │   │   └── main.go
    │   ├── redirector/
    │   │   └── main.go
    │   ├── worker/
    │   │   └── main.go
    │   └── migrate/
    │       └── main.go
    │
    ├── internal/
    │   ├── auth/
    │   ├── workspace/
    │   ├── user/
    │   ├── link/
    │   ├── redirect/
    │   ├── domain/
    │   ├── campaign/
    │   ├── tag/
    │   ├── folder/
    │   ├── tracking/
    │   ├── analytics/
    │   ├── attribution/
    │   ├── conversion/
    │   ├── customer/
    │   ├── webhook/
    │   ├── apikey/
    │   ├── partner/
    │   ├── commission/
    │   ├── billing/
    │   ├── integration/
    │   ├── notification/
    │   └── audit/
    │
    ├── pkg/
    │   ├── httputil/
    │   ├── id/
    │   ├── pagination/
    │   ├── validator/
    │   ├── money/
    │   └── observability/
    │
    ├── platform/
    │   ├── database/
    │   ├── cache/
    │   ├── queue/
    │   ├── analyticsdb/
    │   ├── email/
    │   ├── storage/
    │   └── geoip/
    │
    ├── migrations/
    │
    ├── api/
    │   └── openapi/
    │
    ├── web/
    │
    ├── sdk/
    │   ├── javascript/
    │   └── go/
    │
    ├── deploy/
    │   ├── docker/
    │   └── compose/
    │
    ├── scripts/
    │
    ├── tests/
    │   ├── integration/
    │   └── e2e/
    │
    ├── docs/
    │
    ├── .planning/
    ├── go.mod
    ├── Makefile
    ├── docker-compose.yml
    └── README.md

This structure is a target.

Adapt it intelligently if the existing Go boilerplate already establishes strong conventions.

---

# 38. Internal Domain Package Pattern

Each major domain may follow:

    internal/link/
      model.go
      repository.go
      service.go
      errors.go
      validator.go
      handler.go

However avoid mechanical layering.

Business logic belongs in services/domain functions rather than HTTP handlers.

Handlers:

- parse
- authorize
- validate transport
- invoke domain/application logic
- translate errors
- respond

Handlers should not contain major business rules.

---

# 39. Primary Databases

## PostgreSQL

Source of truth for transactional entities.

Recommended for:

- users
- workspaces
- memberships
- links
- domains
- campaigns
- tags
- API keys
- webhooks
- integrations
- programs
- partners
- commissions
- billing metadata

## Redis

Use for:

- redirect cache
- rate limiting
- temporary state
- sessions where appropriate
- distributed locks where absolutely required
- idempotency cache

Do not use Redis as the authoritative source for link configuration.

## ClickHouse

Recommended analytical datastore for high-volume immutable event data.

Use for:

- clicks
- events
- conversion analytical projections
- aggregate analytics

PostgreSQL may initially store events during early development if doing so significantly reduces complexity.

The architecture must preserve an abstraction that allows ClickHouse introduction before high-scale milestones.

---

# 40. Event Pipeline

Recommended:

    redirector
        ↓
    event bus
        ↓
    consumers
       ├── analytics writer
       ├── attribution processor
       ├── webhook dispatcher
       └── enrichment processor

Possible queue:

- NATS JetStream preferred
- Redis Streams acceptable for a simpler initial deployment

Do not introduce Kafka solely because this is an analytics system.

---

# 41. Event Envelope

Internal events should share a consistent envelope.

Example:

    {
      "id": "evt_...",
      "type": "link.clicked",
      "version": 1,
      "occurred_at": "...",
      "workspace_id": "...",
      "data": {}
    }

Requirements:

- unique event ID
- event version
- timestamp
- producer
- correlation/request IDs where available

Consumers must be idempotent.

---

# 42. Data Consistency

Use strong consistency for:

- workspace membership
- link ownership
- link creation
- custom keys
- domains
- billing entitlements
- API keys

Allow eventual consistency for:

- analytics
- dashboards
- webhooks
- metadata enrichment
- attribution projections where appropriate
- aggregate counters

---

# 43. Link Cache

Suggested cache key:

    redirect:{host}:{path}

Cached structure contains enough information to perform the redirect without another database request.

Cache invalidation occurs when:

- link updated
- link deleted
- link disabled
- domain changed
- targeting changed

TTL remains as a safety mechanism.

---

# 44. Database IDs

Prefer sortable unique IDs such as UUIDv7 or another well-understood sortable identifier.

Do not expose auto-increment database IDs as externally meaningful identifiers.

Potential prefixes:

    usr_
    ws_
    lnk_
    dom_
    cmp_
    clk_
    evt_
    wh_
    key_
    prg_
    ptr_
    com_

Prefixes are optional but useful for developer ergonomics.

---

# 45. Database Design Rules

Every table should intentionally define:

- primary key
- timestamps
- tenant ownership
- unique constraints
- foreign keys
- indexes
- deletion semantics

Do not rely on application code for uniqueness that the database can guarantee.

Important unique constraints may include:

    domain + short_key
    workspace + domain
    workspace + API key prefix

---

# 46. API Architecture

Suggested Go HTTP stack:

- standard net/http where practical
- chi for routing
- OpenAPI-first or strongly OpenAPI-integrated development
- explicit middleware

Avoid large framework dependencies unless there is a compelling project-specific reason.

Middleware:

    recovery
    request ID
    logging
    tracing
    CORS
    authentication
    workspace context
    authorization
    rate limiting

---

# 47. Configuration

Configuration should be environment-driven and validated at startup.

Example variables:

    APP_ENV
    HTTP_ADDR

    DATABASE_URL
    REDIS_URL
    CLICKHOUSE_URL
    NATS_URL

    SESSION_SECRET

    GOOGLE_CLIENT_ID
    GOOGLE_CLIENT_SECRET

    GITHUB_CLIENT_ID
    GITHUB_CLIENT_SECRET

    STRIPE_SECRET_KEY

    EMAIL_PROVIDER_KEY

Never commit secrets.

Provide:

    .env.example

---

# 48. Frontend

Backend remains Go.

Recommended frontend:

    Next.js
    TypeScript
    React
    Tailwind CSS

Reason:

This is an analytics-heavy dashboard SaaS and React's ecosystem is appropriate.

The frontend must communicate with the Go API rather than becoming the primary business-logic backend.

Alternative frontend stacks may be considered only if documented during planning.

---

# 49. Frontend Areas

Application routes should eventually support:

    /login
    /register

    /dashboard
    /links
    /links/:id

    /analytics

    /domains
    /campaigns
    /tags

    /events
    /customers

    /partners
    /programs
    /commissions

    /settings
    /settings/team
    /settings/api
    /settings/webhooks
    /settings/integrations
    /settings/billing

Exact URLs are not mandatory.

---

# 50. UI Product Direction

Do not clone Dub pixel-for-pixel.

The UI should be:

- minimalist
- fast
- information dense
- keyboard friendly
- responsive
- accessible
- polished
- professional

Create an original visual identity.

Use:

- reusable component system
- design tokens
- dark/light themes where appropriate
- loading skeletons
- empty states
- errors
- keyboard shortcuts for high-frequency operations

---

# 51. Observability

Use OpenTelemetry.

Implement:

- structured logging
- tracing
- metrics

Track:

    redirect requests/sec
    redirect latency
    redirect cache hit ratio
    API latency
    API error rate
    database latency
    queue lag
    event ingestion rate
    webhook delivery failures

---

# 52. Background Jobs

Worker system handles:

- analytics persistence
- metadata scraping
- webhook delivery
- email
- domain verification
- event enrichment
- partner calculations
- cleanup
- exports
- scheduled reports

Jobs must have:

- retry policy
- maximum attempts
- idempotency
- observability
- dead-letter strategy

---

# 53. Testing Strategy

## Unit Tests

Business logic.

Especially:

- attribution
- link routing
- permissions
- commissions
- entitlement decisions

## Repository Tests

Database behavior.

## Integration Tests

Test:

- PostgreSQL
- Redis
- queue
- ClickHouse when introduced

Use containers where practical.

## API Tests

Test:

- authentication
- authorization
- validation
- pagination
- idempotency

## End-to-End

Critical flows:

    create workspace
    ↓
    connect/create domain
    ↓
    create link
    ↓
    visit short link
    ↓
    click recorded
    ↓
    conversion recorded
    ↓
    analytics updated

And:

    create program
    ↓
    create partner
    ↓
    partner referral click
    ↓
    sale
    ↓
    commission

---

# 54. Security Requirements

Mandatory:

- TLS in production
- secure session cookies
- CSRF protection where applicable
- password hashing using modern algorithm
- secrets outside repository
- SQL parameterization
- strict URL validation
- SSRF protection
- XSS protection
- rate limits
- API key hashing
- webhook signatures
- authorization at service/API boundary
- workspace isolation tests
- dependency scanning

Destination URL processing is security-sensitive.

Any URL-fetching functionality must explicitly defend against SSRF to localhost, private networks and cloud metadata endpoints.

---

# 55. Privacy

Design for data minimization.

Support eventually:

- configurable data retention
- workspace deletion
- visitor deletion
- analytics export
- consent-aware tracking configuration

Avoid persisting unnecessary raw network identifiers.

---

# 56. Docker Development Environment

Local development should run with one command.

For example:

    docker compose up -d

Services eventually:

    postgres
    redis
    nats
    clickhouse
    mail development service

Application binaries may run inside or outside Docker depending on developer workflow.

---

# 57. CI

GitHub Actions should run:

    gofmt check
    go vet
    staticcheck
    golangci-lint
    unit tests
    integration tests
    frontend lint
    frontend typecheck
    frontend tests
    build

Migration checks should eventually be included.

---

# 58. Deployment

Services must be containerized.

Design for environments:

    local
    test
    staging
    production

Do not hard-code a single cloud provider into core packages.

Initial deployment may use one provider, but infrastructure-specific behavior should remain outside domain code.

---

# 59. Milestone Strategy

Do NOT implement every feature simultaneously.

Build vertically through working product slices.

## Milestone 0 — Codebase Foundation

- inspect existing Go boilerplate
- map current architecture
- document conventions
- finalize package boundaries
- local Docker environment
- PostgreSQL
- Redis
- configuration
- logging
- OpenTelemetry foundation
- migrations
- CI
- health endpoints

Definition of done:

    API starts
    database connects
    Redis connects
    migrations execute
    tests execute
    CI passes

---

## Milestone 1 — Identity + Workspaces

Implement:

- users
- authentication
- sessions
- workspaces
- memberships
- invitations
- roles
- authorization

Definition of done:

A user can register/login, create a workspace and invite another member.

---

## Milestone 2 — Core Link Platform

Implement:

- links
- generated keys
- custom keys
- redirector
- Redis redirect cache
- archive/delete
- search/filter
- basic dashboard

Definition of done:

A user creates a link and receives a working tracked redirect URL.

---

## Milestone 3 — Domains

Implement:

- custom domains
- verification
- domain status
- domain routing
- cache invalidation

Definition of done:

A verified custom domain can serve user links.

---

## Milestone 4 — Tracking Pipeline

Implement:

- click events
- visitor IDs
- queue
- worker
- user-agent parsing
- referrer
- geolocation enrichment
- event persistence

Definition of done:

Redirect traffic produces durable asynchronous click events without affecting redirect availability.

---

## Milestone 5 — Analytics

Implement:

- analytical storage
- dashboard
- timeseries
- geographic analytics
- device analytics
- referrer analytics
- link analytics
- date filters

Definition of done:

Users can meaningfully understand link performance.

---

## Milestone 6 — Conversion Attribution

Implement:

- event API
- browser SDK
- lead events
- purchase events
- visitor identity
- click IDs
- attribution
- revenue analytics

Definition of done:

A purchase can be correctly attributed to a marketing link.

---

## Milestone 7 — Campaign Productivity

Implement:

- campaigns
- tags
- folders
- UTM builder
- QR codes
- bulk link actions
- import/export

---

## Milestone 8 — Developer Platform

Implement:

- API keys
- scopes
- OpenAPI
- rate limiting
- idempotency
- webhooks
- webhook retries
- delivery logs
- Go SDK
- JS/TS SDK

---

## Milestone 9 — Integrations

Start with:

- Stripe conversion ingestion

Then architecture for:

- Shopify
- HubSpot
- Segment
- Slack

Do not implement every integration before validating demand.

---

## Milestone 10 — Partner Platform

Implement:

- programs
- partners
- referral links
- partner attribution
- commissions
- partner dashboard

---

## Milestone 11 — Billing

Implement:

- subscriptions
- usage
- plan limits
- entitlements
- Stripe billing integration

---

## Milestone 12 — Enterprise / Production Hardening

Implement as justified:

- SSO
- SAML
- SCIM
- audit logs
- advanced RBAC
- retention controls
- high-volume optimizations
- security audit
- load testing
- disaster recovery procedures

---

# 60. Non-Goals for Initial Milestones

Do not initially:

- build Kafka infrastructure
- build Kubernetes infrastructure unless deployment requires it
- create dozens of microservices
- create an Elasticsearch cluster
- implement machine-learning attribution
- implement custom TLS infrastructure
- implement dozens of integrations
- duplicate analytics into multiple databases without need

Keep abstractions ready without paying unnecessary complexity upfront.

---

# 61. Expected System Architecture

Logical architecture:

    ┌─────────────────────┐
    │      Web App        │
    │ React / Next.js     │
    └──────────┬──────────┘
               │
               ▼
    ┌─────────────────────┐
    │       Go API        │
    │ Auth / Links / etc. │
    └───┬─────────┬───────┘
        │         │
        ▼         ▼
    PostgreSQL   Redis
        │
        │
        └───────────────┐
                        │
    Internet            │
       │                │
       ▼                │
    ┌─────────────────────┐
    │    Go Redirector    │
    └──────────┬──────────┘
               │
         Redis Cache
               │
               ▼
            Redirect
               │
               └── async event
                       │
                       ▼
               ┌───────────────┐
               │ NATS/Queue    │
               └───────┬───────┘
                       │
             ┌─────────┼───────────┐
             ▼         ▼           ▼
         analytics  attribution  webhooks
          worker       worker      worker
             │
             ▼
         ClickHouse

---

# 62. Important Architectural Rule

There are TWO fundamentally different traffic paths.

### Control Plane

Used for:

- dashboard
- authentication
- link configuration
- domains
- campaigns
- partners
- settings

Runs primarily through:

    web → API → PostgreSQL

### Data Plane

Used for:

- redirects
- click collection
- events
- analytics ingestion

Runs primarily through:

    visitor → redirector → Redis → redirect
                         ↓
                       queue
                         ↓
                      workers
                         ↓
                    ClickHouse

Do not couple data-plane availability to dashboard availability.

This separation is one of the most important architectural decisions in the project.

---

# 63. Coding Standards

Use idiomatic Go.

Prefer:

- explicit dependencies
- small interfaces
- constructor injection
- context.Context
- wrapped errors
- errors.Is / errors.As
- table-driven tests
- standard library where reasonable

Avoid:

- global mutable application state
- giant service structs
- repository interfaces for trivial things purely for ceremony
- unnecessary reflection
- premature generic abstractions
- hidden business logic in middleware
- magic environment variables
- package names like helpers/common/utils for domain logic

---

# 64. Error Design

Create domain errors that can cleanly map into HTTP errors.

Examples:

    ErrLinkNotFound
    ErrKeyAlreadyExists
    ErrDomainNotVerified
    ErrForbidden
    ErrWorkspaceNotFound

Transport adapters translate them into API responses.

Core business packages must not depend on HTTP status codes.

---

# 65. Transaction Strategy

Transactions belong around business operations requiring atomicity.

Examples:

- workspace creation + owner membership
- link creation + related metadata
- partner conversion + commission creation

Do not make repository methods secretly create large transactions.

Transaction boundaries should be visible at the application/service layer.

---

# 66. Documentation

Maintain:

    README.md
    docs/architecture.md
    docs/development.md
    docs/deployment.md
    docs/security.md
    api/openapi/openapi.yaml

Major architecture decisions should use ADRs:

    docs/adr/

Example:

    0001-postgresql-primary-store.md
    0002-redis-redirect-cache.md
    0003-clickhouse-analytics.md
    0004-nats-event-transport.md

---

# 67. GSD Requirements

GSD must treat this document as product intent, not an instruction to immediately generate all files.

Before implementation:

1. map the existing repository,
2. research current best practices,
3. reconcile this spec with the current boilerplate,
4. create PROJECT.md,
5. create REQUIREMENTS.md,
6. create ROADMAP.md,
7. separate v1 requirements from later milestones,
8. identify architectural risks,
9. create phase boundaries,
10. begin implementation only after planning is coherent.

The system should preserve progress under `.planning/`.

---

# 68. Reference Project Policy

Dub can be inspected for understanding:

- feature categories
- workflow ideas
- product expectations
- scaling concerns
- public API concepts
- public documentation
- publicly visible UX ideas

Do NOT:

- copy Dub source files
- translate TypeScript source line-for-line into Go
- reproduce their proprietary design exactly
- copy internal naming without reason
- copy protected assets or branding
- treat their current folder tree as mandatory architecture

When there is a conflict between this specification and the reference repository, prefer this specification unless a documented architectural decision explains otherwise.

---

# 69. Definition of Product Success

A successful initial product must demonstrate the complete attribution loop:

    marketer creates link
           ↓
    visitor clicks short link
           ↓
    redirect occurs immediately
           ↓
    click is recorded asynchronously
           ↓
    visitor reaches destination
           ↓
    visitor signs up / purchases
           ↓
    application sends conversion event
           ↓
    conversion is matched to click
           ↓
    marketer sees conversion + revenue

If this flow is reliable, fast, testable and observable, the core platform is working.

---

# 70. Agent Instructions

When implementing this project:

- inspect before editing
- reuse sound existing code
- do not assume repository structure
- keep modules cohesive
- prefer vertical slices
- write migrations
- add tests
- update documentation
- validate all external input
- preserve backwards compatibility when reasonable
- run tests before considering work complete
- do not leave placeholder production logic
- do not silently swallow errors
- do not introduce dependencies without justification
- do not prematurely optimize
- do not create mock implementations in production paths
- do not copy implementation code from Dub

For major architectural choices, document:

1. problem
2. options
3. decision
4. reasoning
5. consequences

---

# 71. First GSD Objective

The first GSD execution should NOT attempt to build the entire product.

Its objective is:

> Understand the existing Go repository, establish the final architecture, set up infrastructure foundations, and produce an executable roadmap for implementing a production-grade link attribution platform.

The first implementation milestone should finish with a boring, reliable foundation on which the redirect engine and link platform can be built.

---

# 72. Product Name

Use a temporary internal codename until a final brand is selected.

Do not use:

- Dub
- Dub Clone
- Dub Go
- Dub-like branding

The project is an independent product in the same category.

---

# 73. Final Instruction to GSD

Treat the project as a serious production SaaS.

Optimize for:

1. correctness
2. maintainability
3. performance of redirects
4. reliable analytics ingestion
5. clear tenant isolation
6. developer experience
7. testability
8. observability
9. security
10. future scale

Do not optimize for generating the maximum amount of code in the first pass.