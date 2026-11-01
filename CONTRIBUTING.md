# Contributing to MOAJ (Mentor of all Jacks)

Thank you for your interest in contributing to **MOAJ (Mentor of all Jacks)**. MOAJ is a self-hosted, open-core intern management system that converts everyday task workflow movements into transparent, reviewable HR and mentorship evidence.

Before submitting a pull request, please review the repository boundaries, Contributor License Agreement (CLA) policy, build tags, and verification gates below.

---

## 1. Repository Structure & Open-Core Licensing Boundaries

MOAJ uses a **Single-Repository Open-Core** layout (`IMS-GOV-01` §1). Contributions must respect the boundary between the MIT-licensed Community Edition and the proprietary Enterprise Edition (`ee/`):

| Path | Edition & License | Description |
| :--- | :--- | :--- |
| `cmd/moaj/` | Community (MIT) | Single-binary CLI entry point (`moaj serve`, `moaj version`, `moaj help`, and future `worker`, `migrate` subcommands). |
| `internal/` | Community (MIT) | Core domain packages (`config`, `api`, `auth`, `workflow`, `intercept`, `hr`, `db`, `worker`, `notify`, `gdpr`, `realtime`, `observability`, `enterprise`). |
| `db/` | Community (MIT) | Embedded Goose SQL migrations (`db/migrations/`) and SQLc query files (`db/query/`). |
| `ui/` | Community (MIT) | SvelteKit static frontend (`moaj-ui`) embedded into the Go binary. |
| `docs/`, `deploy/` | Community (MIT / CC-BY-4.0) | Architecture Decision Records (`docs/adr/`), governance docs (`docs/governance/`), operations runbooks (`docs/operations/`), and container/monitoring manifests. |
| `ee/` | **Enterprise (Proprietary EULA)** | Enterprise-only modules (`ee/auth/`, `ee/rbac/`, `ee/audit/`, `ee/license/`) governed exclusively by `ee/EULA.md`. |

See `docs/governance/LICENSING_BOUNDARIES.md` for full details on build-tag separation and licensing rules.

---

## 2. Mandatory Contributor License Agreement (CLA) Policy (`IMS-GOV-01` §2)

### Why a DCO Alone Is Insufficient
Relying solely on a Developer Certificate of Origin (`Signed-off-by` / DCO) is legally insufficient for a single-repository open-core architecture. Because the project owner compiles community contributions from `cmd/`, `internal/`, `db/`, and `ui/` directly into both the MIT-licensed Community binary (`bin/moaj`) and the commercial Enterprise Edition binary (`bin/moaj-ee`, linked with `ee/`), the project owner requires explicit rights to distribute contributed code under both licensing regimes.

### CLA Requirement for External Contributors
Prior to merging any pull request from an external contributor:
1. The contributor must execute the **Individual Contributor License Agreement (ICLA)** or, if contributing on behalf of an employer, an authorized representative must execute the **Corporate Contributor License Agreement (CCLA)**.
2. **Grant of Rights:** Under the CLA (`IMS-GOV-01` §2), the contributor retains copyright ownership of their contribution while granting the project owner an **irrevocable, perpetual, worldwide, royalty-free copyright and patent license with the right to sublicense and distribute** the contributed code under both:
   - The open-source **MIT License** (`LICENSE`), and
   - The commercial **MOAJ Enterprise EULA** (`ee/EULA.md`).
3. **Automated PR Gate:** The GitHub Actions pipeline (`.github/workflows/ci-license-gate.yml`) and CLA check verify CLA status on external pull requests before merge.

---

## 3. Development Prerequisites

- **Go Toolchain:** Go `1.26+` (matching `go.mod`, module path `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ` per `D1`).
- **Formatting & Static Analysis:** `gofmt` (standard library), `go vet`, `staticcheck`, and `govulncheck`.
- **Database & Queue (Phase 2+ / Phase 5+):** PostgreSQL 16 and Valkey (BSD-3-Clause).
- **Third-Party Dependency Policy (`NFR-03`):**
  - Distributed runtime dependencies must use approved permissive licenses (`MIT`, `BSD-2-Clause`, `BSD-3-Clause`, `Apache-2.0`, `ISC`, `PostgreSQL`).
  - Strongly reciprocal or source-available copyleft licenses (`GPL-2.0`, `GPL-3.0`, `AGPL-3.0`, `LGPL-*`, `SSPL-*`, `RSAL-*`, `EUPL-*`) are strictly prohibited in distributed artifacts.
  - If a PR modifies `go.mod` or `ui/package.json`, update `docs/governance/DEPENDENCY_LICENSES.md` and ensure `make license-check` passes.

---

## 4. Community vs. Enterprise (`-tags ee`) Build Rules

MOAJ compiles two distinct editions from the same repository using Go build constraints:

1. **Community Edition (Default — `//go:build !ee`):**
   - Built with standard commands (`go build -o bin/moaj ./cmd/moaj` or `make build`).
   - Uses `internal/enterprise/bridge_community.go` safe no-op/Community stubs.
   - **Strict Rule:** No package compiled in the default Community build may import any package under `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/ee/...`. This is enforced by `make boundary-check` and CI.
2. **Enterprise Edition (`-tags ee` — `//go:build ee`):**
   - Built with `go build -tags ee -o bin/moaj-ee ./cmd/moaj` or `make build-ee`.
   - Uses `internal/enterprise/bridge_ee.go` to wire enterprise modules once implemented in Phase 9.
   - Both editions must compile cleanly and pass all tests on every pull request.

---

## 5. Pull Request Verification Gates

Before opening or updating a pull request, run the full local verification suite via the `Makefile` (or equivalent Go commands):

```sh
# 1. Verify formatting (must produce zero unformatted files)
make fmt-check

# 2. Run go vet on both Community and Enterprise build tags
make vet

# 3. Verify Community build never imports ee/
make boundary-check

# 4. Verify dependency licenses against the permissive allowlist
make license-check

# 5. Run unit and race tests for both Community and Enterprise builds
make test
make test-ee
make test-race

# 6. Build both Community and Enterprise binaries (./bin/moaj and ./bin/moaj-ee)
make build
make build-ee
```

### Coding & Privacy Guidelines for Contributors
- **No Secrets or PII in Logs/Metrics:** Never log raw passwords, session tokens, database credentials, cryptographic keys, or free-text intern evaluation notes. Never use `user_id`, `task_id`, raw URLs, or email addresses as Prometheus metric labels (`moaj_*`).
- **Specification Amendments:** Always check `docs/IMPLEMENTATION_PLAN.md` and `docs/adr/` before implementing domain rules. Where the implementation plan amends the original PDF specifications in `docs/specs/`, follow the recorded ADR and implementation plan contract.
