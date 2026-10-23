---
gate: boundary-review
issue: 554
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 52dbebc9-6cf2-40fd-b08d-a3def93c9340
            repository: github.com/xianxu/ariadne
            issue_id: "000554"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/cd3daa1a45a0f36e
            reviewed_head: bf6e38e1ebc285ef74315d90849b2abe2a485db6
            slot: ariadne:4
            requested_at: "2026-10-22T18:36:35.094779Z"
          phase: finished
          pid: 87965
          started_at: "2026-10-22T18:36:35.340666Z"
          finished_at: "2026-10-22T18:39:13.288218Z"
          session_id: 01a164e6-971a-7a12-804f-9c1433a5709e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: bf6e38e1ebc285ef74315d90849b2abe2a485db6
        tracker_head: 9c7687fc6facb9b1f3b2207e41998b06238a60f4
        branch_point: 5f961bcf88d0196ccc33109543d4cb6a5b49361c
        since: "2026-10-22T11:00:55-07:00"
        until: "2026-10-22T11:34:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T11:00:55-07:00"
              end: "2026-10-22T11:34:55-07:00"
              began: "2026-10-22T11:00:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 582fd7aa00ff2de0512c96302a9884b9b8ce63731c2ddb9b75ac6ba718b59b2a
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 34900462
              windows:
                - start: "2026-10-22T11:00:55-07:00"
                  end: "2026-10-22T11:34:55-07:00"
            - harness: codex
              id: 01a164c6-5399-7210-85eb-6fdf6536cdc0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 4eb5a3457f12af4892aaace0b7cd7bbaa379e92c07e80ea4b5811d95b25ba217
              started: "2026-10-22T18:01:21.642Z"
              spawned: true
              bytes: 1213140
              windows:
                - start: "2026-10-22T18:01:21.642Z"
                  end: "2026-10-22T11:34:55-07:00"
      "n": 1
      timestamp: "2026-10-22T11:39:14-07:00"
      dispatched: "2026-10-22T11:36:35-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      recipe: small-diff-review
      isolation: fenced
      reviewed: bf6e38e1ebc285ef74315d90849b2abe2a485db6
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8a68198c-8773-49e3-9fe9-f3629478aab3
            repository: github.com/xianxu/ariadne
            issue_id: "000554"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/bcd26a9fac12e5e3
            reviewed_head: c7778d78873aaa2a2e61e48ef295c6fa181df6bd
            slot: ariadne:4
            requested_at: "2026-10-22T19:43:14.663186Z"
          phase: finished
          pid: 44074
          started_at: "2026-10-22T19:43:15.016283Z"
          finished_at: "2026-10-22T19:44:47.383643Z"
          session_id: 01a16523-a053-7272-ba1e-225b4b43c553
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c7778d78873aaa2a2e61e48ef295c6fa181df6bd
        tracker_head: 376ce12f4b5e1f240c98e2459f28c4ecf7196b97
        branch_point: 5785330cda6bf31d533e4282aecc8f3f9cd632a7
        since: "2026-10-22T11:00:55-07:00"
        until: "2026-10-22T12:41:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T11:00:55-07:00"
              end: "2026-10-22T12:41:18-07:00"
              began: "2026-10-22T11:00:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ca8f6eac53500b2a893e92cd1afec198d717fa9c927e25132b8be6a5d3679b59
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 36299414
              windows:
                - start: "2026-10-22T11:00:55-07:00"
                  end: "2026-10-22T12:41:18-07:00"
            - harness: codex
              id: 01a164c6-5399-7210-85eb-6fdf6536cdc0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef43bad86de970676d085847c8fc9c82fef1c645748abcad50541d330dc690bb
              started: "2026-10-22T18:01:21.642Z"
              spawned: true
              bytes: 1547287
              windows:
                - start: "2026-10-22T18:01:21.642Z"
                  end: "2026-10-22T12:41:18-07:00"
        cuts:
            - dispatched: "2026-10-22T11:36:35-07:00"
              finalized: "2026-10-22T11:39:14-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-22T12:44:48-07:00"
      dispatched: "2026-10-22T12:43:14-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: scope changed after close — run a full `sdlc close`
          detail: workshop/plans/000554-cutover-ownership-recovery-plan.md:79-88 and the issue's new Revisions passage authorize first-claim queue clipping, implemented at cmd/sdlc/actualinputs.go:413-414. No prior finding requested this behavior; it changes which queue evidence hybrid measurement accepts. The fix-only review contract explicitly requires a full sdlc close for added behavior or expanded scope, regardless of the recorded approval.
          family: post-close-scope-review
          round: 2
      recipe: fix-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: de1e85c6-61e6-47ca-8f93-ef0153b3ebde
            repository: github.com/xianxu/ariadne
            issue_id: "000554"
            recipe: small-diff-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2b5b4d29122c0b1a
            reviewed_head: 1b6c0480fcba1f208dd088203bc4ca40fd0324eb
            slot: ariadne:4
            requested_at: "2026-10-22T19:46:45.658987Z"
          phase: finished
          pid: 42319
          started_at: "2026-10-22T19:46:45.86068Z"
          finished_at: "2026-10-22T19:50:35.588632Z"
          session_id: 01a16526-d590-7750-8669-039ea2b9deeb
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1b6c0480fcba1f208dd088203bc4ca40fd0324eb
        tracker_head: 376ce12f4b5e1f240c98e2459f28c4ecf7196b97
        branch_point: 5785330cda6bf31d533e4282aecc8f3f9cd632a7
        since: "2026-10-22T11:00:55-07:00"
        until: "2026-10-22T12:46:00-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T11:00:55-07:00"
              end: "2026-10-22T12:46:00-07:00"
              began: "2026-10-22T11:00:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1bd7cf2bfdd235aa15e0ba8b0f488c918e27e069639f37c0dc5df0f96ad5aa99
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 36464759
              windows:
                - start: "2026-10-22T11:00:55-07:00"
                  end: "2026-10-22T12:46:00-07:00"
            - harness: codex
              id: 01a164c6-5399-7210-85eb-6fdf6536cdc0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef43bad86de970676d085847c8fc9c82fef1c645748abcad50541d330dc690bb
              started: "2026-10-22T18:01:21.642Z"
              spawned: true
              bytes: 1547287
              windows:
                - start: "2026-10-22T18:01:21.642Z"
                  end: "2026-10-22T12:46:00-07:00"
        cuts:
            - dispatched: "2026-10-22T11:36:35-07:00"
              finalized: "2026-10-22T11:39:14-07:00"
              finalizing: true
            - dispatched: "2026-10-22T12:43:14-07:00"
              finalized: "2026-10-22T12:44:48-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-22T12:50:36-07:00"
      dispatched: "2026-10-22T12:46:45-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: This full pinned-range close review covers the added queue behavior. The plan's Revisions passage at lines 93–95 explicitly supersedes the fix-only route. The queue regression independently fails without the attachment-bound fix and passes with it.
          round: 3
      recipe: small-diff-review
      isolation: fenced
      reviewed: 1b6c0480fcba1f208dd088203bc4ca40fd0324eb
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3dc10227-04ff-4f46-85f3-803895d7158c
            repository: github.com/xianxu/ariadne
            issue_id: "000554"
            recipe: small-diff-review
            round: 4
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/979af607b013255c
            reviewed_head: e01e95a065207a21c29d8e74465f5e7aa69f0448
            slot: ariadne:4
            requested_at: "2026-10-22T22:46:48.824316Z"
          phase: finished
          pid: 40402
          started_at: "2026-10-22T22:46:49.021123Z"
          finished_at: "2026-10-22T22:49:19.010875Z"
          session_id: 01a165cb-ad3b-7201-b3a1-33a708c2b0aa
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e01e95a065207a21c29d8e74465f5e7aa69f0448
        tracker_head: 3cf990c135d6768b4f6d8a302340f27f9a119073
        branch_point: 5785330cda6bf31d533e4282aecc8f3f9cd632a7
        since: "2026-10-22T11:00:55-07:00"
        until: "2026-10-22T15:45:56-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T11:00:55-07:00"
              end: "2026-10-22T15:45:56-07:00"
              began: "2026-10-22T11:00:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1177c96b8170e056cbb9bef47ca887ee4a8f46572accca41e6c2e920fa12e38b
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 37113604
              windows:
                - start: "2026-10-22T11:00:55-07:00"
                  end: "2026-10-22T15:45:56-07:00"
            - harness: codex
              id: 01a164c6-5399-7210-85eb-6fdf6536cdc0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef43bad86de970676d085847c8fc9c82fef1c645748abcad50541d330dc690bb
              started: "2026-10-22T18:01:21.642Z"
              spawned: true
              bytes: 1547287
              windows:
                - start: "2026-10-22T18:01:21.642Z"
                  end: "2026-10-22T15:45:56-07:00"
        queue_waits:
            - attempt_id: 87523898ae0044a31e8087b679815bc0
              slot: ariadne:4
              record_digest: 41ef79ce0a331fd81f4fadfdf577cf6965a28fc1868fe3268784f9dfdaf9e612
              start: "2026-10-22T19:41:46.309728Z"
              end: "2026-10-22T22:17:47.911375Z"
              queue_seconds: 9361.601647
              raw_seconds: 9361.601647
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 79.82053333333333
        cuts:
            - dispatched: "2026-10-22T11:36:35-07:00"
              finalized: "2026-10-22T11:39:14-07:00"
              finalizing: true
            - dispatched: "2026-10-22T12:43:14-07:00"
              finalized: "2026-10-22T12:44:48-07:00"
              finalizing: false
            - dispatched: "2026-10-22T12:46:45-07:00"
              finalized: "2026-10-22T12:50:36-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-22T15:49:20-07:00"
      dispatched: "2026-10-22T15:46:48-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan lines 93–95 explicitly supersede the fix-only route; the gate ledger records a full-range review, and this review independently covers the expanded queue behavior. Removing the attachment bound reproduces failures in both native harness fixtures.
          round: 4
      recipe: small-diff-review
      isolation: fenced
      reviewed: e01e95a065207a21c29d8e74465f5e7aa69f0448
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 172d77dd-ffb6-4d7e-b1df-42587a8dc187
            repository: github.com/xianxu/ariadne
            issue_id: "000554"
            recipe: fix-review
            round: 5
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/cf5a13c312e6645c
            reviewed_head: 1cfdc109eade72a3a085caf00e81e6a1f5f08393
            slot: ariadne:4
            requested_at: "2026-10-23T03:15:58.220798Z"
          phase: finished
          pid: 45805
          started_at: "2026-10-23T03:15:58.381979Z"
          finished_at: "2026-10-23T03:17:41.121903Z"
          session_id: 01a166c2-18d2-75b3-8f2d-589009b3bf80
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 1cfdc109eade72a3a085caf00e81e6a1f5f08393
        tracker_head: fb50153e559a94941f7d0e02bad3ed6301571dee
        branch_point: 36544319c0c14a714e70f729ff73d21e8dab0df6
        since: "2026-10-22T11:00:55-07:00"
        until: "2026-10-22T19:56:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T11:00:55-07:00"
              end: "2026-10-22T19:56:10-07:00"
              began: "2026-10-22T11:00:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: bb160eb361947deb51ff84e73df4c7aafda3e5acf4e6bb787452864668147aab
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 39129036
              windows:
                - start: "2026-10-22T11:00:55-07:00"
                  end: "2026-10-22T19:56:10-07:00"
            - harness: codex
              id: 01a164c6-5399-7210-85eb-6fdf6536cdc0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef43bad86de970676d085847c8fc9c82fef1c645748abcad50541d330dc690bb
              started: "2026-10-22T18:01:21.642Z"
              spawned: true
              bytes: 1547287
              windows:
                - start: "2026-10-22T18:01:21.642Z"
                  end: "2026-10-22T19:56:10-07:00"
        queue_waits:
            - attempt_id: 87523898ae0044a31e8087b679815bc0
              slot: ariadne:4
              record_digest: 41ef79ce0a331fd81f4fadfdf577cf6965a28fc1868fe3268784f9dfdaf9e612
              start: "2026-10-22T19:41:46.309728Z"
              end: "2026-10-22T22:17:47.911375Z"
              queue_seconds: 9361.601647
              raw_seconds: 9361.601647
              exact: true
            - attempt_id: 922ce1926f6f91ef863dfe36fade0f14
              slot: ariadne:4
              record_digest: 80011563a9851f51b59035c124bfad8e8e12666fa70e5f184730710809682d91
              start: "2026-10-22T22:46:24.332771Z"
              end: "2026-10-23T00:51:15.224565Z"
              queue_seconds: 7490.891794
              raw_seconds: 7490.891794
              exact: true
            - attempt_id: 5cb73e518d2822c92d0a5eca4d38ff95
              slot: ariadne:4
              record_digest: bc2bfa8ab865407a5c6a3a8cd2c333ab3f66bd4d151dc8de62956c01e7be5326
              start: "2026-10-23T02:55:32.460524Z"
              end: "2026-10-23T02:56:10Z"
              queue_seconds: 37.539476
              raw_seconds: 284.208676
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 196.21703333333335
        cuts:
            - dispatched: "2026-10-22T11:36:35-07:00"
              finalized: "2026-10-22T11:39:14-07:00"
              finalizing: true
            - dispatched: "2026-10-22T12:43:14-07:00"
              finalized: "2026-10-22T12:44:48-07:00"
              finalizing: false
            - dispatched: "2026-10-22T12:46:45-07:00"
              finalized: "2026-10-22T12:50:36-07:00"
              finalizing: true
            - dispatched: "2026-10-22T15:46:48-07:00"
              finalized: "2026-10-22T15:49:20-07:00"
              finalizing: true
      "n": 5
      timestamp: "2026-10-22T20:17:42-07:00"
      dispatched: "2026-10-22T20:15:58-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan lines 93–95 supersede the fix-only route, and the retained full-range close review covers the expanded queue contract. This interdiff adds no further behavior.
          round: 5
      recipe: fix-review
      isolation: fenced
      reviewed: 1cfdc109eade72a3a085caf00e81e6a1f5f08393
      blocked: false
---

# Gate ledger — ariadne#554 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T11:39:14-07:00 (codex) — passed

## Round 2 — 2026-10-22T12:44:48-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `post-close-scope-review` (kind: defect) scope changed after close — run a full `sdlc close`
  workshop/plans/000554-cutover-ownership-recovery-plan.md:79-88 and the issue's new Revisions passage authorize first-claim queue clipping, implemented at cmd/sdlc/actualinputs.go:413-414. No prior finding requested this behavior; it changes which queue evidence hybrid measurement accepts. The fix-only review contract explicitly requires a full sdlc close for added behavior or expanded scope, regardless of the recorded approval.

## Round 3 — 2026-10-22T12:50:36-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — This full pinned-range close review covers the added queue behavior. The plan's Revisions passage at lines 93–95 explicitly supersedes the fix-only route. The queue regression independently fails without the attachment-bound fix and passes with it.

## Round 4 — 2026-10-22T15:49:20-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Plan lines 93–95 explicitly supersede the fix-only route; the gate ledger records a full-range review, and this review independently covers the expanded queue behavior. Removing the attachment bound reproduces failures in both native harness fixtures.

## Round 5 — 2026-10-22T20:17:42-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Plan lines 93–95 supersede the fix-only route, and the retained full-range close review covers the expanded queue contract. This interdiff adds no further behavior.

## Open findings

(none — every finding has been disposed)
