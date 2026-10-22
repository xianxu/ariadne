---
gate: plan-quality
issue: 547
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: d4b2dde1-ab5d-4c06-b276-39eb0b0ebfee
            repository: github.com/xianxu/ariadne
            issue_id: "000547"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: fbe86b2a333463051923fb0cdbaecf80899bbf6f
            slot: ariadne:4
            requested_at: "2026-10-22T12:24:07.338554Z"
          phase: finished
          pid: 92410
          started_at: "2026-10-22T12:24:07.501257Z"
          finished_at: "2026-10-22T12:25:57.700415Z"
          session_id: 01a16391-95d5-7c02-9648-7cc07be1f0f1
          outcome: returned
      "n": 1
      timestamp: "2026-10-22T05:25:58-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Define binding transitions and the intervals they authorize
          detail: 'ARCH-ORDER: The governing producer contract promises multiple bounded train bindings but names only start/bind/end facts, without defining what closes a binding or whether released/unbound gaps contribute coverage. Specify the pure reducer''s legal transitions, authoritative closure evidence and interval semantics for release/reformation, cancellation and terminal outcomes; use reproducible sequence tests through that reducer.'
          family: evidence-transition-contract
          round: 1
        - id: PQ-2
          severity: Important
          title: Specify bounded diagnostic overflow without invalidating measured hours
          detail: 'ARCH-CONSTRAINTS: The plan requests explicit audit bounds but supplies no ceilings or exceeded-budget behavior, and optional coverage shares the existing 64 KiB snapshot limit enforced at cmd/sdlc/internal/gatestate/measurement.go:271–276. Name concrete bounds or existing constants, their basis, and a bounded incomplete-coverage fallback for diagnostic capacity failures that preserves existing accounting evidence and hours/status.'
          family: diagnostic-budget-isolation
          round: 1
        - id: PQ-3
          severity: Important
          title: Replace prose case inventories with function-named test strategies
          detail: The risky-function table mostly names generic seams, and the revised checklist enumerates cases while retaining superseded exclusion/correction strategies in the earlier table. Add a governing strategy table naming each risky function, its adversarial input or event class and its mechanical guard; compress checklist rows to reference those strategies and leave concrete case inventories to executable tests.
          family: function-named-test-contract
          round: 1
      blocked: true
---

# Gate ledger — ariadne#547 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T05:25:58-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `evidence-transition-contract` Define binding transitions and the intervals they authorize
  ARCH-ORDER: The governing producer contract promises multiple bounded train bindings but names only start/bind/end facts, without defining what closes a binding or whether released/unbound gaps contribute coverage. Specify the pure reducer's legal transitions, authoritative closure evidence and interval semantics for release/reformation, cancellation and terminal outcomes; use reproducible sequence tests through that reducer.
- **PQ-2** [Important] `diagnostic-budget-isolation` Specify bounded diagnostic overflow without invalidating measured hours
  ARCH-CONSTRAINTS: The plan requests explicit audit bounds but supplies no ceilings or exceeded-budget behavior, and optional coverage shares the existing 64 KiB snapshot limit enforced at cmd/sdlc/internal/gatestate/measurement.go:271–276. Name concrete bounds or existing constants, their basis, and a bounded incomplete-coverage fallback for diagnostic capacity failures that preserves existing accounting evidence and hours/status.
- **PQ-3** [Important] `function-named-test-contract` Replace prose case inventories with function-named test strategies
  The risky-function table mostly names generic seams, and the revised checklist enumerates cases while retaining superseded exclusion/correction strategies in the earlier table. Add a governing strategy table naming each risky function, its adversarial input or event class and its mechanical guard; compress checklist rows to reference those strategies and leave concrete case inventories to executable tests.

## Open findings

- **PQ-1** [Important] `evidence-transition-contract` Define binding transitions and the intervals they authorize
- **PQ-2** [Important] `diagnostic-budget-isolation` Specify bounded diagnostic overflow without invalidating measured hours
- **PQ-3** [Important] `function-named-test-contract` Replace prose case inventories with function-named test strategies
