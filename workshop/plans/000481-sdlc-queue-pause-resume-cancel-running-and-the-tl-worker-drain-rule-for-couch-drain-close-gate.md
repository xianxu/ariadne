---
gate: boundary-review
issue: 481
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 5b887303-5339-4b9b-83e7-6ca77bfa298f
            repository: github.com/xianxu/ariadne
            issue_id: "000481"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c96aaf0e0e803e05
            reviewed_head: 7bedb19c29d5c52404bd2922a9d8df1e9e454102
            slot: ariadne:3
            requested_at: "2026-10-20T08:55:51.943938Z"
          phase: finished
          pid: 22460
          started_at: "2026-10-20T08:55:52.085778Z"
          finished_at: "2026-10-20T08:58:18.363307Z"
          session_id: 01a15886-335e-7d91-a2b3-e17641373b01
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7bedb19c29d5c52404bd2922a9d8df1e9e454102
        tracker_head: 97ccd9410a99b8cadb917699d8779fb664a986de
        branch_point: 8a44224ff0f867837283368f62c0344bcd322b25
        since: "2026-10-20T01:19:56-07:00"
        until: "2026-10-20T01:55:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:19:56-07:00"
              end: "2026-10-20T01:55:25-07:00"
              began: "2026-10-20T01:19:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c4d5253780abccdf1ce30648f0b45486020f245af078260cd194a77f930957a6
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 70180593
              windows:
                - start: "2026-10-20T01:19:56-07:00"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15865-a812-7ae1-b293-fd0d91059cdb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 6e279d4bf2f8f0edc2ff5186fecf9a261f1445104597c5d55d57a401d8475e40
              started: "2026-10-20T08:20:19.645Z"
              spawned: true
              bytes: 1570280
              windows:
                - start: "2026-10-20T08:20:19.645Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a1586d-3992-7242-ab2c-4427a72c88aa
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5a8d5f93adc68a80491a25a1d6b595b3594c243959b777603d21fd6c28ba5ec
              started: "2026-10-20T08:28:36.266Z"
              spawned: false
              bytes: 710285
              windows:
                - start: "2026-10-20T08:28:36.266Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15871-6263-7d92-8c75-64419f58cf1d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 404c0af4ca27bfe70cd74e03ba6f942cc7bbf8678821ab9511639517491ced4a
              started: "2026-10-20T08:33:08.701Z"
              spawned: false
              bytes: 515584
              windows:
                - start: "2026-10-20T08:33:08.701Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15872-cb08-7721-aad1-06cfa3bdf4f5
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ca135eb5ab2f41048986c73ba8d8265f86e5859e72499657056a1a6fe19621e4
              started: "2026-10-20T08:34:40.935Z"
              spawned: false
              bytes: 469904
              windows:
                - start: "2026-10-20T08:34:40.935Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15873-4a93-7760-99c7-2c400abf6c0a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fbe94f17549253511c8ace0b7a7bdba4c7b8816404600763a10222d53182f8eb
              started: "2026-10-20T08:35:13.533Z"
              spawned: false
              bytes: 424682
              windows:
                - start: "2026-10-20T08:35:13.533Z"
                  end: "2026-10-20T01:55:25-07:00"
      "n": 1
      timestamp: "2026-10-20T01:58:18-07:00"
      dispatched: "2026-10-20T01:55:51-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Durable cancellation intent is not reconciled after partial publication or owner death
          detail: 'cmd/sdlc/internal/queuegate/store.go:244 writes cancel/<id>.json before rewriting the lease, but lease.go:156-168 checks only the lease phase. If the pause process dies between those writes, CancellationRequested returns false and Observe never consumes the persisted intent. Separately, Observe at store.go:333-343 converts a dead interrupting lease into unresolved evidence without retaining the promised interruption record or removing its cancellation intent; prior-boot reconciliation later removes only the unresolved record. This loses interruption diagnostics and leaves permanent cancellation files that can eventually exceed operationFiles'' reader limit. ARCH-ORDER/ARCH-FUNERAL: reconcile exact-ID intents across partial publication and stale-lease recovery, preserve interruption evidence alongside descendant uncertainty, and clean intents after their last reader. Add deterministic crash-window and repeated recovery/boot tests.'
          family: durable-cancellation-reconciliation
          round: 1
        - id: BR-2
          severity: Critical
          title: The Core concepts table names entities and transition behavior absent from the implementation
          detail: 'workshop/plans/000481-queue-drain-plan.md:31-35 lists PauseSet, OperationPhase and Interruption in model.go and says Transition returns requested interruption IDs with owner state protected from direct mutation. The implementation instead exposes Control, string phases and Operation-based interruption records; Transition returns only (Control, error). LegacyObserver at plan line 46 is also implemented as observeLegacyQueue rather than the named entity. Under the required Core concepts cross-check, these are plan/code contradictions. ARCH-PURPOSE/ARCH-ORDER: append a Revisions entry reconciling the table and structural-enforcement claims with the actual API, or implement the promised model.'
          family: core-concept-traceability
          round: 1
        - id: BR-3
          severity: Important
          title: Queue command deadlines stop applying once the control mutex is acquired
          detail: 'cmd/sdlc/internal/queuegate/store.go:67-72 checks context only while acquiring the mutex, then invokes a callback without deadline checks. requestInterruptions at lines 209-258 can perform durable intent writes and lease rewrites for 4096 operations while holding that mutex; Observe likewise scans and reconciles records without checking context inside its transaction. Consequently queuecmd.go:140''s three-second timeout does not enforce the plan''s under-four-second target or the peer''s five-second ceiling. ARCH-CONSTRAINTS: propagate cancellation through long transaction loops, retain recoverable partial progress, and measure representative upper-bound workloads.'
          family: transaction-deadline-enforcement
          round: 1
        - id: BR-4
          severity: Minor
          title: Strict JSON validation is duplicated across the new boundary readers
          detail: 'cmd/sdlc/queue_legacy.go:285-350 repeats the recursive duplicate-key, depth, UTF-8 and trailing-data validation in internal/queuegate/codec.go:17-85. ARCH-DRY: extract a shared strict JSON inspection helper while retaining each reader''s schema-specific decoding.'
          family: shared-json-boundary-validation
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2c8061d6-03b6-42e5-934d-2bbcb01a1a8d
            repository: github.com/xianxu/ariadne
            issue_id: "000481"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/b740070dea487c18
            reviewed_head: 7f5afe02d77fe2015d05fac577f884de7ebbd6d1
            slot: ariadne:3
            requested_at: "2026-10-20T09:15:29.502453Z"
          phase: finished
          pid: 52279
          started_at: "2026-10-20T09:15:29.628808Z"
          finished_at: "2026-10-20T09:18:40.595687Z"
          session_id: 01a15898-2b03-7912-8ea8-213071bb647d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7f5afe02d77fe2015d05fac577f884de7ebbd6d1
        tracker_head: 2a41686cbdf2a2a3a81fb907bb66985c222768af
        branch_point: 8a44224ff0f867837283368f62c0344bcd322b25
        since: "2026-10-20T01:19:56-07:00"
        until: "2026-10-20T02:15:05-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:19:56-07:00"
              end: "2026-10-20T02:15:05-07:00"
              began: "2026-10-20T01:19:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9f491dc07c651c092165856bc947c12c590ab8980806c509e2be3fc6b4f5f5f1
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 71744920
              windows:
                - start: "2026-10-20T01:19:56-07:00"
                  end: "2026-10-20T02:15:05-07:00"
            - harness: codex
              id: 01a15865-a812-7ae1-b293-fd0d91059cdb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 97a621bd384b64f69728dcf6732c6fa28369d28b23bda8a86771436163b5e8b0
              started: "2026-10-20T08:20:19.645Z"
              spawned: true
              bytes: 1806473
              windows:
                - start: "2026-10-20T08:20:19.645Z"
                  end: "2026-10-20T02:15:05-07:00"
            - harness: codex
              id: 01a1586d-3992-7242-ab2c-4427a72c88aa
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5a8d5f93adc68a80491a25a1d6b595b3594c243959b777603d21fd6c28ba5ec
              started: "2026-10-20T08:28:36.266Z"
              spawned: false
              bytes: 710285
              windows:
                - start: "2026-10-20T08:28:36.266Z"
                  end: "2026-10-20T02:15:05-07:00"
            - harness: codex
              id: 01a15871-6263-7d92-8c75-64419f58cf1d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 404c0af4ca27bfe70cd74e03ba6f942cc7bbf8678821ab9511639517491ced4a
              started: "2026-10-20T08:33:08.701Z"
              spawned: false
              bytes: 515584
              windows:
                - start: "2026-10-20T08:33:08.701Z"
                  end: "2026-10-20T02:15:05-07:00"
            - harness: codex
              id: 01a15872-cb08-7721-aad1-06cfa3bdf4f5
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ca135eb5ab2f41048986c73ba8d8265f86e5859e72499657056a1a6fe19621e4
              started: "2026-10-20T08:34:40.935Z"
              spawned: false
              bytes: 469904
              windows:
                - start: "2026-10-20T08:34:40.935Z"
                  end: "2026-10-20T02:15:05-07:00"
            - harness: codex
              id: 01a15873-4a93-7760-99c7-2c400abf6c0a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fbe94f17549253511c8ace0b7a7bdba4c7b8816404600763a10222d53182f8eb
              started: "2026-10-20T08:35:13.533Z"
              spawned: false
              bytes: 424682
              windows:
                - start: "2026-10-20T08:35:13.533Z"
                  end: "2026-10-20T02:15:05-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:55:51-07:00"
              finalized: "2026-10-20T01:58:18-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T02:18:41-07:00"
      dispatched: "2026-10-20T02:15:29-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: lease.go:167 consumes exact-ID durable intents; store.go:381 reconciles them during observation and retains diagnostics before unresolved recovery. Crash-window, repeated-recovery, orphan-proof and boot tests pass. In disposable pinned-head copies, removing owner intent consumption or recovery diagnostic retention made the corresponding regression tests fail.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The plan's Core concepts passages now name Control, TransitionOperation, InterruptionRecord and observeLegacyQueue and accurately describe mutable detached values and store-owned persistence. The Revisions entry at plan line 199 explains the correction; signatures and implementations match the revised M1 inventory.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Transaction loops now check cancellation and preserve applied holds and partial intents. Removing the checks in a disposable copy made both deadline regressions fail after reading all 16 records. The pinned-head capacity test scanned exactly 4096 records and completed pause publication in 222 ms.
          round: 2
        - id: BR-4
          disposition: addressed
          note: queue_legacy.go:286 delegates structural validation to queuegate.StrictObject, also used by typed decoding. The duplicated recursive validator is removed. Shared-helper adversarial tests and legacy malformed-input tests pass; this is a behavior-preserving consolidation.
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 7f5afe02d77fe2015d05fac577f884de7ebbd6d1
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 028e317a-5938-4738-8a59-97ee102a17b1
            repository: github.com/xianxu/ariadne
            issue_id: "000481"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/d96a99e0fc0e7a54
            reviewed_head: 7c307ed8b35875f36791fd65d6256a08ad2a5644
            slot: ariadne:3
            requested_at: "2026-10-20T09:57:26.103159Z"
          phase: finished
          pid: 57488
          started_at: "2026-10-20T09:57:26.258262Z"
          finished_at: "2026-10-20T10:02:22.183238Z"
          session_id: 01a158be-9200-7e20-b3c8-1340920e89cc
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7c307ed8b35875f36791fd65d6256a08ad2a5644
        tracker_head: 0c158439dfa9bdb30b733b79c8c0bfb6aa716bcd
        branch_point: 8a44224ff0f867837283368f62c0344bcd322b25
        since: "2026-10-20T01:19:56-07:00"
        until: "2026-10-20T02:56:15-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:19:56-07:00"
              end: "2026-10-20T02:56:15-07:00"
              began: "2026-10-20T01:19:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 87fa8d932f4d3291cfea007ea9a3da0ed340ca03d3345048088105ce1e5e52a5
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 74061517
              windows:
                - start: "2026-10-20T01:19:56-07:00"
                  end: "2026-10-20T02:56:15-07:00"
            - harness: codex
              id: 01a15865-a812-7ae1-b293-fd0d91059cdb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 93bf6ed81fe046b6ebc147dd0b8e6d0e0a9e90c3b3e06e840c9b0d1a41545688
              started: "2026-10-20T08:20:19.645Z"
              spawned: true
              bytes: 2311426
              windows:
                - start: "2026-10-20T08:20:19.645Z"
                  end: "2026-10-20T02:56:15-07:00"
            - harness: codex
              id: 01a1586d-3992-7242-ab2c-4427a72c88aa
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5a8d5f93adc68a80491a25a1d6b595b3594c243959b777603d21fd6c28ba5ec
              started: "2026-10-20T08:28:36.266Z"
              spawned: false
              bytes: 710285
              windows:
                - start: "2026-10-20T08:28:36.266Z"
                  end: "2026-10-20T02:56:15-07:00"
            - harness: codex
              id: 01a15871-6263-7d92-8c75-64419f58cf1d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 404c0af4ca27bfe70cd74e03ba6f942cc7bbf8678821ab9511639517491ced4a
              started: "2026-10-20T08:33:08.701Z"
              spawned: false
              bytes: 515584
              windows:
                - start: "2026-10-20T08:33:08.701Z"
                  end: "2026-10-20T02:56:15-07:00"
            - harness: codex
              id: 01a15872-cb08-7721-aad1-06cfa3bdf4f5
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ca135eb5ab2f41048986c73ba8d8265f86e5859e72499657056a1a6fe19621e4
              started: "2026-10-20T08:34:40.935Z"
              spawned: false
              bytes: 469904
              windows:
                - start: "2026-10-20T08:34:40.935Z"
                  end: "2026-10-20T02:56:15-07:00"
            - harness: codex
              id: 01a15873-4a93-7760-99c7-2c400abf6c0a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fbe94f17549253511c8ace0b7a7bdba4c7b8816404600763a10222d53182f8eb
              started: "2026-10-20T08:35:13.533Z"
              spawned: false
              bytes: 424682
              windows:
                - start: "2026-10-20T08:35:13.533Z"
                  end: "2026-10-20T02:56:15-07:00"
            - harness: codex
              id: 01a1589b-b325-7bf2-b979-0c0c3d0ee141
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 7a0e775b5c49965500a4db61998033ab8a5d64b58436f2a05e3683fa9b00a0b5
              started: "2026-10-20T09:19:21.466Z"
              spawned: true
              bytes: 1127626
              windows:
                - start: "2026-10-20T09:19:21.466Z"
                  end: "2026-10-20T02:56:15-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:55:51-07:00"
              finalized: "2026-10-20T01:58:18-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:15:29-07:00"
              finalized: "2026-10-20T02:18:41-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-20T03:02:26-07:00"
      dispatched: "2026-10-20T02:57:26-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Admission waits indefinitely on secondary locks while holding the global gate
          detail: 'scripts/test-lock.py:620 takes Tickets.locked''s blocking flock at line 559 inside QueueGate.try_admit''s global transaction. cmd/sdlc/internal/landingline/line.go:234 similarly calls withMutex''s blocking flock at line 559 inside gate(resource). A stopped or delayed allocation-lock holder therefore monopolizes the machine gate: pause/resume cannot acquire it, and unrelated admissions stall. A private foreground reproduction using the production Python admission path confirmed that admission remained blocked and control.lock remained held after 4.2 seconds; releasing the allocation lock immediately completed admission. This is the 2nd finding in family `transaction-deadline-enforcement`. Earlier rounds fixed instances. Do NOT fix only this instance: enforce the rule that every secondary lock acquired inside the global gate is nonblocking, releasing the gate and retrying on contention. Sweep both Go landing and Python weighted admission, preserve global-before-local ordering, and add deterministic contention tests proving pause remains responsive and cancellation terminates the waiter. ARCH-CONSTRAINTS / ARCH-ORDER.'
          family: transaction-deadline-enforcement
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b5fb9a33-5894-4f92-90f4-745b409f477c
            repository: github.com/xianxu/ariadne
            issue_id: "000481"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/bf4f905406079241
            reviewed_head: afb94a8e2be9d2477fbbe7b53d7b8eb7f52152c7
            slot: ariadne:3
            requested_at: "2026-10-20T10:18:49.928213Z"
          phase: finished
          pid: 47961
          started_at: "2026-10-20T10:18:50.062487Z"
          finished_at: "2026-10-20T10:22:53.839615Z"
          session_id: 01a158d2-28b3-7f71-8175-c0c69ed43865
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: afb94a8e2be9d2477fbbe7b53d7b8eb7f52152c7
        tracker_head: 5b852fb1a54ec096eaced18b17f9749319127bbf
        branch_point: 8a44224ff0f867837283368f62c0344bcd322b25
        since: "2026-10-20T01:19:56-07:00"
        until: "2026-10-20T03:18:05-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:19:56-07:00"
              end: "2026-10-20T03:18:05-07:00"
              began: "2026-10-20T01:19:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6f54d25cd0a5083782c5ed841f6915af8f9bd60698d7f8e90b1f9935fe8b289b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 75756128
              windows:
                - start: "2026-10-20T01:19:56-07:00"
                  end: "2026-10-20T03:18:05-07:00"
            - harness: codex
              id: 01a15865-a812-7ae1-b293-fd0d91059cdb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 93bf6ed81fe046b6ebc147dd0b8e6d0e0a9e90c3b3e06e840c9b0d1a41545688
              started: "2026-10-20T08:20:19.645Z"
              spawned: true
              bytes: 2311426
              windows:
                - start: "2026-10-20T08:20:19.645Z"
                  end: "2026-10-20T03:18:05-07:00"
            - harness: codex
              id: 01a1586d-3992-7242-ab2c-4427a72c88aa
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5a8d5f93adc68a80491a25a1d6b595b3594c243959b777603d21fd6c28ba5ec
              started: "2026-10-20T08:28:36.266Z"
              spawned: false
              bytes: 710285
              windows:
                - start: "2026-10-20T08:28:36.266Z"
                  end: "2026-10-20T03:18:05-07:00"
            - harness: codex
              id: 01a15871-6263-7d92-8c75-64419f58cf1d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 404c0af4ca27bfe70cd74e03ba6f942cc7bbf8678821ab9511639517491ced4a
              started: "2026-10-20T08:33:08.701Z"
              spawned: false
              bytes: 515584
              windows:
                - start: "2026-10-20T08:33:08.701Z"
                  end: "2026-10-20T03:18:05-07:00"
            - harness: codex
              id: 01a15872-cb08-7721-aad1-06cfa3bdf4f5
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ca135eb5ab2f41048986c73ba8d8265f86e5859e72499657056a1a6fe19621e4
              started: "2026-10-20T08:34:40.935Z"
              spawned: false
              bytes: 469904
              windows:
                - start: "2026-10-20T08:34:40.935Z"
                  end: "2026-10-20T03:18:05-07:00"
            - harness: codex
              id: 01a15873-4a93-7760-99c7-2c400abf6c0a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fbe94f17549253511c8ace0b7a7bdba4c7b8816404600763a10222d53182f8eb
              started: "2026-10-20T08:35:13.533Z"
              spawned: false
              bytes: 424682
              windows:
                - start: "2026-10-20T08:35:13.533Z"
                  end: "2026-10-20T03:18:05-07:00"
            - harness: codex
              id: 01a1589b-b325-7bf2-b979-0c0c3d0ee141
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 7a0e775b5c49965500a4db61998033ab8a5d64b58436f2a05e3683fa9b00a0b5
              started: "2026-10-20T09:19:21.466Z"
              spawned: true
              bytes: 1127626
              windows:
                - start: "2026-10-20T09:19:21.466Z"
                  end: "2026-10-20T03:18:05-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-20T01:55:51-07:00"
              finalized: "2026-10-20T01:58:18-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-20T02:15:29-07:00"
              finalized: "2026-10-20T02:18:41-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-20T02:57:26-07:00"
              finalized: "2026-10-20T03:02:26-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-20T03:22:54-07:00"
      dispatched: "2026-10-20T03:18:49-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-5
          disposition: addressed
          note: Go landing admission and Python weighted admission now acquire secondary locks nonblockingly under the global gate. Contention tests cover responsive pause, waiter cancellation and retained FIFO position. Independently reverting each fix in temporary copies made its regression test fail; both pass at the pinned head.
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: afb94a8e2be9d2477fbbe7b53d7b8eb7f52152c7
      blocked: false
---

# Gate ledger — ariadne#481 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T01:58:18-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `durable-cancellation-reconciliation` Durable cancellation intent is not reconciled after partial publication or owner death
  cmd/sdlc/internal/queuegate/store.go:244 writes cancel/<id>.json before rewriting the lease, but lease.go:156-168 checks only the lease phase. If the pause process dies between those writes, CancellationRequested returns false and Observe never consumes the persisted intent. Separately, Observe at store.go:333-343 converts a dead interrupting lease into unresolved evidence without retaining the promised interruption record or removing its cancellation intent; prior-boot reconciliation later removes only the unresolved record. This loses interruption diagnostics and leaves permanent cancellation files that can eventually exceed operationFiles' reader limit. ARCH-ORDER/ARCH-FUNERAL: reconcile exact-ID intents across partial publication and stale-lease recovery, preserve interruption evidence alongside descendant uncertainty, and clean intents after their last reader. Add deterministic crash-window and repeated recovery/boot tests.
- **BR-2** [Critical] `core-concept-traceability` The Core concepts table names entities and transition behavior absent from the implementation
  workshop/plans/000481-queue-drain-plan.md:31-35 lists PauseSet, OperationPhase and Interruption in model.go and says Transition returns requested interruption IDs with owner state protected from direct mutation. The implementation instead exposes Control, string phases and Operation-based interruption records; Transition returns only (Control, error). LegacyObserver at plan line 46 is also implemented as observeLegacyQueue rather than the named entity. Under the required Core concepts cross-check, these are plan/code contradictions. ARCH-PURPOSE/ARCH-ORDER: append a Revisions entry reconciling the table and structural-enforcement claims with the actual API, or implement the promised model.
- **BR-3** [Important] `transaction-deadline-enforcement` Queue command deadlines stop applying once the control mutex is acquired
  cmd/sdlc/internal/queuegate/store.go:67-72 checks context only while acquiring the mutex, then invokes a callback without deadline checks. requestInterruptions at lines 209-258 can perform durable intent writes and lease rewrites for 4096 operations while holding that mutex; Observe likewise scans and reconciles records without checking context inside its transaction. Consequently queuecmd.go:140's three-second timeout does not enforce the plan's under-four-second target or the peer's five-second ceiling. ARCH-CONSTRAINTS: propagate cancellation through long transaction loops, retain recoverable partial progress, and measure representative upper-bound workloads.
- **BR-4** [Minor] `shared-json-boundary-validation` Strict JSON validation is duplicated across the new boundary readers
  cmd/sdlc/queue_legacy.go:285-350 repeats the recursive duplicate-key, depth, UTF-8 and trailing-data validation in internal/queuegate/codec.go:17-85. ARCH-DRY: extract a shared strict JSON inspection helper while retaining each reader's schema-specific decoding.

## Round 2 — 2026-10-20T02:18:41-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — lease.go:167 consumes exact-ID durable intents; store.go:381 reconciles them during observation and retains diagnostics before unresolved recovery. Crash-window, repeated-recovery, orphan-proof and boot tests pass. In disposable pinned-head copies, removing owner intent consumption or recovery diagnostic retention made the corresponding regression tests fail.
- BR-2 — addressed — The plan's Core concepts passages now name Control, TransitionOperation, InterruptionRecord and observeLegacyQueue and accurately describe mutable detached values and store-owned persistence. The Revisions entry at plan line 199 explains the correction; signatures and implementations match the revised M1 inventory.
- BR-3 — addressed — Transaction loops now check cancellation and preserve applied holds and partial intents. Removing the checks in a disposable copy made both deadline regressions fail after reading all 16 records. The pinned-head capacity test scanned exactly 4096 records and completed pause publication in 222 ms.
- BR-4 — addressed — queue_legacy.go:286 delegates structural validation to queuegate.StrictObject, also used by typed decoding. The duplicated recursive validator is removed. Shared-helper adversarial tests and legacy malformed-input tests pass; this is a behavior-preserving consolidation.

## Round 3 — 2026-10-20T03:02:26-07:00 (codex) — BLOCKED

### Raised

- **BR-5** [Critical] `transaction-deadline-enforcement` (kind: defect) Admission waits indefinitely on secondary locks while holding the global gate
  scripts/test-lock.py:620 takes Tickets.locked's blocking flock at line 559 inside QueueGate.try_admit's global transaction. cmd/sdlc/internal/landingline/line.go:234 similarly calls withMutex's blocking flock at line 559 inside gate(resource). A stopped or delayed allocation-lock holder therefore monopolizes the machine gate: pause/resume cannot acquire it, and unrelated admissions stall. A private foreground reproduction using the production Python admission path confirmed that admission remained blocked and control.lock remained held after 4.2 seconds; releasing the allocation lock immediately completed admission. This is the 2nd finding in family `transaction-deadline-enforcement`. Earlier rounds fixed instances. Do NOT fix only this instance: enforce the rule that every secondary lock acquired inside the global gate is nonblocking, releasing the gate and retrying on contention. Sweep both Go landing and Python weighted admission, preserve global-before-local ordering, and add deterministic contention tests proving pause remains responsive and cancellation terminates the waiter. ARCH-CONSTRAINTS / ARCH-ORDER.

## Round 4 — 2026-10-20T03:22:54-07:00 (codex) — passed

### Disposed

- BR-5 — addressed — Go landing admission and Python weighted admission now acquire secondary locks nonblockingly under the global gate. Contention tests cover responsive pause, waiter cancellation and retained FIFO position. Independently reverting each fix in temporary copies made its regression test fail; both pass at the pinned head.

## Open findings

(none — every finding has been disposed)
