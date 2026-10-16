---
gate: boundary-review
issue: 408
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 22ab9433-737e-4cb6-a283-c4155d8d806e
            repository: github.com/xianxu/ariadne
            issue_id: "000408"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/abe73d1142956477
            reviewed_head: 7ca9a41743d8986f4ee5211494eac15355da3737
            slot: ariadne:3
            requested_at: "2026-10-16T18:49:51.637226Z"
          phase: finished
          pid: 92349
          started_at: "2026-10-16T18:49:51.865379Z"
          finished_at: "2026-10-16T18:51:24.477328Z"
          session_id: 01a1460c-965f-7601-91f4-1c8509cf8602
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7ca9a41743d8986f4ee5211494eac15355da3737
        tracker_head: 6f2385f2813a3a8a5188e0ee8c8916ab45928252
        branch_point: 1a8c52756b74160f23240dd123ddc25f5b20ed32
        since: "2026-10-16T11:18:56-07:00"
        until: "2026-10-16T11:49:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T11:18:56-07:00"
              end: "2026-10-16T11:49:11-07:00"
              began: "2026-10-16T11:18:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9f5aeeda82ec46044c771a7c574180a5707e7ca03b39595fc067a42243eafe8a
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 4891979
              windows:
                - start: "2026-10-16T11:18:56-07:00"
                  end: "2026-10-16T11:49:11-07:00"
      "n": 1
      timestamp: "2026-10-16T11:51:25-07:00"
      dispatched: "2026-10-16T11:49:51-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Initial-publication recovery does not revalidate archive locations
          detail: 'cmd/sdlc/issuemovedetail.go:134 checks locations only at preflight. If an archive appears afterward, transfer candidate preparation uses DetailBundlePublication (internal/tracker/detailpublish.go:27), which checks only the active destination; newIssueTransferOp''s BeforePublish checks assets only. The resulting candidate can publish an active twin. ARCH-PURPOSE: sweep initial candidate preparation, retry preparation, and recovery replay for both legacy flat history and history/issues; republish candidate preparation already performs the ID-wide check. Add regression coverage that introduces an archive after preflight and verifies refusal without an active write.'
          family: publication-location-revalidation
          round: 1
        - id: BR-2
          severity: Important
          title: Local detail discovery accepts non-Markdown entries
          detail: cmd/sdlc/issuepublish.go:103 passes LocalPathsByID results directly to classifyFamily, whose default class accepts any matching ID prefix. Thus 000012-note.md.bak beside valid details causes a false duplicate refusal. The same unfiltered collection also drives the local-archive refusal at line 119, where an archived backup alone can block initial publication. These are all instances in this window. Filter through issue.ParseFilename before family classification, matching the published resolver, and test backup and non-Markdown entries.
          family: detail-filename-grammar
          round: 1
        - id: BR-3
          severity: Important
          title: Done when omits the local-only archive revision
          detail: The issue's Revisions section at line 104 adds refusal when archived details exist locally but not on main, implemented at issuepublish.go:119 and tested by TestPublishUnpublishedLocalArchiveRefuses. Done when at lines 38–39 covers archive updates, missing done-card details, and twins, but not this distinct state with an open card and one unpublished archive. Append explicit acceptance for refusal, preservation of the local archive, and no active file or main change.
          family: acceptance-contract-completeness
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b223f882-64e4-4c00-a786-1ed9ea257d6e
            repository: github.com/xianxu/ariadne
            issue_id: "000408"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/fc9c54246fc5f9a6
            reviewed_head: 411260a95cd22e012c9e0b74b5e73983b843001b
            slot: ariadne:3
            requested_at: "2026-10-16T19:20:08.85452Z"
          phase: finished
          pid: 57179
          started_at: "2026-10-16T19:20:09.200677Z"
          finished_at: "2026-10-16T19:25:19.349407Z"
          session_id: 01a14628-506c-7f80-b623-fe47c2be186c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 411260a95cd22e012c9e0b74b5e73983b843001b
        tracker_head: ae28755ffab8160eea299b389d4d4504f8b54bf4
        branch_point: 1a8c52756b74160f23240dd123ddc25f5b20ed32
        since: "2026-10-16T11:18:56-07:00"
        until: "2026-10-16T12:19:21-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T11:18:56-07:00"
              end: "2026-10-16T12:19:21-07:00"
              began: "2026-10-16T11:18:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: bbba02ef76fed9f1d7d0578d9b1236a511c26a416cf1dae3072964c7cc8a783a
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 6499375
              windows:
                - start: "2026-10-16T11:18:56-07:00"
                  end: "2026-10-16T12:19:21-07:00"
            - harness: codex
              id: 01a14610-8410-7663-9460-18fdcdccd3b7
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3b1178001718b83b3dac1c9a054c0f02e03d2b4b54b0fb62e84afadd29e83eaf
              started: "2026-10-16T18:54:10.733Z"
              spawned: false
              bytes: 543034
              windows:
                - start: "2026-10-16T18:54:10.733Z"
                  end: "2026-10-16T12:19:21-07:00"
            - harness: codex
              id: 01a14612-6f7e-73e3-9e62-6e216ffbc42b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e76fadc3d4c1495f8287069fcaf5915ab578a73f6c73779b72ac8fc836b7d186
              started: "2026-10-16T18:56:16.333Z"
              spawned: false
              bytes: 429829
              windows:
                - start: "2026-10-16T18:56:16.333Z"
                  end: "2026-10-16T12:19:21-07:00"
            - harness: codex
              id: 01a14613-017f-7ac0-8533-29575dd19183
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e2d0a4e6234ad8f6288995f72564331208e084d09579040a36867d279b567a48
              started: "2026-10-16T18:56:53.793Z"
              spawned: false
              bytes: 376671
              windows:
                - start: "2026-10-16T18:56:53.793Z"
                  end: "2026-10-16T12:19:21-07:00"
        cuts:
            - dispatched: "2026-10-16T11:49:51-07:00"
              finalized: "2026-10-16T11:51:25-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-16T12:25:23-07:00"
      dispatched: "2026-10-16T12:20:08-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The shared BeforePublish hook now rechecks ID-wide locations. All six archive regressions pass on head and fail when this guard is removed, covering both archive layouts across first publication, moved-main retry and receipt recovery.
          round: 2
        - id: BR-2
          disposition: addressed
          note: issuepublish.go:104 filters local candidates through the canonical filename parser before either local decision. All six discovery regressions pass on head and fail when the filter is removed.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The issue's Done when at lines 41–42 explicitly requires local-only archive refusal, byte preservation, unchanged main and no active details. This matches issuepublish.go:121–122 and TestPublishUnpublishedLocalArchiveRefuses, which passed.
          round: 2
      findings:
        - kind: defect
          id: BR-4
          severity: Critical
          title: Terminal-card admission still bypasses publication-time revalidation
          detail: 'cmd/sdlc/issuemovedetail.go:439–451 revalidates locations but never calls requireInitialDetails against the current card; that check remains confined to preflight. In a disposable pinned-head reproduction using archiveRaceTransfer, changing the card to done before transfer.main preparation made Drive return success, publish workshop/issues/000008-spin-off.md and remove the local source. Changing the card after saving the candidate also let recovery succeed; subsequent cleanup archived the newly published file instead of refusing. This violates Spec lines 32–33. This is the 2nd finding in family publication-location-revalidation: do not fix only this instance. State and enforce the shared rule that every unconfirmed initial publication requires both an eligible current card and no existing ID-wide details. Sweep first publication, candidate replay and moved-base retry, with terminal-status regressions and confirmed-publication recovery preserved. ARCH-PURPOSE, ARCH-ORDER.'
          family: publication-location-revalidation
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 21597bc7-f346-4c53-a1c3-cb8a6a7339db
            repository: github.com/xianxu/ariadne
            issue_id: "000408"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e080bcf673647fc0
            reviewed_head: 9da64867cec00ab9d09b1d3391e4988262134d86
            slot: ariadne:3
            requested_at: "2026-10-16T20:00:14.480152Z"
          phase: finished
          pid: 69659
          started_at: "2026-10-16T20:00:14.636648Z"
          finished_at: "2026-10-16T20:02:44.484459Z"
          session_id: 01a1464d-046a-7882-98b1-1940f9bc2dfd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 9da64867cec00ab9d09b1d3391e4988262134d86
        tracker_head: e940f44f382f2564c8fb08340dc4efe4e7b38984
        branch_point: 1a8c52756b74160f23240dd123ddc25f5b20ed32
        since: "2026-10-16T11:18:56-07:00"
        until: "2026-10-16T12:59:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T11:18:56-07:00"
              end: "2026-10-16T12:59:55-07:00"
              began: "2026-10-16T11:18:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6b066eb7b5d8dff65f49c98e1e46d0b95e65dbfdfa76beae300750a894756464
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 7682177
              windows:
                - start: "2026-10-16T11:18:56-07:00"
                  end: "2026-10-16T12:59:55-07:00"
            - harness: codex
              id: 01a14610-8410-7663-9460-18fdcdccd3b7
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3b1178001718b83b3dac1c9a054c0f02e03d2b4b54b0fb62e84afadd29e83eaf
              started: "2026-10-16T18:54:10.733Z"
              spawned: false
              bytes: 543034
              windows:
                - start: "2026-10-16T18:54:10.733Z"
                  end: "2026-10-16T12:59:55-07:00"
            - harness: codex
              id: 01a14612-6f7e-73e3-9e62-6e216ffbc42b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e76fadc3d4c1495f8287069fcaf5915ab578a73f6c73779b72ac8fc836b7d186
              started: "2026-10-16T18:56:16.333Z"
              spawned: false
              bytes: 429829
              windows:
                - start: "2026-10-16T18:56:16.333Z"
                  end: "2026-10-16T12:59:55-07:00"
            - harness: codex
              id: 01a14613-017f-7ac0-8533-29575dd19183
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e2d0a4e6234ad8f6288995f72564331208e084d09579040a36867d279b567a48
              started: "2026-10-16T18:56:53.793Z"
              spawned: false
              bytes: 376671
              windows:
                - start: "2026-10-16T18:56:53.793Z"
                  end: "2026-10-16T12:59:55-07:00"
            - harness: codex
              id: 01a14630-09d3-7033-b75e-5c5236168d01
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 66c411f3b95b6bcd17f1aba8bbc5d5a751ab99654d74132808c9abb5705b76de
              started: "2026-10-16T19:28:36.662Z"
              spawned: false
              bytes: 455228
              windows:
                - start: "2026-10-16T19:28:36.662Z"
                  end: "2026-10-16T12:59:55-07:00"
            - harness: codex
              id: 01a14630-adfa-7cb1-870c-a59856084acb
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ed2077040f1ad0c0fc038e7b140f421eeeb88c1ec0676260ee6b4d7240ecef47
              started: "2026-10-16T19:29:18.589Z"
              spawned: false
              bytes: 421201
              windows:
                - start: "2026-10-16T19:29:18.589Z"
                  end: "2026-10-16T12:59:55-07:00"
        cuts:
            - dispatched: "2026-10-16T11:49:51-07:00"
              finalized: "2026-10-16T11:51:25-07:00"
              finalizing: false
            - dispatched: "2026-10-16T12:20:08-07:00"
              finalized: "2026-10-16T12:25:23-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-16T13:02:45-07:00"
      dispatched: "2026-10-16T13:00:14-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Archive revalidation remains in the shared BeforePublish path; all six archive-layout and publication-phase regression cases passed.
          round: 3
        - id: BR-2
          disposition: addressed
          note: localDetailPath filters through the canonical filename parser before classification; all six backup/non-Markdown regression cases passed.
          round: 3
        - id: BR-3
          disposition: addressed
          note: The pinned issue diff adds the local-only archive acceptance clause, matching TestPublishUnpublishedLocalArchiveRefuses and entry-point parity coverage.
          round: 3
        - id: BR-4
          disposition: addressed
          note: checkInitialPublication reads the current card and calls requireInitialDetails before unconfirmed publication. Removing only that call in a disposable pinned-head copy made all nine terminal-status/phase cases fail with publication accepted. The unchanged head passes, including confirmed-publication recovery.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 9da64867cec00ab9d09b1d3391e4988262134d86
      blocked: false
---

# Gate ledger — ariadne#408 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T11:51:25-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `publication-location-revalidation` Initial-publication recovery does not revalidate archive locations
  cmd/sdlc/issuemovedetail.go:134 checks locations only at preflight. If an archive appears afterward, transfer candidate preparation uses DetailBundlePublication (internal/tracker/detailpublish.go:27), which checks only the active destination; newIssueTransferOp's BeforePublish checks assets only. The resulting candidate can publish an active twin. ARCH-PURPOSE: sweep initial candidate preparation, retry preparation, and recovery replay for both legacy flat history and history/issues; republish candidate preparation already performs the ID-wide check. Add regression coverage that introduces an archive after preflight and verifies refusal without an active write.
- **BR-2** [Important] `detail-filename-grammar` Local detail discovery accepts non-Markdown entries
  cmd/sdlc/issuepublish.go:103 passes LocalPathsByID results directly to classifyFamily, whose default class accepts any matching ID prefix. Thus 000012-note.md.bak beside valid details causes a false duplicate refusal. The same unfiltered collection also drives the local-archive refusal at line 119, where an archived backup alone can block initial publication. These are all instances in this window. Filter through issue.ParseFilename before family classification, matching the published resolver, and test backup and non-Markdown entries.
- **BR-3** [Important] `acceptance-contract-completeness` Done when omits the local-only archive revision
  The issue's Revisions section at line 104 adds refusal when archived details exist locally but not on main, implemented at issuepublish.go:119 and tested by TestPublishUnpublishedLocalArchiveRefuses. Done when at lines 38–39 covers archive updates, missing done-card details, and twins, but not this distinct state with an open card and one unpublished archive. Append explicit acceptance for refusal, preservation of the local archive, and no active file or main change.

## Round 2 — 2026-10-16T12:25:23-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The shared BeforePublish hook now rechecks ID-wide locations. All six archive regressions pass on head and fail when this guard is removed, covering both archive layouts across first publication, moved-main retry and receipt recovery.
- BR-2 — addressed — issuepublish.go:104 filters local candidates through the canonical filename parser before either local decision. All six discovery regressions pass on head and fail when the filter is removed.
- BR-3 — addressed — The issue's Done when at lines 41–42 explicitly requires local-only archive refusal, byte preservation, unchanged main and no active details. This matches issuepublish.go:121–122 and TestPublishUnpublishedLocalArchiveRefuses, which passed.

### Raised

- **BR-4** [Critical] `publication-location-revalidation` (kind: defect) Terminal-card admission still bypasses publication-time revalidation
  cmd/sdlc/issuemovedetail.go:439–451 revalidates locations but never calls requireInitialDetails against the current card; that check remains confined to preflight. In a disposable pinned-head reproduction using archiveRaceTransfer, changing the card to done before transfer.main preparation made Drive return success, publish workshop/issues/000008-spin-off.md and remove the local source. Changing the card after saving the candidate also let recovery succeed; subsequent cleanup archived the newly published file instead of refusing. This violates Spec lines 32–33. This is the 2nd finding in family publication-location-revalidation: do not fix only this instance. State and enforce the shared rule that every unconfirmed initial publication requires both an eligible current card and no existing ID-wide details. Sweep first publication, candidate replay and moved-base retry, with terminal-status regressions and confirmed-publication recovery preserved. ARCH-PURPOSE, ARCH-ORDER.

## Round 3 — 2026-10-16T13:02:45-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Archive revalidation remains in the shared BeforePublish path; all six archive-layout and publication-phase regression cases passed.
- BR-2 — addressed — localDetailPath filters through the canonical filename parser before classification; all six backup/non-Markdown regression cases passed.
- BR-3 — addressed — The pinned issue diff adds the local-only archive acceptance clause, matching TestPublishUnpublishedLocalArchiveRefuses and entry-point parity coverage.
- BR-4 — addressed — checkInitialPublication reads the current card and calls requireInitialDetails before unconfirmed publication. Removing only that call in a disposable pinned-head copy made all nine terminal-status/phase cases fail with publication accepted. The unchanged head passes, including confirmed-publication recovery.

## Open findings

(none — every finding has been disposed)
