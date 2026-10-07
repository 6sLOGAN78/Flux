# Observability

Flux owns a telemetry provider per API, redirector, worker and migrator process.
Applications emit vendor-neutral OTLP over HTTP; monitoring vendors are configured
only at the collector. No vendor key is required to run any role.

## Local collector

The optional Compose profile uses the official contrib collector 0.162.0,
with the immutable multi-platform manifest recorded in `tools.lock.json`.

```sh
docker compose --profile observability up -d otel-collector
docker compose --profile observability config --quiet
docker compose logs otel-collector
```

Validate the checked-in processor configuration with its exact locked binary:

```sh
docker run --rm \
  --mount "type=bind,src=$PWD/deploy/otel-collector.yaml,dst=/etc/otelcol-contrib/config.yaml,readonly" \
  otel/opentelemetry-collector-contrib:0.162.0@sha256:39923a8e431bd1f57be82411999d389fcfe40857492e4365456d97a4c1f74be6 \
  validate --config=/etc/otelcol-contrib/config.yaml
```

Host access is loopback only: HTTP `http://127.0.0.1:4318` and gRPC
`127.0.0.1:4317`. `FLUX_LOCAL_OTLP_HTTP_PORT` and `FLUX_LOCAL_OTLP_GRPC_PORT`
change published host ports. Receiver bindings inside the isolated container are
`0.0.0.0`; Compose does not publish them on LAN interfaces. Applications use the
HTTP base endpoint and append `/v1/traces`, `/v1/metrics`, `/v1/logs` themselves.

Configuration names are dot-nested and cannot be assigned using shell `export`.
For example, from `apps/backend`:

```sh
env 'FLUX_OBSERVABILITY.OTLP.ENABLED=true' \
  'FLUX_OBSERVABILITY.OTLP.ENDPOINT=http://127.0.0.1:4318' \
  go run ./cmd/api
```

The role still needs its usual required dependencies. The collector is never a
readiness dependency and is not listed in another service's `depends_on`.

## Export settings

| Environment name | Default | Accepted bounds |
| --- | --- | --- |
| `FLUX_OBSERVABILITY.OTLP.ENABLED` | `false` | Boolean |
| `FLUX_OBSERVABILITY.OTLP.ENDPOINT` | empty | HTTP(S) base URL; no credentials, query or fragment |
| `FLUX_OBSERVABILITY.OTLP.SAMPLE_RATIO` | `1` | 0 through 1, parent based |
| `FLUX_OBSERVABILITY.OTLP.QUEUE_SIZE` | `256` | 1 through 4096 |
| `FLUX_OBSERVABILITY.OTLP.BATCH_SIZE` | `64` | 1 through queue size |
| `FLUX_OBSERVABILITY.OTLP.EXPORT_TIMEOUT` | `2s` | 1ms through 30s |
| `FLUX_OBSERVABILITY.OTLP.EXPORT_INTERVAL` | `5s` | 1ms through 1m |
| `FLUX_OBSERVABILITY.ENVIRONMENT` | `development` | local, development, test, staging, production |

Disabled export ignores the endpoint, creates no remote exporters and preserves
sanitized structured stdout. Enabled export with an empty endpoint also has no
remote transport. Resource identities are fixed `flux.api`, `flux.redirector`,
`flux.worker`, `flux.migrator`, with an enumerated role and environment. The legacy
service-name setting remains accepted but cannot create arbitrary resource names.

Providers have bounded queues, batches, field counts and export deadlines. HTTP
exporter retries are disabled; telemetry failure does not block HTTP responses or
Redis task processing. Providers flush last, after HTTP/jobs and owned resources
drain, within the role's shared exit deadline. All independent signal providers
are attempted even if another fails. Export errors contain safe operation/stage
classification, never the remote response or opaque driver/provider message.

API readiness checks PostgreSQL and explicitly enabled producer Redis; worker
readiness checks Redis and the local email adapter; redirector readiness has no
undeclared infrastructure requirement. The migrator is a one-shot operation with
no readiness endpoint. It completes authoritative PostgreSQL migrations during
collector outages; a failed final telemetry flush can report exit status 1 after
the migration succeeded. Consult the safe migration version result/schema ledger
before retrying. Migration replay is deterministic.

## Privacy and correlation

HTTP ingress deletes all tracestate and baggage before extracting traceparent.
Only valid trace/span IDs, the approved sampled flag, and validated nonzero UUID
request/correlation IDs survive. Header names are `X-Request-ID` and
`X-Correlation-ID`. Span contexts and injected links are reconstructed with empty
tracestate. Logs and traces retain safe identity; metric labels exclude UUIDs.

The queue envelope contains version, traceparent, request ID and correlation ID.
Welcome-email recipients and names remain necessary private job payload data;
they never enter telemetry. Old tasks without metadata still decode. Legacy tasks
with unsafe metadata/headers are canonicalized through public enqueue/revoke
operations before email execution. The replacement ID is `safe:<original-id>`;
remaining retry budget, queue, timeout, deadline and retention are preserved.
Task identity changes, and delivery remains at least once. Retries persist only
the canonical propagation envelope and safe error messages.

Application exporters use value-aware closed allowlists for every signal,
including span names/status/events/links, logs, scope/resources and metric points.
Queries, destinations, raw paths, query strings, headers, recipients, names,
credentials and opaque error text are removed before stdout and OTLP transport.
Metrics use a closed route/method/status-class/role/outcome/dependency policy, with
128 SDK series per instrument. Unknown routes become `unmatched`; 200 distinct
visitor paths yield one HTTP metric series. Exemplars are disabled and removed.

The collector processes every signal in order:
`memory_limiter → transform/privacy → batch → exporter`. Its transform validates
allowed values, strips arbitrary resources/scopes/free text and private labels,
clears span tracestate, log body/severity/event text and metric descriptions/units,
and removes exemplars. In 0.162.0 OTTL exposes no individual SpanLink context:
the collector removes entire link and event collections to prevent unsupported
nested fields from leaking. Application-side captures prove nonzero HTTP-injected
and worker links have empty tracestate before that boundary. Trace/span/parent IDs,
sampled flags and validated UUID lineage survive collector export; links/events
are intentionally unavailable downstream.

The checked-in debug exporter is a local development sink **after** redaction.
Keep collector operational logging at info: debug OTTL evaluation can expose raw
transform contexts. Production deployments replace the debug exporter with their
vendor exporter after the same processors; authentication, TLS, retention and
vendor routing belong to that collector deployment. No production credentials
are embedded here.

## Deprecated compatibility

`FLUX_OBSERVABILITY.NEW_RELIC.LICENSE_KEY`, `.APP_LOG_FORWARDING_ENABLED`,
`.DISTRIBUTED_TRACING_ENABLED` and `.DEBUG_LOGGING` still parse. They initialize no
vendor client. Nondefault legacy values produce a constant deprecation warning
without serializing their values. Migrate to the OTLP endpoint and configure the
vendor at the collector, then remove these legacy values.

`FLUX_OBSERVABILITY.LOGGING.FORMAT=console` remains accepted for compatibility;
the output is consistently sanitized structured JSON. Other existing logging,
health and role configuration names remain supported.

## Executable verification

From `apps/backend`, with Docker available:

```sh
go test -race ./internal/app \
  -run 'Test(Collector|Observability|TelemetryOutage|TelemetryLeak|TracestateHTTPRedisOTLP|Cardinality)' \
  -count=1
```

Tests validate the exact digest-pinned binary, use real pinned PostgreSQL/Redis,
send real HTTP ingress, execute/retry real Asynq workers through a local fake email
transport, inspect original/replacement/retry payloads, capture actual HTTP OTLP
protobuf before and after the collector, and explicitly decode every
`Span.trace_state` and `Link.trace_state`. An independent dirty three-signal probe
tests collector defenses rather than relying on already sanitized application
data, including secret values under allowed keys, invalid scalar types, and
private metric exemplars. The final cardinality capture requires the cumulative
count of all 200 requests, not an early partial batch. Sentinels are scanned
without printing them in failures, including stdout,
collector debug output and generated OpenAPI/email artifacts. Outage tests cover
all four roles; narrow failing exporters check deadlines and complete provider
shutdown. Test collector host networking is used solely for loopback capture.

The transform syntax is verified against the
[official tagged processor documentation](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.162.0/processor/transformprocessor/README.md)
and the [official release](https://github.com/open-telemetry/opentelemetry-collector-releases/releases/tag/v0.162.0).
