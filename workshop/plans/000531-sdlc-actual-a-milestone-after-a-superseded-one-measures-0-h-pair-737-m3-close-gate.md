---
gate: boundary-review
issue: 531
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a10d8292-4aac-43c6-8e0a-79deb1d7e95b
            repository: github.com/xianxu/ariadne
            issue_id: "000531"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/1a7943ef0e4eab21
            reviewed_head: 930a9c77fdc610a8f1d6d955c6750978bf22b4b8
            slot: ariadne:2
            requested_at: "2026-10-22T17:01:26.829064Z"
          phase: finished
          pid: 48197
          started_at: "2026-10-22T17:01:27.152796Z"
          finished_at: "2026-10-22T17:03:54.413581Z"
          session_id: 01a1648f-7d29-78e3-9cff-a5d796877883
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 930a9c77fdc610a8f1d6d955c6750978bf22b4b8
        tracker_head: 7c9fe37d86caa74969638d6a4cecda2db0a020f0
        branch_point: 891ba3f81f67690364507cf4030a71c60ee681e4
        since: "2026-10-22T06:11:42-07:00"
        until: "2026-10-22T10:00:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T06:11:42-07:00"
              end: "2026-10-22T10:00:25-07:00"
              began: "2026-10-22T06:11:42-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: codex
              id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bf32ab80d135e8b9c297e015dd6fb6b96cf74b5886c8f822c5138db7c879e3f3
              started: "2026-10-20T18:21:54.178Z"
              spawned: false
              bytes: 60097769
              windows:
                - start: "2026-10-22T06:11:42-07:00"
                  end: "2026-10-22T10:00:25-07:00"
            - harness: codex
              id: 01a163bd-be11-7252-9cbe-e62a9fed13bf
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: 0e9cd3d8235a6291af5f8affb1568c0b1859f97ee18e96b05d99b996c2b9ca00
              started: "2026-10-22T13:12:21.961Z"
              spawned: true
              bytes: 782824
              windows:
                - start: "2026-10-22T13:12:21.961Z"
                  end: "2026-10-22T10:00:25-07:00"
            - harness: codex
              id: 01a163c2-3f6f-7202-bd12-605946dc6161
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 89b3e950583cca6f90151c10eabb7fb780153c45969804c56de0c1fbedb9a23e
              started: "2026-10-22T13:17:17.887Z"
              spawned: false
              bytes: 588448
              windows:
                - start: "2026-10-22T13:17:17.887Z"
                  end: "2026-10-22T10:00:25-07:00"
            - harness: codex
              id: 01a163c4-0997-7293-9567-4150d0aa7a2b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 07f180f7034355358cd1372da33eeb07f3f4cd2355a27fab46e8dda69dd5a353
              started: "2026-10-22T13:19:14.648Z"
              spawned: false
              bytes: 440898
              windows:
                - start: "2026-10-22T13:19:14.648Z"
                  end: "2026-10-22T10:00:25-07:00"
            - harness: codex
              id: 01a163c5-a32e-72d2-8cca-8420db1e2121
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 149de2e2a2a0e55e38b649c9c4650b658a8eecbbfcf4c1fbfa1ba3aa2f3dc581
              started: "2026-10-22T13:20:59.914Z"
              spawned: false
              bytes: 393878
              windows:
                - start: "2026-10-22T13:20:59.914Z"
                  end: "2026-10-22T10:00:25-07:00"
            - harness: codex
              id: 01a163c6-13db-7250-b8fc-d997d7c4904b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8f88b959cd4d8955b0458cb6924126b2cd933e31a4d2d3111f0650cb255adc72
              started: "2026-10-22T13:21:28.278Z"
              spawned: false
              bytes: 384312
              windows:
                - start: "2026-10-22T13:21:28.278Z"
                  end: "2026-10-22T10:00:25-07:00"
        queue_waits:
            - attempt_id: f5e4adb1f14c8d5dd1c3330389679e87
              slot: ariadne:2
              record_digest: 7539ab9702724d746ddded5b1bdfa983e5de63c517153ee94eac29a583a3f6d5
              start: "2026-10-22T13:40:24.139117Z"
              end: "2026-10-22T14:39:34.814301Z"
              queue_seconds: 3550.675184
              raw_seconds: 3550.675184
              exact: true
            - attempt_id: 3d171593dcdf291b51b28c47dd4e440b
              slot: ariadne:2
              record_digest: b19b1ddc9f6bdcb300ab4b29818eecf1398c961637de28079790fedab97e1414
              start: "2026-10-22T14:56:18.150905Z"
              end: "2026-10-22T16:40:43.506555Z"
              queue_seconds: 6265.35565
              raw_seconds: 6265.35565
              exact: true
      "n": 1
      timestamp: "2026-10-22T10:03:56-07:00"
      dispatched: "2026-10-22T10:01:26-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Requested milestone validation is skipped when no named boundary is open
          detail: 'cmd/sdlc/actualplan.go:86-98 returns before validating milestone membership whenever rounds are empty or all named boundaries are finalized. Consequently, an absent, unfinalized requested milestone receives the remaining build interval through BuildTimeline instead of refusing, contrary to the plan''s explicit contract at workshop/plans/000531-retired-boundary-accounting-plan.md:39,50 (ARCH-PURPOSE, ARCH-SECURE). The passing TestComputeActualSplitsAtBoundaries already demonstrates this: its fixture has no Plan section, yet requesting M2 returns 25/60 hours. Require pinned membership whenever an explicitly requested milestone lacks finalized historical authority; add refusal regressions for both empty and finalized-only ledgers while preserving whole-issue and retained-replay behavior.'
          family: executable-milestone-authority
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: README omits the new executable Plan prerequisite and retirement workflow
          detail: 'cmd/sdlc/helptext/actual.md:38-43 introduces a user-facing usage contract: committed executable Plan membership controls fresh cuts and missing authority can refuse measurement. README.md:1059-1070 remains unchanged and does not explain this prerequisite or how retiring a milestone row affects replacement accounting. Add concise usage guidance covering committed Plan authority, retirement without clearing review findings, and unchanged saved replay; atlas coverage is already present.'
          family: user-surface-documentation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#531 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T10:03:56-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `executable-milestone-authority` (kind: defect) Requested milestone validation is skipped when no named boundary is open
  cmd/sdlc/actualplan.go:86-98 returns before validating milestone membership whenever rounds are empty or all named boundaries are finalized. Consequently, an absent, unfinalized requested milestone receives the remaining build interval through BuildTimeline instead of refusing, contrary to the plan's explicit contract at workshop/plans/000531-retired-boundary-accounting-plan.md:39,50 (ARCH-PURPOSE, ARCH-SECURE). The passing TestComputeActualSplitsAtBoundaries already demonstrates this: its fixture has no Plan section, yet requesting M2 returns 25/60 hours. Require pinned membership whenever an explicitly requested milestone lacks finalized historical authority; add refusal regressions for both empty and finalized-only ledgers while preserving whole-issue and retained-replay behavior.
- **BR-2** [Important] `user-surface-documentation` (kind: defect) README omits the new executable Plan prerequisite and retirement workflow
  cmd/sdlc/helptext/actual.md:38-43 introduces a user-facing usage contract: committed executable Plan membership controls fresh cuts and missing authority can refuse measurement. README.md:1059-1070 remains unchanged and does not explain this prerequisite or how retiring a milestone row affects replacement accounting. Add concise usage guidance covering committed Plan authority, retirement without clearing review findings, and unchanged saved replay; atlas coverage is already present.

## Open findings

- **BR-1** [Critical] `executable-milestone-authority` (kind: defect) Requested milestone validation is skipped when no named boundary is open
- **BR-2** [Important] `user-surface-documentation` (kind: defect) README omits the new executable Plan prerequisite and retirement workflow
