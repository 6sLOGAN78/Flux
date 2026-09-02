# Configuration

The Flux backend uses a strongly-typed, 12-factor compliant configuration system based on `koanf` and `validator`.

## Architecture

```text
.env (Development)
  ↓
godotenv/autoload (Loads .env into process environment)
  ↓
OS environment variables (Source of truth)
  ↓
koanf environment provider (Parses FLUX_ prefix)
  ↓
Strongly typed Config struct (Nested settings)
  ↓
validator (Enforces required fields / security policies)
```

## Naming Convention

All application environment variables MUST be prefixed with `FLUX_`.
Configuration boundaries are represented by dots (`.`).

Examples:
- `FLUX_SERVER.PORT` maps to `Config.Server.Port`
- `FLUX_DATABASE.HOST` maps to `Config.Database.Host`

## Struct Definition

The root configuration structure is heavily modularized:
- `PrimaryConfig`: Controls runtime identity (`FLUX_PRIMARY.ENV`).
- `ServerConfig`: Defines listener ports, CORS, and timeouts.
- `DatabaseConfig`: Postgres connection settings.
- `RedisConfig`: Caching layer and streaming config.
- `ClerkConfig`: Clerk JWKS/Authentication configuration.
- `StripeConfig`: Webhooks and Billing integration configuration.

## Development vs Production Behavior

- **Development**: The application automatically loads `.env` located in `apps/backend` via `godotenv`. Missing configurations or unsafe defaults (like `whsec_test_secret`) are acceptable for testing purposes.
- **Production**:
  - The process environment defines the variables (no physical `.env` required).
  - Production logic employs a strict **Fail-Closed Strategy**. If `FLUX_PRIMARY.ENV=production` is detected, security-sensitive configurations (like `FLUX_STRIPE.WEBHOOK_SECRET`) MUST be correctly populated. A generic test key will forcibly crash the application during startup to prevent exposure or unauthenticated webhook ingestion.

## Example Configuration

Check `apps/backend/.env.example` for the full schema of available keys.
