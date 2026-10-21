---
gate: boundary-review
issue: 509
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 9d6b9c64-bfb1-40b1-b19d-03a24a67f835
            repository: github.com/xianxu/ariadne
            issue_id: "000509"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/6262fcc90694515f
            reviewed_head: 456fa34aec88b993a6dfbc2636a12884883929b7
            slot: ariadne:3
            requested_at: "2026-10-21T06:05:58.902945Z"
          phase: finished
          pid: 87758
          started_at: "2026-10-21T06:05:59.056938Z"
          finished_at: "2026-10-21T06:09:29.618874Z"
          session_id: 01a15d11-0711-7573-a4a8-e5a9cfc573ab
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 456fa34aec88b993a6dfbc2636a12884883929b7
        tracker_head: b3a3ebc48e9a3b1017329b9d15121047ba2670b0
        branch_point: 6573f0810474e4de9a82fc0a4b7a4c7670c78996
        since: "2026-10-20T22:27:36-07:00"
        until: "2026-10-20T23:05:28-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T22:27:36-07:00"
              end: "2026-10-20T23:05:28-07:00"
              began: "2026-10-20T22:27:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 57529174f7b78b0d8da6464b18f0beb24038f848f09e7e6bded3ac00c9e4d091
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 98550346
              windows:
                - start: "2026-10-20T22:27:36-07:00"
                  end: "2026-10-20T23:05:28-07:00"
            - harness: codex
              id: 01a15d03-86e1-7a71-a2ce-ef4bafdda938
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: d25ada22ff39f9340ef29f894ce65f4ab1fd8f7c0e1eed657d80eb8f1b0a07c8
              started: "2026-10-21T05:51:15.011Z"
              spawned: false
              bytes: 633746
              windows:
                - start: "2026-10-21T05:51:15.011Z"
                  end: "2026-10-20T23:05:28-07:00"
            - harness: codex
              id: 01a15d05-4be6-7122-9be6-0cc770381dfe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 706f06321cf7f4cc2cc648d7f65df581da4cb247366968984dee192fc420a6e2
              started: "2026-10-21T05:53:10.993Z"
              spawned: false
              bytes: 543474
              windows:
                - start: "2026-10-21T05:53:10.993Z"
                  end: "2026-10-20T23:05:28-07:00"
            - harness: codex
              id: 01a15d06-5a0b-7be3-a48a-c58866001c81
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fc671f1f1dbf5643b1b8b18d712eb3b2b0770526c88216a8b8abd0e849ae5d8c
              started: "2026-10-21T05:54:20.122Z"
              spawned: false
              bytes: 451768
              windows:
                - start: "2026-10-21T05:54:20.122Z"
                  end: "2026-10-20T23:05:28-07:00"
            - harness: codex
              id: 01a15d06-d164-7dd0-9b43-b0e9ee056cd1
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 98c45492655fb9461c4189c7583be5d48e47334ad2f33b2480a4f47a2d9be9f5
              started: "2026-10-21T05:54:50.689Z"
              spawned: false
              bytes: 460459
              windows:
                - start: "2026-10-21T05:54:50.689Z"
                  end: "2026-10-20T23:05:28-07:00"
      "n": 1
      timestamp: "2026-10-20T23:09:30-07:00"
      dispatched: "2026-10-20T23:05:58-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: A tracked empty declaration becomes a successful empty selection
          detail: cmd/sdlc/landingchecks_git.go:64-68 passes both an absent declaration and an existing empty/whitespace blob to parseLocalChecks, whose empty-input branch returns an empty declaration. readLandingCheckInputs then returns successfully at lines 102-103 without source-car evidence. This contradicts the plan's rejection of truncated policy input and fabricates an empty selection from invalid configuration (ARCH-SECURE). Distinguish absence from a present blank blob in the new reader, preserve legacy parser behavior, and add regression cases proving blank tracked declarations refuse while genuine absence remains compatible.
          family: policy-input-fails-closed
          round: 1
        - id: BR-2
          severity: Critical
          title: Forced runner cleanup accepts an unconfirmed group termination
          detail: cmd/sdlc/localcheck_resource_unix.go:56-59 returns success after processgroup.Terminate(cmd, true) returns nil, without rechecking group termination. That helper also returns nil when its 100ms repeated-kill budget expires (cmd/sdlc/internal/processgroup/process_unix.go:46-59), so nil does not establish that descendants stopped. The caller then retires the run lease and checkout despite the plan's force-and-join guarantee (ARCH-ORDER, ARCH-FUNERAL). Confirm termination after escalation within a bounded join; if it remains unconfirmed, return a visible cleanup error and preserve the necessary recovery ownership. Add a controllable regression where termination returns nil while the group remains live.
          family: cleanup-requires-confirmed-completion
          round: 1
        - id: BR-3
          severity: Important
          title: README coverage is missing for the M1 configuration and selection surface
          detail: This range introduces version 2 with scope, resource, and paths_file, and changes local-check run --all to exclude landing-scoped checks (cmd/sdlc/localcheck_cmd.go:111-118), but contains no README.md update. README.md:591-593 still describes named checks as running under the machine-wide suite lock, which is false for resource:runner. Update README now with the delivered M1 behavior and explicitly identify automatic enforcement as pending M2; the docs gate applies at each milestone.
          family: user-surface-documentation
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#509 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T23:09:30-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `policy-input-fails-closed` A tracked empty declaration becomes a successful empty selection
  cmd/sdlc/landingchecks_git.go:64-68 passes both an absent declaration and an existing empty/whitespace blob to parseLocalChecks, whose empty-input branch returns an empty declaration. readLandingCheckInputs then returns successfully at lines 102-103 without source-car evidence. This contradicts the plan's rejection of truncated policy input and fabricates an empty selection from invalid configuration (ARCH-SECURE). Distinguish absence from a present blank blob in the new reader, preserve legacy parser behavior, and add regression cases proving blank tracked declarations refuse while genuine absence remains compatible.
- **BR-2** [Critical] `cleanup-requires-confirmed-completion` Forced runner cleanup accepts an unconfirmed group termination
  cmd/sdlc/localcheck_resource_unix.go:56-59 returns success after processgroup.Terminate(cmd, true) returns nil, without rechecking group termination. That helper also returns nil when its 100ms repeated-kill budget expires (cmd/sdlc/internal/processgroup/process_unix.go:46-59), so nil does not establish that descendants stopped. The caller then retires the run lease and checkout despite the plan's force-and-join guarantee (ARCH-ORDER, ARCH-FUNERAL). Confirm termination after escalation within a bounded join; if it remains unconfirmed, return a visible cleanup error and preserve the necessary recovery ownership. Add a controllable regression where termination returns nil while the group remains live.
- **BR-3** [Important] `user-surface-documentation` README coverage is missing for the M1 configuration and selection surface
  This range introduces version 2 with scope, resource, and paths_file, and changes local-check run --all to exclude landing-scoped checks (cmd/sdlc/localcheck_cmd.go:111-118), but contains no README.md update. README.md:591-593 still describes named checks as running under the machine-wide suite lock, which is false for resource:runner. Update README now with the delivered M1 behavior and explicitly identify automatic enforcement as pending M2; the docs gate applies at each milestone.

## Open findings

- **BR-1** [Critical] `policy-input-fails-closed` A tracked empty declaration becomes a successful empty selection
- **BR-2** [Critical] `cleanup-requires-confirmed-completion` Forced runner cleanup accepts an unconfirmed group termination
- **BR-3** [Important] `user-surface-documentation` README coverage is missing for the M1 configuration and selection surface
