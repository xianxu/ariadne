---
gate: boundary-review
issue: 580
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4698faab-b6cf-401f-9716-a2124871edff
            repository: github.com/xianxu/ariadne
            issue_id: "000580"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c5dff2b32ce509f2
            reviewed_head: 8d3d5948cfdfaf4eefa19957d73be00e09442d4e
            slot: ariadne:3
            requested_at: "2026-10-24T03:46:53.993784Z"
          phase: finished
          pid: 71707
          started_at: "2026-10-24T03:46:54.106611Z"
          finished_at: "2026-10-24T03:51:24.903377Z"
          session_id: 01a16c04-c552-7e60-a70d-b66340c5d8f3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8d3d5948cfdfaf4eefa19957d73be00e09442d4e
        tracker_head: fa1f09c67d791c4f7663565d9ed41b3aa2706823
        branch_point: 84b68e2fc5c0773ff38f13771507d93aaa08efd4
        since: "2026-10-23T19:27:24-07:00"
        until: "2026-10-23T20:46:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T19:27:24-07:00"
              end: "2026-10-23T20:46:18-07:00"
              began: "2026-10-23T19:27:24-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: bd04f53300c518ac99422b89e07fdb15d2b4a2fe7104e5014fbcd6b8825614fa
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 143419147
              windows:
                - start: "2026-10-23T19:27:24-07:00"
                  end: "2026-10-23T20:46:18-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.07971666666666667
      "n": 1
      timestamp: "2026-10-23T20:51:25-07:00"
      dispatched: "2026-10-23T20:46:53-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Ownership-only replay accepts a witnessed foreign issue branch
          detail: 'cmd/sdlc/actualbranch.go:311–319 validates branch/foreign-branch identity but accepts ownership-only intervals regardless of their canonical branch ID; measurement_branch.go:85 also leaves this case unchecked. A temporary test overlay relabelled the incident fixture''s foreign-branch interval as ownership-only and widened its saved selected window: resolveOwnedReplay and computeOwnedActualFor accepted it, crediting B with 29 minutes instead of 11. ARCH-SECURE requires contradictory persisted evidence to fail visibly. Reject ownership-only evidence naming any canonical issue branch, matching SelectBranchActivity''s rule, and add a regression through saved-input replay.'
          family: retained-attribution-consistency
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b39d6168-bb8e-4e0c-9df9-40cc0e7692f7
            repository: github.com/xianxu/ariadne
            issue_id: "000580"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2c5f6f4d6f7ba6df
            reviewed_head: dd39e615a6561921075b0066ba0cd09a8e0ca22d
            slot: ariadne:3
            requested_at: "2026-10-24T03:55:28.370329Z"
          phase: finished
          pid: 57059
          started_at: "2026-10-24T03:55:28.469924Z"
          finished_at: "2026-10-24T03:58:03.231204Z"
          session_id: 01a16c0c-9e7c-7b51-95cd-059a51c55db1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: dd39e615a6561921075b0066ba0cd09a8e0ca22d
        tracker_head: 688e4a9c56bbee124adedaaafca5b1fcdb6b5e72
        branch_point: 84b68e2fc5c0773ff38f13771507d93aaa08efd4
        since: "2026-10-23T19:27:24-07:00"
        until: "2026-10-23T20:54:56-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T19:27:24-07:00"
              end: "2026-10-23T20:54:56-07:00"
              began: "2026-10-23T19:27:24-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9bd71bd5a539e0682fdc0e06d70b08a9ac8af2ead25e6a8fc0dcc406c6475306
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 143949265
              windows:
                - start: "2026-10-23T19:27:24-07:00"
                  end: "2026-10-23T20:54:56-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.07971666666666667
        cuts:
            - dispatched: "2026-10-23T20:46:53-07:00"
              finalized: "2026-10-23T20:51:25-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T20:58:04-07:00"
      dispatched: "2026-10-23T20:55:28-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: actualbranch.go:311 rejects ownership-only intervals naming any canonical issue branch. Saved-replay and direct-execution regressions cover both own and foreign branches. Removing only this guard with a temporary Go overlay made both regressions fail; foreign replay again credited 29 minutes instead of 11. The pinned-head tests pass.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: dd39e615a6561921075b0066ba0cd09a8e0ca22d
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e6a37c9e-d3f3-4874-bafb-93a7deee9f74
            repository: github.com/xianxu/ariadne
            issue_id: "000580"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/62aeaff2b3fd0f19
            reviewed_head: 522b79112b8d6b9d89d22d57abf557f007bc235c
            slot: ariadne:3
            requested_at: "2026-10-24T05:18:39.143659Z"
          phase: finished
          pid: 21117
          started_at: "2026-10-24T05:18:39.263886Z"
          finished_at: "2026-10-24T05:24:48.195071Z"
          session_id: 01a16c58-c5e5-7e12-baf7-fb061297f018
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 522b79112b8d6b9d89d22d57abf557f007bc235c
        tracker_head: 733cd47d99bda636d2121144aa03c9c743a7cbd0
        branch_point: 84b68e2fc5c0773ff38f13771507d93aaa08efd4
        since: "2026-10-23T19:27:24-07:00"
        until: "2026-10-23T22:17:53-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T19:27:24-07:00"
              end: "2026-10-23T22:17:53-07:00"
              began: "2026-10-23T19:27:24-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e29c5668ce84501c25e433696a4142adbbb16be987da8cc979c3e00661838f9b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 145361428
              windows:
                - start: "2026-10-23T19:27:24-07:00"
                  end: "2026-10-23T22:17:53-07:00"
        queue_waits:
            - attempt_id: 9ac60fc393d946bc6b53e2b94c11a27b
              slot: ariadne:3
              record_digest: c65011bbd75420874a3547ae97127bff37ed5e5328fae457e4a26b24cc7c9417
              start: "2026-10-24T03:55:08.244205Z"
              end: "2026-10-24T05:00:40.84964Z"
              queue_seconds: 3932.605435
              raw_seconds: 3932.605435
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.07971666666666667
        cuts:
            - dispatched: "2026-10-23T20:46:53-07:00"
              finalized: "2026-10-23T20:51:25-07:00"
              finalizing: false
            - dispatched: "2026-10-23T20:55:28-07:00"
              finalized: "2026-10-23T20:58:04-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-23T22:24:49-07:00"
      dispatched: "2026-10-23T22:18:39-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The retained-proof guard at cmd/sdlc/actualbranch.go:311 remains intact. The concurrent-claim regression passed, including own and foreign canonical branches relabelled ownership-only through saved replay and direct execution.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 522b79112b8d6b9d89d22d57abf557f007bc235c
      blocked: false
---

# Gate ledger — ariadne#580 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T20:51:25-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `retained-attribution-consistency` (kind: defect) Ownership-only replay accepts a witnessed foreign issue branch
  cmd/sdlc/actualbranch.go:311–319 validates branch/foreign-branch identity but accepts ownership-only intervals regardless of their canonical branch ID; measurement_branch.go:85 also leaves this case unchecked. A temporary test overlay relabelled the incident fixture's foreign-branch interval as ownership-only and widened its saved selected window: resolveOwnedReplay and computeOwnedActualFor accepted it, crediting B with 29 minutes instead of 11. ARCH-SECURE requires contradictory persisted evidence to fail visibly. Reject ownership-only evidence naming any canonical issue branch, matching SelectBranchActivity's rule, and add a regression through saved-input replay.

## Round 2 — 2026-10-23T20:58:04-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — actualbranch.go:311 rejects ownership-only intervals naming any canonical issue branch. Saved-replay and direct-execution regressions cover both own and foreign branches. Removing only this guard with a temporary Go overlay made both regressions fail; foreign replay again credited 29 minutes instead of 11. The pinned-head tests pass.

## Round 3 — 2026-10-23T22:24:49-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The retained-proof guard at cmd/sdlc/actualbranch.go:311 remains intact. The concurrent-claim regression passed, including own and foreign canonical branches relabelled ownership-only through saved replay and direct execution.

## Open findings

(none — every finding has been disposed)
