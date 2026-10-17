---
gate: boundary-review
issue: 421
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a1d6bdb4-d325-4746-9420-a9353b50e024
            repository: github.com/xianxu/ariadne
            issue_id: "000421"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c486143584f05acd
            reviewed_head: 8ffb676c1b1458bc3cfbbb07f88bca5f294718a5
            slot: ariadne:2
            requested_at: "2026-10-17T08:38:46.976885Z"
          phase: finished
          pid: 4814
          started_at: "2026-10-17T08:38:47.190071Z"
          finished_at: "2026-10-17T08:44:40.384619Z"
          session_id: 27fb9441-7c8c-4354-af52-9b05f31956e8
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8ffb676c1b1458bc3cfbbb07f88bca5f294718a5
        tracker_head: 70eec81e0f4fa42a2475dfe4ad6f6c9d81bbd487
        branch_point: 3dd288111bc10a361618e7f167abb864bd687b63
        since: "2026-10-16T22:20:44-07:00"
        until: "2026-10-17T01:38:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T22:20:44-07:00"
              end: "2026-10-17T01:38:16-07:00"
              began: "2026-10-16T22:20:44-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3f10042a7e53a0f33ef375f969d8d83cd7a5a9b43d568796f89a44380904d1d7
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 21471084
              windows:
                - start: "2026-10-16T22:20:44-07:00"
                  end: "2026-10-17T01:38:16-07:00"
        excluded_brain_sessions: 2
        excluded_brain_minutes: 102.94318333333334
      "n": 1
      timestamp: "2026-10-17T01:44:42-07:00"
      dispatched: "2026-10-17T01:38:46-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: runStage leaves the new holder's lease on the target when git switch fails, yet reports nothing was staged
          detail: 'stage.go:94 writes the holder''s lease before the switch at :98. A switch error returns "nothing was staged" without restoring the prior lease, so an unchanged target stays held and every other issue''s stage, clear and move refuses. Fix: on switch failure, restore `held` (or remove the record when the slot had none) and add a test that forces the switch to fail.'
          family: failure-path-rollback-of-pre-written-state
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: README update missing for the stage lease refusal and the new --abandoned flag
          detail: README.md:17-24 (Concurrent issue work) documents sdlc stage --build and --clear but not one stage at a time, the refusal naming the holder, or `sdlc stage --clear :0 --abandoned <holder>`. The atlas and --help text were updated; README was not.
          family: readme-missing-new-surface
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: stage --clear --dry-run on a resting target with a held lease prints nothing staged
          detail: 'stage.go:175-187: a real run releases the lease, but the dry run says nothing is staged and doesn''t mention the lease it would release.'
          family: dry-run-report-diverges-from-real-run
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: runUnstage observes the target before taking the stage lock and never re-reads it under the lock
          detail: runStage and runMove re-observe under the lock; runUnstage relies on `to` from before the lock. The lease read under the lock blocks cross-holder races, so the remaining risk is small.
          family: observe-outside-lock
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2b986523-e6d8-482f-abae-cc81d1f6a0e9
            repository: github.com/xianxu/ariadne
            issue_id: "000421"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c2a9a885c9fb8914
            reviewed_head: 767320e0dc23de298cf54b1951404f7efd6f2863
            slot: ariadne:2
            requested_at: "2026-10-17T08:50:03.668699Z"
          phase: finished
          pid: 3284
          started_at: "2026-10-17T08:50:03.787787Z"
          finished_at: "2026-10-17T08:59:04.942463Z"
          session_id: 254a28b6-ee42-445c-8f38-29ef6048e1e0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 767320e0dc23de298cf54b1951404f7efd6f2863
        tracker_head: 82e6334f5978799a7de87c120496656cf09cfdb6
        branch_point: 3dd288111bc10a361618e7f167abb864bd687b63
        since: "2026-10-16T22:20:44-07:00"
        until: "2026-10-17T01:49:35-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T22:20:44-07:00"
              end: "2026-10-17T01:49:35-07:00"
              began: "2026-10-16T22:20:44-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: a89d43514331d6b9529b5cf8d99ad25d4b9a641ca461e2323394af9a05cf62a6
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 21682015
              windows:
                - start: "2026-10-16T22:20:44-07:00"
                  end: "2026-10-17T01:49:35-07:00"
        excluded_brain_sessions: 2
        excluded_brain_minutes: 102.94318333333334
        cuts:
            - dispatched: "2026-10-17T01:38:46-07:00"
              finalized: "2026-10-17T01:44:42-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-17T01:59:06-07:00"
      dispatched: "2026-10-17T01:50:03-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: restoreStageLease at stage.go:99-104; reverting it makes TestStageLeaseSwitchFailureRestores fail (lease Holder:#1 left behind); sibling post-switch failure paths correctly keep the lease.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md Concurrent issue work now names one stage at a time, the refusal naming the holder, and sdlc stage --clear :0 --abandoned <holder>; matches checkLease/checkClearLease behavior.
          round: 2
        - id: BR-3
          disposition: addressed
          note: stage.go:185-188 prints Would release; disabling it makes TestStageLeaseEdges fail with nothing staged output.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: Code reorder is correct (lock, re-resolve, observe at stage.go:145-158) but no test fails without it; add one holding LockStage on a dirty :0 and asserting the lock refusal, not the dirty-target error.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 767320e0dc23de298cf54b1951404f7efd6f2863
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: fe33d539-e8e9-4c72-80d1-26a28e222412
            repository: github.com/xianxu/ariadne
            issue_id: "000421"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/830168282ae0a492
            reviewed_head: 3c3595670b25cdd17754154983e3634a6c603d3d
            slot: ariadne:2
            requested_at: "2026-10-17T09:03:34.613354Z"
          phase: finished
          pid: 11268
          started_at: "2026-10-17T09:03:34.824576Z"
          finished_at: "2026-10-17T09:05:45.841523Z"
          session_id: e896fdae-4fba-490d-9cea-1f7abd6b927f
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 3c3595670b25cdd17754154983e3634a6c603d3d
        tracker_head: fd36c69198cd3279b116e3c3d46f3aa54ab0324c
        branch_point: 3dd288111bc10a361618e7f167abb864bd687b63
        since: "2026-10-16T22:20:44-07:00"
        until: "2026-10-17T02:03:07-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T22:20:44-07:00"
              end: "2026-10-17T02:03:07-07:00"
              began: "2026-10-16T22:20:44-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6e767767347ea9466fd9ecba3bd04dbc10efc6795a726185f454ae47cd2aa701
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 21752383
              windows:
                - start: "2026-10-16T22:20:44-07:00"
                  end: "2026-10-17T02:03:07-07:00"
        excluded_brain_sessions: 2
        excluded_brain_minutes: 102.94318333333334
        cuts:
            - dispatched: "2026-10-17T01:38:46-07:00"
              finalized: "2026-10-17T01:44:42-07:00"
              finalizing: false
            - dispatched: "2026-10-17T01:50:03-07:00"
              finalized: "2026-10-17T01:59:06-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-17T02:05:46-07:00"
      dispatched: "2026-10-17T02:03:34-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: runUnstage locks, re-resolves and then observes (stage.go:146-159). The new TestStageLeaseEdges assertion (stage_test.go:385-390) fails when the observation is moved in front of the lock ("sample:0 is on elsewhere, not a staged commit; nothing to clear") and passes with the fix.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 3c3595670b25cdd17754154983e3634a6c603d3d
      blocked: false
---

# Gate ledger — ariadne#421 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T01:44:42-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `failure-path-rollback-of-pre-written-state` (kind: defect) runStage leaves the new holder's lease on the target when git switch fails, yet reports nothing was staged
  stage.go:94 writes the holder's lease before the switch at :98. A switch error returns "nothing was staged" without restoring the prior lease, so an unchanged target stays held and every other issue's stage, clear and move refuses. Fix: on switch failure, restore `held` (or remove the record when the slot had none) and add a test that forces the switch to fail.
- **BR-2** [Important] `readme-missing-new-surface` (kind: defect) README update missing for the stage lease refusal and the new --abandoned flag
  README.md:17-24 (Concurrent issue work) documents sdlc stage --build and --clear but not one stage at a time, the refusal naming the holder, or `sdlc stage --clear :0 --abandoned <holder>`. The atlas and --help text were updated; README was not.
- **BR-3** [Minor] `dry-run-report-diverges-from-real-run` (kind: defect) stage --clear --dry-run on a resting target with a held lease prints nothing staged
  stage.go:175-187: a real run releases the lease, but the dry run says nothing is staged and doesn't mention the lease it would release.
- **BR-4** [Minor] `observe-outside-lock` (kind: defect) runUnstage observes the target before taking the stage lock and never re-reads it under the lock
  runStage and runMove re-observe under the lock; runUnstage relies on `to` from before the lock. The lease read under the lock blocks cross-holder races, so the remaining risk is small.

## Round 2 — 2026-10-17T01:59:06-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — restoreStageLease at stage.go:99-104; reverting it makes TestStageLeaseSwitchFailureRestores fail (lease Holder:#1 left behind); sibling post-switch failure paths correctly keep the lease.
- BR-2 — addressed — README.md Concurrent issue work now names one stage at a time, the refusal naming the holder, and sdlc stage --clear :0 --abandoned <holder>; matches checkLease/checkClearLease behavior.
- BR-3 — addressed — stage.go:185-188 prints Would release; disabling it makes TestStageLeaseEdges fail with nothing staged output.
- BR-4 — not-addressed — Code reorder is correct (lock, re-resolve, observe at stage.go:145-158) but no test fails without it; add one holding LockStage on a dirty :0 and asserting the lock refusal, not the dirty-target error.

## Round 3 — 2026-10-17T02:05:46-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — runUnstage locks, re-resolves and then observes (stage.go:146-159). The new TestStageLeaseEdges assertion (stage_test.go:385-390) fails when the observation is moved in front of the lock ("sample:0 is on elsewhere, not a staged commit; nothing to clear") and passes with the fix.

## Open findings

(none — every finding has been disposed)
