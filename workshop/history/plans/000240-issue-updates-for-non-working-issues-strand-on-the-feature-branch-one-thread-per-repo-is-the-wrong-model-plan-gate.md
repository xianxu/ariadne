---
gate: plan-quality
issue: 240
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T08:58:14-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Define recovery after publication succeeds but only part of cleanup completes
          detail: 'ARCH-ORDER: Restored tracked files no longer equal published bytes, and removed untracked files fail admission, so repeating the original selection cannot finish under the stated rules. Name the authoritative state/event model and structural transition boundary, define how already-cleaned paths are recognized safely, and specify controllable interruption testing that proves retry neither republishes nor overwrites later edits.'
          family: partial-progress-recovery
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace prose case inventories with named function-level test strategies
          detail: Tasks 1 and 2 enumerate test cases rather than stating an adversarial input class and mechanical guard per risky function. Give strategies for TrunkProsePath, prosePublishDecision, runTrunkPublish and the cleanup/transition component; ARCH-PURE also requires separating lexical eligibility from filesystem type and identity checks.
          family: function-level-test-strategy
          round: 1
        - id: PQ-3
          severity: Important
          title: Remove the single M1 milestone and its redundant milestone-close
          detail: The issue has one atomic implementation boundary; verification and shipping do not constitute a second milestone. Replace M1 with plain checkboxes and use sdlc close for the mandatory review, consistent with AGENTS.md §3.
          family: atomic-review-boundary
          round: 1
        - id: PQ-4
          severity: Minor
          title: State remote time and selected-input bounds
          detail: 'ARCH-CONSTRAINTS: Explicit paths and three CAS attempts do not bound pinned bytes or a stalled remote operation while holding the shared repository lock. State the relevant budgets or supported ranges, their basis, and exceeded-budget behavior; identify how command cancellation and the existing BoundRemote facility apply.'
          family: explicit-operating-envelope
          round: 1
      blocked: true
---

# Gate ledger — ariadne#240 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T08:58:14-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `partial-progress-recovery` Define recovery after publication succeeds but only part of cleanup completes
  ARCH-ORDER: Restored tracked files no longer equal published bytes, and removed untracked files fail admission, so repeating the original selection cannot finish under the stated rules. Name the authoritative state/event model and structural transition boundary, define how already-cleaned paths are recognized safely, and specify controllable interruption testing that proves retry neither republishes nor overwrites later edits.
- **PQ-2** [Important] `function-level-test-strategy` Replace prose case inventories with named function-level test strategies
  Tasks 1 and 2 enumerate test cases rather than stating an adversarial input class and mechanical guard per risky function. Give strategies for TrunkProsePath, prosePublishDecision, runTrunkPublish and the cleanup/transition component; ARCH-PURE also requires separating lexical eligibility from filesystem type and identity checks.
- **PQ-3** [Important] `atomic-review-boundary` Remove the single M1 milestone and its redundant milestone-close
  The issue has one atomic implementation boundary; verification and shipping do not constitute a second milestone. Replace M1 with plain checkboxes and use sdlc close for the mandatory review, consistent with AGENTS.md §3.
- **PQ-4** [Minor] `explicit-operating-envelope` State remote time and selected-input bounds
  ARCH-CONSTRAINTS: Explicit paths and three CAS attempts do not bound pinned bytes or a stalled remote operation while holding the shared repository lock. State the relevant budgets or supported ranges, their basis, and exceeded-budget behavior; identify how command cancellation and the existing BoundRemote facility apply.

## Open findings

- **PQ-1** [Important] `partial-progress-recovery` Define recovery after publication succeeds but only part of cleanup completes
- **PQ-2** [Important] `function-level-test-strategy` Replace prose case inventories with named function-level test strategies
- **PQ-3** [Important] `atomic-review-boundary` Remove the single M1 milestone and its redundant milestone-close
- **PQ-4** [Minor] `explicit-operating-envelope` State remote time and selected-input bounds
