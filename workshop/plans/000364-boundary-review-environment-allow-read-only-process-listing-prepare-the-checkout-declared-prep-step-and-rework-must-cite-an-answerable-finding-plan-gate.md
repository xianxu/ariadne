---
gate: plan-quality
issue: 364
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T00:04:59-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Replace test-case inventories with named function strategies
          detail: Tasks 2 and 3 enumerate test cases but do not name the new unit-tested functions or state one adversarial-input class and mechanical guard per risky function. Name the declaration parser, response admission function and affected existing pure functions; compress the inventories into function-level strategies, including malformed-input fuzzing and reproducible ordering checks where applicable.
          family: function-level-test-strategy
          round: 1
        - id: PQ-2
          severity: Important
          title: Define the lifecycle of appended preparation and protocol diagnostics
          detail: 'ARCH-FUNERAL: the plan caps diagnostics per attempt, but existing writeReviewSidecar appends across invocations and archival does not remove that growth. Specify a writer-enforced retention or superseding policy, or explicitly declare permanent retention with its added cost per dispatch and where that cost is measured; include repeated invalid-response invocations in the contract.'
          family: durable-diagnostic-retention
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-14T00:10:19-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Named function strategies replace Tasks 2 and 3 test inventories, including malformed-input fuzzing, reproducible ordering checks, and mechanical guard verification.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Permanent retention is explicit, with a 256 KiB added diagnostic envelope per dispatch, append-time byte measurement, and repeated failed-invocation verification.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-14T00:12:11-07:00"
      agent: codex
      blocked: false
content_hash: 0fcb9ddc5b61b90d613d3a774e29fd76754eb8f217de5f1d4eef7750d237b408
---

# Gate ledger — ariadne#364 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T00:04:59-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `function-level-test-strategy` Replace test-case inventories with named function strategies
  Tasks 2 and 3 enumerate test cases but do not name the new unit-tested functions or state one adversarial-input class and mechanical guard per risky function. Name the declaration parser, response admission function and affected existing pure functions; compress the inventories into function-level strategies, including malformed-input fuzzing and reproducible ordering checks where applicable.
- **PQ-2** [Important] `durable-diagnostic-retention` Define the lifecycle of appended preparation and protocol diagnostics
  ARCH-FUNERAL: the plan caps diagnostics per attempt, but existing writeReviewSidecar appends across invocations and archival does not remove that growth. Specify a writer-enforced retention or superseding policy, or explicitly declare permanent retention with its added cost per dispatch and where that cost is measured; include repeated invalid-response invocations in the contract.

## Round 2 — 2026-10-14T00:10:19-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Named function strategies replace Tasks 2 and 3 test inventories, including malformed-input fuzzing, reproducible ordering checks, and mechanical guard verification.
- PQ-2 — addressed — Permanent retention is explicit, with a 256 KiB added diagnostic envelope per dispatch, append-time byte measurement, and repeated failed-invocation verification.

## Round 3 — 2026-10-14T00:12:11-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
