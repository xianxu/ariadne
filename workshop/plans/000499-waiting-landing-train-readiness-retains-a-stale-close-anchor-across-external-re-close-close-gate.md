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
---

# Gate ledger — ariadne#499 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T09:10:00-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `pass-scoped-authority` (kind: defect) Production readiness does not share one tracker generation across all checks
  cmd/sdlc/landingcar.go:43 scopes loadIssueRecords, but runLandingPublishGate first calls guardTransferredDetails (landinggate.go:84), whose changedDetails reads env.repo.Snapshot directly (transferguard.go:82). Issue selection subsequently reads through the scoped cache (landingarchive.go:770). An external write between these checks produces different generations within one readiness pass; a disposable-head probe retaining both production checks confirmed distinct tracker SHAs. Route both consumers through one pass-local snapshot and add a production-path regression. The complete affected family comprises this direct-snapshot/scoped-record split, the coherence claim at landingcar.go:40–43, its atlas restatement at atlas/workflow/sdlc-binary.md:455, and the fake-gate coverage at landingcar_records_test.go:88–142, which cannot detect that split. ARCH-PURPOSE: the observation boundary must cover every tracker consumer in the pass.

## Open findings

- **BR-1** [Critical] `pass-scoped-authority` (kind: defect) Production readiness does not share one tracker generation across all checks
