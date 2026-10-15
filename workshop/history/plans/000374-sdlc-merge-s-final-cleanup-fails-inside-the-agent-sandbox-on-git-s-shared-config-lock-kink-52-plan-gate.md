---
gate: plan-quality
issue: 374
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T23:32:52-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Define how cleanup protects branches recreated after its probes
          detail: 'ARCH-ORDER and ARCH-SECURE: the plan promises recreated-ref protection, but removeObsoleteBranchConfig checks absence and occupancy before a separate remove-section operation (cmd/sdlc/branchconfig.go:41-58); another Git process can recreate the branch between them without taking the SDLC lock. State the supported actor/ownership contract and the mechanism protecting that interval, with deterministic boundary-injection verification.'
          family: cleanup-observation-authority
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace test-case and edit inventories with named adversarial strategies
          detail: The durable plan's task rows enumerate cases and restate edits without a strategy per risky production function. Compress them into function-named lines identifying the adversarial input or event-order class and mechanical guard; distinguish direct unit coverage from real-Git integration coverage and retain the full-merge fake-backed regression.
          family: function-test-strategy
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-14T23:35:06-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Defines cooperating versus manual actors, the shared-lock guarantee, its metadata limitation, and deterministic final-boundary injection.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Replaces inventories with named function strategies and mechanical guards, identifies integration coverage, and retains the full merge fake-backed regression.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-14T23:36:29-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Ownership guarantees distinguish serialized SDLC actors from concurrent manual Git, with deterministic final-boundary verification.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Function-specific strategies identify adversarial input classes and mechanical regression guards.
          round: 3
      blocked: false
content_hash: 4b6c41a036af65a32b77e7d81e1bc2f9657837b422c93828688f14c6ed27aa56
---

# Gate ledger — ariadne#374 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T23:32:52-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `cleanup-observation-authority` Define how cleanup protects branches recreated after its probes
  ARCH-ORDER and ARCH-SECURE: the plan promises recreated-ref protection, but removeObsoleteBranchConfig checks absence and occupancy before a separate remove-section operation (cmd/sdlc/branchconfig.go:41-58); another Git process can recreate the branch between them without taking the SDLC lock. State the supported actor/ownership contract and the mechanism protecting that interval, with deterministic boundary-injection verification.
- **PQ-2** [Important] `function-test-strategy` Replace test-case and edit inventories with named adversarial strategies
  The durable plan's task rows enumerate cases and restate edits without a strategy per risky production function. Compress them into function-named lines identifying the adversarial input or event-order class and mechanical guard; distinguish direct unit coverage from real-Git integration coverage and retain the full-merge fake-backed regression.

## Round 2 — 2026-10-14T23:35:06-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Defines cooperating versus manual actors, the shared-lock guarantee, its metadata limitation, and deterministic final-boundary injection.
- PQ-2 — addressed — Replaces inventories with named function strategies and mechanical guards, identifies integration coverage, and retains the full merge fake-backed regression.

## Round 3 — 2026-10-14T23:36:29-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Ownership guarantees distinguish serialized SDLC actors from concurrent manual Git, with deterministic final-boundary verification.
- PQ-2 — addressed — Function-specific strategies identify adversarial input classes and mechanical regression guards.

## Open findings

(none — every finding has been disposed)
