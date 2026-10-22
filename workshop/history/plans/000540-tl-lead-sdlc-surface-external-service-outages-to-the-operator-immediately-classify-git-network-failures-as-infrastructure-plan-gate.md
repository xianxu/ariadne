---
reviewer_runs:
    - version: 1
      intent:
        id: b97d4e25-82e2-4d24-b875-0eacba07c0b1
        repository: github.com/xianxu/ariadne
        issue_id: "000540"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        reviewed_head: bf3c5a6145adbe3c13e1e29ef6882b7d9403d716
        slot: ariadne:1
        requested_at: "2026-10-22T03:22:10.873772Z"
      phase: finished
      pid: 21861
      started_at: "2026-10-22T03:22:10.977298Z"
      finished_at: "2026-10-22T03:22:46.598989Z"
      session_id: 01a161a1-6ba6-7490-b5ac-e6c898495122
      outcome: returned
gate: plan-quality
issue: 540
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 7ef506f7-4424-44e6-8eba-8f07237d1d2b
            repository: github.com/xianxu/ariadne
            issue_id: "000540"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 138354cb2585ce42ad2f129eccb37b119bce72ae
            slot: ariadne:1
            requested_at: "2026-10-22T03:18:33.655412Z"
          phase: finished
          pid: 80673
          started_at: "2026-10-22T03:18:33.75858Z"
          finished_at: "2026-10-22T03:19:39.650892Z"
          session_id: 01a1619e-1b2a-7433-bd4f-759515fc5331
          outcome: returned
      "n": 1
      timestamp: "2026-10-21T20:19:39-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Connect net check shutdown to actual CLI signals.
          detail: 'ARCH-ORDER/ARCH-FUNERAL: the plan promises that probe effects are cancelled and joined before return, but cmd/sdlc/main.go:83–99 only installs signal cancellation for an explicit command allowlist that excludes net check. Name how the new command gains signal ownership, and require a real CLI signal control proving that in-flight utility process groups are terminated and reaped before exit; injected context cancellation alone cannot verify this.'
          family: cli-signal-lifetime-ownership
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace prose case inventories with named risky functions and adversarial strategies.
          detail: The test rows at workshop/plans/000540-network-outage-plan.md:89 and :96 enumerate individual cases while naming files and conceptual types rather than the production functions to unit-test. Compress them into function names with one adversarial-input class and mechanical guard per risky function, including classifier/host parsing, utility-output parsing, report rendering and probe orchestration; use fuzz/property strategies for arbitrary diagnostic and parser input, and controllable ordering for cancellation and lifetime invariants.
          family: function-level-test-strategy
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: de7b4f6f-2406-4ca2-984d-13c6e0efb2c3
            repository: github.com/xianxu/ariadne
            issue_id: "000540"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: bf3c5a6145adbe3c13e1e29ef6882b7d9403d716
            slot: ariadne:1
            requested_at: "2026-10-22T03:20:19.669222Z"
          phase: finished
          pid: 4749
          started_at: "2026-10-22T03:20:19.779995Z"
          finished_at: "2026-10-22T03:20:50.857376Z"
          session_id: 01a1619f-b959-7ec2-8197-3d094d728cb6
          outcome: returned
      "n": 2
      timestamp: "2026-10-21T20:20:51-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Names the exact CLI signal allowlist extension and requires real SIGINT/SIGTERM controls proving utility-group termination and reaping before exit.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Names classifier, host/address parser, renderer and orchestration functions with adversarial strategies, mechanical invariants and controllable ordering.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: e75d8cf8-7d2b-4a5c-bb73-9413225dc4cf
            repository: github.com/xianxu/ariadne
            issue_id: "000540"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: bf3c5a6145adbe3c13e1e29ef6882b7d9403d716
            slot: ariadne:1
            requested_at: "2026-10-22T03:21:41.642316Z"
          phase: finished
          pid: 19355
          started_at: "2026-10-22T03:21:41.738125Z"
          finished_at: "2026-10-22T03:22:09.582022Z"
          session_id: 01a161a0-f963-7761-a38b-3d9423e96e16
          outcome: returned
      "n": 3
      timestamp: "2026-10-21T20:22:09-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      blocked: false
content_hash: 2cada79160d15784be307f6fd2f3bcdcf965bd9b1bc30a8741e632abb17275e4
---

# Gate ledger — ariadne#540 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T20:19:39-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `cli-signal-lifetime-ownership` Connect net check shutdown to actual CLI signals.
  ARCH-ORDER/ARCH-FUNERAL: the plan promises that probe effects are cancelled and joined before return, but cmd/sdlc/main.go:83–99 only installs signal cancellation for an explicit command allowlist that excludes net check. Name how the new command gains signal ownership, and require a real CLI signal control proving that in-flight utility process groups are terminated and reaped before exit; injected context cancellation alone cannot verify this.
- **PQ-2** [Important] `function-level-test-strategy` Replace prose case inventories with named risky functions and adversarial strategies.
  The test rows at workshop/plans/000540-network-outage-plan.md:89 and :96 enumerate individual cases while naming files and conceptual types rather than the production functions to unit-test. Compress them into function names with one adversarial-input class and mechanical guard per risky function, including classifier/host parsing, utility-output parsing, report rendering and probe orchestration; use fuzz/property strategies for arbitrary diagnostic and parser input, and controllable ordering for cancellation and lifetime invariants.

## Round 2 — 2026-10-21T20:20:51-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Names the exact CLI signal allowlist extension and requires real SIGINT/SIGTERM controls proving utility-group termination and reaping before exit.
- PQ-2 — addressed — Names classifier, host/address parser, renderer and orchestration functions with adversarial strategies, mechanical invariants and controllable ordering.

## Round 3 — 2026-10-21T20:22:09-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
