---
gate: plan-quality
issue: 244
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-23T11:55:42-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Tests are listed as scenarios rather than named functions with one adversarial strategy each
          detail: The plan enumerates many cases in prose at plan lines 116-129 and 139-144, but does not name the pure/production functions under test or give one compact adversarial-input strategy per risky function. Compress the scenario inventory into named functions plus strategy lines, covering malformed receipts, CAS races, uncertain pushes, and stale review finalization.
          family: executable-test-strategy
          round: 1
        - id: PQ-2
          severity: Important
          title: The on-disk ownership and receipt contracts are not concrete enough to implement safely
          detail: The plan names ClaimIdentity and a worktree-private receipt store, but does not specify the exact persisted issue field/key, serialized shape, receipt path-resolution rule for linked worktrees versus independent clones, or the migration/compatibility behavior for legacy records. This is a hard-to-reverse ownership and storage boundary; define it before implementation. The existing issue schema currently permits open extra fields at construct/vocabulary/issue.cue:120-126.
          family: durable-state-contract
          round: 1
        - id: PQ-3
          severity: Important
          title: Unlocked review execution lacks an explicit transition model for interruption and stale outcomes
          detail: PreparedReview is named, but the plan does not enumerate its state/event transitions for reviewer cancellation, process death, relock failure, concurrent ledger writes, stale verdicts, or child-process cleanup. ARCH-ORDER requires the legal transitions, effects, bounded reviewer lifetime, and reproducible ordering seams to be explicit before changing change-code/close finalization.
          family: review-state-machine
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-23T11:58:18-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: not-addressed
          note: The named-function strategies were added, but the lossy scenario inventories remain at plan lines 116-129 and 139-144; this is the 2nd finding in family executable-test-strategy.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The plan now specifies the persisted issue field, versioned receipt shape, private Git directory resolution, compatibility behavior, validation, and recovery rules.
          round: 2
        - id: PQ-3
          disposition: not-addressed
          note: The state/event model now covers interruption and stale finalization, but it still does not specify a bounded reviewer lifetime or timeout policy; this is the 2nd finding in family review-state-machine.
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-09-23T12:00:10-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The plan now names each pure or production function under test and gives one compact adversarial strategy and oracle per risky function.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: The plan now enumerates PreparedReview states, interruption events, stale outcomes, relock failure, concurrent ledger writes, cancellation, timeout, process-group cleanup, and bounded reaping.
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-09-23T12:04:47-07:00"
      agent: codex
      blocked: false
      protocol_error: no valid findings block
    - "n": 5
      timestamp: "2026-09-23T12:36:36-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: not-addressed
          note: The plan names test scenarios, but still does not identify the production functions under test or provide one adversarial strategy and mechanical guard for each risky function.
          round: 5
      findings:
        - id: PQ-4
          severity: Important
          title: Define the runtime operating envelope for remote publication and unlocked reviews
          detail: ARCH-CONSTRAINTS requires workload classification, relevant budgets, their basis, and bounded behavior when exceeded. The plan gives retry and timeout constants (lines 40, 100–102) but does not state the expected concurrency/workload, network and repository-size assumptions, lock-wait budget, or behavior when those bounds are exceeded.
          family: operating-envelope
          round: 5
      blocked: false
content_hash: 6f25a9a1f31c7ac17ec2f29741afd3ee4bfb154bad18a8339216b22471cb5e7c
---

# Gate ledger — ariadne#244 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-23T11:55:42-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `executable-test-strategy` Tests are listed as scenarios rather than named functions with one adversarial strategy each
  The plan enumerates many cases in prose at plan lines 116-129 and 139-144, but does not name the pure/production functions under test or give one compact adversarial-input strategy per risky function. Compress the scenario inventory into named functions plus strategy lines, covering malformed receipts, CAS races, uncertain pushes, and stale review finalization.
- **PQ-2** [Important] `durable-state-contract` The on-disk ownership and receipt contracts are not concrete enough to implement safely
  The plan names ClaimIdentity and a worktree-private receipt store, but does not specify the exact persisted issue field/key, serialized shape, receipt path-resolution rule for linked worktrees versus independent clones, or the migration/compatibility behavior for legacy records. This is a hard-to-reverse ownership and storage boundary; define it before implementation. The existing issue schema currently permits open extra fields at construct/vocabulary/issue.cue:120-126.
- **PQ-3** [Important] `review-state-machine` Unlocked review execution lacks an explicit transition model for interruption and stale outcomes
  PreparedReview is named, but the plan does not enumerate its state/event transitions for reviewer cancellation, process death, relock failure, concurrent ledger writes, stale verdicts, or child-process cleanup. ARCH-ORDER requires the legal transitions, effects, bounded reviewer lifetime, and reproducible ordering seams to be explicit before changing change-code/close finalization.

## Round 2 — 2026-09-23T11:58:18-07:00 (codex) — BLOCKED

### Disposed

- PQ-1 — not-addressed — The named-function strategies were added, but the lossy scenario inventories remain at plan lines 116-129 and 139-144; this is the 2nd finding in family executable-test-strategy.
- PQ-2 — addressed — The plan now specifies the persisted issue field, versioned receipt shape, private Git directory resolution, compatibility behavior, validation, and recovery rules.
- PQ-3 — not-addressed — The state/event model now covers interruption and stale finalization, but it still does not specify a bounded reviewer lifetime or timeout policy; this is the 2nd finding in family review-state-machine.

## Round 3 — 2026-09-23T12:00:10-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — The plan now names each pure or production function under test and gives one compact adversarial strategy and oracle per risky function.
- PQ-3 — addressed — The plan now enumerates PreparedReview states, interruption events, stale outcomes, relock failure, concurrent ledger writes, cancellation, timeout, process-group cleanup, and bounded reaping.

## Round 4 — 2026-09-23T12:04:47-07:00 (codex) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 5 — 2026-09-23T12:36:36-07:00 (codex) — passed

### Disposed

- PQ-1 — not-addressed — The plan names test scenarios, but still does not identify the production functions under test or provide one adversarial strategy and mechanical guard for each risky function.

### Raised

- **PQ-4** [Important] `operating-envelope` Define the runtime operating envelope for remote publication and unlocked reviews
  ARCH-CONSTRAINTS requires workload classification, relevant budgets, their basis, and bounded behavior when exceeded. The plan gives retry and timeout constants (lines 40, 100–102) but does not state the expected concurrency/workload, network and repository-size assumptions, lock-wait budget, or behavior when those bounds are exceeded.

## Open findings

- **PQ-1** [Important] `executable-test-strategy` Tests are listed as scenarios rather than named functions with one adversarial strategy each
- **PQ-4** [Important] `operating-envelope` Define the runtime operating envelope for remote publication and unlocked reviews
