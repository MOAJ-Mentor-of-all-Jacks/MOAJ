# MOAJ (Mentor of all Jacks) - Production Implementation Plan

Version: 2.1  
Updated: 2026-11-01  
Status: Phases 0 and 1 are complete and verified against their exit gates below; Phase 2 is next.

## 1. Purpose and authority

Build a self-hosted, open-core intern management system that turns everyday task movements into understandable, reviewable HR evidence. Community must be deployable, recoverable, and monitorable at its first release.

This document replaces the earlier pasted implementation plan. The four PDFs in `docs/specs/` remain the source requirements, subject to the explicit amendments recorded here:

- `IMS Business Requirements Document.pdf` (IMS-BRD-1.0).
- `IMS Feature Specification.pdf`.
- `IMS Full System Architecture.pdf` (Blueprint v1.0).
- `IMS Licensing and Governance Manual.pdf` (IMS-GOV-01).

Do not silently resolve conflicts by mixing incompatible requirements. Record the decision and update the relevant ADR before implementing the affected behavior. The original PDFs are preserved as source records.

The user has authorized two implementation agents to complete the first two phases (Phase 0 and Phase 1). This supersedes the earlier user-types-the-code learning workflow for those phases. Changes remain reviewable and must pass their phase gates. Later phases are outside this implementation request.

## 2. Verified starting point

Repository inspection on 2026-10-31 established:

| Area | Current state |
| --- | --- |
| Git | Already initialized; commit `b4afedd` (`first commit`). Do not run a new repository setup unnecessarily. |
| Go | `module mintern`, `go 1.22`; installed toolchain reports `go1.27.2 windows/amd64`. |
| Entry point | `cmd/ims/main.go` contains practice TCP code with syntax and identifier errors. |
| Config | `go vet ./internal/config` passes; formatting, validation, secret loading, and the `.llc` license extension need correction. |
| Empty Go files | 14 zero-byte files. `go build ./...` fails on empty package files. |
| Infrastructure and UI | Scaffold files exist, but migrations, queries, CI, container setup, and SvelteKit are not implemented. |
| Documentation and licensing | README, CONTRIBUTING, root LICENSE, existing ADRs, and `ee/EULA.md` are empty. |
| Specifications | Root PDFs and `docs/specs/` copies are byte-identical. Consolidate references on `docs/specs/`. |

These are observations, not completed phase gates. Recheck the working tree before each implementation step.

## 3. Decisions and specification amendments

### 3.1 Settled product and engineering direction

| Topic | Revised rule | Supersedes |
| --- | --- | --- |
| Backward transitions | Require a reason on the first and every subsequent backward move. | Trigger A's second-bounce threshold; reconciles the BRD's 100% attribution KPI. |
| Exceptional reasons | Add `ACCIDENTAL_MOVE` and `OTHER`; `OTHER` requires a bounded explanation. | Four-code-only taxonomy. |
| Stale boards | Compare a task-wide version on every mutation; reject stale submissions immediately and return authorized current task state. | Checking only the current column or waiting for a stale timeout. |
| Notification reliability | Atomically commit the business change and durable notification intent in PostgreSQL. Deliver asynchronously. | Insert a deduplication row, then send without durable delivery state. |
| HR presentation | Lead with accomplishments, support needs, exclusions, and concise supporting evidence. Keep raw events and versions in a drill-down. | Presenting telemetry as the primary HR interface. |
| Evidence history | Version definitions that affect interpretation and preserve published review snapshots, with explicit amendments. | Recalculating old reviews from mutable present-day definitions. |
| Monitoring | Community includes metrics, structured logs, dashboards, alert examples, and documented operational procedures. | Treating monitoring as a late or enterprise-only addition. |
| Community release | Installation, workers, backup/restore, upgrades, monitoring, and load validation are part of its release gate. | Releasing after UI completion while packaging remains deferred. |
| Legal drafting | Compare reusable existing open-core templates, then adapt a selected template to MOAJ's requirements. | Treating an unselected template or placeholder as a completed legal package. |

### 3.2 Working defaults and decisions still needed

Working defaults guide design; they are not claims that the user has already approved every detail. Resolve material open decisions at the indicated phase, without blocking unrelated work.

| ID | Default or open decision | Resolve by |
| --- | --- | --- |
| D1 | Resolved: official project name is **MOAJ - Mentor of all Jacks** (`cmd/moaj`, `bin/moaj`, `bin/moaj-ee`, `MOAJ_*` env vars with `IMS_*` legacy aliases, `/etc/moaj/moaj_enterprise.lic`, and `moaj_*` Prometheus metrics), hosted at `https://github.com/MOAJ-Mentor-of-all-Jacks/MOAJ.git` with Go module path `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ`. | Phase 0 |
| D2 | Select and pin a currently supported Go release and CI matrix. The installed version is evidence about this machine, not a support-policy check. | Phase 0 |
| D3 | Use Valkey to align with BRD NFR-03; pin a maintained version and verify Asynq compatibility. | Phase 0; integration proof in Phase 5 |
| D4 | Community authentication: email/password with Argon2id, database sessions, secure cookies, CSRF protection, and rate limits. | Phase 3 |
| D5 | One organization per installation, with multiple departments and cohorts. | Phase 2 |
| D6 | Add missing users, departments, cohorts, membership, intercepts, and blocked-column fields. Add tables incrementally with their features. | Phase 2 onward |
| D7 | Department-defined skill tags with a task/tag join table; each task requires at least one valid tag. | Phase 4 |
| D8 | Use `time_in_prev_state_ms` internally; document the BRD field-name mapping. | Phase 2 |
| D9 | Retention-bounded personnel evidence and controlled erasure; retain only assessed aggregate benchmarks beyond retention. This replaces the indefinite identifiable ledger model. | Design in Phase 2; proof in Phase 7 |
| D10 | Hashing or salt destruction alone is not an anonymity guarantee. Aggregate grouping, small-group suppression, and re-identification assessment remain to be specified. | Design in Phase 2; finalized before Phase 7 gate |
| D11 | Configurable SLA sweep cadence; default 15 minutes. One notification intent per qualifying state-entry event and channel/recipient policy. | Phase 5 |
| D12 | Review checkpoints use the intern's cohort-membership join date, falling back to cohort start. Store timezone and define calendar-day boundaries explicitly. | Phase 6 |
| D13 | Fairness exclusion intervals and revision exclusions need worked examples and approved calculation rules; see section 5. | Before Phase 6 aggregation |
| D14 | Blocked oscillation fires on the third entry within a rolling seven-day window; specify boundary inclusivity in tests. | Phase 4 |
| D15 | Shared 1-on-1 agenda, encrypted private mentor notes, and a human-authored return-offer decision. | Phase 6 |
| D16 | Use a generated installation identifier for license binding. Document clone/restore/rehosting behavior: a copied database also copies this identifier. It is not a hardware or anti-cloning guarantee. | Phase 9 |
| D17 | MIT core direction retained. Choose actual copyright/licensor identity, docs/deploy license mapping, CLA mechanism, enterprise template, and jurisdiction. | Core boundaries in Phase 0; complete EE terms before Phase 9 release |

## 4. Privacy and retention contract

### 4.1 What the application can promise

During a configured retention period, personnel evidence is identifiable and access-controlled. At expiry, a controlled process removes expired personal records and links. Long-term benchmarks use assessed aggregate data, not indefinitely retained individual timelines with renamed users.

Do not describe UUID replacement, hashing, or salt destruction as proof of anonymity. A rare task, exact timestamp, tiny cohort, or recognizable note can identify someone without a name. No unconditional anonymization guarantee is made by this plan.

The ledger is append-only for ordinary application operations during retention. A restricted, audited retention process may erase expired records. This is an explicit amendment to indefinite immutability. Retention exceptions and any legal hold require an explicit policy; neither a blanket hold nor silent indefinite retention is a default.

### 4.2 Data lifecycle

| Data | During retention | At expiry |
| --- | --- | --- |
| Memberships, task assignments, detailed transitions, intercepts, evaluations | Restricted personnel evidence | Erase expired records and identifying links according to the dependency plan. |
| Task text, private notes, `OTHER` explanations | Treat as potentially identifying; minimize collection | Purge expired content, including generated snapshots containing it. |
| User account | May serve multiple active memberships | Do not erase an active account merely because one cohort expires; remove expired membership evidence separately. |
| Department benchmarks | Maintain a documented calculation and privacy policy | Keep only approved aggregates; suppress or combine small/rare groups. |
| Logs, job payloads, caches, notification outbox | Minimize personal content and define short, explicit retention | Expire or scrub; prevent retries from recreating erased information. |
| Backups | Restricted, encrypted where supported, with bounded lifetime | Expire on schedule; replay erasure records after restore before reopening service. |
| External notifications and downloaded exports | Minimize shared content and document destinations | Explain system boundaries; MOAJ cannot silently erase independently held copies. |

Use separate application, migration, and retention privileges. Database triggers and permissions must prevent the normal application role from bypassing append-only rules. Any privileged erasure function needs a fixed safe search path, restricted execution grants, and tests using the actual runtime roles.

The erasure audit must prove operation, time, scope, and outcome without retaining the personal content it erased. Define a protected, minimal erasure record for restore reconciliation with its own retention policy.

Acceptance requires a data inventory, representative re-identification checks, idempotent erasure, cross-cohort safety, and backup-restore verification. Suppression can change displayed benchmarks; do not require every pre-erasure number to remain visible or unchanged.

## 5. Workflow and HR evidence contract

### 5.1 Transitions, intercepts, and corrections

- Every task mutation increments a task-wide `version`.
- A transition request carries the expected version and an idempotency key. Lock the task, check authorization and version, evaluate triggers, then commit the event, task pointer, version, and any durable downstream intent in one transaction.
- A stale version returns `409` with the current authorized task representation. The UI places the card in its authoritative column immediately and explains the conflict. Do not automatically replay the rejected move.
- An intercept without a valid reason returns `428`; no movement is committed. Release the transaction while the user chooses a reason. Recheck authorization, version, and triggers on resubmission.
- Trigger A: every backward move. Trigger B: outgoing duration strictly greater than three times an enabled SLA. Trigger C: third blocked entry within seven days. Define and test how simultaneous triggers are represented, without counting one event multiple times in HR totals.
- Reasons are `KNOWLEDGE_GAP`, `CARELESS_ERROR`, `SCOPE_CREEP`, `THIRD_PARTY_BLOCKER`, `ACCIDENTAL_MOVE`, and `OTHER`.
- `OTHER` requires a length-limited explanation; treat it as sensitive free text and unclassified context, not an automatic negative judgment.
- Recommended authorization rule: an authorized mentor or administrator supplies HR fault attribution. If an intern triggers a transition requiring classification, route it for mentor action; interns must not assign authoritative performance judgments to themselves or peers. Finalize this UX in Phase 4.
- An accidental correction references the mistaken movement. Preserve both events and the correction attribution; exclude confirmed accidental movement/correction pairs from performance revision counts. Do not erase history or allow arbitrary retroactive changes through an unrestricted bypass.
- Create an initial state-entry event when a task is created so first-column durations and nudges have a reliable anchor.
- Validate department/cohort/tag consistency. Serialize workflow-definition edits against transitions or use consistent versioned definitions, so concurrent column reordering cannot change the meaning of a committed event.
- Define terminal/completed columns, reopening, same-column drops, and column archival before calculating completed tasks or cycle time.

### 5.2 Fairness and metric definitions

Keep raw movement counts separate from attributed performance revision counts. Show how exclusions affect the result.

Before implementing aggregation, write examples covering scope changes, third-party delays, overlapping exclusions, accidental moves, `OTHER`, multi-skill tasks, final-column bounces, abandoned tasks, and review-boundary crossings.

The former suggestion to exclude all time until a task moves "past the original column" is not settled: it fails when that column is terminal. Define bounded, explicitly closed exclusion intervals and distinguish scope-rework time from externally blocked time. Merge overlapping intervals to avoid subtracting time twice. Explain which revision counts are excluded as well as which durations.

Define cycle-time start/end, calendar versus working time, completed-task denominator, skill attribution, cohort baseline population, and minimum evidence thresholds. Review windows and fairness rule versions are part of every snapshot's provenance.

### 5.3 Make the review understandable

Default HR display, using illustrative data:

> **Python - Day 30 review**  
> 14 tasks completed. 3 revisions requiring intern improvement. 2 revisions excluded because requirements changed.  
> Support suggested: targeted pairing, based on 3 knowledge-gap interventions across 2 tasks.  
> View supporting tasks.

The primary view answers: what was accomplished, where support is needed, what was outside the intern's control, what evidence supports that interpretation, and what the mentor recommends.

- Show the review window, sample size, missing data, and exclusions. Avoid confident assessments from sparse evidence.
- A knowledge-gap count of at least three per skill in the review window suggests targeted pairing; it does not decide employability.
- Evidence drill-down exposes relevant tasks, reasons, dates, calculation explanations, and definition versions, subject to role permissions.
- Separate system evidence, intern self-appraisal, mentor assessment, and the human return-offer recommendation.
- Version rubrics, calculation rules, and relevant workflow/tag definitions. Snapshot event-time meaning so later edits cannot silently rewrite history.
- Freeze published review evidence with a deterministic cutoff and source provenance. Corrections create an attributed amendment; retention still applies to originals and amendments.
- Do not interpret the Feature Spec's illustrative dashboard as a complete test dataset. Build explicit input fixtures and expected calculations before adopting golden files.

## 6. Architecture and operational contract

### 6.1 Boundaries

Use `api -> application services -> repository interfaces`, with SQLc/pgx adapters in the infrastructure layer. Pure workflow/intercept/HR calculations do not import HTTP or pgx. Transaction orchestration belongs in application services, not pure evaluators.

| Location | Responsibility |
| --- | --- |
| `cmd/moaj` | Wiring and `serve`, `worker`, `migrate` subcommands (`moaj serve`, `moaj version`, `moaj help`). |
| `internal/config` | Validated `MOAJ_*` (and legacy `IMS_*`) environment settings and `_FILE` secrets. |
| `internal/api`, `internal/auth` | HTTP, sessions, authorization, CSRF, request limits. |
| `internal/workflow`, `internal/intercept`, `internal/hr` | Domain rules and evidence calculations, with clearly separated service orchestration. |
| `internal/db` | Generated SQLc access; never edit generated files by hand. |
| `internal/worker`, `internal/notify` | Scheduling, delivery, retries, outbox processing. |
| `internal/gdpr` | Retention planning, restricted erasure, restore reconciliation. |
| `internal/realtime` | Scoped SSE delivery and reconnect/resynchronization. |
| `internal/observability` | Logs, metrics, health, optional trace integration. |
| `internal/enterprise`, `ee/` | Build-tag bridge and enterprise-only implementations. |
| `db/`, `ui/`, `deploy/`, `docs/` | Migrations/queries, embedded SvelteKit UI, operations assets, documentation. |

SvelteKit uses a static adapter and is embedded into the executable. No Node runtime is required in production. Develop a minimal UI in Phase 4 instead of deferring all integration until the end.

### 6.2 Durable delivery and realtime

Business transactions write durable outbox records. Workers deliver pending records, track attempts, and expose failures and delivery age. Use provider idempotency where supported; otherwise document the possible duplicate after a successful send followed by a crash. Do not claim exactly-once external delivery.

Asynq uniqueness and scheduler leadership reduce duplicate work; database constraints and idempotent handlers provide correctness. Include recovery when database commit succeeds but queue enqueue fails. Failed jobs need inspectable state and a controlled replay procedure.

Use PostgreSQL LISTEN/NOTIFY as a wake-up mechanism, not a durable replay log. SSE reconnect must resynchronize from persisted state or a durable cursor. Apply authorization to stream subscriptions and every delivered resource. Avoid leaking peer evidence through event payloads. Configure streaming deadlines separately from ordinary HTTP write timeouts.

### 6.3 Monitoring as a Community feature

Expose protected/network-restricted Prometheus-compatible metrics and structured JSON logs. Supply a versioned Grafana dashboard, scrape examples, alert rules, and troubleshooting instructions. Operators can use an existing monitoring stack; an optional Compose monitoring profile helps local evaluation.

| Area | Required visibility |
| --- | --- |
| HTTP | Request rate, status/error rate, duration histograms, in-flight requests; distinguish expected `409`/`428` from server faults. |
| Database | Connection pool use/waiting, query latency/failures, transaction contention where measurable. |
| Worker/outbox | Queue size, oldest pending age, running work, retry/failure counts, delivery latency. |
| Scheduled jobs | Last successful run, duration, and outcome for SLA sweeps, review snapshots, and retention. |
| Runtime | Memory, goroutines, process CPU where available, restarts and uptime. |
| Realtime | Active streams, reconnects, dropped/backlogged events, resynchronizations. |

Use bounded labels such as route templates, operation names, and outcomes. Never use user IDs, task IDs, raw URLs, notes, or emails as metric labels. Correlate logs with request/job IDs without logging secrets or personnel content.

Health checks distinguish process liveness from dependency readiness. Workers expose their own health and metrics. Alert examples cover sustained server errors/latency, database pool saturation, growing delivery backlog, failed jobs, and overdue scheduled work. Each alert links to a runbook.

Monitoring does not itself establish high-load capacity. Publish load-test conditions and measured results, including the monitoring overhead.

## 7. Development phases and exit gates

Phases are ordered by dependency and usable outcomes, not calendar promises. Work in reviewable changes; independent foundation and runtime work may proceed in parallel under the user's authorization. A gate must pass before declaring the phase complete; experiments may run earlier when they reduce later risk.

### Phase 0 - Repository, decisions, and build foundation

Outcome: a clean, buildable project with explicit ownership and licensing boundaries.

- Preserve existing Git history; add ignore/editor settings and consolidate duplicate PDFs after confirming references.
- Resolve D1-D3. Correct `go.mod`; remove or deliberately populate empty Go files. Replace broken practice code with a minimal compilable entry point.
- Harden config: formatting, helper spelling, `.lic` default, strict validation, environment/log level, `_FILE` secrets and precedence. Fail clearly on malformed production settings.
- Add build/test/generate/UI tasks and CI for formatting, vet, static analysis, vulnerability checks, tests, race detection on a supported runner, and Community/EE build paths.
- Establish dependency-license inventory and an explicit allowlist for distributed dependencies. Separate build tools from shipped artifacts. Do not infer license effects solely from a dependency appearing in a tool manifest.
- Prepare MIT core terms, exact directory boundaries, CONTRIBUTING, SECURITY, and the CLA workflow. Compare template reuse terms before copying. No proprietary release may rely on a placeholder license.
- Write the initial ADRs and record the specification amendments in this plan.

Gate: both build variants compile; available tests and CI pass; no accidental source changes are discarded; core licensing boundaries are reviewable. External CLA enforcement must be configured before accepting external contributions.

**Phase 0 Gate Verification Record (2026-11-01):**
- **Duplicate PDFs**: Confirmed SHA-256 identity between root PDFs and `docs/specs/*.pdf` (`3EF18E63...`, `1CF198E0...`, `A5D9420D...`, `DDF48D6E...`) and removed root copies.
- **Zero-byte files & formatting**: `0` zero-byte files in repository; `gofmt -l .` returned `0` unformatted files.
- **Builds & Community/EE boundary**: `go build -trimpath -o bin/moaj.exe ./cmd/moaj` (`moaj dev (community)`) and `go build -trimpath -tags ee -o bin/moaj-ee.exe ./cmd/moaj` (`moaj dev (enterprise-preview)`) succeeded; `go list -deps ./cmd/moaj` confirmed `0` imports of `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/ee/...`.
- **Static analysis & tests**: `go vet ./...`, `go vet -tags ee ./...`, `go test -count=1 ./...`, and `go test -count=1 -tags ee ./...` all passed (`cmd/moaj`, `internal/api`, `internal/config`, `internal/enterprise`, `internal/observability`). CI workflow configured in `.github/workflows/ci-license-gate.yml` (pending first push to remote).
- **Governance, ADRs, and dependency inventory**: Populated `LICENSE`, `ee/EULA.md`, `CONTRIBUTING.md`, `SECURITY.md`, `docs/governance/LICENSING_BOUNDARIES.md`, `docs/governance/DEPENDENCY_LICENSES.md`, `docs/adr/ADR-01` through `ADR-03`, `Makefile`, and `.github/workflows/ci-license-gate.yml`.
- **Accepted exceptions**: Live GitHub CLA webhook/app binding and final corporate licensor identity/jurisdiction remain tracked under `D17` (`docs/governance/LICENSING_BOUNDARIES.md` §5) prior to accepting external pull requests or distributing Phase 9 Enterprise artifacts.

### Phase 1 - HTTP runtime and observability baseline

Outcome: an inspectable service process with safe lifecycle behavior.

- Implement `moaj serve`, validated startup, server timeouts, graceful shutdown, request IDs, recovery, JSON logs, security headers, body limits, and RFC 9457 errors.
- Add `/healthz`, `/readyz`, and a protected `/metrics` (`moaj_*` metrics); readiness becomes dependency-aware as adapters are introduced.
- Add Community/EE bridge interfaces with safe stubs.
- Start the dashboard and scrape configuration with runtime and HTTP panels; document local startup.

Gate: health/metrics smoke checks and handler tests pass; shutdown drains requests; startup failures are clear; secrets are absent from logs.

**Phase 1 Gate Verification Record (2026-11-01):**
- **Runtime & HTTP handler verification**: `go test -count=1 ./cmd/ims ./internal/config ./internal/api ./internal/enterprise ./internal/observability` (both default and `-tags ee`) passed, covering `/healthz`, `/readyz` (`"scope":"http-runtime"` and `503` drain behavior), `/metrics` loopback vs. constant-time Bearer token auth, RFC 9457 `application/problem+json` error responses, panic recovery without leaking panic values, body size limits (`MaxBytesReader`), in-flight request draining on shutdown, and shutdown deadline cancellation.
- **Secret & PII redaction**: Verified by `TestSecretsDoNotAppearInErrorsOrStructuredLogs`, `TestTelemetryOmitsSecretsAndBoundsLabels`, and `TestServeRejectsInvalidConfigurationBeforeListening` that database URLs, metrics tokens, raw request paths/queries/bodies, untrusted `X-Request-ID` headers, and arbitrary HTTP methods never leak into logs, errors, or Prometheus labels.
- **Monitoring & operations assets**: Created `deploy/monitoring/prometheus.yml`, `deploy/monitoring/alerts.yml`, `deploy/monitoring/grafana-dashboard.json` (`uid: ims-phase1-runtime`, `schemaVersion: 39`), `docs/operations/monitoring.md`, and `README.md`.

### Phase 2 - Persistence, retention-aware schema, and transactions

Outcome: trustworthy data foundations without locking in indefinite personal-data retention.

- Model users, departments, cohorts/memberships, workflow definitions, tasks/versions, skill tags, transitions, intercepts, and installation metadata. Introduce subsequent feature tables in later migrations.
- Specify row scope, initial state entry, definition snapshots, FK deletion behavior, and retention ownership for every personnel table before migrations become a dependency.
- Add FKs/checks/indexes and enforce cross-department consistency. Define a safe strategy for column reordering under uniqueness constraints.
- Implement embedded goose migrations, a pinned-connection advisory lock, SQLc/pgx adapters, pool limits, and transaction interfaces.
- Establish normal application versus migration/retention privileges and append-only protection, including protection from ordinary `TRUNCATE`/DDL access.
- Start real PostgreSQL integration tests. Establish DB metrics and a development database deployment.

Gate: migrations apply on a clean database; supported development rollback/reapply works; runtime-role append-only and scoping tests pass; schema supports the planned erasure dependency order. Production migrations are forward-only with documented recovery.

### Phase 3 - Community authentication and first-run setup

Outcome: a protected installation that can create its first organization and department.

- Implement one-time setup with an installation-local bootstrap secret and atomic first-admin creation; do not let the first arbitrary network visitor claim the instance.
- Add Argon2id passwords, server-side sessions, rotation/revocation, secure cookies, CSRF, rate limits, logout, and a documented account recovery procedure.
- Define INTERN, MENTOR, DEPT_ADMIN, HR_ADMIN, and SYS_ADMIN scopes and assignment rules.
- Enforce authorization at service/query boundaries, including exports and later realtime subscriptions.
- Build minimal setup/login UI and the API contract that will generate the TypeScript client.

Gate: role/resource authorization tests and setup-race tests pass; peer intern data is inaccessible; first-run and recovery paths work without exposing credentials.

### Phase 4 - First usable workflow slice

Outcome: setup -> task -> board movement -> reason capture -> visible evidence works end to end.

- Implement column/task/tag CRUD, completion semantics, task versioning, and transactional movements.
- Implement all three intercept triggers, first-bounce reasons, `OTHER`, and auditable accidental correction behavior.
- Finalize mentor-versus-intern classification UX and test it with the role matrix.
- Add minimal keyboard-accessible board and intercept modal, immediate stale-card reconciliation, and scoped SSE with resync.
- Generate the API client from the versioned OpenAPI contract.
- Instrument transitions, conflicts, intercepts, and realtime without individual identifiers in labels.

Gate: trigger boundary tests cover all branches; two competing writes against one version yield one success and one conflict; A -> B -> A also invalidates stale requests; retries do not duplicate events; a minimal browser E2E path passes. This is an internal demonstration, not a production release.

### Phase 5 - Workers and reliable notifications

Outcome: asynchronous work survives retries, restarts, and enqueue/delivery failures.

- Add `moaj worker`, Asynq/Valkey integration, scheduler leadership, durable outbox polling/recovery, and idempotent job handlers.
- Implement SLA sweeps, neutral Slack/SMTP notifications, delivery state, retry/backoff, failed-job inspection, and replay.
- Put server, worker, PostgreSQL, and Valkey into development Compose with health checks and volumes. A separate worker is explicitly required in the default deployment.
- Add worker health/metrics and queue/outbox dashboard panels and alert rules.

Gate: test Valkey integration, repeated sweeps, scheduler overlap, commit-before-enqueue crashes, worker restarts, send failures, and send-success/ack-loss behavior. No committed intent is silently dropped; duplicate-delivery limits are documented.

### Phase 6 - HR translation and review experience

Outcome: mentors and HR can understand, verify, and act on fair evidence.

- Implement cohort lifecycle, milestone scheduling, versioned JSONB rubrics, 1-5 scales, weights, self-appraisal, and schema validation.
- Resolve D12/D13 with calculation fixtures. Build telemetry aggregation, fairness exclusions, sample-size handling, and competency support suggestions.
- Freeze Day 30/60/90 evidence at defined cutoffs; use unique snapshot keys and deterministic source selection so retries and late jobs cannot change the review window.
- Build concise review cards, evidence drill-down, mentor responses, published amendments, and an explicit human return-offer recommendation.
- Add shared 1-on-1 agendas and encrypted private notes with key configuration, backup, and rotation/recovery procedures.
- Add authorized CSV/JSON exports; protect spreadsheet-targeted exports against formula injection.

Gate: fixture-based calculations and snapshot replay are deterministic; definition edits do not rewrite published reviews; exclusions are explained; role tests cover notes and exports; a mentor can trace every summary claim to evidence. Measure the specification's appraisal-time target with a representative review task.

### Phase 7 - Retention and recovery correctness

Outcome: retention is an implemented data lifecycle, with tested boundaries.

- Finalize retention anchors: BRD post-cohort versus governance post-offboarding needs one documented policy, including early departures and multiple memberships.
- Implement configurable 1-6 year retention, aggregate privacy rules, restricted erasure, non-identifying audit records, and retries/checkpoints.
- Apply retention across tasks, events, notes, reviews/amendments, memberships, jobs, outbox, caches, and application logs as relevant to the data inventory.
- Coordinate erasure with workers and new writes so concurrent processing cannot recreate expired evidence.
- Implement restore reconciliation and document external-export/notification boundaries.
- Add retention success/failure/backlog monitoring and operational procedures.

Gate: erased evidence is absent from active stores and authorized joins; active memberships remain intact; small-group protections work; interrupted erasure resumes safely; a restored backup is reconciled before normal access. Document residual identification risks and retention limitations rather than claiming universal anonymity.

### Phase 8 - Production-ready Community release

Outcome: a new organization can install, operate, monitor, back up, and upgrade Community using the shipped documentation.

- Complete UI accessibility, loading/empty/error states, rubric/column editors, review workflows, and full critical-path E2E coverage.
- Embed static UI assets and migrations; verify production requires no Node runtime.
- Ship a non-root image and Compose configuration with server, worker, PostgreSQL, Valkey, health checks, volumes, and configuration examples. Document TLS/reverse-proxy and secret setup.
- Finish dashboard provisioning, scrape examples, alert rules, and runbooks; test with an actual Prometheus/Grafana setup.
- Test backup/restore, migrations, rollback/recovery guidance, and upgrade compatibility between server/worker/schema versions.
- Define a capacity-test workload: concurrent users and SSE connections, transitions/second, history size, worker backlog, hardware, duration, and latency/error objectives. Run sustained load and dependency-failure tests; report results and bottlenecks.
- Check the source targets of binary <20 MB, idle backend RSS <25 MB, and install-to-first-board <5 minutes. Define measurement conditions and resolve any infeasible target explicitly rather than silently claiming success.
- Publish a tagged Community release with checksums, dependency inventory/SBOM, image signing where configured, changelog, and support/security reporting instructions. Do not claim supply-chain checks that have not been executed.

Gate: a clean machine reaches a working board using only the README/install guide; server and workers operate; critical E2E, authorization, retention, restore, monitoring, and defined load gates pass. No open critical correctness/security defects. This is the first production Community release.

### Phase 9 - Enterprise edition

Outcome: commercial controls add enterprise governance while preserving recovery and access to existing records.

- Complete the template comparison and adapted EE license/CLA package, including source reuse terms, directory coverage, actual licensor, jurisdiction, and the spec's required clauses before distribution or sale.
- Implement RS256 verification with an algorithm allowlist, required claims, canonical domain and installation binding, active INTERN seat counting, and enabled features. Vendor signing keys stay in a separate private system.
- Specify and test expired versus invalid/tampered licenses, 14-day expiry grace, boot and periodic checks, quota changes, database clones, rehosting, and recovery behavior. Invalid signatures do not receive an expiry grace by accident.
- Implement SAML/OIDC, SCIM users/groups with explicit cohort mapping, PBAC scoping, and SIEM audit delivery using durable outbox mechanics.
- Ensure admin recovery and authorized CSV/JSON exports remain available after commercial entitlement expiry. License checks must not depend exclusively on a functioning job queue.
- Extend monitoring and load tests to EE behavior without adding personnel data to metrics.

Gate: Community's dependency graph/build contains no enterprise implementation; both editions pass relevant tests; entitlements and tenant/department scoping are enforced; license expiry and identity-provider outages preserve tested recovery/export paths; commercial terms are complete.

### Phase 10 - Release automation and measured scaling

Outcome: repeatable releases and capacity improvements based on operational evidence.

- Expand multi-architecture artifacts, automated release/SBOM/signing workflows, provenance, and compatibility testing.
- Add tracing integration where needed to diagnose cross-service latency; preserve the telemetry privacy policy.
- Validate multiple API/worker replicas, scheduler failover, SSE routing/resync, bounded queues, pool sizing, and overload behavior.
- Maintain capacity baselines and performance regression tests; optimize measured bottlenecks before introducing new infrastructure.

Gate: reproducible release procedures and documented capacity/failover results on declared infrastructure. This phase does not defer the minimum operating requirements of Phase 8.

## 8. Cross-cutting completion rules

- Tests progress from pure domain tests to PostgreSQL/Valkey integration and browser E2E. Run relevant gates as features land; do not defer all testing to release.
- All requests and jobs have bounded work, cancellation/deadlines, validated input, and scoped authorization. Use parameterized SQL and redact secrets.
- Keep OpenAPI, migrations, metric definitions, dashboard panels, and documentation synchronized with changed behavior.
- Use integration tests for privilege boundaries, concurrent writes, crash recovery, erasure, and migration behavior; line coverage alone is not a correctness guarantee.
- Preserve the Community/Enterprise boundary and check actual shipped dependency licenses for each build variant. Scan the UI bundle and included assets too.
- Requirement traceability is maintained through the phase map below and more granular test references as tests are implemented.
- Do not mark a phase complete because files exist; record the commands, measured results, and any accepted exceptions for its gate.

## 9. Requirement-to-phase map

| Requirement | Primary implementation phase(s) |
| --- | --- |
| BR-WF-01 custom states; BR-WF-02 skill tags | 2, 4 |
| BR-WF-03 ledger, amended for retention | 2, 4, 7 |
| BR-WF-04 SLA nudges | 5 |
| BR-INT-01/02/03 triggers, modal, amended taxonomy | 4; interpretation in 6 |
| BR-HR-01 cohort lifecycle | 2, 6 |
| BR-HR-02 rubrics/self-appraisal; BR-HR-03 evidence | 6 |
| BR-HR-04 1-on-1 and return-offer workflow | 6 |
| BR-EE-01/02/03/04 licensing, identity, PBAC, SIEM | 9 |
| NFR-01 embedded single binary | 1, 4, 8 |
| NFR-02 relational + JSONB | 2, 6 |
| NFR-03 dependency licensing | 0 and every release |
| NFR-04 retention, revised privacy contract | 2, 7, 8 |
| NFR-05 license boundaries and CLA | 0, 9 |
| Governance role isolation | 3 and every new endpoint/job/export/stream |
| Governance grace/recovery/export guarantee | Community exports in 6; EE behavior in 9 |
| New requirement: Community monitoring and operability | 1 onward; release gate in 8 |

## 10. Next implementation step

Finish and verify the authorized Phase 0 foundation and Phase 1 HTTP runtime work for **MOAJ (Mentor of all Jacks)**. The next development phase after these gates is Phase 2: persistence, retention-aware schema, and transaction boundaries.

Before implementing any affected feature, resolve its explicit design gate: retention-aware schema in Phase 2, classification/correction UX in Phase 4, fairness and review definitions in Phase 6, aggregate privacy thresholds in Phase 7, and final commercial terms in Phase 9.

## 11. Reference notes

- [EDPB Guidelines 01/2025 on pseudonymisation, consultation version](https://www.edpb.europa.eu/system/files/2025-01/edpb_guidelines_202501_pseudonymisation_en.pdf): supports the distinction between pseudonymisation and anonymity; it is not certification of this application's privacy design.
- [Prometheus instrumentation guidance](https://prometheus.io/docs/practices/instrumentation/): service, processing, batch-job, and label-cardinality considerations.
- [Asynq API documentation](https://pkg.go.dev/github.com/hibiken/asynq): uniqueness and retry behavior must be understood independently of business idempotency.
- [Redis license history](https://redis.io/legal/licenses/): 7.4-7.8 and 8+ have different licensing options. MOAJ's default remains Valkey, subject to version-specific dependency review.

Check current upstream versions, APIs, security advisories, and template reuse terms when implementation reaches them. This roadmap does not freeze unverified third-party assumptions.
