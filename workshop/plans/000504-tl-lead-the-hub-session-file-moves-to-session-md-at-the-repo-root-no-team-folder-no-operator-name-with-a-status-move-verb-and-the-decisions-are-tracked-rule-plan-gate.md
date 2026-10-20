---
gate: plan-quality
issue: 504
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 156f78ce-d411-46f5-8abb-81767f4a0c9d
            repository: github.com/xianxu/ariadne
            issue_id: "000504"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: eeff4ddce3819d247495ebac07e74da2cf2ea5df
            slot: ariadne:4
            requested_at: "2026-10-20T17:00:58.300806Z"
          phase: finished
          pid: 20815
          started_at: "2026-10-20T17:00:58.391507Z"
          finished_at: "2026-10-20T17:02:33.469316Z"
          session_id: 01a15a42-5379-7c20-88c4-31004da81ecc
          outcome: returned
      "n": 1
      timestamp: "2026-10-20T10:02:37-07:00"
      agent: codex
      reviewer: 'codex (openai; author openai); degraded: same-family by choice'
      findings:
        - id: PQ-1
          severity: Important
          title: Define how unregistered legacy session state survives migration
          detail: 'ARCH-ORDER/ARCH-PURPOSE: the plan promises legacy routing but Couch Rename refuses unregistered sources, and legacy shadow/sync state is path-specific. Specify a state-preserving Couch import before rename, or an explicit registration prerequisite with refusal before effects; copying only the seen set does not preserve the promised unread and owed state.'
          family: migration-state-continuity
          round: 1
        - id: PQ-2
          severity: Important
          title: Adapt index rebasing for moving parents and preserve unrelated metadata
          detail: 'ARCH-DRY/ARCH-EVOLVE/ARCH-SECURE: cmd/tl-lead/migrate.go:357 uses one directory for both input and output coordinates and rewrites every frontmatter - doc: line. Specify old/new coordinates and rewriting confined to the actual discussions index so root relocation resolves correctly and unknown metadata survives unchanged.'
          family: relocation-transform-contract
          round: 1
        - id: PQ-3
          severity: Important
          title: Replace prose test inventories with named function strategies
          detail: Tasks 1–3 enumerate test cases instead of naming every risky production function with its adversarial input class and mechanical guard. Compress these inventories into function-level strategies, retaining controlled interruption and stateful reconciliation for the executor.
          family: function-level-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#504 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T10:02:37-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `migration-state-continuity` Define how unregistered legacy session state survives migration
  ARCH-ORDER/ARCH-PURPOSE: the plan promises legacy routing but Couch Rename refuses unregistered sources, and legacy shadow/sync state is path-specific. Specify a state-preserving Couch import before rename, or an explicit registration prerequisite with refusal before effects; copying only the seen set does not preserve the promised unread and owed state.
- **PQ-2** [Important] `relocation-transform-contract` Adapt index rebasing for moving parents and preserve unrelated metadata
  ARCH-DRY/ARCH-EVOLVE/ARCH-SECURE: cmd/tl-lead/migrate.go:357 uses one directory for both input and output coordinates and rewrites every frontmatter - doc: line. Specify old/new coordinates and rewriting confined to the actual discussions index so root relocation resolves correctly and unknown metadata survives unchanged.
- **PQ-3** [Important] `function-level-test-strategy` Replace prose test inventories with named function strategies
  Tasks 1–3 enumerate test cases instead of naming every risky production function with its adversarial input class and mechanical guard. Compress these inventories into function-level strategies, retaining controlled interruption and stateful reconciliation for the executor.

## Open findings

- **PQ-1** [Important] `migration-state-continuity` Define how unregistered legacy session state survives migration
- **PQ-2** [Important] `relocation-transform-contract` Adapt index rebasing for moving parents and preserve unrelated metadata
- **PQ-3** [Important] `function-level-test-strategy` Replace prose test inventories with named function strategies
