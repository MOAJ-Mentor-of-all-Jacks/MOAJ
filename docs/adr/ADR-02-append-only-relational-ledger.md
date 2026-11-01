# ADR-02: Append-Only Relational Transition Ledger & Retention-Bounded Privacy Contract

- **Status:** Accepted (Amended by `docs/IMPLEMENTATION_PLAN.md` v2.1)
- **Date:** 2026-11-01
- **References:** `IMS-BRD-1.0` (`BR-WF-03`, `NFR-02`, `NFR-04`), `IMS Full System Architecture.pdf` (Blueprint v1.0 §2), `IMS-GOV-01` (§5), `docs/IMPLEMENTATION_PLAN.md` (§3.1, §3.2 `D8`–`D10`, §4, §5.1, §6.2)

---

## 1. Context

Traditional task trackers overwrite a mutable `status` column when a card moves across a Kanban board, discarding the historical timing, bounce-back sequence, and context needed for objective intern performance reviews. Conversely, storing unstructured event streams in a NoSQL document store sacrifices relational foreign-key integrity, transactional consistency between task state and audit history, and strict role-based query scoping.

At the same time, because **MOAJ (Mentor of all Jacks)** records granular workflow friction (backward transitions, SLA overruns, and mentor-attributed `KNOWLEDGE_GAP` or `CARELESS_ERROR` classifications), the database legally constitutes an **Automated Employee Evaluation & Personnel File System** subject to GDPR Article 5(1)(e) (Storage Limitation) and regional labor laws.

During Phase 0 specification reconciliation (`docs/IMPLEMENTATION_PLAN.md`), several conflicts and gaps in the initial PDFs were identified and resolved:
1. Field naming divergence between the BRD (`time_spent_in_previous_ms`) and the Architecture Blueprint (`time_in_prev_state_ms`).
2. Missing initial state-entry anchor when a task is first created.
3. Stale concurrent board mutations and reliable notification delivery.
4. Overstated privacy claims in the initial specs that replacing `intern_id` with a salted hash (`SHA-256`) while keeping indefinite task/event timelines guarantees irreversible GDPR anonymity.

---

## 2. Decision

### 2.1 Relational Append-Only Transition Ledger in PostgreSQL 16
1. **Transactional State + Ledger Commit:** Every task movement executes inside a single PostgreSQL ACID transaction that locks the task row (`SELECT ... FOR UPDATE`), validates authorization and the expected task-wide `version`, updates the current state pointer on `tasks`, increments `tasks.version`, and inserts an immutable row into `transition_ledger`.
2. **Initial State-Entry Anchor:** Creating a task inserts an initial state-entry event into `transition_ledger` (`from_column_id = NULL`, `to_column_id = initial_column_id`, `time_in_prev_state_ms = 0`) so that first-column residence time and SLA nudges (`BR-WF-04`) have an exact, deterministic timestamp anchor.
3. **Canonical Duration Field Name (`D8`):** The database schema and internal Go domain use `time_in_prev_state_ms BIGINT NOT NULL` (from Blueprint v1.0). External API documentation and BRD traceability maps explicitly document that `time_in_prev_state_ms` implements BRD `BR-WF-03` (`time_spent_in_previous_ms`).

---

## 3. Specification Amendments Recorded in This ADR

### 3.1 Task-Wide Optimistic Concurrency (`HTTP 409 Conflict`)
- **Supersedes:** Checking only whether `current_column_id` matches, or waiting for a timeout on stale board state (`docs/IMPLEMENTATION_PLAN.md` §3.1 & §5.1).
- **Amended Rule:** Every mutation on a task increments its integer `version`. Transition requests must supply both the expected `version` and an `Idempotency-Key`. Even if a card moves `Column A -> Column B -> Column A`, `version` has advanced; a competing request carrying the old `version` is immediately rejected with `HTTP 409 Conflict` and returns the current authorized task representation so the UI reconciles immediately without replaying the stale move.

### 3.2 Atomic Durable Notification Outbox (`docs/IMPLEMENTATION_PLAN.md` §3.1 & §6.2)
- **Supersedes:** Inserting a bare deduplication key and firing external Slack/SMTP calls without durable delivery tracking.
- **Amended Rule:** Whenever a transition or SLA sweep generates a notification intent, the intent is committed atomically into a PostgreSQL durable outbox table within the same transaction as the state change. Background workers (`moaj worker`) poll/process pending outbox rows asynchronously with idempotency keys, retry/backoff state, and inspectable failure metrics.

### 3.3 Privacy, Retention, and Privilege-Separated Erasure Contract (`D9`, `D10`, `docs/IMPLEMENTATION_PLAN.md` §4)
- **Supersedes:** The claim in Blueprint v1.0 / `IMS-GOV-01` §5.1 that replacing `intern_id` with a salted one-way hash (`SHA-256`) while retaining individual task titles, timestamps, and transition histories indefinitely makes records "100% anonymous." As supported by EDPB Guidelines 01/2025 on pseudonymisation, rare task descriptions, exact timestamps, small cohorts, or free-text `OTHER` explanations can still re-identify an individual without a name or UUID.
- **Amended Data Lifecycle & Retention Contract:**
  1. **Bounded Retention Window (`D9`):** Identifiable personnel evidence (memberships, task assignments, detailed transitions, intercept events, evaluations, private mentor notes, and `OTHER` explanations) is retained only during the configured retention period (`1–6 years`).
  2. **Privilege Separation:** The database enforces three distinct privilege roles:
     - **Application Runtime Role:** Granted `SELECT, INSERT` (no `UPDATE`, `DELETE`, `TRUNCATE`, or `DDL`) on `transition_ledger` and `intercept_events`, enforced by both SQL privileges and append-only database triggers.
     - **Migration Role:** Used exclusively by `moaj migrate` under an advisory lock.
     - **Restricted Retention/Erasure Role:** Executes controlled, audited expiry erasure via a security-definer/restricted pathway with a fixed `search_path`.
  3. **Controlled Expiry Erasure (`D10`):** At retention expiry, the controlled retention process purges expired personal records, task text, private mentor notes, `OTHER` explanations, and published snapshot details, while preserving active user accounts that belong to other still-active cohorts.
  4. **Assessed Aggregate Benchmarks & Small-Group Suppression:** Only approved cohort/department aggregate benchmarks are retained beyond the retention window, with small-group suppression rules so tiny cohorts cannot leak individual identity.
  5. **Backup Restore Reconciliation:** Minimal, non-identifying erasure audit markers are retained so that if a database backup taken shortly before an erasure run is restored, pending erasures are replayed before reopening normal service access.

---

## 4. Consequences

### Positive
- **Tamper-Evident Operational History:** During the active retention period, neither application bugs nor compromised standard application queries can silently alter or delete historical workflow transitions.
- **Defensible Privacy Posture:** Aligns with GDPR Article 5(1)(e) and Article 22 by combining strict role isolation, human-in-the-loop return-offer decisions, privilege-separated erasure of expired personal data, and small-group benchmark suppression.
- **Zero Lost Notifications or Silent Conflicts:** Optimistic concurrency (`version` -> `409`) and the transactional PostgreSQL outbox guarantee consistent board state and reliable downstream notifications.

### Trade-offs & Mitigations
- **Schema & Role Complexity in Phase 2/Phase 7:** Foreign keys, deletion dependencies, and role permissions must be designed upfront in Phase 2 so that Phase 7 retention erasure can cleanly remove expired records without breaking referential integrity or active multi-cohort users.
