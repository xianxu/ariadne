---
gate: plan-quality
issue: 481
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: df1b498b-3374-465a-adea-0221858cf8c2
            repository: github.com/xianxu/ariadne
            issue_id: "000481"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: f84663b0189f2dbf127b7b9cfdc89727eaac9a7a
            slot: ariadne:3
            requested_at: "2026-10-20T08:28:34.89736Z"
          phase: finished
          pid: 30225
          started_at: "2026-10-20T08:28:35.113983Z"
          finished_at: "2026-10-20T08:30:00.253602Z"
          session_id: 01a1586d-3992-7242-ab2c-4427a72c88aa
          outcome: returned
      "n": 1
      timestamp: "2026-10-20T01:30:00-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Define the cancellation and descendant-lifetime boundary for direct make runs
          detail: 'ARCH-ORDER: plan lines 70–72 reuse a watcher that currently observes only make''s PID, while managed process-group cancellation exists only for Go-launched suites. Specify which component consumes direct-make cancellation intents, what safely authorizes termination, and what evidence keeps the operation active or unknown if make exits before its descendants; name the production seam and deterministic owner-death/child-survival verification strategy.'
          family: cancellation-lifetime-ownership
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace prose case inventories with production functions and adversarial strategies
          detail: 'ARCH-PURE: the M1/M2 checklists at plan lines 139–155 enumerate test cases and Test* names without consistently naming the production functions being unit-tested. Compress these inventories into one strategy line per risky production function, identifying its adversarial input or event class and mechanical guard; retain integration boundaries and verification commands.'
          family: function-level-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#481 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T01:30:00-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `cancellation-lifetime-ownership` Define the cancellation and descendant-lifetime boundary for direct make runs
  ARCH-ORDER: plan lines 70–72 reuse a watcher that currently observes only make's PID, while managed process-group cancellation exists only for Go-launched suites. Specify which component consumes direct-make cancellation intents, what safely authorizes termination, and what evidence keeps the operation active or unknown if make exits before its descendants; name the production seam and deterministic owner-death/child-survival verification strategy.
- **PQ-2** [Important] `function-level-test-strategy` Replace prose case inventories with production functions and adversarial strategies
  ARCH-PURE: the M1/M2 checklists at plan lines 139–155 enumerate test cases and Test* names without consistently naming the production functions being unit-tested. Compress these inventories into one strategy line per risky production function, identifying its adversarial input or event class and mechanical guard; retain integration boundaries and verification commands.

## Open findings

- **PQ-1** [Important] `cancellation-lifetime-ownership` Define the cancellation and descendant-lifetime boundary for direct make runs
- **PQ-2** [Important] `function-level-test-strategy` Replace prose case inventories with production functions and adversarial strategies
