---
reviewer_runs:
    - version: 1
      intent:
        id: 1d39494c-e30c-472b-a81e-3c7baef96a18
        repository: github.com/xianxu/ariadne
        issue_id: "000452"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        reviewed_head: e57faee8f28e83431876a90ea527f070fe28e74b
        slot: ariadne:3
        requested_at: "2026-10-19T01:16:28.061696Z"
      phase: finished
      pid: 34422
      started_at: "2026-10-19T01:16:28.390931Z"
      finished_at: "2026-10-19T01:16:53.255059Z"
      session_id: 01a151bb-4195-78e2-97a7-b858be9075ec
      outcome: returned
gate: plan-quality
issue: 452
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: a54fc91c-01e1-48d7-af59-f54fab923508
            repository: github.com/xianxu/ariadne
            issue_id: "000452"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: 5e77523c1a8934e44478d1d23199aa84d193c4f0
            slot: ariadne:3
            requested_at: "2026-10-19T01:13:17.698308Z"
          phase: finished
          pid: 996
          started_at: "2026-10-19T01:13:17.812477Z"
          finished_at: "2026-10-19T01:14:11.130726Z"
          session_id: 01a151b8-57ff-7551-be32-48930d0927c7
          outcome: returned
      "n": 1
      timestamp: "2026-10-18T18:14:11-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Cached grant evidence must account for activation as well as expiry
          detail: The plan caches active IDs and expiry deadlines until the source changes, but ValidGrants also rejects grants whose Granted time is still in the future (cmd/sdlc/internal/landingline/order.go:64). Such a grant can become valid without a log change and remain invisible indefinitely. Define a Go-derived temporal refresh boundary or equivalent authoritative revalidation, and use a controllable-clock strategy to verify unchanged-source validity transitions (ARCH-DRY, ARCH-ORDER).
          family: temporal-cache-validity
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace prose case inventories with named functions and adversarial strategies
          detail: Task 1 and Task 2 enumerate test cases while leaving the new production unit-test functions unnamed. Name the subject-binding, protocol-decoding, cache-validity and ordering functions, with one adversarial-input class and mechanical guard per risky function; compress the case inventories into those strategies and retain the real-helper integration boundary (ARCH-PURE).
          family: function-level-test-strategy
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 79d77ff7-0eaf-4682-9281-a9d23a8d9671
            repository: github.com/xianxu/ariadne
            issue_id: "000452"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: e57faee8f28e83431876a90ea527f070fe28e74b
            slot: ariadne:3
            requested_at: "2026-10-19T01:14:49.92353Z"
          phase: finished
          pid: 67531
          started_at: "2026-10-19T01:14:50.188237Z"
          finished_at: "2026-10-19T01:15:13.219413Z"
          session_id: 01a151b9-c33a-7220-be50-1faab4d207c0
          outcome: returned
      "n": 2
      timestamp: "2026-10-18T18:15:14-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Go-derived refresh_at covers activation and expiry; injected-clock strategies verify unchanged-source validity transitions without duplicating grant policy.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Named production functions now have adversarial strategies and mechanical guards, while real-helper integration remains the admission and lifetime boundary.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: c4c1c700-5e7d-48fa-b592-c14d736ed123
            repository: github.com/xianxu/ariadne
            issue_id: "000452"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: e57faee8f28e83431876a90ea527f070fe28e74b
            slot: ariadne:3
            requested_at: "2026-10-19T01:16:00.324987Z"
          phase: finished
          pid: 17539
          started_at: "2026-10-19T01:16:00.659248Z"
          finished_at: "2026-10-19T01:16:24.863742Z"
          session_id: 01a151ba-d5e9-78c1-aab9-a87098ddc125
          outcome: returned
      "n": 3
      timestamp: "2026-10-18T18:16:25-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      blocked: false
content_hash: b9bac28d1f3a6758e4e438c280024dc8121ccfdaffe83a88b265d8086ddeba5d
---

# Gate ledger — ariadne#452 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T18:14:11-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `temporal-cache-validity` Cached grant evidence must account for activation as well as expiry
  The plan caches active IDs and expiry deadlines until the source changes, but ValidGrants also rejects grants whose Granted time is still in the future (cmd/sdlc/internal/landingline/order.go:64). Such a grant can become valid without a log change and remain invisible indefinitely. Define a Go-derived temporal refresh boundary or equivalent authoritative revalidation, and use a controllable-clock strategy to verify unchanged-source validity transitions (ARCH-DRY, ARCH-ORDER).
- **PQ-2** [Important] `function-level-test-strategy` Replace prose case inventories with named functions and adversarial strategies
  Task 1 and Task 2 enumerate test cases while leaving the new production unit-test functions unnamed. Name the subject-binding, protocol-decoding, cache-validity and ordering functions, with one adversarial-input class and mechanical guard per risky function; compress the case inventories into those strategies and retain the real-helper integration boundary (ARCH-PURE).

## Round 2 — 2026-10-18T18:15:14-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Go-derived refresh_at covers activation and expiry; injected-clock strategies verify unchanged-source validity transitions without duplicating grant policy.
- PQ-2 — addressed — Named production functions now have adversarial strategies and mechanical guards, while real-helper integration remains the admission and lifetime boundary.

## Round 3 — 2026-10-18T18:16:25-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
