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

## Open findings

- **BR-1** [Critical] `publication-location-revalidation` Initial-publication recovery does not revalidate archive locations
- **BR-2** [Important] `detail-filename-grammar` Local detail discovery accepts non-Markdown entries
- **BR-3** [Important] `acceptance-contract-completeness` Done when omits the local-only archive revision
