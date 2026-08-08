# Flux Backend API Service (`apps/backend`)

The **Flux Backend** is an Echo v4 REST API service written in Go 1.25. It provides short link redirection, cache management, analytics ingestion, tenant RBAC, and attribution modeling.

---

## 🏗 Package & Directory Architecture

Strictly following the **Single Responsibility Principle (SRP)** and layered internal architecture:

```text
apps/backend/
├── cmd/
│   ├── api/                   ← Echo HTTP server entrypoint (main.go)
│   └── cron/                  ← Asynchronous background worker entrypoint
├── internal/
│   ├── config/                ← Koanf environment configuration loader (config.go)
│   ├── database/              ← PostgreSQL pgxpool initialization & Tern migrations
│   ├── db/                    ← Primary/Replica cluster query router & cache invalidator
│   ├── errs/                  ← Domain error definitions & HTTP error status mappers
│   ├── handler/               ← Echo HTTP controllers (health, redirect, analytics)
│   ├── lib/                   ← Reusable helpers (AWS S3, email, crypto utils)
│   ├── logger/                ← Structured Zerolog logger setup & New Relic APM tracing
│   ├── middleware/            ← HTTP middlewares (Auth, Rate Limiting, Request ID)
│   ├── model/                 ← Pure domain DTOs & entities (link, category, redirect)
│   ├── modules/               ← 16 Decoupled Business Logic Subsystems
│   ├── repository/            ← SQL & Redis persistence contracts & implementations
│   ├── router/                ← Echo v4 system & API v1 route registration
│   ├── server/                ← Server lifecycle & dependency injection container
│   ├── service/               ← Framework-agnostic business logic services
│   ├── testing/               ← Testcontainers-go integration test suites
│   └── validation/            ← Request payload validation rules & slug format checking
└── pkg/
    ├── base62/                ← Base62 short-code encoder & decoder
    └── sqlerr/                ← PostgreSQL error code & unique constraint trapper
```

---

## 📦 16 Decoupled Business Modules (`internal/modules/`)

Every domain subsystem is self-contained within `internal/modules/`:

1. **`abtest`**: Weighted traffic splitter & conversion rate winner evaluator.
2. **`analytics`**: Non-blocking `ClickHouseBatchWriter` & click event batching.
3. **`attribution`**: Multi-touch attribution calculator (First, Last, Linear, Time-Decay, U-Shaped).
4. **`billing`**: Stripe webhook processing & usage limit checker.
5. **`campaign`**: UTM parameter injection, sanitization & campaign stats aggregator.
6. **`deeplink`**: iOS AASA & Android AssetLinks manifest generators.
7. **`domain`**: Custom domain CNAME DNS verifier & ACME SSL status tracker.
8. **`integration`**: GA4 Measurement Protocol & Zapier webhook formatters.
9. **`notification`**: Multi-channel notification dispatcher (Slack Block Kit & Discord Embeds).
10. **`ogmeta`**: Social crawler bot detector (`IsSocialBot`) & OpenGraph card generator.
11. **`publicapi`**: API key generation, scope checking, and OAuth2 token issuer.
12. **`qr`**: Custom SVG/PNG QR code renderer with logo embedding.
13. **`queue`**: Asynchronous worker pool, exponential backoff retries & DLQ handler.
14. **`redirect`**: Dynamic routing rules (GeoIP, Device type, Time window matching).
15. **`tenant`**: Multi-tenant RBAC role hierarchy & permission override engine.
16. **`webhook`**: HMAC-SHA256 signed webhook delivery engine.

---

## ⚡ Running & Development

### 1. Environment Configuration
Configuration is loaded from `.env` or environment variables:
```env
FLUX_SERVER.PORT="8080"
FLUX_DATABASE.HOST="localhost"
FLUX_DATABASE.PORT="5432"
FLUX_REDIS.ADDRESS="localhost:6379"
```

### 2. Start API Server
```bash
go run ./cmd/api/main.go
```

### 3. Run Test Suite
```bash
go test -v ./...
```
