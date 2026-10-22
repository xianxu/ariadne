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

## Open findings

(none — every finding has been disposed)
