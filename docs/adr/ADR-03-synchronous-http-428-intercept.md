# ADR-03: Synchronous `HTTP 428 Precondition Required` Workflow Intercepts & First-Bounce Attribution

- **Status:** Accepted (Amended by `docs/IMPLEMENTATION_PLAN.md` v2.1)
- **Date:** 2026-11-01
- **References:** `IMS-BRD-1.0` (`BR-INT-01`, `BR-INT-02`, `BR-INT-03`), `IMS Feature Specification.pdf` (§1.3), `IMS Full System Architecture.pdf` (Blueprint v1.0 §3), `docs/IMPLEMENTATION_PLAN.md` (§3.1, §3.2 `D14`, §5.1, §5.2)

---

## 1. Context

A core differentiator of **MOAJ (Mentor of all Jacks)** is capturing structured root-cause attribution at the exact moment workflow friction occurs—such as when a task bounces backward from code review to in-progress, exceeds 300% of its column SLA, or repeatedly oscillates into a `Blocked` state.

If root-cause capture is implemented as an optional post-hoc comment or asynchronous prompt after the card has already moved, users routinely dismiss or skip the prompt, resulting in unattributed revisions and subjective end-of-internship evaluations.

At the same time, Phase 0 review (`docs/IMPLEMENTATION_PLAN.md` §3.1 & §5.1) identified critical contradictions in the original PDFs that required explicit resolution:
1. **First-Bounce vs. Second-Bounce Contradiction:** `IMS-BRD-1.0` KPI requires **100% of backward task movements** to carry a structured reason code, whereas Blueprint v1.0 §3.1 pseudo-code gated Trigger A on `task.revision_count >= 1` (which skipped the first backward bounce).
2. **Rigid 4-Code Taxonomy:** The original four codes (`KNOWLEDGE_GAP`, `CARELESS_ERROR`, `SCOPE_CREEP`, `THIRD_PARTY_BLOCKER`) left no honest classification for accidental drag-and-drop mistakes (`ACCIDENTAL_MOVE`) or legitimate edge cases (`OTHER`).
3. **Self-Incrimination / Authorization Ambiguity:** Allowing an intern to assign authoritative HR fault codes (`KNOWLEDGE_GAP` vs. `CARELESS_ERROR`) to themselves or peers creates skewed incentives and governance risk.

---

## 2. Decision

### 2.1 Synchronous `HTTP 428 Precondition Required` Handshake
When a client submits `PATCH /api/v1/tasks/{id}/transition`:
1. The application service locks the task, verifies caller authorization and expected task `version` (returning `409 Conflict` if stale), and evaluates the three deterministic friction triggers against the proposed transition:
   - **Trigger A (Backward Regression):** `target_column.position_index < current_column.position_index`.
   - **Trigger B (Extreme SLA Overrun):** `current_column.sla_limit_hours` is enabled (`> 0`) and `time_in_prev_state_ms > (sla_limit_hours * 3600 * 1000 * 3.0)` (strictly greater than 300% of SLA).
   - **Trigger C (Blocked Oscillation, `D14`):** Moving into a column marked `is_blocked_state = true` for the **third (or subsequent) time within a rolling 7-day window** (boundary inclusivity explicitly tested in Phase 4).
2. **Abort & Challenge (`HTTP 428`):** If one or more triggers fire and the request payload does not contain a valid `intercept_justification`, the server **rolls back / releases the database transaction without committing the move** and responds with `HTTP 428 Precondition Required` (using an RFC 9457 Problem Details payload listing the active trigger code(s), required taxonomy options, and current task version).
3. **Resubmission & Re-Verification:** Because no database lock is held while the user interacts with the `InterceptModal.svelte` UI, the resubmitted `PATCH` request (now carrying `intercept_justification`) re-runs full authorization, `version` concurrency check, and trigger evaluation before atomically committing the `tasks` update, `transition_ledger` row, and `intercept_events` record.

---

## 3. Specification Amendments Recorded in This ADR

### 3.1 First-Bounce Trigger A Rule (`docs/IMPLEMENTATION_PLAN.md` §3.1 & §5.1)
- **Supersedes:** Blueprint v1.0 §3.1 `AND task.revision_count >= 1` ("second bounce") condition.
- **Amended Rule:** Trigger A fires on the **first and every subsequent backward transition** (`target_column.position_index < current_column.position_index`). Every backward movement requires a structured justification reason, directly satisfying the BRD's 100% backward-bounce attribution KPI.

### 3.2 Expanded 6-Reason Taxonomy & Auditable Accidental Corrections
- **Supersedes:** The four-code-only taxonomy (`KNOWLEDGE_GAP`, `CARELESS_ERROR`, `SCOPE_CREEP`, `THIRD_PARTY_BLOCKER`).
- **Amended Taxonomy (6 Codes):**
  1. `KNOWLEDGE_GAP` — Intern requires mentoring or pairing on an unfamiliar concept/skill (counts toward competency support recommendations).
  2. `CARELESS_ERROR` — Preventable oversight against established standards (counts toward attributed performance revisions).
  3. `SCOPE_CREEP` — Requirements changed after task start (excluded from negative intern performance revisions; exclusion interval tracked per §5.2).
  4. `THIRD_PARTY_BLOCKER` — External dependency, upstream outage, or vendor delay outside the intern's control (excluded from negative intern performance revisions).
  5. `ACCIDENTAL_MOVE` — Mistaken drag-and-drop or status click being corrected. References the mistaken movement event; both the original movement and correction remain in the append-only ledger for auditability, but confirmed accidental movement/correction pairs are excluded from performance revision counts.
  6. `OTHER` — Requires a bounded free-text explanation (`max 280 chars`). Treated as sensitive, unclassified context for mentor review (never automatically counted as a negative fault judgment, and purged at retention expiry).

### 3.3 Simultaneous Triggers & Single-Count HR Attribution
- When a single transition satisfies multiple triggers simultaneously (for example, moving backward out of a column after exceeding 300% of its SLA), the `428` challenge reports all triggered conditions and records which triggers fired on the resulting `intercept_events` entry, but HR summary totals never double-count a single transition event as multiple revisions.

### 3.4 Role Authorization for Fault Classification
- Authoritative HR fault classifications (`KNOWLEDGE_GAP`, `CARELESS_ERROR`, `SCOPE_CREEP`, `THIRD_PARTY_BLOCKER`) are supplied by an authorized `MENTOR` or administrator (`DEPT_ADMIN`, `HR_ADMIN`, `SYS_ADMIN`). Interns (`INTERN`) must never assign authoritative fault judgments to peers, nor be forced into self-incriminating fault labels when requesting help; if an intern-initiated transition requires mentor classification, the workflow routes/prompts accordingly (finalized in Phase 4 UX and tested against the role matrix).

---

## 4. Consequences

### Positive
- **100% Structured Attribution Without Orphaned Moves:** No backward transition, 300% SLA overrun, or 3x blocked oscillation can enter the ledger without an explicit, reviewable classification.
- **No Long-Lived Database Locks:** Releasing the transaction on `HTTP 428` and re-verifying `version` on resubmission prevents idle browser modals from exhausting PostgreSQL connections or holding row locks.
- **Fairness & Data Honesty:** Adding `ACCIDENTAL_MOVE` and `OTHER` prevents users from polluting `KNOWLEDGE_GAP` or `SCOPE_CREEP` metrics when correcting a mistaken drag-and-drop.

### Trade-offs & Mitigations
- **Expected `428` & `409` Responses in HTTP Telemetry:** Because `428` and `409` are normal domain control-flow signals rather than server failures, `internal/observability` (`moaj_http_requests_total`) and Prometheus alert rules explicitly separate `409`/`428` outcomes from `5xx` server errors.
