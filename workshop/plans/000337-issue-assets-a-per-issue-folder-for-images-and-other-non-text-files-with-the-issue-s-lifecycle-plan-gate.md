---
gate: plan-quality
issue: 337
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-13T13:11:21-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Adapt existing issue-discovery readers before adding binary subtrees
          detail: 'ARCH-CONSTRAINTS / ARCH-PURPOSE: finishLandedLeftovers reads the entire issues subtree through Files (cmd/sdlc/externalmerge.go:217), and migrationInventory does likewise before filtering direct issue files (cmd/sdlc/issuemigrate.go:167). Files loads every descendant blob under a shared 32 MiB output ceiling (cmd/sdlc/internal/gitx/snapshot.go:62; boundedoutput.go:16), so several individually valid bundles can break lifecycle cleanup and legacy tracker migration; include these readers in the scope and select issue documents before loading bytes, with an aggregate-volume regression strategy.'
          family: bounded-reader-scope
          round: 1
        - id: PQ-2
          severity: Important
          title: Define attachment failure and restart transitions before implementing the writer
          detail: 'ARCH-ORDER / ARCH-SECURE: the plan promises refusals leave files unchanged (plan:63), but describes an installed asset surviving interruption and only temporary-file removal on ordinary failure (plan:115–121); it does not specify what happens when body validation or rename fails after installation. Name the transition function and state/event rules, distinguish pre-existing bytes from bytes owned by this attempt, and define rollback versus retained recovery state plus the evidence permitting restart, including how recoverable residue differs from the unlinked files that otherwise refuse.'
          family: explicit-recovery-transitions
          round: 1
        - id: PQ-3
          severity: Important
          title: Replace prose test-case inventories with named functions and adversarial strategies
          detail: Tasks 1–3 enumerate test cases (plan:140–149, 160–168, 179–186), while the core table names types and concepts rather than the functions to unit-test. Compress these inventories into named parser, bundle-delta, policy-decision, attachment-transition, publication-decision, and archive-planner functions, each with one strategy line identifying its adversarial input class and mechanical guard; use fuzzing for malformed input and reproducible event-sequence testing for recovery.
          family: function-level-test-strategy
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-13T13:13:36-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Metadata-first discovery, a subtree-reader rule, and an aggregate-volume regression cover the affected reader class.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: NextAttach defines ownership-aware rollback, retained state, uncertain-outcome reconciliation, and exact-byte restart adoption.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Named functions now have adversarial strategies and mechanical guards, including fuzzing and reproducible recovery sequences.
          round: 2
      blocked: false
content_hash: 82de980021ee7c409cce19204476c71303574888d619bab466bed6e30d090d63
---

# Gate ledger — ariadne#337 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T13:11:21-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `bounded-reader-scope` Adapt existing issue-discovery readers before adding binary subtrees
  ARCH-CONSTRAINTS / ARCH-PURPOSE: finishLandedLeftovers reads the entire issues subtree through Files (cmd/sdlc/externalmerge.go:217), and migrationInventory does likewise before filtering direct issue files (cmd/sdlc/issuemigrate.go:167). Files loads every descendant blob under a shared 32 MiB output ceiling (cmd/sdlc/internal/gitx/snapshot.go:62; boundedoutput.go:16), so several individually valid bundles can break lifecycle cleanup and legacy tracker migration; include these readers in the scope and select issue documents before loading bytes, with an aggregate-volume regression strategy.
- **PQ-2** [Important] `explicit-recovery-transitions` Define attachment failure and restart transitions before implementing the writer
  ARCH-ORDER / ARCH-SECURE: the plan promises refusals leave files unchanged (plan:63), but describes an installed asset surviving interruption and only temporary-file removal on ordinary failure (plan:115–121); it does not specify what happens when body validation or rename fails after installation. Name the transition function and state/event rules, distinguish pre-existing bytes from bytes owned by this attempt, and define rollback versus retained recovery state plus the evidence permitting restart, including how recoverable residue differs from the unlinked files that otherwise refuse.
- **PQ-3** [Important] `function-level-test-strategy` Replace prose test-case inventories with named functions and adversarial strategies
  Tasks 1–3 enumerate test cases (plan:140–149, 160–168, 179–186), while the core table names types and concepts rather than the functions to unit-test. Compress these inventories into named parser, bundle-delta, policy-decision, attachment-transition, publication-decision, and archive-planner functions, each with one strategy line identifying its adversarial input class and mechanical guard; use fuzzing for malformed input and reproducible event-sequence testing for recovery.

## Round 2 — 2026-10-13T13:13:36-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Metadata-first discovery, a subtree-reader rule, and an aggregate-volume regression cover the affected reader class.
- PQ-2 — addressed — NextAttach defines ownership-aware rollback, retained state, uncertain-outcome reconciliation, and exact-byte restart adoption.
- PQ-3 — addressed — Named functions now have adversarial strategies and mechanical guards, including fuzzing and reproducible recovery sequences.

## Open findings

(none — every finding has been disposed)
