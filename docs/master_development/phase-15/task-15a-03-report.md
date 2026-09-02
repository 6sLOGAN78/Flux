# Phase 15A-03: Webhook Retry Queue & Dead-Letter Queue

## Architecture Overview
The 15A-03 implementation adds robust resilience to the outbound webhook pipeline by establishing a background retry queue powered directly by PostgreSQL and a dedicated `WebhookRetryWorker`.

**Lifecycle:**
1. If the initial `WebhookWorker` delivery fails (network error, timeout, 5xx, or 429), it explicitly records the full payload and sets the status to `retrying` with a mathematically scheduled `next_attempt_at`.
2. The `WebhookRetryWorker` continuously polls PostgreSQL for `status = 'retrying' AND next_attempt_at <= NOW()`.
3. Workers claim pending retries transactionally using `FOR UPDATE SKIP LOCKED`.
4. The HTTP attempt is repeated exactly as in 15A-02 (including SSRF protection and HMAC regeneration).
5. If the delivery succeeds, it transitions to `success`.
6. If the delivery fails, the `attempt_count` is incremented. If it hits `FLUX_WEBHOOK_MAX_RETRIES` (default 5), it is permanently transitioned to `dead_letter`. Otherwise, it calculates a new exponential backoff delay and is placed back in `retrying`.

## Key Decisions & Constraints Enforced
*   **Database Minimization**: Repurposed `webhook_deliveries` to support the queue. Added `payload JSONB` and `next_attempt_at TIMESTAMPTZ` with a filtered index. This avoids needing an entirely separate queue table or external queue broker.
*   **Tenant Isolation & Webhook Deactivation**: If a webhook is deactivated or deleted *while* a retry is pending in the queue, the retry worker explicitly re-validates the webhook status. Deleted/deactivated webhooks result in an immediate drop to `dead_letter`.
*   **At-Least-Once Delivery semantics**: We guarantee a best effort delivery. We do *not* guarantee exactly-once delivery. If the process crashes immediately after receiving a 200 OK from a customer but before updating the DB, the system may re-deliver the event. Customers should deduplicate using `X-Flux-Signature` or the JSON `id` property.

## Retry Formula & Jitter
The delay is calculated via an exponential backoff formula:
```text
Delay = InitialDelay * (2 ^ (Attempt - 1))
```
A 20% randomized jitter is then applied to the calculated delay to prevent retry storms:
```text
FinalDelay = Delay - (Delay * 0.2) + Random(0, Delay * 0.4)
```

## Error Classification
**Retryable Errors:**
* `5xx` Server Errors (except `501`)
* `429` Too Many Requests
* `408` Request Timeout
* Connection resets, DNS lookup failures, TCP timeouts

**Permanent Errors (Dead-Letter immediately):**
* `400`, `401`, `403`, `404`, `410`, `422` Client Errors
* SSRF blocks (Private IP resolution attempts)
* Webhook configuration deletion or deactivation

## Verification
*   `TestCalculateRetryDelay`: Unit tested mathematical boundaries and cap enforcement (`RetryMaxDelay`).
*   `TestIsRetryableError`: Unit tested HTTP classification code mapping.
*   `TestWebhookRetryWorker_Integration`: Confirmed complete end-to-end functionality utilizing `testcontainers`. Confirmed the worker correctly scales failures into `retrying`, succeeds when the endpoint comes back online, and correctly drops to `dead_letter` when attempt exhaustion is reached. Validated atomic DB locking mechanics.
