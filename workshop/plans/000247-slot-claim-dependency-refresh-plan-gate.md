---
gate: plan-quality
issue: 247
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-23T22:50:53-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Name the unit-tested functions and replace prose test-case inventories with strategy lines
          detail: The plan lists scenarios across lines 103–106 and 113–127, but does not name the concrete functions to unit-test or give one adversarial-input/mechanical-guard strategy per risky function. Compress these into function names such as the eligibility validator, phase transition function, Git observation parser, preflight validator, apply revalidator, and compile orchestration seam, each with its malformed/racy input class and guard.
          family: test-strategy-contract
          round: 1
        - id: PQ-2
          severity: Important
          title: Define the Git fake's conformance path and cadence
          detail: The plan names a stateful fake behind GitRunner and real Git fixtures, but ARCH-MOCK also requires a live or scheduled conformance check comparing the fake with the real Git binary. Specify which behaviors are modeled, how the shared seam is exercised by both flows, and when the conformance check runs; otherwise the fake can silently drift from Git semantics.
          family: external-double-conformance
          round: 1
        - id: PQ-3
          severity: Important
          title: Make the Pair project update an explicit coordinated dependency or remove it from the implementation plan
          detail: Task 4 requires updating `../pair/workshop/projects/couch-slots-v2.md`, but issue frontmatter declares no dependency and the plan gives no exact section, ownership, commit/publication protocol, or coordination rule. The phrase “without touching concurrent work there” is not an executable write policy; either define the cross-repo handoff or leave the project sweep to the SDLC close gate.
          family: cross-repo-artifact-ownership
          round: 1
      blocked: true
---

# Gate ledger — ariadne#247 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-23T22:50:53-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `test-strategy-contract` Name the unit-tested functions and replace prose test-case inventories with strategy lines
  The plan lists scenarios across lines 103–106 and 113–127, but does not name the concrete functions to unit-test or give one adversarial-input/mechanical-guard strategy per risky function. Compress these into function names such as the eligibility validator, phase transition function, Git observation parser, preflight validator, apply revalidator, and compile orchestration seam, each with its malformed/racy input class and guard.
- **PQ-2** [Important] `external-double-conformance` Define the Git fake's conformance path and cadence
  The plan names a stateful fake behind GitRunner and real Git fixtures, but ARCH-MOCK also requires a live or scheduled conformance check comparing the fake with the real Git binary. Specify which behaviors are modeled, how the shared seam is exercised by both flows, and when the conformance check runs; otherwise the fake can silently drift from Git semantics.
- **PQ-3** [Important] `cross-repo-artifact-ownership` Make the Pair project update an explicit coordinated dependency or remove it from the implementation plan
  Task 4 requires updating `../pair/workshop/projects/couch-slots-v2.md`, but issue frontmatter declares no dependency and the plan gives no exact section, ownership, commit/publication protocol, or coordination rule. The phrase “without touching concurrent work there” is not an executable write policy; either define the cross-repo handoff or leave the project sweep to the SDLC close gate.

## Open findings

- **PQ-1** [Important] `test-strategy-contract` Name the unit-tested functions and replace prose test-case inventories with strategy lines
- **PQ-2** [Important] `external-double-conformance` Define the Git fake's conformance path and cadence
- **PQ-3** [Important] `cross-repo-artifact-ownership` Make the Pair project update an explicit coordinated dependency or remove it from the implementation plan
