# MOAJ (Mentor of all Jacks) Enterprise Edition License & Reservation of Rights (`ee/EULA.md`)

**Document Reference:** `IMS-GOV-01` (Open-Core Licensing, Legal & Governance Manual)  
**Scope:** Everything under `ee/` (`ee/auth/`, `ee/rbac/`, `ee/audit/`, `ee/license/`) and any binary compiled with `-tags ee` (`bin/moaj-ee`)  
**Copyright:** Copyright (c) 2026 MOAJ Project Owner (see Decision `D17` in `docs/IMPLEMENTATION_PLAN.md`). All Rights Reserved.

---

## 1. Proprietary Reservation of Rights & Exclusion from Root MIT License

1. **Exclusion from MIT License:** All source code, files, configuration, cryptographic keys, and assets located within the top-level `ee/` directory are **proprietary** and are **explicitly excluded** from the root `LICENSE` (MIT License) governing the MOAJ Community Edition.
2. **No Implied Production Rights:** Absent a valid, executed commercial Enterprise Subscription Agreement and an active cryptographic license key (`/etc/moaj/moaj_enterprise.lic`) issued by the Licensor, **no right or license is granted** to execute, deploy in production, modify, create derivative works from, sublicense, or redistribute any code under `ee/` or any binary compiled with the `ee` Go build tag (`-tags ee`).

---

## 2. Subdirectory Boundary Split (`IMS-GOV-01` Section 1)

Within `ee/`, rights are further partitioned according to `IMS-GOV-01` Section 1:

| Repository Path | Governing License | Permitted Inspection / Usage Rights | Commercial & Technical Scope |
| :--- | :--- | :--- | :--- |
| `ee/auth/`, `ee/rbac/` | MOAJ Enterprise EULA (Proprietary) | **Source-available for customer security inspection** and audit only; production execution requires an active signed RS256 enterprise license key. Zero modification or redistribution rights. | Gates SAML 2.0, OIDC, SCIM 2.0 automated user provisioning, and Policy-Based Access Control (PBAC). |
| `ee/audit/`, `ee/license/` | MOAJ Enterprise EULA (Proprietary) | **Zero modification or redistribution rights; tamper-protected.** | Gates SIEM log streaming (Datadog/Splunk) and the offline RS256 cryptographic JWT license verification engine. |

---

## 3. Mandatory Commercial Provisions (`IMS-GOV-01` Sections 2, 3, and 4)

### 3.1 Mandatory Anti-Circumvention Clause (`IMS-GOV-01` §4.2)
Any modification of the source code within `ee/license/`, removal or alteration of Go build constraints (`//go:build ee` or `//go:build !ee`), bypassing of `internal/enterprise` bridge interfaces, or replacement of the embedded RSA-2048 (`RS256`) public key to bypass seat limits, module gates, or expiration checks constitutes an immediate material breach of contract and willful copyright infringement.

### 3.2 Offline Cryptographic License Enforcement (`RS256`)
Enterprise entitlements are enforced completely offline without external network calls ("phoning home") using Asymmetric RSA-2048 (`RS256`) signed JSON Web Tokens (`MOAJ_LICENSE_FILE`, defaulting to `/etc/moaj/moaj_enterprise.lic`). The self-hosted `moaj` binary verifies the signature at startup and on a periodic schedule (every 6 hours, without depending exclusively on a functioning job queue) against the hardcoded public key compiled into the Enterprise Edition binary. Invalid or tampered signatures are rejected immediately and **never** receive an expiration grace period.

### 3.3 Instance Fingerprinting, Domain Binding, and Clone/Restore Behavior (`IMS-GOV-01` §4 & Decision `D16`)
- **Binding:** Each enterprise license token binds to the configured `MOAJ_APP_DOMAIN` (or legacy alias `IMS_APP_DOMAIN` / `APP_DOMAIN`) and the installation identifier stored in PostgreSQL (cluster/installation UUID per Decision `D16`).
- **Clone / Restore / Rehosting Behavior (`D16`):** Because the installation identifier is persisted in PostgreSQL, restoring a database backup or cloning a database to another environment copies this installation identifier. It is an operational compliance binding rather than a hardware DRM or anti-cloning guarantee. Deploying a single production license token simultaneously across multiple isolated production clusters violates the commercial license and triggers a read-only compliance banner on unauthorized instances.

### 3.4 Active Intern Seat Quota Definition (`IMS-GOV-01` §4)
- **Counted Seats:** Only active users assigned the `INTERN` role within an **open** cohort (`max_intern_seats`) count toward the licensed seat quota.
- **Zero-Cost Roles:** Mentors (`MENTOR`), department administrators (`DEPT_ADMIN`), HR administrators (`HR_ADMIN`), and system administrators (`SYS_ADMIN`), as well as archived/closed cohort records within their retention window, are provisioned at **zero seat cost** to encourage organization-wide mentoring participation and preserve historical compliance access.

### 3.5 Expiration Grace Period & Fail-Safe Data Access Guarantee (`IMS-GOV-01` §4)
- **14-Day Grace Period:** Upon expiration (`exp`) of a validly signed enterprise license token, the installation enters a **14-day warning state** (`grace: 14d`) during which enterprise features remain operational while administrators renew the subscription.
- **Post-Grace Fail-Safe & Anti-Hostage Guarantee:** If the license remains unrenewed after the 14-day grace period expires:
  1. Enterprise SSO (`SAML 2.0` / `OIDC`) and automated `SCIM 2.0` provisioning fall back to local administrator recovery mode.
  2. **Customer Data Access & Export Guarantee:** Existing tasks, transition ledgers, and HR evaluation records remain readable and exportable (via authorized CSV/JSON compliance exports) so customers are **never held hostage** from their personnel or compliance records.

---

## 4. Decision `D17` Pre-Release Legal Gate (Phase 9 Requirement)

> **Status (`D17` in `docs/IMPLEMENTATION_PLAN.md`):**
> This file establishes the mandatory technical, boundary, and commercial specification provisions required in Phase 0. **No proprietary Enterprise Edition binary or release may be distributed or sold under placeholder legal text.**
>
> Before the Phase 9 Enterprise Edition gate can pass:
> 1. Reusable open-core commercial license templates (evaluated in `docs/governance/LICENSING_BOUNDARIES.md`) must be reviewed for reuse terms and adapted into the final binding commercial EULA.
> 2. The final legal entity name (Licensor identity), governing law, venue/jurisdiction, warranty disclaimers, and limitation of liability schedules must be finalized and recorded here.
