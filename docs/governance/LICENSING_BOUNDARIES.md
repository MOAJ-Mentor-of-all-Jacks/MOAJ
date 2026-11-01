# Open-Core Licensing Boundaries & Template Comparison (`LICENSING_BOUNDARIES.md`)

**Document Reference:** `IMS-GOV-01` §1–§4 & `docs/IMPLEMENTATION_PLAN.md` §3.1, §3.2 (`D17`), §7 (Phase 0 & Phase 9)  
**Status:** Active Phase 0 Governance Record (Pre-Release Commercial Terms Gate Tracked Under `D17`)

---

## 1. Purpose

**MOAJ (Mentor of all Jacks)** uses a **Single-Repository Open-Core** structure to support frictionless bottom-up adoption of the Community Edition while isolating commercial enterprise governance, identity, and compliance capabilities inside the top-level `ee/` directory.

This document defines:
1. The authoritative directory-to-license boundary mapping across the repository.
2. The Go build-tag isolation mechanism (`!ee` vs. `ee`) and automated CI enforcement ensuring the Community Edition binary (`bin/moaj`) never links or imports proprietary `ee/` packages.
3. A comparison of reusable open-core legal and CLA templates—including their upstream reuse terms—and the explicit checklist required to close Decision `D17` prior to any Phase 9 Enterprise Edition distribution.

---

## 2. Authoritative Directory-to-License Mapping (`IMS-GOV-01` §1)

| Repository Path | Governing License | License File | Permitted Rights | Commercial & Technical Boundary |
| :--- | :--- | :--- | :--- | :--- |
| `cmd/`, `internal/`, `db/`, `ui/` | **MIT License** (Permissive Open-Source) | `LICENSE` | Free use, modification, self-hosting, private forking, and redistribution with copyright notice. | Core Kanban workflow state machine (`cmd/moaj`), append-only `transition_ledger`, `HTTP 428` friction intercepts, SLA nudges, 30/60/90-day HR evaluation snapshots, and Prometheus/JSON observability (`moaj_*`). |
| `docs/`, `deploy/`, `sqlc.yaml`, `Makefile` | **CC-BY-4.0 / MIT License** | `LICENSE` | Public distribution and modification with attribution. | Architecture Decision Records (`docs/adr/`), governance & operations manuals, Docker/Compose manifests, Prometheus/Grafana monitoring configs, and SQLc configuration. |
| `ee/auth/`, `ee/rbac/` | **MOAJ Enterprise EULA** (Proprietary) | `ee/EULA.md` | **Source-available for customer security inspection** and audit only; production execution requires an active signed RS256 enterprise license (`/etc/moaj/moaj_enterprise.lic`). Zero modification or redistribution rights. | SAML 2.0, OIDC, SCIM 2.0 automated user/cohort provisioning, and fine-grained Policy-Based Access Control (PBAC). |
| `ee/audit/`, `ee/license/` | **MOAJ Enterprise EULA** (Proprietary) | `ee/EULA.md` | **Zero modification or redistribution rights; tamper-protected.** | SIEM audit log streaming (Datadog/Splunk) and the offline RSA-2048 (`RS256`) cryptographic JWT license verification engine. |

---

## 3. Build-Tag Isolation & Community Graph Verification

### 3.1 Build Tag Architecture (`internal/enterprise/`)
To guarantee clean compile-time separation without preprocessor hacks or runtime dynamic plugin loading:

- **`internal/enterprise/bridge.go`** (no build constraint): Defines the pure Go `Bridge` interface (`Edition()`, `LicenseFile()`, `ValidateStartup()`, and future hooks for SSO/SCIM/PBAC/SIEM) consumed by `cmd/moaj` and `internal/api`.
- **`internal/enterprise/bridge_community.go`** (`//go:build !ee`): Compiled by default into all standard `go build` / `go test` invocations. Provides the MIT-licensed Community Edition implementation with zero dependencies on `ee/`.
- **`internal/enterprise/bridge_ee.go`** (`//go:build ee`): Compiled only when `-tags ee` is explicitly passed to the Go toolchain (`bin/moaj-ee`). Wires the proprietary `ee/` packages (`ee/license`, `ee/auth`, `ee/rbac`, `ee/audit`) in Phase 9.
- **Files under `ee/**`**: Every `.go` file added under `ee/` must begin with `//go:build ee`.

### 3.2 Automated Boundary Verification (`make boundary-check`)
Every pull request and build runs a deterministic dependency graph inspection on the default Community build:

```sh
go list -deps ./cmd/moaj
```

If any package in the default (`!ee`) transitive dependency graph matches `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ/ee` (or `ee/...`), `make boundary-check` and `.github/workflows/ci-license-gate.yml` fail the build immediately.

---

## 4. Open-Core Legal & CLA Template Comparison (`D17`)

Per `docs/IMPLEMENTATION_PLAN.md` §3.1 ("Compare reusable existing open-core templates, then adapt a selected template to MOAJ's requirements"), the table below compares industry-standard open-core licensing, directory boundary, and CLA templates alongside their reuse terms:

| Template / Pattern | Primary Scope | Upstream License / Reuse Terms | Strengths for MOAJ | Adaptation Notes & Limitations |
| :--- | :--- | :--- | :--- | :--- |
| **Apache Software Foundation ICLA v2.0 & CCLA** (via `cla-assistant.io` or `cla-bot`) | Contributor License Agreement (Individual & Corporate) | Publicly reusable legal template; Apache-2.0 / permissive adaptation standard across open-core projects (provided "Apache" trademarks are removed). | Industry-standard wording familiar to corporate legal departments; grants irrevocable, perpetual, worldwide, royalty-free copyright & patent license with sublicensing rights. | Must explicitly add the dual-distribution sublicense clause (`IMS-GOV-01` §2) authorizing distribution under both the root `MIT License` and the `MOAJ Enterprise EULA`. |
| **Mattermost / GitLab Single-Repo Open-Core Boundary Model** (`LICENSE` + `ee/LICENSE`) | Single-repository directory split between MIT/Apache core and `/ee` commercial directory | Directory-isolation architecture is an industry convention; specific commercial subscription texts are proprietary to their vendors and must not be copied verbatim without permission. | Cleanly informs users in the root `LICENSE` that everything outside `ee/` is open-source while `ee/` is governed by `ee/EULA.md`. Matches `IMS-GOV-01` §1 directly. | MOAJ uses its own drafted `ee/EULA.md` provisions derived from `IMS-GOV-01` rather than copying third-party proprietary subscription agreements verbatim. |
| **PolyForm Shield / Fair Source / Elastic v2 (ELv2)** | Source-available non-compete licenses | PolyForm licenses are freely reusable verbatim (CC0 / public domain dedication for the license text itself); ELv2 is reusable with attribution. | Clear, concise legal drafting for source-available repositories. | **Not a full substitute for `ee/EULA.md`:** PolyForm/ELv2 focus on managed-service non-compete rather than per-seat (`max_intern_seats`) RS256 cryptographic key gating, 14-day grace periods, and `IMS-GOV-01` §4.2 anti-circumvention rules. Best used as structural inspiration only. |
| **Creative Commons Attribution 4.0 (`CC-BY-4.0`)** | Documentation & architecture specs (`docs/`, `deploy/`) | Standard public license by Creative Commons; freely usable by reference. | Explicitly required by `IMS-GOV-01` §1 (`CC-BY-4.0 / MIT`) for documentation and orchestration specifications. | Dual-licensing `docs/` and `deploy/` under `CC-BY-4.0 / MIT` allows seamless reuse of Docker Compose / Prometheus configs (`MIT`) alongside architectural prose (`CC-BY-4.0`). |

### Selected Direction for MOAJ
1. **CLA Mechanism:** Adopt an **Apache-derived ICLA/CCLA** with an explicit open-core dual-distribution clause (`MIT License` + `MOAJ Enterprise EULA`), enforced on GitHub pull requests via automated CLA verification (`cla-assistant` or repository CLA allowlist gate).
2. **Repository Boundary Notice:** Use the explicit root `LICENSE` (MIT) + `ee/EULA.md` directory split with automated `go list -deps ./cmd/moaj` verification.
3. **Documentation & Deploy Assets:** Govern `docs/` and `deploy/` under dual `CC-BY-4.0 / MIT` terms as specified in `IMS-GOV-01` §1.

---

## 5. Checklist to Close Decision `D17` Before Phase 9 Release

Phase 0 establishes the technical build-tag boundary, dependency allowlist, CLA requirement, and specification-aligned `ee/EULA.md`. Before any Phase 9 Enterprise Edition artifact is distributed or sold, the following items must be completed and verified:

- [ ] **Final Licensor Legal Entity:** Replace `MOAJ Contributors / Project Owner (see D17 in docs/IMPLEMENTATION_PLAN.md)` in `LICENSE` and `ee/EULA.md` with the registered corporate entity or copyright holder name.
- [ ] **Governing Law & Jurisdiction:** Specify the governing law, courts/venue, and dispute resolution clause in `ee/EULA.md` and the CLA text.
- [ ] **Hosted CLA Bot Configuration:** Connect the live GitHub App / webhook (`cla-assistant.io` or signed CLA database) to `github.com/MOAJ-Mentor-of-all-Jacks/MOAJ` before merging any external contributor PRs.
- [ ] **Commercial Master Subscription Schedule:** Attach commercial warranty, indemnification, SLA support tiers, and limitation-of-liability schedules reviewed by legal counsel to `ee/EULA.md`.
- [ ] **Production RS256 Key Isolation:** Confirm the vendor RSA-2048 private signing key resides in an isolated offline signing vault (`IMS-GOV-01` §3) and only the public key is embedded in `ee/license/`.
