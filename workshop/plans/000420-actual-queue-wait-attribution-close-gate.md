---
gate: boundary-review
issue: 420
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 42c1741a-9784-4319-b090-3f12c52210e2
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/eb5e6ff16e87a774
            reviewed_head: fe2155eda0572e95b2e5243c164698decddfcf3a
            slot: ariadne:1
            requested_at: "2026-10-17T09:46:59.009541Z"
          phase: finished
          pid: 45129
          started_at: "2026-10-17T09:46:59.129733Z"
          finished_at: "2026-10-17T09:51:31.199571Z"
          session_id: 01a14941-ec21-71e2-8cb9-f6e498416133
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: fe2155eda0572e95b2e5243c164698decddfcf3a
        tracker_head: e989ed358facd8920cdc63dc245c4430d85beace
        branch_point: 0a1f6a454d8192e798fade61106cedede1774795
        since: "2026-10-17T00:53:13-07:00"
        until: "2026-10-17T02:45:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T00:53:13-07:00"
              end: "2026-10-17T02:45:41-07:00"
              began: "2026-10-17T00:53:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 12aac8a3cd1d5022f28fa23ae553df3e399267a2dafacfc3593c66195d984778
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 52193671
              windows:
                - start: "2026-10-17T00:53:13-07:00"
                  end: "2026-10-17T02:45:41-07:00"
            - harness: codex
              id: 01a148dd-d9dc-7b41-8ff5-37fb3d72ae46
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 87ed2689fe0f0571ac966d9c43d9baed2e6b74b7b96fef94d0b83b5589259f28
              started: "2026-10-17T07:57:41.883Z"
              spawned: false
              bytes: 603729
              windows:
                - start: "2026-10-17T07:57:41.883Z"
                  end: "2026-10-17T02:45:41-07:00"
            - harness: codex
              id: 01a148df-e9ee-7061-9d8d-4bd9f106aedf
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f4e79066b008f729a28d08b863f1a8623c12b7287055544ff958dd0ca78bc478
              started: "2026-10-17T07:59:56.888Z"
              spawned: false
              bytes: 450509
              windows:
                - start: "2026-10-17T07:59:56.888Z"
                  end: "2026-10-17T02:45:41-07:00"
            - harness: codex
              id: 01a148e1-33c7-7f21-8981-aa85a762079f
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6e71a79536593227c575c21f9b579b4c1fefab82e87d3e63525cdf0f96f255d0
              started: "2026-10-17T08:01:21.375Z"
              spawned: false
              bytes: 459559
              windows:
                - start: "2026-10-17T08:01:21.375Z"
                  end: "2026-10-17T02:45:41-07:00"
            - harness: codex
              id: 01a148e2-2b49-7230-8d38-9b246eb10ef1
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f9e3d27cc800f5eacab8e808fdf5b07f5ab675fd515d4ae79355dac67d6873e9
              started: "2026-10-17T08:02:24.452Z"
              spawned: false
              bytes: 452872
              windows:
                - start: "2026-10-17T08:02:24.452Z"
                  end: "2026-10-17T02:45:41-07:00"
            - harness: codex
              id: 01a148e2-90ae-7eb0-8e51-7f557d118b67
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6fa9d67313428b1a35c9c30f44ffbb549985879b80f5cfc31b2601cbfc37c256
              started: "2026-10-17T08:02:50.754Z"
              spawned: false
              bytes: 393000
              windows:
                - start: "2026-10-17T08:02:50.754Z"
                  end: "2026-10-17T02:45:41-07:00"
            - harness: codex
              id: 01a148e7-2483-7b31-8228-d6ebed426406
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: c9dd33e0dfde2d70c72ba1a88b8e941aaca55a0b99026745a8bdedd23be5e3fb
              started: "2026-10-17T08:07:50.283Z"
              spawned: true
              bytes: 930638
              windows:
                - start: "2026-10-17T08:07:50.283Z"
                  end: "2026-10-17T02:45:41-07:00"
        queue_waits:
            - attempt_id: 56ea53f8bb5c486b2893c1a4c559c90b
              slot: ariadne:1
              record_digest: bbe728c9cb16e05f781896337d7de57a7f43b48590c48a05fe0d5ea8a491ed88
              start: "2026-10-17T08:35:47.220135416Z"
              end: "2026-10-17T08:58:47.970797Z"
              queue_seconds: 1380.750661584
              raw_seconds: 1380.750661584
              exact: false
            - attempt_id: a7d113f3e8f9ba1aea28d0c7763c013a
              slot: ariadne:1
              record_digest: a87e3a5eb82e1a266e0a302dc7837817bf2a98304b095b62a5bdadd915b1a502
              start: "2026-10-17T09:16:34.852656042Z"
              end: "2026-10-17T09:36:53.266773Z"
              queue_seconds: 1218.414116958
              raw_seconds: 1218.414116958
              exact: false
        excluded_brain_sessions: 1
        excluded_brain_minutes: 18.58065
      "n": 1
      timestamp: "2026-10-17T02:51:32-07:00"
      dispatched: "2026-10-17T02:46:59-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Snapshot filtering drops sources required to replay queue exclusions
          detail: 'cmd/sdlc/actualmeasurement.go:65-79 retains only sources contributing activity after subtraction, while retaining every QueueWait. When a wait removes all activity from its source and another source contributes work, measurement succeeds but drops the wait''s binding transcript. A temporary regression reproduced successful measurement followed by replay failure: "queue attempt aabb has no owning wait invocation". Retain all queue-binding source prefixes and provenance, including sources contributing zero corrected activity (ARCH-PURPOSE).'
          family: retained-evidence-covers-replay-dependencies
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: A still-running poll finalizes yielded work as completed evidence
          detail: cmd/sdlc/internal/activetime/queue.go:138 preserves running polls only for toolWait. A yielded stage command followed by a poll returning "still running" instead loses its yielded association and reaches result(), which emits a work span without completion evidence. A temporary regression produced an 11-minute work span from an unfinished invocation. Preserve running state for both kinds and emit work only after confirmed completion; cover repeated polls, missing completion and eventual completion through an explicit transition model (ARCH-ORDER, ARCH-SECURE).
          family: work-evidence-requires-confirmed-completion
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: Queue-aware corrections fail to establish calibration trust
          detail: cmd/sdlc/actualinputs.go:353 introduces MeasurementMethodQueueV1, but cmd/sdlc/actualcorrection.go:820 sets WindowTrusted only for MeasurementMethodEngagementV1. Correcting an untrusted sample with ownership-proven queue evidence therefore leaves it excluded from calibration. Substituting QueueV1 into the existing TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust reproduced "correction trust false, want true". Include the new method in this consumer and parameterize the regression across both ownership-proven methods (ARCH-PURPOSE).
          family: measurement-method-consumer-completeness
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 9d6d4e65-fa0b-4d35-a569-6dcab083436f
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c1e66e49e965bbea
            reviewed_head: 1b66ce929c739364bea85940917c7d571fbf5bde
            slot: ariadne:1
            requested_at: "2026-10-17T11:12:57.16443Z"
          phase: finished
          pid: 41574
          started_at: "2026-10-17T11:12:57.377331Z"
          finished_at: "2026-10-17T11:16:38.5981Z"
          session_id: 01a14990-a2f0-7c21-a6c0-ed96d348eab8
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 1b66ce929c739364bea85940917c7d571fbf5bde
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 409e789d83910de366d11b8a536c2c6d8dc9900e
        since: "2026-10-17T00:53:13-07:00"
        until: "2026-10-17T04:11:02-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T00:53:13-07:00"
              end: "2026-10-17T04:11:02-07:00"
              began: "2026-10-17T00:53:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6c6248729d9fa9d99222821fb0a6753768e1599f740bd2ab153d70c86e124d56
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 54803092
              windows:
                - start: "2026-10-17T00:53:13-07:00"
                  end: "2026-10-17T04:11:02-07:00"
            - harness: codex
              id: 01a148dd-d9dc-7b41-8ff5-37fb3d72ae46
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 87ed2689fe0f0571ac966d9c43d9baed2e6b74b7b96fef94d0b83b5589259f28
              started: "2026-10-17T07:57:41.883Z"
              spawned: false
              bytes: 603729
              windows:
                - start: "2026-10-17T07:57:41.883Z"
                  end: "2026-10-17T04:11:02-07:00"
            - harness: codex
              id: 01a148df-e9ee-7061-9d8d-4bd9f106aedf
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f4e79066b008f729a28d08b863f1a8623c12b7287055544ff958dd0ca78bc478
              started: "2026-10-17T07:59:56.888Z"
              spawned: false
              bytes: 450509
              windows:
                - start: "2026-10-17T07:59:56.888Z"
                  end: "2026-10-17T04:11:02-07:00"
            - harness: codex
              id: 01a148e1-33c7-7f21-8981-aa85a762079f
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6e71a79536593227c575c21f9b579b4c1fefab82e87d3e63525cdf0f96f255d0
              started: "2026-10-17T08:01:21.375Z"
              spawned: false
              bytes: 459559
              windows:
                - start: "2026-10-17T08:01:21.375Z"
                  end: "2026-10-17T04:11:02-07:00"
            - harness: codex
              id: 01a148e2-2b49-7230-8d38-9b246eb10ef1
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f9e3d27cc800f5eacab8e808fdf5b07f5ab675fd515d4ae79355dac67d6873e9
              started: "2026-10-17T08:02:24.452Z"
              spawned: false
              bytes: 452872
              windows:
                - start: "2026-10-17T08:02:24.452Z"
                  end: "2026-10-17T04:11:02-07:00"
            - harness: codex
              id: 01a148e2-90ae-7eb0-8e51-7f557d118b67
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6fa9d67313428b1a35c9c30f44ffbb549985879b80f5cfc31b2601cbfc37c256
              started: "2026-10-17T08:02:50.754Z"
              spawned: false
              bytes: 393000
              windows:
                - start: "2026-10-17T08:02:50.754Z"
                  end: "2026-10-17T04:11:02-07:00"
            - harness: codex
              id: 01a148e7-2483-7b31-8228-d6ebed426406
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: c9dd33e0dfde2d70c72ba1a88b8e941aaca55a0b99026745a8bdedd23be5e3fb
              started: "2026-10-17T08:07:50.283Z"
              spawned: true
              bytes: 930638
              windows:
                - start: "2026-10-17T08:07:50.283Z"
                  end: "2026-10-17T04:11:02-07:00"
        queue_waits:
            - attempt_id: 56ea53f8bb5c486b2893c1a4c559c90b
              slot: ariadne:1
              record_digest: bbe728c9cb16e05f781896337d7de57a7f43b48590c48a05fe0d5ea8a491ed88
              start: "2026-10-17T08:35:47.220135416Z"
              end: "2026-10-17T08:58:47.970797Z"
              queue_seconds: 1380.750661584
              raw_seconds: 1380.750661584
              exact: false
            - attempt_id: a7d113f3e8f9ba1aea28d0c7763c013a
              slot: ariadne:1
              record_digest: a87e3a5eb82e1a266e0a302dc7837817bf2a98304b095b62a5bdadd915b1a502
              start: "2026-10-17T09:16:34.852656042Z"
              end: "2026-10-17T09:36:53.266773Z"
              queue_seconds: 1218.414116958
              raw_seconds: 1218.414116958
              exact: false
            - attempt_id: 503a82ece642f2f419ca5b101781061f
              slot: ariadne:1
              record_digest: 4d4f9044ac11be97bb6090fdd6a676685d797480a66df5c3ed7e5d96479d0e50
              start: "2026-10-17T09:59:28.755618042Z"
              end: "2026-10-17T10:22:32.123528Z"
              queue_seconds: 1383.367909958
              raw_seconds: 1383.367909958
              exact: false
            - attempt_id: cf8cc13fe24cabe130e0ee7fc67749d9
              slot: ariadne:1
              record_digest: 90404e1bca67d3455b21fb816152063fa4fe84990fd0caa77c2553d908886276
              start: "2026-10-17T10:40:51.956892583Z"
              end: "2026-10-17T11:02:35.838925Z"
              queue_seconds: 1303.882032417
              raw_seconds: 1303.882032417
              exact: false
        excluded_brain_sessions: 1
        excluded_brain_minutes: 18.58065
        cuts:
            - dispatched: "2026-10-17T02:46:59-07:00"
              finalized: "2026-10-17T02:51:32-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-17T04:16:39-07:00"
      dispatched: "2026-10-17T04:12:57-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: actualmeasurement.go retains QueueBindingSourceIDs alongside contributing sources. TestActualQueueMeasurementRetainsBindingSource passes at head and fails when that retention is removed; it also verifies successful replay.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'queue.go:75-91 and 138-150 still treat any poll result lacking a recognized running marker as completion, deleting the yielded association before emitting work. A temporary pinned-head regression yielded stage work at minute 1, then polled at minute 10: transport-error, empty, and unknown-process responses at minute 11 each produced an 11-minute toolWork span without completion evidence. The running-poll regression passes and fails without its fix, but the required confirmed-completion rule remains unenforced. Preserve unconfirmed outcomes through an explicit transition model and test failed/empty polls followed by missing and eventual terminal results (ARCH-ORDER, ARCH-SECURE, ARCH-PURPOSE).'
          round: 2
        - id: BR-3
          disposition: addressed
          note: actualcorrection.go:820 includes QueueV1 in ownership-proven calibration trust. The parameterized legacy/engagement/queue regression passes at head; removing QueueV1 makes its queue case fail with correction trust false, want true. The production measurement-method consumer sweep found no remaining omission.
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 710c6dda-e9a2-4dbd-9e86-c260fe6244b4
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/dbf706a03bbfb60e
            reviewed_head: 96da94cc8a46984bb2a8265cf991bb93cffb2b0f
            slot: ariadne:1
            requested_at: "2026-10-17T11:49:22.097679Z"
          phase: finished
          pid: 82020
          started_at: "2026-10-17T11:49:22.265368Z"
          finished_at: "2026-10-17T11:52:29.480943Z"
          session_id: 01a149b1-f84e-72b0-ae76-73827379e39e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 96da94cc8a46984bb2a8265cf991bb93cffb2b0f
        tracker_head: beffdc90c4c28c72051d300c3570a00d31384efd
        branch_point: 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774
        since: "2026-10-17T00:53:13-07:00"
        until: "2026-10-17T04:47:53-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T00:53:13-07:00"
              end: "2026-10-17T04:47:53-07:00"
              began: "2026-10-17T00:53:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: fa3299ce46e00ebf6d68bc86d9d13dc47d6e3aea2c9fcc59b110cac0eaab2b1f
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 55578692
              windows:
                - start: "2026-10-17T00:53:13-07:00"
                  end: "2026-10-17T04:47:53-07:00"
            - harness: codex
              id: 01a148dd-d9dc-7b41-8ff5-37fb3d72ae46
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 87ed2689fe0f0571ac966d9c43d9baed2e6b74b7b96fef94d0b83b5589259f28
              started: "2026-10-17T07:57:41.883Z"
              spawned: false
              bytes: 603729
              windows:
                - start: "2026-10-17T07:57:41.883Z"
                  end: "2026-10-17T04:47:53-07:00"
            - harness: codex
              id: 01a148df-e9ee-7061-9d8d-4bd9f106aedf
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f4e79066b008f729a28d08b863f1a8623c12b7287055544ff958dd0ca78bc478
              started: "2026-10-17T07:59:56.888Z"
              spawned: false
              bytes: 450509
              windows:
                - start: "2026-10-17T07:59:56.888Z"
                  end: "2026-10-17T04:47:53-07:00"
            - harness: codex
              id: 01a148e1-33c7-7f21-8981-aa85a762079f
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6e71a79536593227c575c21f9b579b4c1fefab82e87d3e63525cdf0f96f255d0
              started: "2026-10-17T08:01:21.375Z"
              spawned: false
              bytes: 459559
              windows:
                - start: "2026-10-17T08:01:21.375Z"
                  end: "2026-10-17T04:47:53-07:00"
            - harness: codex
              id: 01a148e2-2b49-7230-8d38-9b246eb10ef1
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f9e3d27cc800f5eacab8e808fdf5b07f5ab675fd515d4ae79355dac67d6873e9
              started: "2026-10-17T08:02:24.452Z"
              spawned: false
              bytes: 452872
              windows:
                - start: "2026-10-17T08:02:24.452Z"
                  end: "2026-10-17T04:47:53-07:00"
            - harness: codex
              id: 01a148e2-90ae-7eb0-8e51-7f557d118b67
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 6fa9d67313428b1a35c9c30f44ffbb549985879b80f5cfc31b2601cbfc37c256
              started: "2026-10-17T08:02:50.754Z"
              spawned: false
              bytes: 393000
              windows:
                - start: "2026-10-17T08:02:50.754Z"
                  end: "2026-10-17T04:47:53-07:00"
            - harness: codex
              id: 01a148e7-2483-7b31-8228-d6ebed426406
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: c9dd33e0dfde2d70c72ba1a88b8e941aaca55a0b99026745a8bdedd23be5e3fb
              started: "2026-10-17T08:07:50.283Z"
              spawned: true
              bytes: 930638
              windows:
                - start: "2026-10-17T08:07:50.283Z"
                  end: "2026-10-17T04:47:53-07:00"
        queue_waits:
            - attempt_id: 56ea53f8bb5c486b2893c1a4c559c90b
              slot: ariadne:1
              record_digest: bbe728c9cb16e05f781896337d7de57a7f43b48590c48a05fe0d5ea8a491ed88
              start: "2026-10-17T08:35:47.220135416Z"
              end: "2026-10-17T08:58:47.970797Z"
              queue_seconds: 1380.750661584
              raw_seconds: 1380.750661584
              exact: false
            - attempt_id: a7d113f3e8f9ba1aea28d0c7763c013a
              slot: ariadne:1
              record_digest: a87e3a5eb82e1a266e0a302dc7837817bf2a98304b095b62a5bdadd915b1a502
              start: "2026-10-17T09:16:34.852656042Z"
              end: "2026-10-17T09:36:53.266773Z"
              queue_seconds: 1218.414116958
              raw_seconds: 1218.414116958
              exact: false
            - attempt_id: 503a82ece642f2f419ca5b101781061f
              slot: ariadne:1
              record_digest: 4d4f9044ac11be97bb6090fdd6a676685d797480a66df5c3ed7e5d96479d0e50
              start: "2026-10-17T09:59:28.755618042Z"
              end: "2026-10-17T10:22:32.123528Z"
              queue_seconds: 1383.367909958
              raw_seconds: 1383.367909958
              exact: false
            - attempt_id: cf8cc13fe24cabe130e0ee7fc67749d9
              slot: ariadne:1
              record_digest: 90404e1bca67d3455b21fb816152063fa4fe84990fd0caa77c2553d908886276
              start: "2026-10-17T10:40:51.956892583Z"
              end: "2026-10-17T11:02:35.838925Z"
              queue_seconds: 1303.882032417
              raw_seconds: 1303.882032417
              exact: false
            - attempt_id: a1e58a0cc16a26a44f96927be6753621
              slot: ariadne:1
              record_digest: ffcab54e24b6fd05473b5f75485d0dda06dc8259427cb1e542dca3d87a07fe89
              start: "2026-10-17T11:18:58.330199666Z"
              end: "2026-10-17T11:37:57.218311Z"
              queue_seconds: 1138.888111334
              raw_seconds: 1138.888111334
              exact: false
        excluded_brain_sessions: 1
        excluded_brain_minutes: 18.58065
        cuts:
            - dispatched: "2026-10-17T02:46:59-07:00"
              finalized: "2026-10-17T02:51:32-07:00"
              finalizing: false
            - dispatched: "2026-10-17T04:12:57-07:00"
              finalized: "2026-10-17T04:16:39-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-17T04:52:30-07:00"
      dispatched: "2026-10-17T04:49:22-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: QueueBindingSourceIDs preserves binding transcripts even after all their activity is subtracted. TestActualQueueMeasurementRetainsBindingSource passes and verifies successful replay.
          round: 3
        - id: BR-2
          disposition: not-addressed
          note: 'cmd/sdlc/internal/activetime/queue.go:133-142 registers polls without duplicate-ID validation and distinguishes polls from original results by comparing call IDs. A temporary pinned-head regression yielded stage command "stage" into session 7, then polled that session using call_id "stage" with output "unknown process". The parser emitted an 11-minute toolWork span without completion evidence. Existing unconfirmed-poll tests pass at head and fail against the pre-fix implementation, but omit this identity collision. This repeats family work-evidence-requires-confirmed-completion. Enforce the rule across every call registration: identities must be unique, poll status must be explicit, and yielded work may complete only on confirmed terminal evidence. Implement the requested transition model and cover duplicate IDs, concurrent poll aliases, and late results rather than adding another instance-specific guard (ARCH-ORDER, ARCH-SECURE, ARCH-PURPOSE).'
          round: 3
        - id: BR-3
          disposition: addressed
          note: actualcorrection.go includes QueueV1 in ownership-proven calibration trust. The parameterized legacy, engagement, and queue correction-trust regression passes; the production method-consumer sweep found no remaining omission.
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4311457e-973d-4b8e-bd74-ef51ce51e9db
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c22bf0eed05685f4
            reviewed_head: 3290d6f0f592307fb36466b54a2e112bc81f1517
            slot: ariadne:1
            requested_at: "2026-10-17T13:02:39.137015Z"
          phase: finished
          pid: 45750
          started_at: "2026-10-17T13:02:39.349373Z"
          finished_at: "2026-10-17T13:07:22.532649Z"
          session_id: 01a149f5-1100-76f3-aed9-894eae2881cc
          outcome: returned
      "n": 4
      timestamp: "2026-10-17T06:07:23-07:00"
      dispatched: "2026-10-17T06:02:39-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: TestActualQueueMeasurementRetainsBindingSource passes; temporarily removing binding-source retention makes it fail because the zero-contribution source disappears.
          round: 4
        - id: BR-2
          disposition: not-addressed
          note: The poll-based completion path was removed, but the confirmed-completion rule remains violated at cmd/sdlc/internal/activetime/queue.go:135. Claude tool_result content is passed as raw JSON, and is_error is ignored. Temporary regressions with empty, unknown-process and transport-error content all emit completed work. An end-to-end fixture with is_error:true and content:"unknown process" restores 11 minutes during a recorded 60-minute queue instead of refusing measurement. This remains family work-evidence-requires-confirmed-completion (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER).
          round: 4
        - id: BR-3
          disposition: addressed
          note: TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust passes; temporarily removing queue-v1 from the trust condition makes its queue case fail with false trust.
          round: 4
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#420 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T02:51:32-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `retained-evidence-covers-replay-dependencies` (kind: defect) Snapshot filtering drops sources required to replay queue exclusions
  cmd/sdlc/actualmeasurement.go:65-79 retains only sources contributing activity after subtraction, while retaining every QueueWait. When a wait removes all activity from its source and another source contributes work, measurement succeeds but drops the wait's binding transcript. A temporary regression reproduced successful measurement followed by replay failure: "queue attempt aabb has no owning wait invocation". Retain all queue-binding source prefixes and provenance, including sources contributing zero corrected activity (ARCH-PURPOSE).
- **BR-2** [Critical] `work-evidence-requires-confirmed-completion` (kind: defect) A still-running poll finalizes yielded work as completed evidence
  cmd/sdlc/internal/activetime/queue.go:138 preserves running polls only for toolWait. A yielded stage command followed by a poll returning "still running" instead loses its yielded association and reaches result(), which emits a work span without completion evidence. A temporary regression produced an 11-minute work span from an unfinished invocation. Preserve running state for both kinds and emit work only after confirmed completion; cover repeated polls, missing completion and eventual completion through an explicit transition model (ARCH-ORDER, ARCH-SECURE).
- **BR-3** [Important] `measurement-method-consumer-completeness` (kind: defect) Queue-aware corrections fail to establish calibration trust
  cmd/sdlc/actualinputs.go:353 introduces MeasurementMethodQueueV1, but cmd/sdlc/actualcorrection.go:820 sets WindowTrusted only for MeasurementMethodEngagementV1. Correcting an untrusted sample with ownership-proven queue evidence therefore leaves it excluded from calibration. Substituting QueueV1 into the existing TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust reproduced "correction trust false, want true". Include the new method in this consumer and parameterize the regression across both ownership-proven methods (ARCH-PURPOSE).

## Round 2 — 2026-10-17T04:16:39-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — actualmeasurement.go retains QueueBindingSourceIDs alongside contributing sources. TestActualQueueMeasurementRetainsBindingSource passes at head and fails when that retention is removed; it also verifies successful replay.
- BR-2 — not-addressed — queue.go:75-91 and 138-150 still treat any poll result lacking a recognized running marker as completion, deleting the yielded association before emitting work. A temporary pinned-head regression yielded stage work at minute 1, then polled at minute 10: transport-error, empty, and unknown-process responses at minute 11 each produced an 11-minute toolWork span without completion evidence. The running-poll regression passes and fails without its fix, but the required confirmed-completion rule remains unenforced. Preserve unconfirmed outcomes through an explicit transition model and test failed/empty polls followed by missing and eventual terminal results (ARCH-ORDER, ARCH-SECURE, ARCH-PURPOSE).
- BR-3 — addressed — actualcorrection.go:820 includes QueueV1 in ownership-proven calibration trust. The parameterized legacy/engagement/queue regression passes at head; removing QueueV1 makes its queue case fail with correction trust false, want true. The production measurement-method consumer sweep found no remaining omission.

## Round 3 — 2026-10-17T04:52:30-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — QueueBindingSourceIDs preserves binding transcripts even after all their activity is subtracted. TestActualQueueMeasurementRetainsBindingSource passes and verifies successful replay.
- BR-2 — not-addressed — cmd/sdlc/internal/activetime/queue.go:133-142 registers polls without duplicate-ID validation and distinguishes polls from original results by comparing call IDs. A temporary pinned-head regression yielded stage command "stage" into session 7, then polled that session using call_id "stage" with output "unknown process". The parser emitted an 11-minute toolWork span without completion evidence. Existing unconfirmed-poll tests pass at head and fail against the pre-fix implementation, but omit this identity collision. This repeats family work-evidence-requires-confirmed-completion. Enforce the rule across every call registration: identities must be unique, poll status must be explicit, and yielded work may complete only on confirmed terminal evidence. Implement the requested transition model and cover duplicate IDs, concurrent poll aliases, and late results rather than adding another instance-specific guard (ARCH-ORDER, ARCH-SECURE, ARCH-PURPOSE).
- BR-3 — addressed — actualcorrection.go includes QueueV1 in ownership-proven calibration trust. The parameterized legacy, engagement, and queue correction-trust regression passes; the production method-consumer sweep found no remaining omission.

## Round 4 — 2026-10-17T06:07:23-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — TestActualQueueMeasurementRetainsBindingSource passes; temporarily removing binding-source retention makes it fail because the zero-contribution source disappears.
- BR-2 — not-addressed — The poll-based completion path was removed, but the confirmed-completion rule remains violated at cmd/sdlc/internal/activetime/queue.go:135. Claude tool_result content is passed as raw JSON, and is_error is ignored. Temporary regressions with empty, unknown-process and transport-error content all emit completed work. An end-to-end fixture with is_error:true and content:"unknown process" restores 11 minutes during a recorded 60-minute queue instead of refusing measurement. This remains family work-evidence-requires-confirmed-completion (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER).
- BR-3 — addressed — TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust passes; temporarily removing queue-v1 from the trust condition makes its queue case fail with false trust.

## Open findings

- **BR-2** [Critical] `work-evidence-requires-confirmed-completion` (kind: defect) A still-running poll finalizes yielded work as completed evidence
