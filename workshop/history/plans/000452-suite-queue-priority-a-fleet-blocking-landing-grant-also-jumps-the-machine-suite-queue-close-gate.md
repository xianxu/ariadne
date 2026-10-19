---
gate: boundary-review
issue: 452
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f5c17809-02e2-4770-9fba-e8fe1f55f378
            repository: github.com/xianxu/ariadne
            issue_id: "000452"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/bde00c2a014aaac8
            reviewed_head: eb44436e9e8781ac40a4888c228e2734f7ebd1bb
            slot: ariadne:3
            requested_at: "2026-10-19T01:55:58.79483Z"
          phase: finished
          pid: 26372
          started_at: "2026-10-19T01:55:58.922231Z"
          finished_at: "2026-10-19T01:58:17.126507Z"
          session_id: 01a151df-6c7d-76d2-a333-c1c4b3edc0b2
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: eb44436e9e8781ac40a4888c228e2734f7ebd1bb
        tracker_head: 259fad25d00c8cea8109a1945805e0db5c9cd2e4
        branch_point: 3a0012b449ff8488eb53b938ec591ca996ad46aa
        since: "2026-10-18T18:04:52-07:00"
        until: "2026-10-18T18:55:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T18:04:52-07:00"
              end: "2026-10-18T18:55:41-07:00"
              began: "2026-10-18T18:04:52-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: dc5c8a0f7f7151193bd5e612d0f6f1a745f79440219b0a45a46682164e04ca00
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 18792573
              windows:
                - start: "2026-10-18T18:04:52-07:00"
                  end: "2026-10-18T18:55:41-07:00"
            - harness: codex
              id: 01a151b0-f8b3-7730-a424-98e3136b9a71
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 8483af07cd1ba4c234c40e04d6f8796a30c38e21c7860acd2964fb4cc9d600d8
              started: "2026-10-19T01:05:15.019Z"
              spawned: true
              bytes: 359776
              windows:
                - start: "2026-10-19T01:05:15.019Z"
                  end: "2026-10-18T18:55:41-07:00"
            - harness: codex
              id: 01a151b8-57ff-7551-be32-48930d0927c7
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 36c35d4580b57893d808aed912be4ecd713da952be5f0c4032dff58ee2c0df86
              started: "2026-10-19T01:13:18.875Z"
              spawned: false
              bytes: 584817
              windows:
                - start: "2026-10-19T01:13:18.875Z"
                  end: "2026-10-18T18:55:41-07:00"
            - harness: codex
              id: 01a151b9-c33a-7220-be50-1faab4d207c0
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a4d69228c94995b43b2bcab25b568ceeb35a6e6fc1b7c1ebb9cf163635b2de9d
              started: "2026-10-19T01:14:52.08Z"
              spawned: false
              bytes: 439055
              windows:
                - start: "2026-10-19T01:14:52.08Z"
                  end: "2026-10-18T18:55:41-07:00"
            - harness: codex
              id: 01a151ba-d5e9-78c1-aab9-a87098ddc125
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 365008833dfdceb7743cc9db944adcf43b61d5c519bc3d692d96f6c444a4b907
              started: "2026-10-19T01:16:01.964Z"
              spawned: false
              bytes: 482666
              windows:
                - start: "2026-10-19T01:16:01.964Z"
                  end: "2026-10-18T18:55:41-07:00"
            - harness: codex
              id: 01a151bb-4195-78e2-97a7-b858be9075ec
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 2c1eea224dff3bb937a7b45f1be9020d212153d2b69bac9256152da741802f1e
              started: "2026-10-19T01:16:29.512Z"
              spawned: false
              bytes: 435228
              windows:
                - start: "2026-10-19T01:16:29.512Z"
                  end: "2026-10-18T18:55:41-07:00"
            - harness: codex
              id: 01a151bc-0da4-76f1-a140-454d67a715bb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 2a996dc73813f0a9293760a0e4596abba1c271eb9d730521011ed9af577020c2
              started: "2026-10-19T01:17:21.309Z"
              spawned: true
              bytes: 648877
              windows:
                - start: "2026-10-19T01:17:21.309Z"
                  end: "2026-10-18T18:55:41-07:00"
        queue_waits:
            - attempt_id: b7e282640d6d444a91c06c0b1c46c866
              slot: ariadne:3
              record_digest: e38f2f0efdc228e3766648cd6a3e0961fb5d68accd147a34920ea15efeba0ee6
              start: "2026-10-19T01:33:20.996589Z"
              end: "2026-10-19T01:40:06.363644Z"
              queue_seconds: 405.367055
              raw_seconds: 405.367055
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.11616666666666667
      "n": 1
      timestamp: "2026-10-18T18:58:18-07:00"
      dispatched: "2026-10-18T18:55:58-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Concurrent ticket removal can crash another suite's admission
          detail: scripts/test-lock.py:144-145 opens paths returned by live_tickets without handling disappearance, while Tickets.release at lines 165-171 unlinks tickets without the allocation mutex. An admitted, cancelled, or timed-out waiter can therefore remove its ticket between another waiter's scan and open. I reproduced FileNotFoundError by scheduling release at that boundary. Serialize removal with election or tolerate vanished tickets, and add a deterministic regression covering this interleaving (ARCH-ORDER).
          family: queue-ticket-lifecycle-serialization
          round: 1
        - id: BR-2
          severity: Critical
          title: Malformed ticket subjects crash admission instead of falling back to FIFO
          detail: scripts/test-lock.py:224-232 accepts an absolute common_dir containing an embedded NUL; GrantEvidence.deadlines then passes it to os.stat at line 265, whose ValueError escapes the OSError handler. I reproduced this with common_dir="/tmp/review\u0000bad". A malformed live ticket can consequently crash every participating waiter that reads it. Validate persisted subject paths before filesystem use and visibly discard invalid evidence; add malformed-ticket regressions (ARCH-SECURE).
          family: persisted-subject-validation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a0b91489-4dad-4250-8aab-73d1fc311e20
            repository: github.com/xianxu/ariadne
            issue_id: "000452"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a54cf661db7e568a
            reviewed_head: cbd3a63740a3c87de7c8657da7a7fda0e1d14cf3
            slot: ariadne:3
            requested_at: "2026-10-19T02:04:21.661143Z"
          phase: finished
          pid: 22517
          started_at: "2026-10-19T02:04:21.991358Z"
          finished_at: "2026-10-19T02:06:47.499899Z"
          session_id: 01a151e7-1a59-7613-bf29-30afc1f04833
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: cbd3a63740a3c87de7c8657da7a7fda0e1d14cf3
        tracker_head: 259fad25d00c8cea8109a1945805e0db5c9cd2e4
        branch_point: 3a0012b449ff8488eb53b938ec591ca996ad46aa
        since: "2026-10-18T18:04:52-07:00"
        until: "2026-10-18T19:02:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T18:04:52-07:00"
              end: "2026-10-18T19:02:55-07:00"
              began: "2026-10-18T18:04:52-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: cccec54e6fcbad1456f0b691fe712ba32710a3b3efa776effd31dd8762a3e46b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 19166236
              windows:
                - start: "2026-10-18T18:04:52-07:00"
                  end: "2026-10-18T19:02:55-07:00"
            - harness: codex
              id: 01a151b0-f8b3-7730-a424-98e3136b9a71
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 8483af07cd1ba4c234c40e04d6f8796a30c38e21c7860acd2964fb4cc9d600d8
              started: "2026-10-19T01:05:15.019Z"
              spawned: true
              bytes: 359776
              windows:
                - start: "2026-10-19T01:05:15.019Z"
                  end: "2026-10-18T19:02:55-07:00"
            - harness: codex
              id: 01a151b8-57ff-7551-be32-48930d0927c7
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 36c35d4580b57893d808aed912be4ecd713da952be5f0c4032dff58ee2c0df86
              started: "2026-10-19T01:13:18.875Z"
              spawned: false
              bytes: 584817
              windows:
                - start: "2026-10-19T01:13:18.875Z"
                  end: "2026-10-18T19:02:55-07:00"
            - harness: codex
              id: 01a151b9-c33a-7220-be50-1faab4d207c0
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a4d69228c94995b43b2bcab25b568ceeb35a6e6fc1b7c1ebb9cf163635b2de9d
              started: "2026-10-19T01:14:52.08Z"
              spawned: false
              bytes: 439055
              windows:
                - start: "2026-10-19T01:14:52.08Z"
                  end: "2026-10-18T19:02:55-07:00"
            - harness: codex
              id: 01a151ba-d5e9-78c1-aab9-a87098ddc125
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 365008833dfdceb7743cc9db944adcf43b61d5c519bc3d692d96f6c444a4b907
              started: "2026-10-19T01:16:01.964Z"
              spawned: false
              bytes: 482666
              windows:
                - start: "2026-10-19T01:16:01.964Z"
                  end: "2026-10-18T19:02:55-07:00"
            - harness: codex
              id: 01a151bb-4195-78e2-97a7-b858be9075ec
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 2c1eea224dff3bb937a7b45f1be9020d212153d2b69bac9256152da741802f1e
              started: "2026-10-19T01:16:29.512Z"
              spawned: false
              bytes: 435228
              windows:
                - start: "2026-10-19T01:16:29.512Z"
                  end: "2026-10-18T19:02:55-07:00"
            - harness: codex
              id: 01a151bc-0da4-76f1-a140-454d67a715bb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 2a996dc73813f0a9293760a0e4596abba1c271eb9d730521011ed9af577020c2
              started: "2026-10-19T01:17:21.309Z"
              spawned: true
              bytes: 648877
              windows:
                - start: "2026-10-19T01:17:21.309Z"
                  end: "2026-10-18T19:02:55-07:00"
        queue_waits:
            - attempt_id: b7e282640d6d444a91c06c0b1c46c866
              slot: ariadne:3
              record_digest: e38f2f0efdc228e3766648cd6a3e0961fb5d68accd147a34920ea15efeba0ee6
              start: "2026-10-19T01:33:20.996589Z"
              end: "2026-10-19T01:40:06.363644Z"
              queue_seconds: 405.367055
              raw_seconds: 405.367055
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.11616666666666667
        cuts:
            - dispatched: "2026-10-18T18:55:58-07:00"
              finalized: "2026-10-18T18:58:18-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T19:07:05-07:00"
      dispatched: "2026-10-18T19:04:21-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: scripts/test/test-lock-priority.test.py:302 deterministically releases an older ticket after enumeration. It passes at head and raises FileNotFoundError with the pre-fix implementation loaded in memory.
          round: 2
        - id: BR-2
          disposition: addressed
          note: scripts/test/test-lock-priority.test.py:330 exercises malformed persisted subjects, verifies visible fallback and prevents filesystem access. It passes at head and raises ValueError for the embedded-NUL case with the pre-fix implementation.
          round: 2
      findings:
        - kind: defect
          id: BR-3
          severity: Critical
          title: Core concepts table claims a modification absent from the pinned range
          detail: workshop/plans/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue-plan.md:26 lists cmd/sdlc/helpercapability.go under a modified entity, but the pinned diff contains no change to that file. The later revision correctly explains that existing capability checks remain unchanged. Correct the table to distinguish the modified localcheck_run.go binding from unchanged capability infrastructure, and add a Revisions entry reconciling it. This is a prose-only correction; no runtime change or wording-presence test is required.
          family: core-concept-change-traceability
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 76215204-6e74-4de4-8cec-b8672a76154f
            repository: github.com/xianxu/ariadne
            issue_id: "000452"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/268c07b459af1e60
            reviewed_head: 0e36a07b0b8ff23f565224549d685ef68a7c4e36
            slot: ariadne:3
            requested_at: "2026-10-19T02:09:21.910567Z"
          phase: finished
          pid: 78552
          started_at: "2026-10-19T02:09:22.091911Z"
          finished_at: "2026-10-19T02:11:17.553782Z"
          session_id: 01a151eb-ae31-70f1-afed-5955ac4e6690
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 0e36a07b0b8ff23f565224549d685ef68a7c4e36
        tracker_head: 259fad25d00c8cea8109a1945805e0db5c9cd2e4
        branch_point: 3a0012b449ff8488eb53b938ec591ca996ad46aa
        since: "2026-10-18T18:04:52-07:00"
        until: "2026-10-18T19:08:33-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T18:04:52-07:00"
              end: "2026-10-18T19:08:33-07:00"
              began: "2026-10-18T18:04:52-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 8e514da41390be916a40f32dac0b76bf1fb84266a956ede2e82402d62c9c6278
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 19344523
              windows:
                - start: "2026-10-18T18:04:52-07:00"
                  end: "2026-10-18T19:08:33-07:00"
            - harness: codex
              id: 01a151b0-f8b3-7730-a424-98e3136b9a71
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 8483af07cd1ba4c234c40e04d6f8796a30c38e21c7860acd2964fb4cc9d600d8
              started: "2026-10-19T01:05:15.019Z"
              spawned: true
              bytes: 359776
              windows:
                - start: "2026-10-19T01:05:15.019Z"
                  end: "2026-10-18T19:08:33-07:00"
            - harness: codex
              id: 01a151b8-57ff-7551-be32-48930d0927c7
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 36c35d4580b57893d808aed912be4ecd713da952be5f0c4032dff58ee2c0df86
              started: "2026-10-19T01:13:18.875Z"
              spawned: false
              bytes: 584817
              windows:
                - start: "2026-10-19T01:13:18.875Z"
                  end: "2026-10-18T19:08:33-07:00"
            - harness: codex
              id: 01a151b9-c33a-7220-be50-1faab4d207c0
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a4d69228c94995b43b2bcab25b568ceeb35a6e6fc1b7c1ebb9cf163635b2de9d
              started: "2026-10-19T01:14:52.08Z"
              spawned: false
              bytes: 439055
              windows:
                - start: "2026-10-19T01:14:52.08Z"
                  end: "2026-10-18T19:08:33-07:00"
            - harness: codex
              id: 01a151ba-d5e9-78c1-aab9-a87098ddc125
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 365008833dfdceb7743cc9db944adcf43b61d5c519bc3d692d96f6c444a4b907
              started: "2026-10-19T01:16:01.964Z"
              spawned: false
              bytes: 482666
              windows:
                - start: "2026-10-19T01:16:01.964Z"
                  end: "2026-10-18T19:08:33-07:00"
            - harness: codex
              id: 01a151bb-4195-78e2-97a7-b858be9075ec
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 2c1eea224dff3bb937a7b45f1be9020d212153d2b69bac9256152da741802f1e
              started: "2026-10-19T01:16:29.512Z"
              spawned: false
              bytes: 435228
              windows:
                - start: "2026-10-19T01:16:29.512Z"
                  end: "2026-10-18T19:08:33-07:00"
            - harness: codex
              id: 01a151bc-0da4-76f1-a140-454d67a715bb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 2a996dc73813f0a9293760a0e4596abba1c271eb9d730521011ed9af577020c2
              started: "2026-10-19T01:17:21.309Z"
              spawned: true
              bytes: 648877
              windows:
                - start: "2026-10-19T01:17:21.309Z"
                  end: "2026-10-18T19:08:33-07:00"
        queue_waits:
            - attempt_id: b7e282640d6d444a91c06c0b1c46c866
              slot: ariadne:3
              record_digest: e38f2f0efdc228e3766648cd6a3e0961fb5d68accd147a34920ea15efeba0ee6
              start: "2026-10-19T01:33:20.996589Z"
              end: "2026-10-19T01:40:06.363644Z"
              queue_seconds: 405.367055
              raw_seconds: 405.367055
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.11616666666666667
        cuts:
            - dispatched: "2026-10-18T18:55:58-07:00"
              finalized: "2026-10-18T18:58:18-07:00"
              finalizing: false
            - dispatched: "2026-10-18T19:04:21-07:00"
              finalized: "2026-10-18T19:07:05-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-18T19:11:19-07:00"
      dispatched: "2026-10-18T19:09:21-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The controlled ticket-removal regression passes at head and fails with FileNotFoundError against the pre-fix helper loaded in memory.
          round: 3
        - id: BR-2
          disposition: addressed
          note: The malformed-subject regression passes at head and fails with embedded-NUL ValueError against the pre-fix helper loaded in memory; validation prevents invalid identities reaching filesystem access.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Plan lines 26–27 now distinguish the modified localcheck_run.go binding from unchanged helpercapability.go; Task 2 and the Revisions entry at line 141 agree with the pinned implementation diff. This is a prose-only correction.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 0e36a07b0b8ff23f565224549d685ef68a7c4e36
      blocked: false
---

# Gate ledger — ariadne#452 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T18:58:18-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `queue-ticket-lifecycle-serialization` Concurrent ticket removal can crash another suite's admission
  scripts/test-lock.py:144-145 opens paths returned by live_tickets without handling disappearance, while Tickets.release at lines 165-171 unlinks tickets without the allocation mutex. An admitted, cancelled, or timed-out waiter can therefore remove its ticket between another waiter's scan and open. I reproduced FileNotFoundError by scheduling release at that boundary. Serialize removal with election or tolerate vanished tickets, and add a deterministic regression covering this interleaving (ARCH-ORDER).
- **BR-2** [Critical] `persisted-subject-validation` Malformed ticket subjects crash admission instead of falling back to FIFO
  scripts/test-lock.py:224-232 accepts an absolute common_dir containing an embedded NUL; GrantEvidence.deadlines then passes it to os.stat at line 265, whose ValueError escapes the OSError handler. I reproduced this with common_dir="/tmp/review\u0000bad". A malformed live ticket can consequently crash every participating waiter that reads it. Validate persisted subject paths before filesystem use and visibly discard invalid evidence; add malformed-ticket regressions (ARCH-SECURE).

## Round 2 — 2026-10-18T19:07:05-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — scripts/test/test-lock-priority.test.py:302 deterministically releases an older ticket after enumeration. It passes at head and raises FileNotFoundError with the pre-fix implementation loaded in memory.
- BR-2 — addressed — scripts/test/test-lock-priority.test.py:330 exercises malformed persisted subjects, verifies visible fallback and prevents filesystem access. It passes at head and raises ValueError for the embedded-NUL case with the pre-fix implementation.

### Raised

- **BR-3** [Critical] `core-concept-change-traceability` (kind: defect) Core concepts table claims a modification absent from the pinned range
  workshop/plans/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue-plan.md:26 lists cmd/sdlc/helpercapability.go under a modified entity, but the pinned diff contains no change to that file. The later revision correctly explains that existing capability checks remain unchanged. Correct the table to distinguish the modified localcheck_run.go binding from unchanged capability infrastructure, and add a Revisions entry reconciling it. This is a prose-only correction; no runtime change or wording-presence test is required.

## Round 3 — 2026-10-18T19:11:19-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The controlled ticket-removal regression passes at head and fails with FileNotFoundError against the pre-fix helper loaded in memory.
- BR-2 — addressed — The malformed-subject regression passes at head and fails with embedded-NUL ValueError against the pre-fix helper loaded in memory; validation prevents invalid identities reaching filesystem access.
- BR-3 — addressed — Plan lines 26–27 now distinguish the modified localcheck_run.go binding from unchanged helpercapability.go; Task 2 and the Revisions entry at line 141 agree with the pinned implementation diff. This is a prose-only correction.

## Open findings

(none — every finding has been disposed)
