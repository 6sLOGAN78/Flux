# Phase 13 Audit: Multi-Touch Attribution

## 1. Existing Analytics Event Model
Currently, `analytics_events` captures link clicks with properties like `event_id`, `timestamp`, `workspace_id`, `campaign_id`, and UTM dimensions. 
However, it only stores `event_type = 'link.redirect'`. There is no mechanism to store conversion events or revenue.

## 2. Visitor Identification & ip_hash
The existing model relies on `ip_hash` for unique visitor counting. 
**Audit Finding:** `ip_hash` is entirely insufficient for Multi-Touch Attribution (MTA). MTA windows span up to 90 days. An IP address is highly volatile (DHCP, mobile networks) and often shared (NAT). Using it to link a click today with a conversion 30 days from now will yield severely corrupted attribution data.

## 3. Link -> Campaign Relationships
The system successfully resolves `campaign_id` at the time of the click (click-time attribution snapping) and writes it into the `analytics_events` row. This provides a solid foundation for attribution.

## 4. ClickHouse Schema & Query Architecture
`analytics_events` is ordered by `(workspace_id, link_id, timestamp)`. 
**Audit Finding:** Grouping clicks into a user journey requires querying by a click identifier. Since the primary key does not include `event_id`, querying `WHERE event_id IN (...)` will result in a full partition scan for that workspace. We will need a Data Skipping Index (Bloom Filter) on `event_id` to make these lookups efficient during attribution resolution.

## 5. Existing Code (`calculator.go`)
`apps/backend/internal/modules/attribution/calculator.go` implements 5 models (First-Touch, Last-Touch, Linear, Time-Decay, Position-Based) entirely in-memory. It expects to be fed an array of `Conversion` structs, each containing a nested array of `Touchpoint` structs.
**Audit Finding:** The logic works perfectly. The architectural challenge is purely data retrieval: fetching the `conversions` and their linked `analytics_events` efficiently from ClickHouse, formatting them into these structs, and running them through the calculator.

## 6. Edge Cases & Missing Pieces
* **Attribution Bridging:** There is currently no way to link a redirect on `flux.to` with a conversion on a customer's destination domain (`customer.com`). Browsers block cross-site cookies.
* **Missing Feature:** We must append a unique identifier (e.g., `?flux_cid=<event_id>`) to the destination URL during the redirect. The customer must install a JavaScript tracking snippet on their site to capture this CID and send it back during a conversion event.

## 7. What is Broken / Missing
* Missing `conversions` table in ClickHouse.
* Missing URL parameter appending in `RedirectHandler`.
* Missing public REST API for ingesting conversions from client browsers.
* Missing Data Skipping Index on `analytics_events.event_id`.
