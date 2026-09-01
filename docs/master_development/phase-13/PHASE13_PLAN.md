# Phase 13 Implementation Plan: Multi-Touch Attribution

## 1. Readiness
Phase 13 is ready for implementation. Phase 11 (Campaigns) and Phase 12 (Custom Domains) are fully verified, providing the necessary ClickHouse table structures and routing logic to build upon.

## 2. Proposed Architecture & Data Flow
To achieve cross-domain attribution without relying on third-party cookies or the volatile `ip_hash`:
1. **URL Decoration:** The `RedirectHandler` will append `?flux_cid=<event_id>` to the target destination URL before issuing the 301/302 redirect.
2. **Client-Side Tracking Pixel:** Customers embed a lightweight Flux JS pixel on their website. This pixel detects `flux_cid` in the URL and stores it in a first-party cookie (`_flux_cids`, storing an array of recent clicks).
3. **Conversion Firing:** When a user checks out or signs up, the pixel calls `flux.track('conversion', { revenue: 99.99 })`. The pixel automatically attaches the stored `_flux_cids`.
4. **Ingestion API:** A new public endpoint `POST /api/v1/events/track` authenticates the request via a public Workspace Client ID and queues the conversion event to a Redis Stream.
5. **ClickHouse Persistence:** A new ClickHouse consumer batches these into a new `conversions` table.

## 3. Data Model Changes (ClickHouse)
**Table: `conversions`**
```sql
CREATE TABLE IF NOT EXISTS conversions (
    conversion_id String,
    workspace_id String,
    timestamp DateTime64(3, 'UTC'),
    conversion_name String,
    revenue Float64,
    currency String,
    click_ids Array(String), -- The stored flux_cids
    visitor_id String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (workspace_id, timestamp);
```

**Index on `analytics_events`:**
```sql
ALTER TABLE analytics_events ADD INDEX idx_event_id event_id TYPE bloom_filter(0.01) GRANULARITY 1;
```

## 4. Attribution Engine & API
**Endpoint:** `GET /api/v1/analytics/attribution?workspace_id=...&model=linear&from=...&to=...`
1. Fetch all `conversions` for the workspace in the timeframe.
2. Extract all unique `click_ids` from those conversions.
3. Fetch `analytics_events` matching those `click_ids` (fast due to bloom filter).
4. Map the data into the Go `Conversion` and `Touchpoint` structs.
5. Feed them into the existing `calculator.Calculate(conversions, model, halfLife)`.
6. Return the aggregated campaign revenue and conversion counts.

## 5. Supported Attribution Models
The product will support all 5 models currently in `calculator.go`:
* **First-Touch:** Good for understanding top-of-funnel brand awareness.
* **Last-Touch:** Good for understanding immediate conversion drivers.
* **Linear:** Good for long B2B sales cycles with multiple equal touchpoints.
* **Time-Decay:** Good for weighting recent interactions higher while acknowledging past touches.
* **Position-Based (U-Shaped):** Good for hybrid analysis (40% first, 40% last, 20% middle).

## 6. Frontend Plan
* **`AttributionPage.tsx`:** A dashboard allowing users to select an attribution model from a dropdown and view a table comparing `Campaign Name`, `Attributed Conversions`, and `Attributed Revenue`.
* Includes a date-range picker and model comparison toggles.

## 7. Security / Privacy Concerns
* **Public Tracking API:** The conversion API must be rate-limited and accept a public `Client ID` rather than a secret key, as it runs in the browser.
* **Data Isolation:** Conversions must be strictly tied to the `workspace_id` associated with the Client ID. When joining clicks, we must ensure `analytics_events.workspace_id == conversions.workspace_id` to prevent customers from querying other tenants' click IDs.
* **GDPR/CCPA:** First-party cookies storing random UUIDs (`flux_cid`) generally fall under necessary tracking if tied to an explicit conversion, but the JS pixel should provide a `flux.optOut()` method for consent banners.

## 8. Edge Cases Handled
* **Missing UTM / Direct Traffic:** If a user converts without any `flux_cid`, the conversion is recorded but attributed to "Direct / None".
* **Attribution Window:** The API will limit the `analytics_events` lookup to clicks that occurred within 90 days prior to the conversion timestamp.
* **Deleted Campaigns:** Because attribution is snapped to the event at click-time, deleted campaigns will still accurately reflect historical revenue.

## 9. Testing Plan
* **Unit Tests:** Verify `RedirectHandler` successfully appends `flux_cid` while respecting existing query parameters and hash fragments.
* **Integration Tests:** Verify ClickHouse conversion insertion and bloom filter utilization.
* **E2E Tests:** Simulate a multi-touch journey (Click A -> Click B -> Conversion) and assert the API returns 50/50 revenue splits on the Linear model.

## 10. Recommended Implementation Order
1. ClickHouse Migrations (Conversions table & Bloom Filter).
2. Redirect Handler Update (CID Appending).
3. Public Tracking API & Redis Stream (Ingestion).
4. Attribution API Engine (Querying & Calculator wiring).
5. Frontend Analytics UI.
