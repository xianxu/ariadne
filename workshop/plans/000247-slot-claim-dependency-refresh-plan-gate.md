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
    - "n": 2
      timestamp: "2026-09-23T22:53:30-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The plan names concrete unit-tested functions and gives adversarial-input/mechanical-guard strategies at plan lines 95–106.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The plan defines the stateful fake, shared GitRunner seam, semantic comparisons, and conformance cadence at lines 108–118.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Task 4 removes the peer-project write and delegates project ticking to the SDLC close gate at line 153.
          round: 2
      findings:
        - id: PQ-4
          severity: Important
          title: Define a strict refresh discovery path instead of assuming Restore refuses missing checkouts
          detail: The plan states that `Restore(ctx, root, true)` “refuses a missing checkout” at plan line 37, but the existing implementation treats dry-run missing repositories as an incomplete graph and records them in `Result.Missing` (`cmd/weave/internal/acquire/acquire.go:240-248`, `cmd/weave/internal/acquire/acquire.go:326-330`); existing CLI code reports those entries and continues (`cmd/weave/main.go:476-483`). Specify whether refresh adds a strict discovery option/wrapper that rejects `Result.Missing`, or changes Restore semantics while preserving ordinary compile/dependency behavior, and name the regression tests for both paths.
          family: existing-behavior-grounding
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-23T22:54:54-07:00"
      agent: codex
      dispose:
        - id: PQ-4
          disposition: addressed
          note: The plan defines strict discoverRefresh behavior that rejects errors and nonempty Result.Missing before fetch/update, with named regression tests preserving ordinary Restore semantics.
          round: 3
      blocked: false
content_hash: bf3233b99690bf88e0c29e709a33850a7ebae5648aacc4b3b60aa61597e486f8
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

## Round 2 — 2026-09-23T22:53:30-07:00 (codex) — BLOCKED

### Disposed

- PQ-1 — addressed — The plan names concrete unit-tested functions and gives adversarial-input/mechanical-guard strategies at plan lines 95–106.
- PQ-2 — addressed — The plan defines the stateful fake, shared GitRunner seam, semantic comparisons, and conformance cadence at lines 108–118.
- PQ-3 — addressed — Task 4 removes the peer-project write and delegates project ticking to the SDLC close gate at line 153.

### Raised

- **PQ-4** [Important] `existing-behavior-grounding` Define a strict refresh discovery path instead of assuming Restore refuses missing checkouts
  The plan states that `Restore(ctx, root, true)` “refuses a missing checkout” at plan line 37, but the existing implementation treats dry-run missing repositories as an incomplete graph and records them in `Result.Missing` (`cmd/weave/internal/acquire/acquire.go:240-248`, `cmd/weave/internal/acquire/acquire.go:326-330`); existing CLI code reports those entries and continues (`cmd/weave/main.go:476-483`). Specify whether refresh adds a strict discovery option/wrapper that rejects `Result.Missing`, or changes Restore semantics while preserving ordinary compile/dependency behavior, and name the regression tests for both paths.

## Round 3 — 2026-09-23T22:54:54-07:00 (codex) — passed

### Disposed

- PQ-4 — addressed — The plan defines strict discoverRefresh behavior that rejects errors and nonempty Result.Missing before fetch/update, with named regression tests preserving ordinary Restore semantics.

## Open findings

(none — every finding has been disposed)
