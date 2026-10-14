---
gate: plan-quality
issue: 370
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-14T11:40:36-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Legacy merge can execute a suite while retaining the whole-command lock
          detail: 'Task 2 incorrectly says legacy merge has no CI/suite wait. cmd/sdlc/merge.go:570 supplies a local runner to runCIGate, which executes required local checks at cmd/sdlc/cigate.go:602–605. ARCH-PURPOSE and ARCH-CONSTRAINTS: release the transaction lock around that execution too, and specify revalidation and a regression proof for this path.'
          family: lock-scope-covers-all-waiting-paths
          round: 1
        - id: PQ-2
          severity: Important
          title: Define main-drift recovery between replay and integration push
          detail: 'The proposed existing-event loop cannot accept a main observation in integrationPush: cmd/sdlc/integrationstate.go:46–48 rejects it after replay acceptance at lines 79–84. ARCH-ORDER: specify the transition, unpublished-commit disposition, retry accounting and crash recovery; another unpublished merge would violate integrationResume''s remote-first-parent requirement at cmd/sdlc/landingintegrate.go:58–60.'
          family: drift-recovery-transition-contract
          round: 1
        - id: PQ-3
          severity: Important
          title: Replace the test-case inventory with named risky-function strategies
          detail: Tasks 1 and 3 enumerate test scenarios and fixture procedures without naming the functions to unit-test. Name transitionIntegration and the affected orchestration/revalidation functions, distinguish pure unit proofs from integration proofs, and give each risky function one strategy line stating its adversarial input or event class and mechanical guard.
          family: function-level-test-strategy
          round: 1
        - id: PQ-4
          severity: Minor
          title: State the design's non-goals explicitly
          detail: Add a short statement identifying intentionally excluded work and why, such as changing suite scheduling, introducing cross-machine serialization, or adding a durable journal. This makes the intended lock-scoping boundary reviewable without inferring exclusions from the implementation checklist.
          family: explicit-scope-exclusions
          round: 1
      blocked: true
---

# Gate ledger — ariadne#370 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T11:40:36-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `lock-scope-covers-all-waiting-paths` Legacy merge can execute a suite while retaining the whole-command lock
  Task 2 incorrectly says legacy merge has no CI/suite wait. cmd/sdlc/merge.go:570 supplies a local runner to runCIGate, which executes required local checks at cmd/sdlc/cigate.go:602–605. ARCH-PURPOSE and ARCH-CONSTRAINTS: release the transaction lock around that execution too, and specify revalidation and a regression proof for this path.
- **PQ-2** [Important] `drift-recovery-transition-contract` Define main-drift recovery between replay and integration push
  The proposed existing-event loop cannot accept a main observation in integrationPush: cmd/sdlc/integrationstate.go:46–48 rejects it after replay acceptance at lines 79–84. ARCH-ORDER: specify the transition, unpublished-commit disposition, retry accounting and crash recovery; another unpublished merge would violate integrationResume's remote-first-parent requirement at cmd/sdlc/landingintegrate.go:58–60.
- **PQ-3** [Important] `function-level-test-strategy` Replace the test-case inventory with named risky-function strategies
  Tasks 1 and 3 enumerate test scenarios and fixture procedures without naming the functions to unit-test. Name transitionIntegration and the affected orchestration/revalidation functions, distinguish pure unit proofs from integration proofs, and give each risky function one strategy line stating its adversarial input or event class and mechanical guard.
- **PQ-4** [Minor] `explicit-scope-exclusions` State the design's non-goals explicitly
  Add a short statement identifying intentionally excluded work and why, such as changing suite scheduling, introducing cross-machine serialization, or adding a durable journal. This makes the intended lock-scoping boundary reviewable without inferring exclusions from the implementation checklist.

## Open findings

- **PQ-1** [Important] `lock-scope-covers-all-waiting-paths` Legacy merge can execute a suite while retaining the whole-command lock
- **PQ-2** [Important] `drift-recovery-transition-contract` Define main-drift recovery between replay and integration push
- **PQ-3** [Important] `function-level-test-strategy` Replace the test-case inventory with named risky-function strategies
- **PQ-4** [Minor] `explicit-scope-exclusions` State the design's non-goals explicitly
