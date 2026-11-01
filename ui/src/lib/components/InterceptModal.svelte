<!--
  MOAJ (Mentor of all Jacks) - Synchronous HTTP 428 Workflow Intercept Modal Scaffold
  Implementation Target: Phase 4 (First usable workflow slice)
  Reference: docs/adr/ADR-03-synchronous-http-428-intercept.md & docs/IMPLEMENTATION_PLAN.md §5.1

  Planned Phase 4 capabilities:
    * Triggered when PATCH /api/v1/tasks/{id}/transition returns HTTP 428 Precondition Required
    * Captures the 6-reason taxonomy (KNOWLEDGE_GAP, CARELESS_ERROR, SCOPE_CREEP,
      THIRD_PARTY_BLOCKER, ACCIDENTAL_MOVE, OTHER with bounded max-280-char explanation)
    * Enforces role-appropriate attribution UX (mentor/admin classification vs. intern workflow)
-->
<div class="intercept-modal-scaffold" aria-hidden="true"></div>
