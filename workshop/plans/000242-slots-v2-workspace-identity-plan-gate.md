---
gate: plan-quality
issue: 242
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-22T22:31:26-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Replace prose test-case inventories with named adversarial testing strategies.
          detail: Tasks 1–4 enumerate cases without the required strategy per risky function. Name ParseWorktrees, ParseAddress, SlotPath, Classify, NormalizeVantage, and Resolve as test targets, giving each its adversarial input class and mechanical guard; compress the case lists while retaining production-consumer coverage. In particular, specify reproducible between-read mutation injection for Resolve and malformed-input/property testing for the pure core (ARCH-PURE, ARCH-SECURE, ARCH-ORDER).
          family: function-level-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#242 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-22T22:31:26-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `function-level-test-strategy` Replace prose test-case inventories with named adversarial testing strategies.
  Tasks 1–4 enumerate cases without the required strategy per risky function. Name ParseWorktrees, ParseAddress, SlotPath, Classify, NormalizeVantage, and Resolve as test targets, giving each its adversarial input class and mechanical guard; compress the case lists while retaining production-consumer coverage. In particular, specify reproducible between-read mutation injection for Resolve and malformed-input/property testing for the pure core (ARCH-PURE, ARCH-SECURE, ARCH-ORDER).

## Open findings

- **PQ-1** [Important] `function-level-test-strategy` Replace prose test-case inventories with named adversarial testing strategies.
