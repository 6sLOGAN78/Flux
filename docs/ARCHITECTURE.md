# ARCHITECTURE.md — Protask Master Monorepo & Technology Reference Blueprint

> **Source**: Complete Protask Architecture & Exhaustive Technology Reference specification.

---

## 1. Package Managers, Orchestrators & Runtimes

| Tool / Manager | Version | Category | Purpose & Usage |
| :--- | :--- | :--- | :--- |
| **Bun** | `1.2.13` | JS/TS Package Manager | Package installation, script runner, monorepo workspace linking (`packageManager` in `package.json`). |
| **Turborepo** | `^2.5.5` | Monorepo Orchestrator | Pipeline manager executing parallel `dev`, `build`, `typecheck`, `lint`, and `clean` tasks across apps & packages (`turbo.json`). |
| **Go** | `1.24.5` | Backend Language Runtime | Compiles and executes the backend server binaries (`go.mod`). |
| **Go Modules** | Native | Go Package Manager | Manages Go dependencies and versioning via `go.mod` and `go.sum`. |
| **Taskfile (Go Task)** | `Taskfile.yml` | Task Runner | Automates database migrations, builds, and backend scripts (`taskfile.dev`). |
| **Docker & Docker Compose** | Docker v28+ | Containerization | Multi-stage Docker builds and `docker-compose.yml` orchestrating Postgres, Redis, Go Backend, and Nginx Frontend. |

---

## 2. Third-Party Cloud Services & External Integrations

| Service / Provider | Package / SDK | Location | Purpose & Functionality |
| :--- | :--- | :--- | :--- |
| **Clerk Auth** | `github.com/clerk/clerk-sdk-go/v2` | `apps/backend` | Validates JWT Bearer tokens and handles backend user identity context (`middleware/auth.go`). |
| **Clerk Frontend** | `@clerk/clerk-react ^5.38.1` & `@clerk/themes` | `apps/frontend` | Client-side authentication hooks (`useAuth`), sign-in/sign-up modals, and theme styling. |
| **New Relic APM** | `github.com/newrelic/go-agent/v3` | `apps/backend` | Complete Application Performance Monitoring & distributed tracing (`middleware/tracing.go`). |
| **New Relic Integrations** | `nrecho-v4`, `nrpgx5`, `nrredis-v9`, `nrpkgerrors` | `apps/backend` | Automatically instruments Echo routes, PostgreSQL queries, Redis calls, and error stack traces. |
| **New Relic Log Forwarding** | `zerologWriter` | `apps/backend` | Links Zerolog JSON log events directly to New Relic APM transaction traces (`logger/logger.go`). |
| **AWS S3 / S3-Compatible** | `github.com/aws/aws-sdk-go-v2/service/s3` | `apps/backend` | File attachment storage, generating S3 presigned upload & download URLs (`service/s3.go`). |
| **Resend Email API** | `github.com/resend/resend-go/v2` | `apps/backend` | Dispatching transactional emails (notifications, welcome emails). |
| **React Email** | `react-email 4.0.2` & `@react-email/components` | `packages/emails` | JSX-rendered email templates and interactive preview server (`bun run dev` on port 3001). |

---

## 3. Database, Cache & Background Queues

| Component | Library / Driver | Usage & Description |
| :--- | :--- | :--- |
| **PostgreSQL 16** | `postgres:16-alpine` | Core relational SQL database for todos, categories, comments, and attachments. |
| **PostgreSQL Driver** | `github.com/jackc/pgx/v5` | High-performance PostgreSQL driver and connection pool (`pgxpool`). |
| **Database Migrations** | `github.com/jackc/tern/v2` | Manages versioned raw SQL migration files (`001_setup.sql`, `002_todo.sql`, `003_attachments.sql`). |
| **DB Logger** | `github.com/jackc/pgx-zerolog` | Adapts `pgx` query execution logs into Zerolog formatted output. |
| **Redis 8** | `redis:8-alpine` | In-memory key-value cache and job queue storage. |
| **Redis Client** | `github.com/redis/go-redis/v9` | Go Redis client for caching and session management. |
| **Background Task Queue** | `github.com/hibiken/asynq` | Asynchronous worker processing backed by Redis for background jobs (emails, cleanup, scheduled tasks). |

---

## 4. Contract-First API & Schema Pipeline

| Tool / Package | Package / Import | Role in Schema Pipeline |
| :--- | :--- | :--- |
| **Zod** | `zod ^3.25.76` | Shared domain validation schemas (`ZTodo`, `ZCategory`, `ZComment`) in `packages/zod`. |
| **Zod OpenAPI Plugin** | `@anatine/zod-openapi ^2.2.7` | Extends Zod definitions with OpenAPI 3.0 schema metadata attributes (`extendZodWithOpenApi`). |
| **ts-rest Core** | `@ts-rest/core ^3.52.1` | Defines type-safe HTTP contracts (`path`, `method`, `query`, `body`, `responses`) in `packages/openapi`. |
| **ts-rest OpenAPI Generator** | `@ts-rest/open-api ^3.52.1` | Converts `@ts-rest` contracts into OpenAPI 3.0 JSON (`packages/openapi/src/gen.ts`). |
| **Scalar API Reference** | `@scalar/api-reference` | Modern interactive Swagger UI viewer loaded in backend `static/openapi.html` served at `GET /docs`. |
| **Go Request Validator** | `github.com/go-playground/validator/v10` | Validates incoming JSON payloads and URL params in Echo handlers to mirror Zod constraints. |
| **TS Contract Execution** | `tsx ^4.19.3` & `tsc-alias` | Executes `gen.ts` script to generate and copy `openapi.json` into `apps/backend/static/openapi.json`. |

---

## 5. Backend Libraries & Utilities (`apps/backend`)

| Library / Module | Import Path | Purpose |
| :--- | :--- | :--- |
| **Echo Web Framework** | `github.com/labstack/echo/v4` | Web router, context handler, and global HTTP middlewares (`internal/router`). |
| **Rate Limiter** | `golang.org/x/time/rate` | In-memory token bucket rate limiting middleware for API routes. |
| **Config Loader** | `github.com/knadh/koanf/v2` & `providers/env` | Environment variable configuration loader supporting structured config structs. |
| **Dotenv Parser** | `github.com/joho/godotenv` | Loads `.env` environment variables into process runtime. |
| **UUID Library** | `github.com/google/uuid` | Generates RFC 4122 compliant UUID v4 identifiers. |
| **Zerolog Logging** | `github.com/rs/zerolog` | Zero-allocation JSON structured logging with log levels (debug, info, warn, error). |
| **Cobra CLI** | `github.com/spf13/cobra` | CLI command framework powering `protask-server` and `protask-cron` binaries. |
| **Testify** | `github.com/stretchr/testify` | Unit test assertion framework (`assert`, `require`). |
| **Testcontainers** | `github.com/testcontainers/testcontainers-go` | Programmatically provisions Postgres & Redis Docker containers during Go integration testing. |

---

## 6. Frontend Libraries & Styling (`apps/frontend`)

### Core Framework & State
* **React 19** (`react ^19.1.0`, `react-dom ^19.1.0`): Modern UI rendering engine.
* **Vite 7** (`vite ^7.0.4`, `@vitejs/plugin-react`): Lightning-fast frontend build tool and dev server.
* **React Router v7** (`react-router-dom ^7.7.1`): Client-side SPA routing (`createBrowserRouter`, `PublicRoute`, `ProtectedRoute`).
* **TanStack React Query** (`@tanstack/react-query ^5.84.1`, `@tanstack/react-query-devtools`): Server-state management, caching, deduplication, and devtools.
* **Axios** (`axios ^1.11.0`): Promise-based HTTP client custom-configured for `ts-rest` client integration.

### UI Component Primitives (Radix UI / shadcn/ui)
* `@radix-ui/react-accordion`, `@radix-ui/react-alert-dialog`, `@radix-ui/react-aspect-ratio`, `@radix-ui/react-avatar`, `@radix-ui/react-checkbox`, `@radix-ui/react-collapsible`, `@radix-ui/react-context-menu`, `@radix-ui/react-dialog`, `@radix-ui/react-dropdown-menu`, `@radix-ui/react-hover-card`, `@radix-ui/react-label`, `@radix-ui/react-menubar`, `@radix-ui/react-navigation-menu`, `@radix-ui/react-popover`, `@radix-ui/react-progress`, `@radix-ui/react-radio-group`, `@radix-ui/react-scroll-area`, `@radix-ui/react-select`, `@radix-ui/react-separator`, `@radix-ui/react-slider`, `@radix-ui/react-slot`, `@radix-ui/react-switch`, `@radix-ui/react-tabs`, `@radix-ui/react-toggle`, `@radix-ui/react-toggle-group`, `@radix-ui/react-tooltip`.

### Form Validation & Utilities
* **React Hook Form** (`react-hook-form ^7.62.0`): High-performance form state management.
* **Hookform Resolvers** (`@hookform/resolvers ^5.2.1`): Connects React Hook Form with Zod schemas.

### Styling & Animation
* **Tailwind CSS v4** (`tailwindcss ^4.1.11`, `@tailwindcss/vite`): Next-gen utility-first CSS engine.
* **CVA** (`class-variance-authority ^0.7.1`): Variant-based component styling.
* **Clsx & Tailwind Merge** (`clsx`, `tailwind-merge`): Safely combines dynamic Tailwind class names without conflicts.
* **Animate CSS** (`tw-animate-css`): Animation utility classes.

### Specialized UI Elements
* **Lucide Icons** (`lucide-react`): Icon set for modern web applications.
* **Recharts** (`recharts 2.15.4`): Interactive charts & statistical graphics (Dashboard metrics).
* **Cmdk** (`cmdk ^1.1.1`): Fast command palette search menu.
* **Sonner** (`sonner ^2.0.7`): Toast notification system.
* **Vaul** (`vaul ^1.1.2`): Mobile-friendly drawer / bottom-sheet dialogs.
* **Embla Carousel** (`embla-carousel-react ^8.6.0`): Smooth carousel slider component.
* **Date Picker** (`react-day-picker ^9.8.1`, `date-fns ^4.1.0`): Calendar date selection and date manipulation.
* **OTP Input** (`input-otp ^1.4.2`): One-time password input fields.
* **Resizable Panels** (`react-resizable-panels ^3.0.4`): Split-pane resizable layout windows.
* **Theme Switching** (`next-themes ^0.4.6`): Dark/light mode theme management.
* **Pattern Matching** (`ts-pattern ^5.8.0`): Pattern matching utility for complex React state branching.

---

## 7. Complete Project Directory Layout

```
protask/
├── package.json               # Bun workspace config & Turborepo script commands
├── bun.lock                   # Bun lockfile
├── turbo.json                 # Turborepo task pipeline definitions
├── docker-compose.yml         # Container orchestration (Postgres, Redis, Backend, Frontend)
├── ARCHITECTURE.md            # Architecture blueprint & technology reference
│
├── apps/
│   ├── backend/               # Go (Echo v4) Service
│   │   ├── cmd/               # Executable main entrypoints
│   │   │   ├── protask/       # Server binary (main.go)
│   │   │   └── cron/          # Cron job binary (main.go)
│   │   ├── internal/          # Application modules
│   │   │   ├── config/        # Environment settings parser (koanf)
│   │   │   ├── database/      # Postgres connection pool & Tern migrations
│   │   │   │   └── migrations/# Versioned SQL migration files (.sql)
│   │   │   ├── handler/       # Echo HTTP controllers
│   │   │   ├── middleware/    # Auth (Clerk), Tracing (NewRelic), CORS, Rate Limit
│   │   │   ├── repository/    # Raw SQL persistence queries
│   │   │   ├── router/        # Echo routes setup (/api/v1, /docs, /status)
│   │   │   └── service/       # Core business logic layer
│   │   ├── static/            # Static assets (openapi.json, openapi.html for Scalar UI)
│   │   ├── Taskfile.yml       # Go task automation scripts
│   │   ├── Dockerfile         # Multi-stage Go build Dockerfile
│   │   └── go.mod             # Go module manifest & dependencies
│   │
│   └── frontend/              # React 19 + Vite Application
│       ├── src/
│       │   ├── api/           # ts-rest client configuration & React Query hooks
│       │   ├── components/    # Radix UI + shadcn/ui components
│       │   ├── config/        # Environment variable validator (Zod)
│       │   ├── hooks/         # Custom React hooks
│       │   ├── pages/         # Dashboard, Todos, Categories, Settings, Auth pages
│       │   ├── main.tsx       # Entrypoint (ClerkProvider, QueryClientProvider)
│       │   └── router.tsx     # React Router v7 routes
│       ├── nginx.conf         # Nginx server configuration for SPA routing
│       ├── Dockerfile         # Bun build + Nginx runtime Dockerfile
│       ├── package.json       # Frontend dependencies & scripts
│       └── vite.config.ts     # Vite configuration
│
└── packages/                  # Shared Monorepo Shared Packages
    ├── zod/                   # Single Source of Truth Zod schemas & types
    │   ├── src/               # Schema definitions (todo, category, comment)
    │   └── package.json       # @protask/zod package manifest
    ├── openapi/               # ts-rest contract definitions & spec generator
    │   ├── src/
    │   │   ├── contracts/     # ts-rest HTTP contract routes
    │   │   └── gen.ts         # Script building openapi.json
    │   └── package.json       # @protask/openapi package manifest
    └── emails/                # React Email templates
        ├── src/templates/     # JSX email templates
        └── package.json       # @protask/emails package manifest
```

<!-- KNOWLEDGE_GRAPH_NAVIGATION_START -->
---
### 🧭 Knowledge Graph & Navigation
| Dimension | Link / Reference |
| :--- | :--- |
| **Parent** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) |
| **Previous** | [ROADMAP.md](file:///home/logan78/Desktop/flux/ROADMAP.md) |
| **Next** | [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md) |
| **Children** | [adr/0001-modular-monolith-foundation.md](file:///home/logan78/Desktop/flux/adr/0001-modular-monolith-foundation.md), [api/openapi_v1_core.yaml](file:///home/logan78/Desktop/flux/api/openapi_v1_core.yaml), [database/postgres_master_schema.sql](file:///home/logan78/Desktop/flux/database/postgres_master_schema.sql) |
| **Dependencies** | [PRODUCT.md](file:///home/logan78/Desktop/flux/PRODUCT.md) |
| **Related Documents** | [ai/SECURITY.md](file:///home/logan78/Desktop/flux/ai/SECURITY.md), [ai/PERFORMANCE.md](file:///home/logan78/Desktop/flux/ai/PERFORMANCE.md) |
| **Navigation Hub** | [docs/INDEX.md](file:///home/logan78/Desktop/flux/docs/INDEX.md) \| [AGENTS.md](file:///home/logan78/Desktop/flux/AGENTS.md) |
---
<!-- KNOWLEDGE_GRAPH_NAVIGATION_END -->

