# Security Policy (`SECURITY.md`)

**MOAJ (Mentor of all Jacks)** processes sensitive personnel evaluation records, mentorship notes, and organizational workflow data. Security, role isolation, secret hygiene, and data privacy are core architectural requirements across both Community and Enterprise editions.

---

## 1. Supported Versions & Development Status

MOAJ is currently under active staged development according to `docs/IMPLEMENTATION_PLAN.md`:

| Version / Phase | Status | Security Updates |
| :--- | :--- | :--- |
| `main` (Phase 0 – Phase 1 foundation) | Active development | Supported via repository patches |
| Community `v1.0.x` (Target: Phase 8 gate) | Planned first production release | Full security advisory & patch support upon release |
| Enterprise `v1.0.x-ee` (Target: Phase 9 gate) | Planned commercial release | Full security advisory & patch support upon release |

---

## 2. Reporting a Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues, discussions, or pull requests.**

If you discover a security vulnerability, authorization bypass, cryptographic flaw, or personnel data leak:

1. **Private Reporting Channel:**
   - Use **GitHub Private Vulnerability Reporting** (`Security` -> `Report a vulnerability` on the `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ` repository) so the maintainers can triage and remediate the issue confidentially.
2. **What to Include:**
   - Affected component or package path (e.g., `cmd/moaj`, `internal/api`, `internal/config`, `internal/auth`, `ee/license`).
   - Build variant (`Community` default build vs. `Enterprise` `-tags ee` build).
   - Step-by-step reproduction instructions or proof-of-concept test case.
   - Potential impact on confidentiality (e.g., cross-intern data exposure), integrity (e.g., ledger tampering), or availability.
3. **Response & Coordinated Disclosure Expectations:**
   - **Acknowledgment:** Within **48 hours** of submission.
   - **Triage & Remediation Plan:** Within **5 business days**.
   - **Coordinated Disclosure:** We ask reporters to maintain confidentiality until a patch is released and operators have had a reasonable window to upgrade. Credit will be given in the security advisory unless anonymity is requested.

---

## 3. Core Security Architecture Principles

All code in MOAJ must adhere to the following non-negotiable security invariants:

### 3.1 Secret Handling & `_FILE` Support
- Production secrets (`MOAJ_DATABASE_URL`, `MOAJ_VALKEY_ADDR`, `MOAJ_METRICS_TOKEN`, `MOAJ_LICENSE_FILE`, and legacy `IMS_*` aliases) must never be hardcoded in source files or committed to version control.
- `internal/config` supports reading sensitive values from mounted files via the `*_FILE` environment variable convention (for example, `MOAJ_DATABASE_URL_FILE=/run/secrets/db_url` or `MOAJ_METRICS_TOKEN_FILE=/run/secrets/metrics_token`), compatible with Docker/Kubernetes secret mounts.
- Setting both a direct variable and its `*_FILE` variant simultaneously is rejected as an explicit configuration conflict to prevent ambiguous credential precedence.
- In `production` mode (`MOAJ_ENV=production`), insecure default credentials, wildcard CORS origins (`*`), and unprotected `/metrics` endpoints are rejected at startup.

### 3.2 Zero Secrets & Zero PII in Logs and Metrics
- **Structured Logs (`log/slog`):** Logs correlate activity using request IDs and job IDs. Passwords, bearer tokens, session cookies, connection strings, and personnel evaluation text (`OTHER` explanations, private 1-on-1 mentor notes) must never be written to application logs.
- **Prometheus Metrics (`/metrics`, `moaj_*`):** Metrics (`moaj_build_info`, `moaj_process_start_time_seconds`, `moaj_http_requests_total`, `moaj_http_request_duration_seconds`, `moaj_http_requests_in_flight`) use strictly bounded cardinality labels (route templates, HTTP methods, status codes, operation names). High-cardinality or identifying values (`user_id`, `intern_id`, `task_id`, email addresses, raw URLs) are strictly forbidden in metric labels.
- **Protected Telemetry:** The `/metrics` endpoint requires bearer token authentication (`MOAJ_METRICS_TOKEN` validated via constant-time `subtle.ConstantTimeCompare`).

### 3.3 Strict Role Isolation (`IMS-GOV-01` §5.3)
- Interns (`INTERN`) have full visibility into their own Kanban board, tasks, and 30/60/90-day evaluation rubrics (preventing opaque or "secret surveillance" practices), but are strictly prohibited from reading or mutating any peer intern's tasks, revision counts, intercept classifications, or evaluations.
- Mentors (`MENTOR`) and administrators (`DEPT_ADMIN`, `HR_ADMIN`, `SYS_ADMIN`) are scoped to their authorized organizational and department boundaries at the service and query layer—including HTTP APIs, SSE realtime streams, and CSV/JSON exports.
- Interns cannot self-assign authoritative HR fault classifications when a workflow intercept (`HTTP 428 Precondition Required`) fires; fault attribution is restricted to authorized mentors/administrators.

### 3.4 Append-Only Ledger Integrity & Privilege-Separated Retention Erasure
- During an intern cohort's active retention window (1–6 years), `transition_ledger` and `intercept_events` are append-only for the normal application database role (`SELECT`, `INSERT` only; no `UPDATE`, `DELETE`, `TRUNCATE`, or `DDL`).
- Database migrations and retention-expiry erasure run under separate, least-privilege database roles.
- At retention expiry, controlled erasure removes expired personal records, free-text notes, and identifying links rather than relying on naive UUID hashing alone (see `docs/IMPLEMENTATION_PLAN.md` §4 and `docs/adr/ADR-02-append-only-relational-ledger.md`).

### 3.5 Offline Enterprise License Verification (`RS256`)
- Enterprise Edition installations operate in air-gapped environments without external network calls ("phoning home").
- License tokens (`/etc/moaj/moaj_enterprise.lic`) are verified offline using RSA-2048 (`RS256`) with a strict algorithm allowlist. Invalid or tampered signatures never receive an expiration grace period, while expired valid tokens enter a 14-day grace period and preserve admin recovery and HR compliance export access thereafter.
