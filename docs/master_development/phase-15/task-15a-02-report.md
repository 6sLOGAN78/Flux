# Phase 15A-02: Async Webhook Delivery Worker

## Architecture Overview
The 15A-02 implementation establishes an asynchronous delivery pipeline for canonical Flux events to customer webhook endpoints. The architecture is deliberately decoupled from the synchronous `link.redirect` ingestion path.

**Flow:**
1. Events are seamlessly pushed to the `analytics:events` Redis stream by the existing `RedisAnalyticsPublisher`.
2. A new `WebhookWorker` consumes these events using a dedicated `analytics-webhooks` consumer group.
3. Upon consuming a canonical `AnalyticsEvent`, the worker looks up active webhooks subscribed to the specific `EventType` for the event's `WorkspaceID`.
4. Matching webhooks are packaged with the event into a standard JSON payload.
5. The worker signs the raw JSON using HMAC-SHA256 and the customer's `whsec_` secret.
6. The delivery job is dispatched to a bounded local goroutine pool (`w.concurrency`) which executes an SSRF-safe HTTP POST.
7. Delivery outcomes (success, timeout, SSRF block) are logged minimally in the newly created `webhook_deliveries` PostgreSQL table.

## Key Decisions & Constraints Enforced
*   **Tenant Isolation**: Implemented strictly. No event can be dispatched to a webhook belonging to a different workspace, ensured by foreign key validation `GetActiveWebhooksForWorkspace`.
*   **SSRF Protection**: Custom `net.Dialer` explicitly intercepts private IP address connections before data transfer. Loopback, RFC1918, Link-local, and Metadata IP spaces are permanently blocked. Automatic HTTP redirects are explicitly disabled.
*   **Signature Scheme**: The `X-Flux-Signature` header leverages `v1=<hex-hmac-sha256(raw-body, secret)>`. Verification helper functions execute in constant time to prevent timing attacks.
*   **Database Minimization**: A `webhook_deliveries` table was created exclusively to serve the future 15A-03 Retry queue. No scheduling or Dead Letter Queue abstractions were built yet.

## Delivery Payload schema
```json
{
  "id": "evt_...",
  "type": "link.redirect",
  "created_at": "2026-09-02T12:00:00Z",
  "data": {
     // Canonical AnalyticsEvent
  }
}
```

## Future (15A-03)
The delivery worker records failed payloads. The subsequent task must implement the dead-letter-queue exponential backoff consumer which routinely picks up `webhook_deliveries` in a `status != success` state.

## Verification
* `TestWebhookWorker_Integration`: Confirmed complete end-to-end functionality utilizing `testcontainers` running PostgreSQL and Redis. Confirmed event matching exclusivity (unsubscribed events drop cleanly).
* `TestWebhookWorker_SSRF`: Validated rejection of protected IPs (`127.0.0.1`, `169.254.169.254`, etc.).
* `TestWebhookWorker_Signing`: Validated HMAC-SHA256 correctness and modification corruption checks.
