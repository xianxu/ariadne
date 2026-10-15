---
gate: plan-quality
issue: 385
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 82fe347d-6a17-4401-8aeb-8dd84d3a7145
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: 54d188bdb3a6991d89439cdf0e98679dc38472ed
            slot: ariadne:3
            requested_at: "2026-10-15T16:45:45.034795Z"
          phase: finished
          pid: 71059
          started_at: "2026-10-15T16:45:45.143508Z"
          finished_at: "2026-10-15T16:46:55.399744Z"
          session_id: 01a14074-981e-74f2-af8c-a6ddf477fc6d
          outcome: returned
      "n": 1
      timestamp: "2026-10-15T09:46:55-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Latest workspace activity does not establish the invoking session
          detail: The 09:47 revision can stamp session A's commit with session B's model when both share a cwd and B has newer activity; neither a timestamp tie nor conflicting metadata is required. Under ARCH-SECURE and ARCH-PURPOSE, require launch-bound session evidence or preserve unknown model/effort when competing sessions cannot be excluded; state and test that policy.
          family: invoking-identity-requires-provenance
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace test-case inventories with named functions and mechanical strategies
          detail: Tasks 1–5 enumerate test cases instead of naming each risky production function and its adversarial input class plus mechanical guard. Compress these lists into function-level strategies for record parsing, SessionState transitions, decimal parsing/addition, NormalizeResponses, AggregateUsage, ProjectMeter, and identity selection; use seeded fuzzing, generated sequences, or invariant/property checks as appropriate.
          family: function-level-adversarial-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#385 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T09:46:55-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `invoking-identity-requires-provenance` Latest workspace activity does not establish the invoking session
  The 09:47 revision can stamp session A's commit with session B's model when both share a cwd and B has newer activity; neither a timestamp tie nor conflicting metadata is required. Under ARCH-SECURE and ARCH-PURPOSE, require launch-bound session evidence or preserve unknown model/effort when competing sessions cannot be excluded; state and test that policy.
- **PQ-2** [Important] `function-level-adversarial-test-strategy` Replace test-case inventories with named functions and mechanical strategies
  Tasks 1–5 enumerate test cases instead of naming each risky production function and its adversarial input class plus mechanical guard. Compress these lists into function-level strategies for record parsing, SessionState transitions, decimal parsing/addition, NormalizeResponses, AggregateUsage, ProjectMeter, and identity selection; use seeded fuzzing, generated sequences, or invariant/property checks as appropriate.

## Open findings

- **PQ-1** [Important] `invoking-identity-requires-provenance` Latest workspace activity does not establish the invoking session
- **PQ-2** [Important] `function-level-adversarial-test-strategy` Replace test-case inventories with named functions and mechanical strategies
