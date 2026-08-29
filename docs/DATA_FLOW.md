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
 ↓ (SELECT destination_url FROM links)
RedirectHandler
 ↓ (HTTP 301/302)
Browser
```

## 3. Analytics (Broken)
```text
Browser (GET /xyz123)
 ↓
RedirectHandler
 ↓
(No click event emitted!)
```
