# ADR-01: Go + SQLc + Embedded SvelteKit Single-Binary Architecture

- **Status:** Accepted (Amended by `docs/IMPLEMENTATION_PLAN.md` v2.1)
- **Date:** 2026-11-01
- **References:** `IMS-BRD-1.0` (`NFR-01`, `NFR-02`, `NFR-03`), `IMS Full System Architecture.pdf` (Blueprint v1.0 §1–§2), `docs/IMPLEMENTATION_PLAN.md` (§3.1, §3.2 `D1`–`D3`, §6.1, §6.3)

---

## 1. Context

**MOAJ (Mentor of all Jacks)** targets mid-market engineering leads, mentors, and HR teams that require a self-hosted workflow and intern appraisal platform without complex microservice orchestration or heavy runtime memory overhead.

The source specifications (`IMS-BRD-1.0` and Blueprint v1.0) establish the following architectural goals:
- **`NFR-01` (Single-Binary Deployment):** Ship the HTTP API, asynchronous background worker, database migration runner, and static web frontend inside a single compiled executable (`moaj`) with no Node.js runtime required in production.
- **`NFR-02` (Relational + JSONB Persistence):** Use PostgreSQL 16 as the primary ACID relational database for strict workflow state transitions and append-only event ledgers, paired with `JSONB` columns for department-customizable 30/60/90-day evaluation rubrics.
- **`NFR-03` (Permissive Open-Source Stack):** Ensure all runtime dependencies and infrastructure components carry permissive open-source licenses (`MIT`, `BSD`, `Apache-2.0`, `PostgreSQL`) so the codebase can be distributed under both the root MIT License (Community Edition) and the commercial Enterprise EULA (`ee/EULA.md`).

During Phase 0 architecture review (`docs/IMPLEMENTATION_PLAN.md`), three additional foundation decisions (`D1`–`D3`) and a key operational scope amendment (Community observability) required formal recording.

---

## 2. Decision

### 2.1 Core Application & Persistence Stack
1. **Language & Runtime:** Use **Go** (`go 1.26.0` in `go.mod`, Decision `D2`) with the standard library `net/http` router, `log/slog` structured JSON logger, and `embed` package.
2. **Project Name & Module Path (`D1`):** Adopt the official project name **MOAJ (Mentor of all Jacks)** across documentation, CLI (`cmd/moaj`, `bin/moaj`, `bin/moaj-ee`), environment variables (`MOAJ_*` with `IMS_*` legacy aliases), enterprise license path (`/etc/moaj/moaj_enterprise.lic`), Prometheus metrics (`moaj_*`), and the Go module path `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ` in `go.mod` matching `https://github.com/MOAJ-Mentor-of-all-Jacks/MOAJ.git`.
3. **Database Access (`SQLc` + `pgx/v5`):** Write explicit, parameterized SQL queries in `db/query/*.sql` and compile them into type-safe Go code in `internal/db/` via **SQLc** (`sqlc.yaml`), rather than using reflection-heavy ORMs (`GORM`, `Ent`). Schema migrations live in `db/migrations/*.sql` and are embedded into the binary via `goose`.
4. **Embedded Frontend (`SvelteKit` Static Adapter):** Build the frontend in `ui/` (`moaj-ui`) using **SvelteKit** with `@sveltejs/adapter-static`, embedding the compiled static bundle directly into the Go binary (`NFR-01`).
5. **Single Binary Subcommands (`cmd/moaj`):** The `moaj` executable provides explicit operational entry points (`moaj serve`, `moaj version`, `moaj help`, and future `moaj worker`, `moaj migrate`) from one binary artifact while allowing the API server and background worker to run as separate processes in production Compose/container deployments.
6. **Open-Core Build-Tag Boundary (`!ee` vs. `ee`):** Isolate proprietary enterprise features (`SAML`, `OIDC`, `SCIM`, `PBAC`, `SIEM`, `RS256` licensing) inside `ee/` behind `//go:build ee` constraints and the `internal/enterprise` bridge interface so the default Community build (`//go:build !ee`) never imports `ee/`.

### 2.2 Resolution of `D1`, `D2`, and `D3` (Valkey over Redis)
- **`D1` (Project Naming & Module Identity):** Official project name is **MOAJ - Mentor of all Jacks** (`cmd/moaj`, `bin/moaj`, `MOAJ_*` env vars, `moaj_*` metrics) with Go module path `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ` matching the GitHub organization remote.
- **`D2` (Go Version & CI Matrix):** Pin `go 1.26.0` in `go.mod` and enforce Linux CI verification (`gofmt`, `go vet`, `staticcheck`, `govulncheck`, `go test -race` for both Community and `-tags ee` builds).
- **`D3` (Valkey `BSD-3-Clause` over Redis):** Where the original Blueprint referenced Redis 7 alongside `hibiken/asynq`, Redis 7.4+ switched to non-permissive `RSALv2`/`SSPLv1` (and Redis 8 added `AGPLv3`), conflicting with `NFR-03`. MOAJ adopts **Valkey** (`BSD-3-Clause`, `valkey/valkey:8.0-alpine`) as its supported in-memory store for `asynq` queues and scheduled tasks, preserving full wire-protocol compatibility under a permissive BSD license.

---

## 3. Specification Amendments Recorded in This ADR

### 3.1 Community Monitoring & Operability Amendment (`docs/IMPLEMENTATION_PLAN.md` §3.1 & §6.3)
- **Original Specification Gap:** The initial specifications treated external telemetry mainly as an Enterprise SIEM feature (`BR-EE-04`) and deferred operational packaging until the end of development.
- **Amended Rule:** Prometheus-compatible `/metrics` (`moaj_build_info`, `moaj_process_start_time_seconds`, `moaj_http_requests_total`, `moaj_http_request_duration_seconds`, `moaj_http_requests_in_flight`, protected by `MOAJ_METRICS_TOKEN` bearer authentication), structured JSON logs (`log/slog`), process liveness (`/healthz`) vs. dependency readiness (`/readyz`), versioned Grafana dashboards (`deploy/monitoring/grafana/`), Prometheus scrape/alert rules (`deploy/monitoring/prometheus/`), and operational runbooks (`docs/operations/RUNBOOKS.md`) are **core Community Edition features** starting in Phase 1.
- **Privacy Invariant on Telemetry:** Metric labels must remain strictly bounded (route templates, HTTP methods, status codes, operation names) and must never include `user_id`, `intern_id`, `task_id`, email addresses, or raw URLs.

---

## 4. Consequences

### Positive
- **Zero Production Runtime Bloat:** Operators deploy a single statically linked Go binary (`moaj`) plus PostgreSQL 16 and Valkey, with no Node.js or external ORM runtime dependencies.
- **Compile-Time SQL & Boundary Safety:** SQL query mismatches fail at code-generation/compile time, and `make boundary-check` proves the Community binary contains zero proprietary `ee/` packages.
- **License Cleanliness (`NFR-03`):** Adopting Valkey (`BSD-3-Clause`) and enforcing `make license-check` eliminates copyleft/SSPL contamination risks across both Community and Enterprise builds.
- **Day-One Operability:** Community operators gain immediate visibility into HTTP latency, expected `409`/`428` workflow responses vs. `5xx` faults, database pool health, and worker outbox backlogs.

### Trade-offs & Mitigations
- **SQLc Boilerplate for Dynamic Queries:** Complex analytical queries require carefully authored SQL in `db/query/` rather than dynamic query builders; mitigated by PostgreSQL 16 CTEs, window functions, and explicit repository interfaces in `internal/`.
- **Measured Footprint Verification (Phase 8 Gate):** Source targets (`<20 MB` binary, `<25 MB` idle RSS, `<5 min` install-to-first-board) will be explicitly benchmarked under documented conditions in Phase 8 rather than assumed without measurement.
