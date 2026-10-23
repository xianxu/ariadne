---
gate: plan-quality
issue: 577
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 99245dfb-6bb3-4029-ae4e-c7672bdfe000
            repository: github.com/xianxu/ariadne
            issue_id: "000577"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: e4735e34887623c43d297e37dace427aa1f177d4
            slot: ariadne:4
            requested_at: "2026-10-23T19:47:12.675792Z"
          phase: finished
          pid: 81105
          started_at: "2026-10-23T19:47:12.862253Z"
          finished_at: "2026-10-23T19:48:54.172731Z"
          session_id: 01a16a4d-9bee-7212-bb04-23295dc0dd86
          outcome: returned
      "n": 1
      timestamp: "2026-10-23T12:48:54-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Cover local-check identity and PR discovery before replacing its status sink
          detail: 'ARCH-PURPOSE / ARCH-SECURE: runLocalCheckCmd and runLocalCheckRepublish call detectRepo, which rejects the planned private origin, and perform GitHub PR discovery before reaching postLocalCheckRecord (cmd/sdlc/localcheck_cmd.go:221, :277, :316; cmd/sdlc/fetch.go:89). Extend the plan with validated bench identity/dependency construction for both commands, preserving production behavior and existing guards; require a real CLI private-origin test whose GitHub sentinel rejects every call, rather than testing only the injected recorder.'
          family: external-dependency-boundary-completeness
          round: 1
      blocked: true
---

# Gate ledger — ariadne#577 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T12:48:54-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `external-dependency-boundary-completeness` Cover local-check identity and PR discovery before replacing its status sink
  ARCH-PURPOSE / ARCH-SECURE: runLocalCheckCmd and runLocalCheckRepublish call detectRepo, which rejects the planned private origin, and perform GitHub PR discovery before reaching postLocalCheckRecord (cmd/sdlc/localcheck_cmd.go:221, :277, :316; cmd/sdlc/fetch.go:89). Extend the plan with validated bench identity/dependency construction for both commands, preserving production behavior and existing guards; require a real CLI private-origin test whose GitHub sentinel rejects every call, rather than testing only the injected recorder.

## Open findings

- **PQ-1** [Important] `external-dependency-boundary-completeness` Cover local-check identity and PR discovery before replacing its status sink
