# MOAJ — Mentor of all Jacks

**MOAJ (Mentor of all Jacks)** is a self-hosted, open-core Intern Management System that turns everyday Kanban task movements into understandable, reviewable, and fair HR performance evidence.

Built as a single Go binary (`moaj`) with an embedded static UI (arriving in Phase 4/8) and strict privacy-by-design controls, MOAJ is engineered so that **Community Edition** is fully deployable, monitorable, and recoverable on its own.

---

## Architecture & Open-Core Boundaries

MOAJ follows a strict physical and compile-time separation between its MIT-licensed Community core and proprietary Enterprise extensions:

| Path | License | Responsibility |
| --- | --- | --- |
| `cmd/moaj/` | MIT | Process entry point (`serve`, `version`, `help`; `worker` and `migrate` in later phases). |
| `internal/config/` | MIT | Validated environment configuration (`MOAJ_*` with legacy `IMS_*` aliases), timeout bounds, and `_FILE` secret loading. |
| `internal/api/` | MIT | HTTP server, graceful drain lifecycle, security headers, request limits, and RFC 9457 Problem Details. |
| `internal/observability/` | MIT | Structured JSON logging (`log/slog`) and instance-scoped Prometheus metrics (`moaj_*`). |
| `internal/enterprise/` | MIT | Compile-time bridge interface (`//go:build !ee` vs `//go:build ee`). Neither stub grants paid capabilities without Phase 9 license verification. |
| `db/`, `ui/` | MIT | SQL migrations/queries (Phase 2) and static SvelteKit frontend (Phase 4). |
| `deploy/`, `docs/` | MIT | Prometheus/Grafana monitoring assets, operational runbooks, ADRs, and governance documentation. |
| `ee/` | Proprietary (`ee/EULA.md`) | Enterprise-only SAML/OIDC, SCIM, PBAC, SIEM, and RS256 `.lic` verification (Phase 9). Never imported by Community builds. |

---

## Implementation Status

| Phase | Scope | Status |
| --- | --- | --- |
| **Phase 0** | Repository foundation, licensing boundaries, ADRs, CI, and strict config loader | **Complete** |
| **Phase 1** | HTTP runtime (`moaj serve`), graceful drain, RFC 9457 errors, `/healthz`, `/readyz`, `/metrics`, Prometheus/Grafana baseline | **Complete** |
| **Phase 2** | PostgreSQL persistence, retention-aware schema, append-only ledger triggers, and SQLc adapters | Planned |
| **Phase 3** | Community authentication (Argon2id, sessions, CSRF, RBAC) and atomic first-run setup | Planned |
| **Phase 4** | First usable workflow slice: Kanban board, task versioning (`409`), and synchronous intercepts (`428`) | Planned |
| **Phase 5** | Background workers (`moaj worker`), Valkey/Asynq integration, and durable notification outbox | Planned |
| **Phase 6–10** | HR translation & review snapshots, GDPR retention erasure, Community release packaging, and Enterprise Edition | Planned |

See [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md) for full phase exit gates and specification amendments.

---

## Prerequisites & Building

- **Go**: `1.26+` (see `go.mod`; module `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ`)

### Build Community Edition (Default)

```bash
go build -o bin/moaj ./cmd/moaj
./bin/moaj version
# Output: moaj dev (community)
```

### Build with Enterprise Bridge Tag

```bash
go build -tags ee -o bin/moaj-ee ./cmd/moaj
./bin/moaj-ee version
# Output: moaj dev (enterprise-preview)
```

### Run Tests

```bash
go test ./...
go test -tags ee ./...
```

---

## Running Locally

### CLI Subcommands

- `moaj serve` *(default when no arguments are passed)*: Validates configuration and starts the HTTP runtime.
- `moaj version`: Prints the version and build edition (`community` or `enterprise-preview`) without loading environment configuration.
- `moaj help`: Prints command usage.

### Configuration Reference

MOAJ validates all settings at startup and fails fast on empty explicit variables or malformed values.

Secret-capable settings support `*_FILE` variants (up to 64 KiB, trailing `\r\n` stripped, mutually exclusive with the direct environment variable and any legacy alias). Note that `MOAJ_LICENSE_FILE` is already a filesystem path and does not use a `_FILE` suffix.

| Variable | Legacy Aliases | `_FILE` Support | Default | Description |
| --- | --- | --- | --- | --- |
| `MOAJ_ENV` | `IMS_ENV` | Yes | `development` | Deployment environment: `development`, `test`, or `production`. |
| `MOAJ_LOG_LEVEL` | `IMS_LOG_LEVEL` | Yes | `INFO` | Minimum structured JSON log level: `DEBUG`, `INFO`, `WARN`, or `ERROR`. |
| `MOAJ_APP_DOMAIN` | `IMS_APP_DOMAIN`, `APP_DOMAIN` | Yes | `localhost` | Public hostname (without scheme or path). Must not be `localhost` / loopback when `MOAJ_ENV=production`. |
| `MOAJ_HTTP_ADDR` | `IMS_HTTP_ADDR`, `HTTP_ADDR` | Yes | `127.0.0.1:8000` | TCP `host:port` bind address (`1–65535`). |
| `MOAJ_DATABASE_URL` | `IMS_DATABASE_URL`, `DATABASE_URL` | Yes | *(empty)* | PostgreSQL connection URL (`postgres://` or `postgresql://`). Validated when set; required starting in Phase 2. |
| `MOAJ_VALKEY_ADDR` | `IMS_VALKEY_ADDR`, `REDIS_ADDR` | Yes | `localhost:6379` | Valkey `host:port` address (`1–65535`). Required by workers starting in Phase 5. |
| `MOAJ_LICENSE_FILE` | `IMS_LICENSE_FILE`, `MINTERN_LICENSE_FILE` | No *(is a path)* | `/etc/moaj/moaj_enterprise.lic` | Path to the Enterprise `.lic` file (verified in Phase 9). |
| `MOAJ_METRICS_TOKEN` | `IMS_METRICS_TOKEN` | Yes | *(empty)* | Bearer token protecting `/metrics`. Must be `>= 32` non-whitespace characters when `MOAJ_ENV=production` or when `MOAJ_HTTP_ADDR` binds beyond loopback. |
| `MOAJ_HTTP_READ_TIMEOUT_SEC` | `IMS_HTTP_READ_TIMEOUT_SEC`, `HTTP_READ_TIMEOUT_SEC` | Yes | `10` | Maximum duration (seconds) for reading an entire request. |
| `MOAJ_HTTP_READ_HEADER_TIMEOUT_SEC` | `IMS_HTTP_READ_HEADER_TIMEOUT_SEC` | Yes | `5` | Maximum duration (seconds) allowed to read request headers. |
| `MOAJ_HTTP_WRITE_TIMEOUT_SEC` | `IMS_HTTP_WRITE_TIMEOUT_SEC`, `HTTP_WRITE_TIMEOUT_SEC` | Yes | `15` | Maximum duration (seconds) before timing out writes of the response. |
| `MOAJ_HTTP_IDLE_TIMEOUT_SEC` | `IMS_HTTP_IDLE_TIMEOUT_SEC`, `HTTP_IDLE_TIMEOUT_SEC` | Yes | `60` | Maximum duration (seconds) to wait for the next request on keep-alive connections. |
| `MOAJ_HTTP_SHUTDOWN_TIMEOUT_SEC` | `IMS_HTTP_SHUTDOWN_TIMEOUT_SEC`, `HTTP_SHUTDOWN_TIMEOUT_SEC` | Yes | `10` | Grace period (seconds) to drain active HTTP requests on `SIGINT` / `SIGTERM`. |
| `MOAJ_HTTP_MAX_BODY_BYTES` | `IMS_HTTP_MAX_BODY_BYTES` | Yes | `1048576` | Maximum allowed request body size in bytes (default 1 MiB). |

---

## Smoke Checks (`/healthz`, `/readyz`, `/metrics`)

With `moaj serve` running on the default loopback address (`127.0.0.1:8000`):

```bash
# 1. Process liveness check
curl -i http://127.0.0.1:8000/healthz
# HTTP/1.1 200 OK
# {"status":"ok"}

# 2. Readiness check (discloses Phase 1 scope)
curl -i http://127.0.0.1:8000/readyz
# HTTP/1.1 200 OK
# {"scope":"http-runtime","status":"ready"}

# 3. Prometheus metrics (unauthenticated on loopback in development when MOAJ_METRICS_TOKEN is unset)
curl -s http://127.0.0.1:8000/metrics | grep -E "^moaj_(build_info|process_start_time_seconds|http_)"
```

When `MOAJ_METRICS_TOKEN` is configured:

```bash
curl -s -H "Authorization: Bearer $MOAJ_METRICS_TOKEN" http://127.0.0.1:8000/metrics | head -n 25
```

---

## Documentation Index

- **Implementation Plan & Amendments**: [`docs/IMPLEMENTATION_PLAN.md`](docs/IMPLEMENTATION_PLAN.md)
- **Operations & Monitoring Runbook**: [`docs/operations/monitoring.md`](docs/operations/monitoring.md)
- **Monitoring Assets**:
  - Prometheus scrape config: [`deploy/monitoring/prometheus.yml`](deploy/monitoring/prometheus.yml)
  - Prometheus alert rules: [`deploy/monitoring/alerts.yml`](deploy/monitoring/alerts.yml)
  - Grafana dashboard (`moaj-phase1-runtime`): [`deploy/monitoring/grafana-dashboard.json`](deploy/monitoring/grafana-dashboard.json)
- **Architecture Decision Records**: [`docs/adr/`](docs/adr/)
- **Governance & Licensing**: [`docs/governance/`](docs/governance/), [`LICENSE`](LICENSE), [`CONTRIBUTING.md`](CONTRIBUTING.md), [`SECURITY.md`](SECURITY.md)
- **Source Specifications**: [`docs/specs/`](docs/specs/)
