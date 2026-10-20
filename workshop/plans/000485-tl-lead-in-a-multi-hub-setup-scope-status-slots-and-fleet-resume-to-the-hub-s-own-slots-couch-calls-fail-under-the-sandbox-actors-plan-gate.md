---
gate: plan-quality
issue: 485
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 6901956d-57f8-45eb-ba2a-c3f2a9cbe9ce
            repository: github.com/xianxu/ariadne
            issue_id: "000485"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: ea23ff3b9caa3eeb487e632d5c9f62e36b70b2dd
            slot: ariadne:2
            requested_at: "2026-10-20T20:44:39.805193Z"
          phase: finished
          pid: 17973
          started_at: "2026-10-20T20:44:39.925678Z"
          finished_at: "2026-10-20T20:45:38.497182Z"
          session_id: 01a15b0f-1fb0-7c61-8192-ad8d7feab469
          outcome: returned
      "n": 1
      timestamp: "2026-10-20T13:45:38-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Define the evidence that distinguishes confirmed refusal from uncertainty
          detail: 'Task 3 distinguishes timeout and explicit uncertainty from confirmed refusal, but leaves signal termination and unrecognized or truncated nonzero output undefined. Under ARCH-SECURE/ORDER, name the outcome-classification function and its confirmation rule: after possible admission, absent recognizable terminal evidence must remain uncertain, retain diagnostics, permit only read-only observation and stop the batch.'
          family: outcome-requires-confirming-evidence
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace test-case inventories with named functions and adversarial strategies
          detail: Tasks 1–3 enumerate scope, freshness and resume cases in prose, contrary to this gate's test-plan contract. Compress these into the functions being tested and one strategy line per risky function, naming the adversarial input or event class and mechanical guard; retain controlled completion barriers, output retention, hidden-hold protection and selected-read counting as strategy-level guarantees.
          family: function-level-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#485 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T13:45:38-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `outcome-requires-confirming-evidence` Define the evidence that distinguishes confirmed refusal from uncertainty
  Task 3 distinguishes timeout and explicit uncertainty from confirmed refusal, but leaves signal termination and unrecognized or truncated nonzero output undefined. Under ARCH-SECURE/ORDER, name the outcome-classification function and its confirmation rule: after possible admission, absent recognizable terminal evidence must remain uncertain, retain diagnostics, permit only read-only observation and stop the batch.
- **PQ-2** [Important] `function-level-test-strategy` Replace test-case inventories with named functions and adversarial strategies
  Tasks 1–3 enumerate scope, freshness and resume cases in prose, contrary to this gate's test-plan contract. Compress these into the functions being tested and one strategy line per risky function, naming the adversarial input or event class and mechanical guard; retain controlled completion barriers, output retention, hidden-hold protection and selected-read counting as strategy-level guarantees.

## Open findings

- **PQ-1** [Important] `outcome-requires-confirming-evidence` Define the evidence that distinguishes confirmed refusal from uncertainty
- **PQ-2** [Important] `function-level-test-strategy` Replace test-case inventories with named functions and adversarial strategies
