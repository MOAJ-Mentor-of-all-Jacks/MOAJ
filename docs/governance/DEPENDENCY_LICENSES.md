# Third-Party Dependency License Inventory & Policy (`DEPENDENCY_LICENSES.md`)

**Document Reference:** `IMS-BRD-1.0` (`NFR-03`), `IMS-GOV-01` §5.4 & §6, `docs/IMPLEMENTATION_PLAN.md` §3.2 (`D1`–`D3`) & §7 (Phase 0)  
**Status:** Active & Enforced via `make license-check` and `.github/workflows/ci-license-gate.yml`

---

## 1. Foundational Decisions (`D1`, `D2`, `D3`)

Per `docs/IMPLEMENTATION_PLAN.md` §3.2 and §7 (Phase 0), the following foundational naming, module, toolchain, and infrastructure licensing decisions are resolved and recorded:

| Decision ID | Resolution | Rationale & Compliance Impact |
| :--- | :--- | :--- |
| **`D1`** (Project Name & Module Path) | Official Name: **MOAJ - Mentor of all Jacks** (`cmd/moaj`, `bin/moaj`, `MOAJ_*` env vars, `/etc/moaj/moaj_enterprise.lic`, `moaj_*` metrics); Go Module Path: **`github.com/MOAJ-Mentor-of-all-Jacks/MOAJ`** | Establishes the official product identity **MOAJ (Mentor of all Jacks)** and aligns `go.mod` with the canonical GitHub organization repository `https://github.com/MOAJ-Mentor-of-all-Jacks/MOAJ.git` (`IMS_*` environment variables are retained as legacy aliases). |
| **`D2`** (Go Toolchain & CI Policy) | **`go 1.26.0` in `go.mod`**; CI verifies on supported stable Go (`1.26.x`) on Linux with `-race` enabled | Local environment reports `go1.27.2 windows/amd64`, which is treated as machine evidence rather than upstream support policy. Pinning `go 1.26.0` in `go.mod` and running Linux CI with `-race` ensures reproducible builds across environments. |
| **`D3`** (In-Memory Queue Broker) | **Valkey (`BSD-3-Clause`, pinned `valkey/valkey:8.0-alpine` for Phase 5)** selected over Redis (`RSALv2` / `SSPLv1` / `AGPLv3`) | `IMS-BRD-1.0` `NFR-03` requires that all core infrastructure components use permissive open-source licenses (`MIT`, `Apache-2.0`, `BSD`, `PostgreSQL`). Redis 7.4+ transitioned to dual `RSALv2`/`SSPLv1` (and Redis 8 added `AGPLv3`), which are non-permissive/copyleft. **Valkey** is the Linux Foundation-backed `BSD-3-Clause` fork that maintains wire-protocol compatibility with Redis 7.2 and full compatibility with Go's `github.com/hibiken/asynq` (`MIT`) and `github.com/redis/go-redis/v9` (`BSD-2-Clause`), to be integration-tested in Phase 5. |

---

## 2. Distributed Dependencies vs. Non-Distributed Build/CI Tools

As required by `docs/IMPLEMENTATION_PLAN.md` §7 (Phase 0: *"Separate build tools from shipped artifacts. Do not infer license effects solely from a dependency appearing in a tool manifest"*), MOAJ strictly distinguishes between two categories of third-party software:

### 2.1 Distributed Runtime Dependencies (Strict Permissive Gate)
- **Definition:** Any Go module (`go.mod`), JavaScript/TypeScript package (`ui/package.json`), CSS/font asset, or embedded binary artifact that is compiled into, linked with, or bundled inside the shipped `moaj` single binary (Community `bin/moaj` or Enterprise `bin/moaj-ee`) or container image.
- **Legal Impact:** Because the `moaj` binary is distributed under both the MIT License (Community) and the proprietary MOAJ Enterprise EULA (`ee/EULA.md`), every distributed runtime dependency **must** carry an approved permissive license with zero reciprocal/copyleft viral obligations.

### 2.2 Non-Distributed Build, Code-Generation, and CI Tools
- **Definition:** Standalone developer and CI executables invoked during build, code generation, linting, or testing whose own source code is **not** linked into the compiled `moaj` binary (unless a runtime library package is separately imported in `go.mod`).
- **Examples:**
  - `sqlc` (`MIT`) — generates pure Go source code in `internal/db/` from SQL queries.
  - `staticcheck` (`MIT`) — static analysis CLI run in CI.
  - `govulncheck` (`BSD-3-Clause`) — Go vulnerability scanner CLI run in CI.
  - `google/go-licenses` (`Apache-2.0`) — license verification CLI run in CI.
  - `goose` (`MIT`) — when imported as a library for embedded migrations in Phase 2, `github.com/pressly/goose/v3` (`MIT`) becomes a distributed runtime dependency and will be added to the table in Section 4.

---

## 3. License Classification Policy (`IMS-GOV-01` §5.4 & §6)

### 3.1 Approved Permissive Licenses (Allowlisted for Distributed Dependencies)
Distributed Go modules and UI packages may only use the following SPDX license categories:
- `MIT`
- `BSD-2-Clause`
- `BSD-3-Clause`
- `Apache-2.0`
- `ISC`
- `PostgreSQL`
- `0BSD` / `Unlicense` / `CC0-1.0`

### 3.2 Prohibited Copyleft & Source-Available Licenses (Build-Breaking)
Any pull request introducing a distributed dependency under any of the following licenses **fails the CI build immediately**:
- **Strong Reciprocal Copyleft:** `GPL-2.0`, `GPL-2.0-only`, `GPL-2.0-or-later`, `GPL-3.0`, `GPL-3.0-only`, `GPL-3.0-or-later`, `AGPL-3.0`, `AGPL-3.0-only`, `AGPL-3.0-or-later`
- **Weak / Library Copyleft (unless explicitly reviewed and exempted for dynamic-only use, which does not apply to static Go binaries):** `LGPL-2.1`, `LGPL-3.0`, `MPL-2.0`, `EPL-2.0`, `EUPL-1.2`
- **Restrictive Source-Available / Non-Compete:** `SSPL-1.0`, `RSALv2`, `BUSL-1.1`

---

## 4. Current Distributed Go Module Inventory (`go.mod` — Phase 0 / Phase 1)

Every direct and indirect module present in `go.mod` as of Phase 0/Phase 1 has been audited and verified against upstream repository `LICENSE` files:

| Module Path | Version | Dependency Type | SPDX License | Upstream Repository / Notes |
| :--- | :--- | :--- | :--- | :--- |
| `github.com/prometheus/client_golang` | `v1.25.0` | Direct | `Apache-2.0` | Prometheus Go client library (`internal/observability`, `internal/api`). |
| `github.com/beorn7/perks` | `v1.0.1` | Indirect | `MIT` | Effective computation of quantile histograms for Prometheus. |
| `github.com/cespare/xxhash/v2` | `v2.3.0` | Indirect | `MIT` | Fast 64-bit `xxHash` implementation used by Prometheus metrics registry. |
| `github.com/munnerz/goautoneg` | `v0.0.0-20191010083416-a7dc8b61c822` | Indirect | `BSD-3-Clause` | HTTP `Accept` header content negotiation for Prometheus scrape handlers. |
| `github.com/prometheus/client_model` | `v0.6.3` | Indirect | `Apache-2.0` | Data model artifacts for Prometheus metrics. |
| `github.com/prometheus/common` | `v0.72.0` | Indirect | `Apache-2.0` | Shared Prometheus Go utilities and exposition format support. |
| `github.com/prometheus/procfs` | `v0.22.0` | Indirect | `Apache-2.0` | `/proc` filesystem collector for process CPU/memory/FD telemetry. |
| `golang.org/x/sys` | `v0.48.0` | Indirect | `BSD-3-Clause` | Go supplementary low-level OS syscall packages. |
| `google.golang.org/protobuf` | `v1.36.12` | Indirect | `BSD-3-Clause` | Go support for Protocol Buffers used by `prometheus/client_model`. |

### Core Infrastructure Stack Licenses (`NFR-03`)

| Component | Planned / Active Version | SPDX License | Role in MOAJ |
| :--- | :--- | :--- | :--- |
| **Go Standard Library & Runtime** | `1.26+` | `BSD-3-Clause` | Single-binary HTTP server (`moaj`), `log/slog`, `crypto/*`, `embed`. |
| **PostgreSQL** | `16` | `PostgreSQL` | Primary ACID relational store, `JSONB` rubrics, `LISTEN/NOTIFY`. |
| **Valkey** | `8.0` | `BSD-3-Clause` | Redis-compatible in-memory store backing Asynq queues & schedules (Phase 5). |
| **SvelteKit (`@sveltejs/kit`, `@sveltejs/adapter-static`)** | `2.x` (Phase 4) | `MIT` | Static single-page frontend (`moaj-ui`) embedded into the Go binary. |

---

## 5. Verification Procedure

1. **Deterministic Allowlist Check (`make license-check`):**
   - Queries `go list -m all` and verifies that every resolved module in `go.mod` is explicitly listed in the approved permissive inventory above. Any unreviewed module causes `make license-check` to fail with instructions to verify its SPDX license and update this document.
2. **Automated Upstream License Classification (`go-licenses`):**
   - In CI (`.github/workflows/ci-license-gate.yml`), `go-licenses check ./cmd/moaj --disallowed_types=forbidden,restricted,reciprocal,unknown` inspects the license files of all transitive packages compiled into the binary.
3. **Frontend Package Gate (Phase 4+):**
   - When `ui/` dependencies are installed in Phase 4, `npx license-checker --production --onlyAllow "MIT;ISC;BSD-2-Clause;BSD-3-Clause;Apache-2.0;0BSD;Unlicense"` enforces the identical permissive policy on the embedded UI bundle.
