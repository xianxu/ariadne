---
gate: boundary-review
issue: 499
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d61398c3-b61a-4fb1-bcb4-ff7a94a2538e
            repository: github.com/xianxu/ariadne
            issue_id: "000499"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/5981a55a8c1b48a1
            reviewed_head: 8d74ea7a3132393ba070737dc0a1ac87b0db6439
            slot: ariadne:5
            requested_at: "2026-10-20T16:06:10.711306Z"
          phase: finished
          pid: 63085
          started_at: "2026-10-20T16:06:10.803729Z"
          finished_at: "2026-10-20T16:09:59.41483Z"
          session_id: 01a15a10-2969-7022-8fd8-5bd98abcb468
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8d74ea7a3132393ba070737dc0a1ac87b0db6439
        tracker_head: 4aaad437d0c3a4263480ddc638301ead2aa87e39
        branch_point: e725110a66b5585debff6181e2bf0b0fc7f27d26
        since: "2026-10-20T08:58:55-07:00"
        until: "2026-10-20T09:05:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T08:58:55-07:00"
              end: "2026-10-20T09:05:55-07:00"
              began: "2026-10-20T08:58:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: c1d21344f258cd98ecc6d49c95f777f411358591387520341d81dff49f570a68
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 1545643
              windows:
                - start: "2026-10-20T08:58:55-07:00"
                  end: "2026-10-20T09:05:55-07:00"
      "n": 1
      timestamp: "2026-10-20T09:10:00-07:00"
      dispatched: "2026-10-20T09:06:10-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Production readiness does not share one tracker generation across all checks
          detail: 'cmd/sdlc/landingcar.go:43 scopes loadIssueRecords, but runLandingPublishGate first calls guardTransferredDetails (landinggate.go:84), whose changedDetails reads env.repo.Snapshot directly (transferguard.go:82). Issue selection subsequently reads through the scoped cache (landingarchive.go:770). An external write between these checks produces different generations within one readiness pass; a disposable-head probe retaining both production checks confirmed distinct tracker SHAs. Route both consumers through one pass-local snapshot and add a production-path regression. The complete affected family comprises this direct-snapshot/scoped-record split, the coherence claim at landingcar.go:40–43, its atlas restatement at atlas/workflow/sdlc-binary.md:455, and the fake-gate coverage at landingcar_records_test.go:88–142, which cannot detect that split. ARCH-PURPOSE: the observation boundary must cover every tracker consumer in the pass.'
          family: pass-scoped-authority
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7908d70b-e7b9-4487-a8a0-4ae3ae82ef2a
            repository: github.com/xianxu/ariadne
            issue_id: "000499"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2d5da2fbaecd8d0c
            reviewed_head: 2af32ce2021a9388d45e21bde7f3605ed048d590
            slot: ariadne:5
            requested_at: "2026-10-20T16:34:31.787767Z"
          phase: finished
          pid: 12714
          started_at: "2026-10-20T16:34:31.948835Z"
          finished_at: "2026-10-20T16:37:48.712515Z"
          session_id: 01a15a2a-1f2d-77c0-b058-08d326a43b8b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2af32ce2021a9388d45e21bde7f3605ed048d590
        tracker_head: 9888952f070da1a83c0cdacdcc77550d300ed5c4
        branch_point: e725110a66b5585debff6181e2bf0b0fc7f27d26
        since: "2026-10-20T08:58:55-07:00"
        until: "2026-10-20T09:33:47-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T08:58:55-07:00"
              end: "2026-10-20T09:33:47-07:00"
              began: "2026-10-20T08:58:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: b072cf34ca693176e376d9c27191230b2347d8499968d29e10b090bdbff0e820
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 2746987
              windows:
                - start: "2026-10-20T08:58:55-07:00"
                  end: "2026-10-20T09:33:47-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 1.3874166666666667
        cuts:
            - dispatched: "2026-10-20T09:06:10-07:00"
              finalized: "2026-10-20T09:10:00-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T09:37:49-07:00"
      dispatched: "2026-10-20T09:34:31-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: transferguard.go:35,82 now uses scoped records; issuerecord.go:68 shares repository authority across detail directories. TestCarReadinessPublishChecksShareGeneration passes at HEAD and fails when the former direct-snapshot reader is restored in a disposable checkout.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 2af32ce2021a9388d45e21bde7f3605ed048d590
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 229c3b1c-a254-427a-b245-a34d8a3544c4
            repository: github.com/xianxu/ariadne
            issue_id: "000499"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/96d8265027933b3a
            reviewed_head: 266a62dd1139bd222ea7594044201286fce91246
            slot: ariadne:5
            requested_at: "2026-10-20T17:02:55.597056Z"
          phase: finished
          pid: 44044
          started_at: "2026-10-20T17:02:55.692262Z"
          finished_at: "2026-10-20T17:04:05.007868Z"
          session_id: 01a15a44-1e01-7d41-b35a-294a272a0957
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 266a62dd1139bd222ea7594044201286fce91246
        tracker_head: df38ba5af9f8fadf91afcff3e7ca665ab07b107c
        branch_point: e725110a66b5585debff6181e2bf0b0fc7f27d26
        since: "2026-10-20T08:58:55-07:00"
        until: "2026-10-20T10:02:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T08:58:55-07:00"
              end: "2026-10-20T10:02:31-07:00"
              began: "2026-10-20T08:58:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 563146a64efac983ca0ddd595a784050e321c4063a48b36fc6704a5636e59e5c
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 3616105
              windows:
                - start: "2026-10-20T08:58:55-07:00"
                  end: "2026-10-20T10:02:31-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 28.503233333333334
        cuts:
            - dispatched: "2026-10-20T09:06:10-07:00"
              finalized: "2026-10-20T09:10:00-07:00"
              finalizing: false
            - dispatched: "2026-10-20T09:34:31-07:00"
              finalized: "2026-10-20T09:37:49-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-20T10:04:05-07:00"
      dispatched: "2026-10-20T10:02:55-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Scoped production authority remains intact: TestCarReadinessPublishChecksShareGeneration passes at HEAD. The root-binding correction preserves that shared reader while preventing WF_ISSUES_DIR from selecting a peer repository''s authority.'
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 266a62dd1139bd222ea7594044201286fce91246
      blocked: false
---

# Gate ledger — ariadne#499 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T09:10:00-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `pass-scoped-authority` (kind: defect) Production readiness does not share one tracker generation across all checks
  cmd/sdlc/landingcar.go:43 scopes loadIssueRecords, but runLandingPublishGate first calls guardTransferredDetails (landinggate.go:84), whose changedDetails reads env.repo.Snapshot directly (transferguard.go:82). Issue selection subsequently reads through the scoped cache (landingarchive.go:770). An external write between these checks produces different generations within one readiness pass; a disposable-head probe retaining both production checks confirmed distinct tracker SHAs. Route both consumers through one pass-local snapshot and add a production-path regression. The complete affected family comprises this direct-snapshot/scoped-record split, the coherence claim at landingcar.go:40–43, its atlas restatement at atlas/workflow/sdlc-binary.md:455, and the fake-gate coverage at landingcar_records_test.go:88–142, which cannot detect that split. ARCH-PURPOSE: the observation boundary must cover every tracker consumer in the pass.

## Round 2 — 2026-10-20T09:37:49-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — transferguard.go:35,82 now uses scoped records; issuerecord.go:68 shares repository authority across detail directories. TestCarReadinessPublishChecksShareGeneration passes at HEAD and fails when the former direct-snapshot reader is restored in a disposable checkout.

## Round 3 — 2026-10-20T10:04:05-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Scoped production authority remains intact: TestCarReadinessPublishChecksShareGeneration passes at HEAD. The root-binding correction preserves that shared reader while preventing WF_ISSUES_DIR from selecting a peer repository's authority.

## Open findings

(none — every finding has been disposed)
