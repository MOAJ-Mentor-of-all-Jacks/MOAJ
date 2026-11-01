# MOAJ (Mentor of all Jacks) Operations & Monitoring Guide

This document defines the Phase 1 observability contract for **MOAJ (Mentor of all Jacks)**, including health and readiness semantics, Prometheus metric definitions, label-cardinality and privacy protections, structured JSON logging rules, and alert troubleshooting runbooks.

Monitoring is a first-class Community capability (per `docs/IMPLEMENTATION_PLAN.md` Section 6.3). Every operator can inspect service health, HTTP behavior, and runtime resource usage without an Enterprise license.

---

## 1. Health, Readiness, and Metrics Endpoints

| Endpoint | Methods | Purpose | Phase 1 Behavior |
| --- | --- | --- | --- |
| `/healthz` | `GET`, `HEAD` | Process liveness probe | Returns `200 OK` with `{"status":"ok"}` as long as the HTTP server process is alive—even while draining during graceful shutdown. Do not point load-balancer traffic routing solely at `/healthz` if graceful drain isolation is required. |
| `/readyz` | `GET`, `HEAD` | Traffic readiness probe | Returns `200 OK` with `{"status":"ready","scope":"http-runtime"}` during normal operation. Immediately switches to `503 Service Unavailable` (`application/problem+json`) when shutdown draining begins (`BeginShutdown()`). In Phase 2+, `/readyz` expands to verify PostgreSQL and Valkey connectivity. |
| `/metrics` | `GET`, `HEAD` | Prometheus & OpenMetrics scrape endpoint | Exposes application, Go runtime, and OS process metrics from an instance-scoped Prometheus registry. Protected by loopback/token authorization rules (see below). |

### `/metrics` Authorization Rules

1. **Token Configured (`MOAJ_METRICS_TOKEN` non-empty)**:
   - Every scrape request (including from loopback) must provide `Authorization: Bearer <token>`.
   - Token verification compares SHA-256 digests using `crypto/subtle.ConstantTimeCompare`.
   - Missing or invalid credentials return `401 Unauthorized` with `WWW-Authenticate: Bearer realm="metrics"` and an RFC 9457 Problem Details body.
2. **Token Unset (`MOAJ_METRICS_TOKEN` empty — allowed in `development` / `test` on loopback only)**:
   - Startup validation rejects an empty `MOAJ_METRICS_TOKEN` when `MOAJ_ENV=production` or when `MOAJ_HTTP_ADDR` binds to a non-loopback address (such as `0.0.0.0:8000` or `:8000`).
   - When `MOAJ_METRICS_TOKEN` is unset in local development, `/metrics` inspects the socket peer (`r.RemoteAddr`) directly—never `X-Forwarded-For` or `X-Real-IP`—and rejects non-loopback peers with `403 Forbidden`.

---

## 2. Metric Catalog (Phase 1 Baseline)

### Application Metrics (`internal/observability/metrics.go`)

| Metric Name | Type | Labels | Description |
| --- | --- | --- | --- |
| `moaj_build_info` | Gauge | `version`, `edition` | Constant `1` identifying the running binary version and edition (`community` or `enterprise-preview`). |
| `moaj_process_start_time_seconds` | Gauge | *(none)* | Unix epoch timestamp (seconds) recorded when the `Metrics` registry is initialized. Used for cross-platform uptime (`time() - moaj_process_start_time_seconds`) and restart detection (`changes(moaj_process_start_time_seconds[15m])`). |
| `moaj_http_requests_total` | Counter | `route`, `method`, `status` | Total completed HTTP requests. |
| `moaj_http_request_duration_seconds` | Histogram | `route`, `method` | HTTP request execution latency in seconds using standard Prometheus buckets (`0.005` to `10` seconds). |
| `moaj_http_requests_in_flight` | Gauge | *(none)* | Number of HTTP requests currently executing inside the server middleware. |

### Standard Runtime Collectors

Each `observability.Metrics` registry also registers:
- **`collectors.NewGoCollector()`**: Exports `go_goroutines`, `go_memstats_alloc_bytes`, `go_memstats_heap_inuse_bytes`, `go_gc_duration_seconds`, and Go runtime scheduler statistics.
- **`collectors.NewProcessCollector()`**: Exports `process_cpu_seconds_total`, `process_resident_memory_bytes`, and `process_open_fds` on supported OS platforms (Linux/procfs).

### Label Cardinality & Privacy Policy

To prevent cardinality explosions and protect intern/mentor privacy (`docs/IMPLEMENTATION_PLAN.md` Sections 4 and 6.3):
- **`route`**: Bounded to registered `http.ServeMux` route patterns (e.g., `"/healthz"`, `"/readyz"`, `"/metrics"`) or the constant `"unmatched"` when no pattern matches (`""` or catch-all `"/"`). Raw request URLs and path parameters are **never** emitted as metric labels.
- **`method`**: Normalized via `observability.MethodLabel` to standard HTTP verbs (`GET`, `HEAD`, `POST`, `PUT`, `PATCH`, `DELETE`, `OPTIONS`, `CONNECT`, `TRACE`) or `"OTHER"`. Arbitrary client-supplied method strings cannot create new time series.
- **`status`**: Numeric HTTP status code (`200`, `404`, `409`, `428`, `500`, etc.).
- **Prohibited Labels**: User IDs, intern/mentor names, emails, task IDs, cohort IDs, query strings, client IP addresses, and free-text notes (`OTHER` intercept explanations) must never appear in metric labels.

### Distinguishing Expected Workflow Outcomes from Faults

In MOAJ's workflow engine (Phase 4+):
- **`409 Conflict`** indicates a stale board version when two users mutate a task concurrently.
- **`428 Precondition Required`** indicates a synchronous workflow intercept (Trigger A backward move, Trigger B >3x SLA breach, or Trigger C blocked oscillation) prompting for a structured reason code.
- Neither `409` nor `428` is a system fault. Dashboards (`deploy/monitoring/grafana-dashboard.json`) and alert rules (`deploy/monitoring/alerts.yml`) separate `409|428` domain responses from other `4xx` client errors and `5xx` server faults.

---

## 3. Structured JSON Logging & Request Correlation

All logs are written as single-line JSON records to `stderr` via Go's `log/slog.JSONHandler` (`internal/observability/logger.go`).

- **Server-Generated Request IDs**: Every HTTP request receives a cryptographically random 128-bit (32-character hex) ID generated by `crypto/rand`. Any client-supplied `X-Request-ID` header is ignored and overwritten so untrusted input cannot forge log correlation or inject fields.
- **Correlation**: The generated ID is returned to the client in the `X-Request-ID` response header, included as `request_id` in RFC 9457 Problem Details responses (`application/problem+json`), and logged as `request_id` on every request log line.
- **Redaction by Construction**:
  - Request access logs emit only `request_id`, `method` (normalized), `route` (matched pattern or `"unmatched"`), `status`, `duration_ms`, and response `bytes`.
  - Raw URL paths, query parameters, request headers (`Authorization`, `Cookie`), request bodies, and panic payloads are never logged.
  - `config.Config` implements `slog.LogValuer` (`LogValue()`) so accidental logging of the configuration struct emits only safe booleans (`database_configured`, `metrics_auth_configured`) and non-secret metadata.

---

## 4. Local Monitoring Quickstart

1. Start the MOAJ HTTP runtime:
   ```bash
   go run ./cmd/moaj serve
   ```
2. Start Prometheus using the shipped configuration:
   ```bash
   prometheus --config.file=deploy/monitoring/prometheus.yml
   ```
3. Import `deploy/monitoring/grafana-dashboard.json` into Grafana and select your Prometheus data source.

---

## 5. Alert Runbooks

The following sections correspond to the `runbook_url` annotations in `deploy/monitoring/alerts.yml`.

### `MOAJTargetDown` {#moajtargetdown}

- **Condition**: `up{job="moaj"} == 0` for 1 minute.
- **Impact**: Either the `moaj` HTTP runtime process is down/unreachable, or Prometheus cannot authenticate/connect to `/metrics`.
- **Triage Steps**:
  1. Check whether the `moaj` process or container is running and inspect its `stderr` JSON logs for `"startup or server failed"`.
  2. Probe liveness directly from the host:
     ```bash
     curl -i http://127.0.0.1:8000/healthz
     ```
  3. If `/healthz` returns `200 OK` but Prometheus still reports `up == 0`, test the `/metrics` endpoint with your configured bearer token:
     ```bash
     curl -i -H "Authorization: Bearer $MOAJ_METRICS_TOKEN" http://127.0.0.1:8000/metrics
     ```
  4. Verify that `MOAJ_METRICS_TOKEN` (or `MOAJ_METRICS_TOKEN_FILE`) matches the token in Prometheus `authorization.credentials_file`.

### `MOAJHighServerFaultRate` {#moajhighserverfaultrate}

- **Condition**: HTTP `5xx` responses exceed 1% of total requests over 5 minutes (while request rate > 0.1 req/s). Expected `409`/`428` workflow responses and `4xx` client errors do not trigger this alert.
- **Impact**: Users are experiencing internal server errors (`500`) or service unavailability (`503`).
- **Triage Steps**:
  1. Check whether the instance is stuck in a prolonged drain state (`GET /readyz` returning `503`).
  2. Filter `stderr` JSON logs for `"status":500`, `"status":503`, or `"msg":"HTTP handler panicked"`:
     ```bash
     jq 'select(.status >= 500 or .msg == "HTTP handler panicked")' /var/log/moaj/moaj.log
     ```
  3. Use the `request_id` from affected client error responses (`application/problem+json`) to locate the exact server log entry.
  4. In Phase 2+, verify PostgreSQL connectivity and pool saturation metrics.

### `MOAJHighRequestLatencyP95` {#moajhighrequestlatencyp95}

- **Condition**: 95th percentile latency on any bounded `route` exceeds `500ms` over 5 minutes.
- **Impact**: Degraded responsiveness on API or health endpoints; risk of hitting `MOAJ_HTTP_WRITE_TIMEOUT_SEC` (default `15s`).
- **Triage Steps**:
  1. Open the **HTTP Request Latency (p50 / p95 / p99 by Route)** panel in Grafana (`uid: moaj-phase1-runtime`) to identify which `route` is slow.
  2. Inspect **Process CPU & GC Pause Duration** and **Memory Allocation & Resident Heap** to rule out host CPU throttling or memory pressure.
  3. Correlate with **In-Flight HTTP Requests** to determine whether latency is caused by traffic saturation or downstream dependency contention (PostgreSQL/Valkey in Phases 2 and 5).

### `MOAJFrequentProcessRestarts` {#moajfrequentprocessrestarts}

- **Condition**: `changes(moaj_process_start_time_seconds[15m]) > 2`.
- **Impact**: Crash loop, OOM kills, failing liveness probes, or repeated configuration/deployment restarts.
- **Triage Steps**:
  1. Inspect container orchestrator / systemd exit codes for the `moaj` service (`OOMKilled`, `exit code 1`, `SIGTERM`).
  2. Check JSON logs immediately preceding each startup message (`"msg":"HTTP runtime listening"`) for `"msg":"startup or server failed"` or `"msg":"HTTP shutdown started"`.
  3. Compare `process_resident_memory_bytes` and `go_memstats_heap_inuse_bytes` against container memory limits.

### `MOAJHighInFlightRequests` {#moajhighinflightrequests}

- **Condition**: `moaj_http_requests_in_flight > 100` for 5 minutes.
- **Impact**: Request queuing or stalled handlers holding connections open.
- **Triage Steps**:
  1. Compare `moaj_http_requests_in_flight` with `go_goroutines` and incoming request rate (`rate(moaj_http_requests_total[5m])`).
  2. Check whether slow clients or large payloads are holding connections; verify `MOAJ_HTTP_READ_HEADER_TIMEOUT_SEC` (default `5s`), `MOAJ_HTTP_READ_TIMEOUT_SEC` (default `10s`), and `MOAJ_HTTP_WRITE_TIMEOUT_SEC` (default `15s`).
  3. In Phase 4+, note that long-lived SSE streams (`/api/v1/events`) will be tracked via dedicated realtime stream gauges so ordinary HTTP saturation remains distinct from active event subscriptions.
