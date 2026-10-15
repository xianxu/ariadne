---
reviewer_runs:
    - version: 1
      intent:
        id: 5ba00362-88ca-4e02-8b9b-0a134cd49da3
        repository: github.com/xianxu/ariadne
        issue_id: "000385"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        reviewed_head: 6144acd333d630e7486568433df3ebdc2e55d2ca
        slot: ariadne:3
        requested_at: "2026-10-15T16:52:50.036788Z"
      phase: finished
      pid: 84018
      started_at: "2026-10-15T16:52:50.133236Z"
      finished_at: "2026-10-15T16:53:23.638409Z"
      session_id: 01a1407b-1424-7c31-964f-c6839b6c5951
      outcome: returned
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
    - reviewer_runs:
        - version: 1
          intent:
            id: 7919318e-2d9e-498f-a811-bdc2fedf444c
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: 6144acd333d630e7486568433df3ebdc2e55d2ca
            slot: ariadne:3
            requested_at: "2026-10-15T16:51:11.246719Z"
          phase: finished
          pid: 80762
          started_at: "2026-10-15T16:51:11.342656Z"
          finished_at: "2026-10-15T16:51:39.609665Z"
          session_id: 01a14079-9246-7ea3-91fb-c457e3886da9
          outcome: returned
      "n": 2
      timestamp: "2026-10-15T09:51:39-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The superseding policy requires validated launch binding or unambiguous complete evidence, rejects recency selection among competing sessions, and tests session/activity permutations.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The function-level strategy table names parsing, state transitions, decimal arithmetic, normalization, aggregation, projection, and identity selection with adversarial inputs and mechanical guards.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: c432253b-7986-4175-bc53-1d0f8bffa868
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: 6144acd333d630e7486568433df3ebdc2e55d2ca
            slot: ariadne:3
            requested_at: "2026-10-15T16:52:24.252014Z"
          phase: finished
          pid: 83000
          started_at: "2026-10-15T16:52:24.348972Z"
          finished_at: "2026-10-15T16:52:49.073313Z"
          session_id: 01a1407a-af6f-7152-833b-a43cafac9231
          outcome: returned
      "n": 3
      timestamp: "2026-10-15T09:52:49-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: The final revision supersedes recency selection with invocation evidence, validated session binding, and conservative ambiguity handling.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Named production functions have adversarial strategies and mechanical guards rather than prose test-case inventories.
          round: 3
      blocked: false
content_hash: 458f2e439b08160311fcc16fee1570d8fcd56cf774635b263c3a770401bc9927
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

## Round 2 — 2026-10-15T09:51:39-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — The superseding policy requires validated launch binding or unambiguous complete evidence, rejects recency selection among competing sessions, and tests session/activity permutations.
- PQ-2 — addressed — The function-level strategy table names parsing, state transitions, decimal arithmetic, normalization, aggregation, projection, and identity selection with adversarial inputs and mechanical guards.

## Round 3 — 2026-10-15T09:52:49-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — The final revision supersedes recency selection with invocation evidence, validated session binding, and conservative ambiguity handling.
- PQ-2 — addressed — Named production functions have adversarial strategies and mechanical guards rather than prose test-case inventories.

## Open findings

(none — every finding has been disposed)
