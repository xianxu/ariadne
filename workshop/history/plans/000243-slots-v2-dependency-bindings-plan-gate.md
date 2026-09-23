---
gate: plan-quality
issue: 243
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-23T10:20:21-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Name the production functions under test and compress the prose test inventory
          detail: The adversarial verification section (plan:129-138) enumerates scenarios but does not identify the functions that will be unit-tested or provide one strategy line per risky function. Replace the case inventory with named production functions, each paired with its malformed/adversarial input class and mechanical guard.
          family: test-strategy-contract
          round: 1
        - id: PQ-2
          severity: Important
          title: Define the exact Couch-facing provisioning contract
          detail: 'The plan names `weave compile` and `weave dependencies` but leaves the callable contract underspecified: invocation inputs, required cwd/context, outputs, exit/error classes, contention behavior, and retry semantics are not concrete enough for Couch to implement against. Specify the exact command/API and recovery mapping.'
          family: couch-provisioning-contract
          round: 1
        - id: PQ-3
          severity: Important
          title: Specify schema-v2 handling for older and malformed workspace state
          detail: Changing the machine contract from schema v1 to v2 without defining read compatibility or rejection behavior leaves persisted/older state ambiguous. State whether v1, truncated, malformed, and unknown-version documents are accepted, migrated, or rejected, and name the production parser and tests covering those paths.
          family: schema-version-compatibility
          round: 1
        - id: PQ-4
          severity: Minor
          title: State the cadence for live fake-versus-real conformance
          detail: ARCH-MOCK is addressed with stateful fakes and real acceptance tests, but the plan does not say when live conformance runs to detect Git/tool behavior drift.
          family: external-conformance-cadence
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-23T10:22:21-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The adversarial verification table now names production functions and gives one strategy/mechanical guard per risky function.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The plan specifies commands, cwd, inputs, outputs, exit behavior, contention, failure mapping, and retry behavior for Couch.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: The plan establishes that workspace JSON is live output rather than a persisted input protocol, with v1 treated as stale and unsupported/malformed consumer input rejected.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Git fake-versus-real conformance is scheduled in the normal package suite on seam changes and CI, with product acceptance before close and after relevant later changes.
          round: 2
      blocked: false
content_hash: e1edda55e7c1e4024b45b5835db1894c434f3e009f4f7465300a3d3b5f7c148b
---

# Gate ledger — ariadne#243 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-23T10:20:21-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `test-strategy-contract` Name the production functions under test and compress the prose test inventory
  The adversarial verification section (plan:129-138) enumerates scenarios but does not identify the functions that will be unit-tested or provide one strategy line per risky function. Replace the case inventory with named production functions, each paired with its malformed/adversarial input class and mechanical guard.
- **PQ-2** [Important] `couch-provisioning-contract` Define the exact Couch-facing provisioning contract
  The plan names `weave compile` and `weave dependencies` but leaves the callable contract underspecified: invocation inputs, required cwd/context, outputs, exit/error classes, contention behavior, and retry semantics are not concrete enough for Couch to implement against. Specify the exact command/API and recovery mapping.
- **PQ-3** [Important] `schema-version-compatibility` Specify schema-v2 handling for older and malformed workspace state
  Changing the machine contract from schema v1 to v2 without defining read compatibility or rejection behavior leaves persisted/older state ambiguous. State whether v1, truncated, malformed, and unknown-version documents are accepted, migrated, or rejected, and name the production parser and tests covering those paths.
- **PQ-4** [Minor] `external-conformance-cadence` State the cadence for live fake-versus-real conformance
  ARCH-MOCK is addressed with stateful fakes and real acceptance tests, but the plan does not say when live conformance runs to detect Git/tool behavior drift.

## Round 2 — 2026-09-23T10:22:21-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — The adversarial verification table now names production functions and gives one strategy/mechanical guard per risky function.
- PQ-2 — addressed — The plan specifies commands, cwd, inputs, outputs, exit behavior, contention, failure mapping, and retry behavior for Couch.
- PQ-3 — addressed — The plan establishes that workspace JSON is live output rather than a persisted input protocol, with v1 treated as stale and unsupported/malformed consumer input rejected.
- PQ-4 — addressed — Git fake-versus-real conformance is scheduled in the normal package suite on seam changes and CI, with product acceptance before close and after relevant later changes.

## Open findings

(none — every finding has been disposed)
