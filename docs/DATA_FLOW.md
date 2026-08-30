# Data Flow

## 1. Link Shortening (Working)
```text
Frontend (useCreateLink)
 ↓ (POST /api/v1/links)
LinksHandler (Echo)
 ↓
LinkService
 ↓ (Generates random string instead of Base62)
LinkRepository
 ↓
PostgreSQL (INSERT INTO links)
```

## 2. Redirect Resolution (Working, Sub-optimal)
```text
Browser (GET /xyz123)
 ↓
RedirectHandler
 ↓
RedirectService
 ↓ (Skips Redis cache because it is nil)
PostgresRedirectRepository
## 2. Redirect Resolution (Cache-Aside)
```text
Browser (GET /xyz123)
 ↓
RedirectHandler
 ↓
Redis (Cache GET)
 ├── HIT → Return 301/302
 └── MISS → Query PostgreSQL
             ↓
            Redis (Cache SET, TTL: 24h)
             ↓
            Return 301/302
```

## 3. Analytics (Functional)
```text
Browser (GET /xyz123)
 ↓
RedirectHandler (Retrieves Link via PG)
 ↓
AnalyticsEvent generated
 ↓
Bounded Go Channel (AnalyticsPublisher)
 ↓ (Async)
Redis Stream (analytics:events)
 ↓
ClickHouse Consumer (XREADGROUP)
 ↓ (Batch Insert)
ClickHouse (MergeTree)
```
