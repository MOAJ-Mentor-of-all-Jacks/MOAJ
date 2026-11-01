-- MOAJ (Mentor of all Jacks) - Core Workflow & Append-Only Transition Ledger Schema
-- Implementation Target: Phase 2 (Persistence, retention-aware schema, and transactions)
-- Reference: docs/adr/ADR-02-append-only-relational-ledger.md & docs/IMPLEMENTATION_PLAN.md §7 (Phase 2)
--
-- Planned Phase 2 tables and invariants:
--   * installation_metadata, departments, users, cohorts, cohort_memberships
--   * workflow_columns, skill_tags, tasks (with version concurrency counter), task_skill_tags
--   * transition_ledger (append-only during retention; time_in_prev_state_ms; initial state-entry anchor)
--   * intercept_events (6-reason taxonomy: KNOWLEDGE_GAP, CARELESS_ERROR, SCOPE_CREEP,
--     THIRD_PARTY_BLOCKER, ACCIDENTAL_MOVE, OTHER)
--   * Privilege separation between runtime application role, migration role, and retention-erasure role

-- +goose Up
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
