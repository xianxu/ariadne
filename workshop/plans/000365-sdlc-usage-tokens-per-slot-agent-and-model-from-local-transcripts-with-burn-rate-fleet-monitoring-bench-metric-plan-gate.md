---
gate: plan-quality
issue: 365
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T08:53:47-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Define reviewer accounting persistence independently of gate advancement.
          detail: 'ARCH-ORDER: existing planning and boundary paths refuse persistence after cancellation, stale validation, or lock reacquisition failure, so adding optional fields to their final records does not establish the promised capture of every launched attempt. Name the durable attempt states, write timing, lock/snapshot interaction, and recovery or explicit coverage policy for process death and retry; test accounting retention without findings, round-cap, or pass-through advancement.'
          family: accounting-evidence-survives-interruption
          round: 1
        - id: PQ-2
          severity: Important
          title: Give explicit reviewer bindings precedence over worker ownership eligibility.
          detail: 'ARCH-PURPOSE: in-place plan and estimate reviewers run in an owned checkout and can satisfy the same root/window eligibility as workers. Define exclusive role assignment and conflicting-binding behavior, and add a production selection/aggregation invariant proving worker plus reviewer subtotals count each response exactly once.'
          family: attribution-roles-are-exclusive
          round: 1
        - id: PQ-3
          severity: Minor
          title: Name the live conformance cadence for reviewer identity protocols.
          detail: 'ARCH-MOCK: recorded startup preambles and fake Run tests establish fixtures, but the plan supplies no recurring comparison with live Claude/Codex identity emission. Name an opt-in conformance check and its cadence, such as supported CLI upgrades, so protocol drift is detected beyond the initial acceptance smoke.'
          family: external-model-conformance-cadence
          round: 1
      blocked: true
---

# Gate ledger — ariadne#365 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T08:53:47-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `accounting-evidence-survives-interruption` Define reviewer accounting persistence independently of gate advancement.
  ARCH-ORDER: existing planning and boundary paths refuse persistence after cancellation, stale validation, or lock reacquisition failure, so adding optional fields to their final records does not establish the promised capture of every launched attempt. Name the durable attempt states, write timing, lock/snapshot interaction, and recovery or explicit coverage policy for process death and retry; test accounting retention without findings, round-cap, or pass-through advancement.
- **PQ-2** [Important] `attribution-roles-are-exclusive` Give explicit reviewer bindings precedence over worker ownership eligibility.
  ARCH-PURPOSE: in-place plan and estimate reviewers run in an owned checkout and can satisfy the same root/window eligibility as workers. Define exclusive role assignment and conflicting-binding behavior, and add a production selection/aggregation invariant proving worker plus reviewer subtotals count each response exactly once.
- **PQ-3** [Minor] `external-model-conformance-cadence` Name the live conformance cadence for reviewer identity protocols.
  ARCH-MOCK: recorded startup preambles and fake Run tests establish fixtures, but the plan supplies no recurring comparison with live Claude/Codex identity emission. Name an opt-in conformance check and its cadence, such as supported CLI upgrades, so protocol drift is detected beyond the initial acceptance smoke.

## Open findings

- **PQ-1** [Important] `accounting-evidence-survives-interruption` Define reviewer accounting persistence independently of gate advancement.
- **PQ-2** [Important] `attribution-roles-are-exclusive` Give explicit reviewer bindings precedence over worker ownership eligibility.
- **PQ-3** [Minor] `external-model-conformance-cadence` Name the live conformance cadence for reviewer identity protocols.
