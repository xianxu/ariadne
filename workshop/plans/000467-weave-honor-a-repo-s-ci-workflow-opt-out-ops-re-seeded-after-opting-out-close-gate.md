---
gate: boundary-review
issue: 467
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c42da31e-714f-417a-a2e6-912695ffe2cd
            repository: github.com/xianxu/ariadne
            issue_id: "000467"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/0d420d9deab70d8f
            reviewed_head: e4ee6277a482202da97576459c5847c1de49eaa3
            slot: ariadne:5
            requested_at: "2026-10-19T17:08:29.005976Z"
          phase: finished
          pid: 50974
          started_at: "2026-10-19T17:08:29.224535Z"
          finished_at: "2026-10-19T17:16:43.564066Z"
          session_id: 8cbfc921-a127-4421-93f0-2b3e2bde87b8
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: e4ee6277a482202da97576459c5847c1de49eaa3
        tracker_head: 49ef52a131f6a3a4626ecdc96dda5a9719b99fbb
        branch_point: 931a4ad6898a9883e1324701e92860bca401f2bc
        since: "2026-10-19T10:00:05-07:00"
        until: "2026-10-19T10:07:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:00:05-07:00"
              end: "2026-10-19T10:07:52-07:00"
              began: "2026-10-19T10:00:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 2b1c35f1-9bee-4064-83d0-2a67ebaaf586
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 2bd7e323c9afec4a3a4df4756216af4328250c20a682f81001356fc4700855db
              started: "2026-10-19T16:59:53.709Z"
              spawned: false
              bytes: 3038888
              windows:
                - start: "2026-10-19T10:00:05-07:00"
                  end: "2026-10-19T10:07:52-07:00"
      "n": 1
      timestamp: "2026-10-19T10:16:46-07:00"
      dispatched: "2026-10-19T10:08:29-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: Decline removes and reports ancestor internal rows that never participate in the leaf
          detail: layer.Decline filters every ancestor row regardless of intent.Selected, so a decline over only internal ancestor rows reports "opted out of N row(s)" and suppresses the stale-decline warning. Skip rows where !intent.Selected(in.Visibility, false).
          family: decline-ignores-visibility-selection
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: A decline row with two path arguments silently declines the second one
          detail: ParseManifest sets Target to fields[2], and Decline matches on Target, so `decline a b` declines b, not a. Warn on or reject a second argument for decline.
          family: manifest-row-arity-unchecked
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Plan docs step says construct/manifest.md was updated, but it is not in the diff
          detail: manifest.md does not list verbs, so nothing is missing from the docs. Correct the plan wording in a Revisions entry.
          family: plan-claims-undelivered-doc
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: e4ee6277a482202da97576459c5847c1de49eaa3
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 578d9075-a485-4c8b-b692-6f61abf8b221
            repository: github.com/xianxu/ariadne
            issue_id: "000467"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4ee93fd71497e996
            reviewed_head: c3fca0ece7d29f05d3b014a3030590748cd7de60
            slot: ariadne:5
            requested_at: "2026-10-19T18:44:44.442076Z"
          phase: finished
          pid: 81150
          started_at: "2026-10-19T18:44:44.539006Z"
          finished_at: "2026-10-19T18:46:37.282221Z"
          session_id: 13159420-1ebd-4f26-b816-ad2179142b7f
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c3fca0ece7d29f05d3b014a3030590748cd7de60
        tracker_head: 5ab2120b88b7b3d614471399ee21ee2d07acb9d0
        branch_point: 931a4ad6898a9883e1324701e92860bca401f2bc
        since: "2026-10-19T10:00:05-07:00"
        until: "2026-10-19T10:17:35-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:00:05-07:00"
              end: "2026-10-19T10:17:35-07:00"
              began: "2026-10-19T10:00:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 2b1c35f1-9bee-4064-83d0-2a67ebaaf586
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 54c301439b80ed805ddd72b3470bbd778f8fd6ec77d1c88fe67be9a485376555
              started: "2026-10-19T16:59:53.709Z"
              spawned: false
              bytes: 4199678
              windows:
                - start: "2026-10-19T10:00:05-07:00"
                  end: "2026-10-19T10:17:35-07:00"
        cuts:
            - dispatched: "2026-10-19T10:08:29-07:00"
              finalized: "2026-10-19T10:16:46-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-19T11:46:47-07:00"
      dispatched: "2026-10-19T11:44:44-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: declinedBy returns nil when intent.Selected is false (decline.go:65). TestDeclineSkipsAncestorInternalRows fails with the guard reverted.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The path now comes from Source and a differing Target goes to Extra; the row is inert and printDeclined warns. TestDeclineWithTwoPathsDeclinesNothing fails with the guard reverted.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The Revisions entry says the docs step updated the base.manifest header and atlas base-layer.md and ci-merge-check.md, not construct/manifest.md, matching the prior diff.
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: c3fca0ece7d29f05d3b014a3030590748cd7de60
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 35933142-781a-4bf1-8d78-697479168ca7
            repository: github.com/xianxu/ariadne
            issue_id: "000467"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4e1185258025cbe7
            reviewed_head: 23435183559f8b6c88a41ff3d52fbcb5103fade6
            slot: ariadne:5
            requested_at: "2026-10-19T18:47:50.389049Z"
          phase: finished
          pid: 16097
          started_at: "2026-10-19T18:47:50.514011Z"
          finished_at: "2026-10-19T18:59:39.011754Z"
          session_id: 424364b5-bb8a-4f4e-8d99-25bef6e853cc
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 23435183559f8b6c88a41ff3d52fbcb5103fade6
        tracker_head: 143bf488c653e126c3ee5a25faa24fea50534a85
        branch_point: 931a4ad6898a9883e1324701e92860bca401f2bc
        since: "2026-10-19T10:00:05-07:00"
        until: "2026-10-19T11:47:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T10:00:05-07:00"
              end: "2026-10-19T11:47:18-07:00"
              began: "2026-10-19T10:00:05-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 2b1c35f1-9bee-4064-83d0-2a67ebaaf586
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 819e55b9496933831fe18808378255c408fd0cfdd6412862485f0d724db6d73e
              started: "2026-10-19T16:59:53.709Z"
              spawned: false
              bytes: 4616708
              windows:
                - start: "2026-10-19T10:00:05-07:00"
                  end: "2026-10-19T11:47:18-07:00"
        queue_waits:
            - attempt_id: 734b3a70411a62c801e715582c0f8f34
              slot: ariadne:5
              record_digest: 972e9433b6772efb25c985b3c2c1a32d0215cf369bd087a20b16a681fe6f0d93
              start: "2026-10-19T17:17:49.379266Z"
              end: "2026-10-19T17:55:30.29124Z"
              queue_seconds: 2260.911974
              raw_seconds: 2260.911974
              exact: true
        cuts:
            - dispatched: "2026-10-19T10:08:29-07:00"
              finalized: "2026-10-19T10:16:46-07:00"
              finalizing: true
            - dispatched: "2026-10-19T11:44:44-07:00"
              finalized: "2026-10-19T11:46:47-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-19T11:59:40-07:00"
      dispatched: "2026-10-19T11:47:50-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      recipe: fix-review
      isolation: checkout
      reviewed: 23435183559f8b6c88a41ff3d52fbcb5103fade6
      blocked: false
---

# Gate ledger — ariadne#467 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T10:16:46-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `decline-ignores-visibility-selection` (kind: defect) Decline removes and reports ancestor internal rows that never participate in the leaf
  layer.Decline filters every ancestor row regardless of intent.Selected, so a decline over only internal ancestor rows reports "opted out of N row(s)" and suppresses the stale-decline warning. Skip rows where !intent.Selected(in.Visibility, false).
- **BR-2** [Minor] `manifest-row-arity-unchecked` (kind: defect) A decline row with two path arguments silently declines the second one
  ParseManifest sets Target to fields[2], and Decline matches on Target, so `decline a b` declines b, not a. Warn on or reject a second argument for decline.
- **BR-3** [Minor] `plan-claims-undelivered-doc` (kind: defect) Plan docs step says construct/manifest.md was updated, but it is not in the diff
  manifest.md does not list verbs, so nothing is missing from the docs. Correct the plan wording in a Revisions entry.

## Round 2 — 2026-10-19T11:46:47-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — declinedBy returns nil when intent.Selected is false (decline.go:65). TestDeclineSkipsAncestorInternalRows fails with the guard reverted.
- BR-2 — addressed — The path now comes from Source and a differing Target goes to Extra; the row is inert and printDeclined warns. TestDeclineWithTwoPathsDeclinesNothing fails with the guard reverted.
- BR-3 — addressed — The Revisions entry says the docs step updated the base.manifest header and atlas base-layer.md and ci-merge-check.md, not construct/manifest.md, matching the prior diff.

## Round 3 — 2026-10-19T11:59:40-07:00 (claude) — passed

## Open findings

(none — every finding has been disposed)
