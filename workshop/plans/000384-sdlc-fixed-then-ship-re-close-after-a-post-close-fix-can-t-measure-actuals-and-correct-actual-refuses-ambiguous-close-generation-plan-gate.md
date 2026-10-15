---
gate: plan-quality
issue: 384
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-15T02:50:04-07:00"
      agent: codex
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Compress test inventories into strategies for named risky functions
          detail: workshop/plans/000384-reclose-actuals-plan.md:67–68 enumerates test cases that this gate explicitly rejects; replace those inventories with one adversarial-input-class and mechanical-guard strategy per risky function. At line 85, name the actual resolver and close functions instead of “Owned resolver and close caller,” retaining the production lifecycle regression and mutation checks.
          family: function-level-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#384 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T02:50:04-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `function-level-test-strategy` (kind: defect) Compress test inventories into strategies for named risky functions
  workshop/plans/000384-reclose-actuals-plan.md:67–68 enumerates test cases that this gate explicitly rejects; replace those inventories with one adversarial-input-class and mechanical-guard strategy per risky function. At line 85, name the actual resolver and close functions instead of “Owned resolver and close caller,” retaining the production lifecycle regression and mutation checks.

## Open findings

- **PQ-1** [Important] `function-level-test-strategy` (kind: defect) Compress test inventories into strategies for named risky functions
