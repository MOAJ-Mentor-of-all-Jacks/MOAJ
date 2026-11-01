-- MOAJ (Mentor of all Jacks) - SQLc Queries for Workflow Tasks, Append-Only Ledger & Intercepts
-- Implementation Target: Phase 2 (Persistence, retention-aware schema, and transactions)
-- Reference: docs/adr/ADR-01-go-sqlc-single-binary.md & docs/adr/ADR-02-append-only-relational-ledger.md

-- name: PingDatabase :one
SELECT 1::int AS ok;
