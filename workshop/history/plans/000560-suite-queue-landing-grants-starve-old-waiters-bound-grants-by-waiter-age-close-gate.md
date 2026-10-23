---
gate: boundary-review
issue: 560
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1393cf07-8a1f-484f-9881-ff17abc9cade
            repository: github.com/xianxu/ariadne
            issue_id: "000560"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c8a72681fe917250
            reviewed_head: 532929d548a9badfc18ccf1b433180af93f27037
            slot: ariadne:1
            requested_at: "2026-10-22T22:35:57.133123Z"
          phase: finished
          pid: 28161
          started_at: "2026-10-22T22:35:57.315073Z"
          finished_at: "2026-10-22T22:37:55.750063Z"
          session_id: 01a165c1-bb70-7232-bccf-8e3580a32ec0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 532929d548a9badfc18ccf1b433180af93f27037
        tracker_head: 3cf990c135d6768b4f6d8a302340f27f9a119073
        branch_point: 25024b5ed8de7bc930f357fbabb3a6ecf1771566
        since: "2026-10-22T15:10:08-07:00"
        until: "2026-10-22T15:34:24-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T15:10:08-07:00"
              end: "2026-10-22T15:34:24-07:00"
              began: "2026-10-22T15:10:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c4b2c592132e29135bf2601caf9ca0479fa22470fce89c438ff88616b62b86e9
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 85303558
              windows:
                - start: "2026-10-22T15:10:08-07:00"
                  end: "2026-10-22T15:34:24-07:00"
            - harness: codex
              id: 01a165b1-7420-76d0-9388-7e1db75d6bbb
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 105f5888afdff6a6716dcfe162a277ef768bb20028d153f5888f10c9260bc49f
              started: "2026-10-22T22:18:11.291Z"
              spawned: false
              bytes: 484280
              windows:
                - start: "2026-10-22T22:18:11.291Z"
                  end: "2026-10-22T15:34:24-07:00"
            - harness: codex
              id: 01a165b3-9b97-74f1-a0f9-38b1ea01b74b
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c522c77761c3d2cc37ee19988377032c51b3ce23067339d16ea0ec12094dd048
              started: "2026-10-22T22:20:32.51Z"
              spawned: false
              bytes: 400738
              windows:
                - start: "2026-10-22T22:20:32.51Z"
                  end: "2026-10-22T15:34:24-07:00"
            - harness: codex
              id: 01a165b4-777c-76d0-9bd9-a3cee9f18d7c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 79e0e08d33019943ed84d1b118279b8f2b2c6b8e66c760c3d9f96be0de9f171e
              started: "2026-10-22T22:21:28.727Z"
              spawned: false
              bytes: 430135
              windows:
                - start: "2026-10-22T22:21:28.727Z"
                  end: "2026-10-22T15:34:24-07:00"
            - harness: codex
              id: 01a165b4-f501-7a82-b548-af05786c8e99
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: bd377ec0fcd1ebacdd77ea5cb3d4335a55e75044000c6f4d82e94916abbdb30d
              started: "2026-10-22T22:22:00.927Z"
              spawned: false
              bytes: 394041
              windows:
                - start: "2026-10-22T22:22:00.927Z"
                  end: "2026-10-22T15:34:24-07:00"
      "n": 1
      timestamp: "2026-10-22T15:37:56-07:00"
      dispatched: "2026-10-22T15:35:57-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Status cancellation leaves the grant provider running
          detail: cmd/sdlc/queue_suite_status.go:101 kills only the inspector's process group, but query_priority at scripts/test-lock.py:919–920 launches the provider in a separate session. Killing the inspector bypasses its provider-cleanup finally block, leaving the provider and potentially its descendants alive after status returns. A private fixture using the actual inspect command reproduced distinct helper/provider groups and confirmed provider survival after the observer-equivalent SIGKILL. Preserve bounded ownership across both groups and add a regression through observeSuiteQueue using the real inspector with a blocked provider; cancellation and deadline return must leave neither provider nor descendants running. ARCH-ORDER, ARCH-FUNERAL.
          family: subprocess-cancellation-ownership
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 16dbd975-796a-4e14-b8cf-4525f885dba1
            repository: github.com/xianxu/ariadne
            issue_id: "000560"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/9248e5e836107ba6
            reviewed_head: 37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47
            slot: ariadne:1
            requested_at: "2026-10-22T22:46:04.537279Z"
          phase: finished
          pid: 3222
          started_at: "2026-10-22T22:46:04.70985Z"
          finished_at: "2026-10-22T22:47:58.442906Z"
          session_id: 01a165cb-0048-7b72-9571-8b44d014c773
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47
        tracker_head: 3cf990c135d6768b4f6d8a302340f27f9a119073
        branch_point: 25024b5ed8de7bc930f357fbabb3a6ecf1771566
        since: "2026-10-22T15:10:08-07:00"
        until: "2026-10-22T15:44:57-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T15:10:08-07:00"
              end: "2026-10-22T15:44:57-07:00"
              began: "2026-10-22T15:10:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 034ed55acf2126b5d3759a245389c17bb8fc278d1405d79b81d04b1dbd9ce231
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 85884671
              windows:
                - start: "2026-10-22T15:10:08-07:00"
                  end: "2026-10-22T15:44:57-07:00"
            - harness: codex
              id: 01a165b1-7420-76d0-9388-7e1db75d6bbb
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 105f5888afdff6a6716dcfe162a277ef768bb20028d153f5888f10c9260bc49f
              started: "2026-10-22T22:18:11.291Z"
              spawned: false
              bytes: 484280
              windows:
                - start: "2026-10-22T22:18:11.291Z"
                  end: "2026-10-22T15:44:57-07:00"
            - harness: codex
              id: 01a165b3-9b97-74f1-a0f9-38b1ea01b74b
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c522c77761c3d2cc37ee19988377032c51b3ce23067339d16ea0ec12094dd048
              started: "2026-10-22T22:20:32.51Z"
              spawned: false
              bytes: 400738
              windows:
                - start: "2026-10-22T22:20:32.51Z"
                  end: "2026-10-22T15:44:57-07:00"
            - harness: codex
              id: 01a165b4-777c-76d0-9bd9-a3cee9f18d7c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 79e0e08d33019943ed84d1b118279b8f2b2c6b8e66c760c3d9f96be0de9f171e
              started: "2026-10-22T22:21:28.727Z"
              spawned: false
              bytes: 430135
              windows:
                - start: "2026-10-22T22:21:28.727Z"
                  end: "2026-10-22T15:44:57-07:00"
            - harness: codex
              id: 01a165b4-f501-7a82-b548-af05786c8e99
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: bd377ec0fcd1ebacdd77ea5cb3d4335a55e75044000c6f4d82e94916abbdb30d
              started: "2026-10-22T22:22:00.927Z"
              spawned: false
              bytes: 394041
              windows:
                - start: "2026-10-22T22:22:00.927Z"
                  end: "2026-10-22T15:44:57-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.1865
        cuts:
            - dispatched: "2026-10-22T15:35:57-07:00"
              finalized: "2026-10-22T15:37:56-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-22T15:48:00-07:00"
      dispatched: "2026-10-22T15:46:04-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: scripts/test-lock.py:924 keeps supervised providers in the Go-owned group; queue_suite_status.go:103–107 collects that group on cancellation and completion. TestSuiteQueueRealInspectorOwnsProviderTree passes all four modes. In a disposable pinned-head copy, restoring detached sessions makes cancellation and deadline tests fail on surviving provider and descendant processes. The disposable copy was removed.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47
      blocked: false
---

# Gate ledger — ariadne#560 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T15:37:56-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `subprocess-cancellation-ownership` (kind: defect) Status cancellation leaves the grant provider running
  cmd/sdlc/queue_suite_status.go:101 kills only the inspector's process group, but query_priority at scripts/test-lock.py:919–920 launches the provider in a separate session. Killing the inspector bypasses its provider-cleanup finally block, leaving the provider and potentially its descendants alive after status returns. A private fixture using the actual inspect command reproduced distinct helper/provider groups and confirmed provider survival after the observer-equivalent SIGKILL. Preserve bounded ownership across both groups and add a regression through observeSuiteQueue using the real inspector with a blocked provider; cancellation and deadline return must leave neither provider nor descendants running. ARCH-ORDER, ARCH-FUNERAL.

## Round 2 — 2026-10-22T15:48:00-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — scripts/test-lock.py:924 keeps supervised providers in the Go-owned group; queue_suite_status.go:103–107 collects that group on cancellation and completion. TestSuiteQueueRealInspectorOwnsProviderTree passes all four modes. In a disposable pinned-head copy, restoring detached sessions makes cancellation and deadline tests fail on surviving provider and descendant processes. The disposable copy was removed.

## Open findings

(none — every finding has been disposed)
