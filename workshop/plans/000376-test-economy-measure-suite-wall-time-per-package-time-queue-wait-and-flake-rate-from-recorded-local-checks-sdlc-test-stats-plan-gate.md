---
gate: plan-quality
issue: 376
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T20:55:08-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Define publication and cleanup ownership across concurrent suite attempts
          detail: The plan promises unique logs, atomic publication and crash-residue cleanup without specifying store placement or how cleanup distinguishes an active writer from an abandoned attempt. Define the attempt publication states and permitted cleanup, isolate retained records from existing 20-head pruning, and name the synchronization or ownership mechanism; verify concurrent publication and cleanup with controllable ordering. Standalone runs call runLocalCheck directly (cmd/sdlc/localcheck_cmd.go:117), so the plan cannot assume repo-wide serialization. ARCH-ORDER, ARCH-FUNERAL.
          family: durable-publication-ownership
          round: 1
        - id: PQ-2
          severity: Important
          title: Specify the trust boundary and validation policy for attempt records
          detail: The new schema includes timestamps, durations, outcome, identity and log paths, but the plan does not say how truncated, hand-edited or unsupported-version records become validated SuiteAttempt values. Name that parsing boundary and its rejection/coverage policy, including path confinement for any record-derived filesystem access or deletion; malformed records must never become successful-run or flake evidence. ARCH-SECURE.
          family: persisted-input-validation
          round: 1
        - id: PQ-3
          severity: Important
          title: Replace test-case inventories with named functions and adversarial strategies
          detail: Tasks 1–3 enumerate cases but do not name the production parser, aggregator or store-policy functions being tested. Compress those inventories into function names plus one strategy line per risky function, identifying the malformed-input or event-sequence class and its mechanical guard; retain controlled-order integration verification for publication and pruning.
          family: function-level-test-strategy
          round: 1
        - id: PQ-4
          severity: Minor
          title: Explicitly reuse the existing failure-extraction behavior
          detail: The proposed shared parser covers recorder and legacy reader, but omits failingTests in cmd/sdlc/suitefailures.go:11, already used by baseline diagnosis. Reuse it or extract one failure parser consumed by all three paths so package/test identity rules cannot drift. ARCH-DRY.
          family: shared-behavior-single-source
          round: 1
        - id: PQ-5
          severity: Minor
          title: Give the statistics reader and recorder an operating envelope
          detail: Thirty-day retention bounds age rather than record count, failed-log bytes or reader memory. State an expected fleet/run/log-size range with its basis and visible behavior when exceeded, including malformed oversized input; identify an acceptable recorder overhead rather than implying retention alone bounds resource use. ARCH-CONSTRAINTS.
          family: explicit-operating-envelope
          round: 1
      blocked: true
---

# Gate ledger — ariadne#376 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T20:55:08-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `durable-publication-ownership` Define publication and cleanup ownership across concurrent suite attempts
  The plan promises unique logs, atomic publication and crash-residue cleanup without specifying store placement or how cleanup distinguishes an active writer from an abandoned attempt. Define the attempt publication states and permitted cleanup, isolate retained records from existing 20-head pruning, and name the synchronization or ownership mechanism; verify concurrent publication and cleanup with controllable ordering. Standalone runs call runLocalCheck directly (cmd/sdlc/localcheck_cmd.go:117), so the plan cannot assume repo-wide serialization. ARCH-ORDER, ARCH-FUNERAL.
- **PQ-2** [Important] `persisted-input-validation` Specify the trust boundary and validation policy for attempt records
  The new schema includes timestamps, durations, outcome, identity and log paths, but the plan does not say how truncated, hand-edited or unsupported-version records become validated SuiteAttempt values. Name that parsing boundary and its rejection/coverage policy, including path confinement for any record-derived filesystem access or deletion; malformed records must never become successful-run or flake evidence. ARCH-SECURE.
- **PQ-3** [Important] `function-level-test-strategy` Replace test-case inventories with named functions and adversarial strategies
  Tasks 1–3 enumerate cases but do not name the production parser, aggregator or store-policy functions being tested. Compress those inventories into function names plus one strategy line per risky function, identifying the malformed-input or event-sequence class and its mechanical guard; retain controlled-order integration verification for publication and pruning.
- **PQ-4** [Minor] `shared-behavior-single-source` Explicitly reuse the existing failure-extraction behavior
  The proposed shared parser covers recorder and legacy reader, but omits failingTests in cmd/sdlc/suitefailures.go:11, already used by baseline diagnosis. Reuse it or extract one failure parser consumed by all three paths so package/test identity rules cannot drift. ARCH-DRY.
- **PQ-5** [Minor] `explicit-operating-envelope` Give the statistics reader and recorder an operating envelope
  Thirty-day retention bounds age rather than record count, failed-log bytes or reader memory. State an expected fleet/run/log-size range with its basis and visible behavior when exceeded, including malformed oversized input; identify an acceptable recorder overhead rather than implying retention alone bounds resource use. ARCH-CONSTRAINTS.

## Open findings

- **PQ-1** [Important] `durable-publication-ownership` Define publication and cleanup ownership across concurrent suite attempts
- **PQ-2** [Important] `persisted-input-validation` Specify the trust boundary and validation policy for attempt records
- **PQ-3** [Important] `function-level-test-strategy` Replace test-case inventories with named functions and adversarial strategies
- **PQ-4** [Minor] `shared-behavior-single-source` Explicitly reuse the existing failure-extraction behavior
- **PQ-5** [Minor] `explicit-operating-envelope` Give the statistics reader and recorder an operating envelope
