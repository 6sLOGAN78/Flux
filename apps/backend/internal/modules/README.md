# Backend Internal Subsystem Modules (`apps/backend/internal/modules`)

This directory contains the 16 decoupled business logic domain modules for the Flux backend:

---

## 📂 Subsystem Reference

| Module | Package | Responsibilities |
| :--- | :--- | :--- |
| **`abtest`** | `abtest` | Multi-variant A/B traffic split distribution & conversion rate evaluation. |
| **`analytics`** | `analytics` | Non-blocking `ClickHouseBatchWriter` & click event batching. |
| **`attribution`** | `attribution` | Multi-touch attribution models (First, Last, Linear, Time-Decay, U-Shaped). |
| **`billing`** | `billing` | Stripe webhook handling & subscription usage limit checking. |
| **`campaign`** | `campaign` | UTM parameter injection, sanitization & campaign stats aggregation. |
| **`deeplink`** | `deeplink` | Apple Universal Links (`aasa`) & Android App Links (`assetlinks.json`) builder. |
| **`domain`** | `domain` | Custom domain CNAME DNS verifier & ACME SSL certificate status tracker. |
| **`integration`** | `integration` | GA4 Measurement Protocol & Zapier event payload formatters. |
| **`notification`** | `notification` | Multi-channel alert dispatcher (Slack Block Kit & Discord Embeds). |
| **`ogmeta`** | `ogmeta` | Bot crawler detection (`IsSocialBot`) & OpenGraph card HTML generator. |
| **`publicapi`** | `publicapi` | Public API key validation, scope checking, and OAuth2 token issuer. |
| **`qr`** | `qr` | Custom SVG/PNG QR code rendering with logo placement. |
| **`queue`** | `queue` | Asynchronous worker pool, retries with exponential backoff & DLQ. |
| **`redirect`** | `redirect` | Dynamic smart routing rules (GeoIP, Device type, Time window matching). |
| **`tenant`** | `tenant` | Multi-tenant RBAC role hierarchy & custom permission override engine. |
| **`webhook`** | `webhook` | HMAC-SHA256 signed webhook delivery engine with retry backoff. |

---

## 🧪 Testing Subsystem Modules
Each module is tested independently:
```bash
go test -v ./internal/modules/...
```
