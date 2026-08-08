# Flux — Multi-Tier Link Management & Enterprise Attribution Platform

[![Go Version](https://img.shields.io/badge/Go-1.25.0-00ADD8?style=flat&logo=go)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=flat&logo=redis)](https://redis.io/)
[![ClickHouse](https://img.shields.io/badge/ClickHouse-24.1-FFCC00?style=flat&logo=clickhouse)](https://clickhouse.com/)
[![Docker](https://img.shields.io/badge/Docker-Enabled-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Flux** is a high-performance, production-grade open-source URL shortening, dynamic link routing, real-time analytics, SaaS multi-tenancy, and multi-touch attribution platform built in Go.

Designed as a clean **Modular Monolith**, Flux delivers sub-millisecond link redirection, high-throughput time-series analytics, custom domain DNS automation, QR code generation, and enterprise marketing attribution.

---

## 🌟 Feature Overview

### 1. Core Link Management
- **High-Speed Base62 URL Shortener**: $62^6 \approx 56.8 \text{ billion}$ unique collision-resistant short codes.
- **Cache-Aside Redirect Engine**: Sub-millisecond $HTTP\ 301/302$ redirects using Redis caching with database fallback.
- **Link Controls**: Expiration dates (**HTTP 410 Gone**), soft-deletion, status disabling (**HTTP 403**), and bcrypt password protection (**HTTP 401**).
- **Protocol & Slug Validation**: Rejects dangerous schemes (`javascript:`, `data:`, `file:`) and guards system-reserved slugs (`login`, `api`, `docs`, `health`, etc.).

### 2. Growth & Marketing Engine
- **Multi-Touch Attribution Engine**: First-Touch, Last-Touch, Linear, Time-Decay (7-day half-life), and Position-Based (U-Shaped 40/40/20) revenue & conversion attribution models.
- **A/B Test Traffic Splitter**: Weighted multi-variant split testing with automatic conversion winner evaluation.
- **Smart Redirect Router**: Dynamic routing rules based on GeoIP country, device type (Mobile/Desktop/Tablet), and time windows.
- **Deep Linking Engine**: Apple Universal Links (`apple-app-site-association`) and Android App Links (`assetlinks.json`) generator with fallback web handlers.
- **Custom Domains & SSL**: Automated CNAME and TXT DNS challenge verification with ACME TLS/SSL certificate lifecycle checks.
- **Social Preview Cards (OG Meta)**: Dynamic OpenGraph & Twitter Card HTML generator with social crawler bot detection (`IsSocialBot`).
- **QR Code Studio**: Custom SVG/PNG QR code rendering with custom hex colors, logo embedding, and async batch creation.
- **UTM Campaign Builder**: Automatic parameter injection, sanitization, and campaign performance aggregation.

### 3. SaaS & Enterprise Capabilities
- **Multi-Tenant Workspaces & RBAC**: Organization hierarchies, workspace scopes, and role permissions (`owner`, `admin`, `member`, `viewer`).
- **Stripe Billing Integration**: Tier limit enforcement (Free/Pro/Enterprise) and webhook event processors.
- **Public API & OAuth Tokens**: API key issuance (`flx_live_...`), scope checks, and OAuth2 bearer token generation.
- **Webhook Delivery Engine**: Asynchronous HMAC-SHA256 event delivery with exponential backoff retries.
- **Multi-Channel Notifications**: Real-time alert dispatcher supporting Slack Block Kit, Discord Embeds, and in-app notifications.
- **Redis Rate Limiting**: Distributed Sliding Window ZSET rate limiter returning **HTTP 429 Too Many Requests**.

---

## 🏗 System Architecture

Flux follows a decoupled **Modular Monolith** architecture with a **Polyglot Persistence Layer**:

```
                                  Internet Traffic
                                         │
                                   HTTPS (Port 443)
                                         │
                               Nginx / Load Balancer
                                         │
                 ┌───────────────────────┴───────────────────────┐
                 │                                               │
           React Frontend                                  Go Echo Backend
         (apps/frontend)                                  (apps/backend)
                                                                 │
                 ┌───────────────────────┬───────────────────────┤
                 ▼                       ▼                       ▼
            PostgreSQL                 Redis                 ClickHouse
       (Transactional DB)        (Cache & Rate Limit)    (Time-Series Click Stream)
```

### 💾 Polyglot Database Design
- **PostgreSQL**: Stores transactional ACID entities (`links`, `users`, `tenants`, `custom_domains`, `campaigns`, `webhooks`).
- **Redis**: Provides sub-millisecond Cache-Aside link resolution, Sliding Window ZSET rate limiting, and L1/L2 pub/sub cache invalidation.
- **ClickHouse**: High-throughput time-series analytics engine using `ReplacingMergeTree` to store append-only click event facts without locking PostgreSQL.

---

## 📁 Repository Structure

```text
flux/
├── apps/
│   ├── backend/               ← Go 1.25 Echo v4 REST API server & domain modules
│   │   ├── cmd/
│   │   │   ├── api/           ← Server entrypoint (main.go)
│   │   │   └── cron/          ← Background job worker entrypoint
│   │   ├── internal/          ← Layered backend packages & domain modules
│   │   │   ├── config/        ← Application configuration loader
│   │   │   ├── database/      ← PostgreSQL pool & Tern migration scripts
│   │   │   ├── db/            ← Read replica cluster query router
│   │   │   ├── errs/          ← Domain error types & HTTP error mappers
│   │   │   ├── handler/       ← Echo HTTP controllers (redirect, health, analytics)
│   │   │   ├── lib/           ← Common utilities & AWS helpers
│   │   │   ├── logger/        ← Structured Zerolog logger & New Relic tracing
│   │   │   ├── middleware/    ← Echo middlewares (Auth, Rate Limit, Tracing, Request ID)
│   │   │   ├── model/         ← Pure domain DTOs & entities (link, category, redirect)
│   │   │   ├── modules/       ← 16 Decoupled Business Logic Modules
│   │   │   ├── repository/    ← Data persistence interfaces & implementations
│   │   │   ├── router/        ← System & V1 API route registars
│   │   │   ├── server/        ← Echo server lifecycle initializer
│   │   │   ├── service/       ← Framework-agnostic business logic services
│   │   │   ├── testing/       ← Testcontainers-go integration test helpers
│   │   │   └── validation/    ← Input validation utilities
│   │   ├── pkg/
│   │   │   ├── base62/        ← Base62 short-code encoder & decoder
│   │   │   └── sqlerr/        ← PostgreSQL error code & unique constraint trapper
│   │   └── static/            ← OpenAPI documentation (openapi.html, openapi.json)
│   └── frontend/              ← React 19 / Vite / TypeScript dashboard web app
│
├── api/                       ← OpenAPI v1-v4 & AsyncAPI specification contracts
├── database/                  ← PostgreSQL master schema & ClickHouse analytics schema
├── ops/                       ← Infrastructure specs (Docker, Kubernetes, Terraform, CI/CD)
├── packages/                  ← Workspace shared npm packages (@flux/openapi, @flux/zod)
└── docker-compose.yml         ← Multi-container local orchestration
```

---

## ⚡ Quick Start & Development Setup

### 1. Prerequisites
- **Go**: `1.25+`
- **Docker**: Docker Desktop or Docker Engine
- **Node.js / Bun**: `Node 20+` or `Bun 1.1+`

### 2. Local Infrastructure (Docker Compose)
Start PostgreSQL and Redis containers locally:

```bash
docker compose up -d postgres redis
```

### 3. Run the Go API Backend
```bash
cd apps/backend
go run ./cmd/api/main.go
```
The server will start on `http://localhost:8080`.

### 4. Interactive OpenAPI / Swagger Documentation
Open your browser at:
👉 **[http://localhost:8080/docs](http://localhost:8080/docs)**

---

## 🧪 Testing & Code Quality

Per our strict testing guidelines, all features are covered by unit and integration test suites.

```bash
# Run all Go backend unit & module tests
cd apps/backend
go test -v ./...
```

### Run Static Analysis Linter
```bash
cd apps/backend
go vet ./...
```

---

## 🔌 API Reference Overview

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/:slug` | Resolves short code via Redis Cache-Aside Engine & redirects |
| `GET` | `/health` | Live PostgreSQL connection health check |
| `GET` | `/docs` | Interactive Swagger/OpenAPI Documentation UI |
| `POST` | `/api/v1/urls` | Shortens a URL (accepts custom alias & expiration) |
| `GET` | `/api/v1/urls` | Lists user links (with search, category filter, and offset pagination) |
| `PATCH` | `/api/v1/urls/:id` | Updates link destination, title, or category |
| `DELETE` | `/api/v1/urls/:id` | Soft-deletes a link (preserves ClickHouse click history) |
| `GET` | `/api/v1/analytics/summary` | Aggregated platform performance summary |
| `GET` | `/api/v1/analytics/links/:id` | Time-series click metrics for a specific link |
| `GET` | `/api/v1/enterprise/attribution` | Multi-touch attribution calculation by model |

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
