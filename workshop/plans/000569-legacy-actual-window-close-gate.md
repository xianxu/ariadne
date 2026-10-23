---
gate: boundary-review
issue: 569
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 76de1c9f-4d21-4ad9-8008-975baebf257e
            repository: github.com/xianxu/ariadne
            issue_id: "000569"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/38cb31914bca19b2
            reviewed_head: 96cd72ed248907d4cdee3a026054be87cbd47a69
            slot: ariadne:3
            requested_at: "2026-10-23T09:05:06.531615Z"
          phase: finished
          pid: 52738
          started_at: "2026-10-23T09:05:06.625098Z"
          finished_at: "2026-10-23T09:07:20.249297Z"
          session_id: 01a16801-bd55-7ad2-9952-0cd55678e92a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 96cd72ed248907d4cdee3a026054be87cbd47a69
        tracker_head: 223c5b5cb78d7ccd01a0f49a9ad0fc03d2d58712
        branch_point: b10a950e2693b79fe69b1d4279f1b752c9d48162
        since: "2026-10-23T01:30:48-07:00"
        until: "2026-10-23T02:04:07-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T01:30:48-07:00"
              end: "2026-10-23T02:04:07-07:00"
              began: "2026-10-23T01:30:48-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 5765fd206046bbba738dde2527961aa4e20bfebf552d67fb5c0aa6f113056f65
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 119177838
              windows:
                - start: "2026-10-23T01:30:48-07:00"
                  end: "2026-10-23T02:04:07-07:00"
      "n": 1
      timestamp: "2026-10-23T02:07:20-07:00"
      dispatched: "2026-10-23T02:05:06-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      recipe: small-diff-review
      isolation: fenced
      reviewed: 96cd72ed248907d4cdee3a026054be87cbd47a69
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f1dc5fac-41f0-49b3-adad-7e763e482c2b
            repository: github.com/xianxu/ariadne
            issue_id: "000569"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/92da1d8b41348ec6
            reviewed_head: 662b28ccbfc4c6e2fc16a3d3f07cee1446196c23
            slot: ariadne:3
            requested_at: "2026-10-23T09:18:45.317596Z"
          phase: finished
          pid: 10671
          started_at: "2026-10-23T09:18:45.420206Z"
          finished_at: "2026-10-23T09:23:00.772908Z"
          session_id: 01a1680e-3bf2-7263-9dc6-3d743cd31bfd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 662b28ccbfc4c6e2fc16a3d3f07cee1446196c23
        tracker_head: 7c760f31cb00aa357e9703043e7973b7694ace23
        branch_point: b10a950e2693b79fe69b1d4279f1b752c9d48162
        since: "2026-10-23T01:30:48-07:00"
        until: "2026-10-23T02:18:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T01:30:48-07:00"
              end: "2026-10-23T02:18:01-07:00"
              began: "2026-10-23T01:30:48-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 2cc7d7111b1038d167f2d3d16d36d32fa7a0ec187152a2ba6580328caf057cf8
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 119623953
              windows:
                - start: "2026-10-23T01:30:48-07:00"
                  end: "2026-10-23T02:18:01-07:00"
        cuts:
            - dispatched: "2026-10-23T02:05:06-07:00"
              finalized: "2026-10-23T02:07:20-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-23T02:23:01-07:00"
      dispatched: "2026-10-23T02:18:45-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Modern evidence saved after the selected close escapes the archive-bound guard
          detail: 'cmd/sdlc/legacycompletion.go:187-192 checks only selected.CloseCommit when present. A modern or unreadable ledger saved at archival, or in a later ledger-only commit before cutover, is therefore ignored and loadAdoptedExclusions permits the disjoint skip. Both timings were reproduced with independent damaged-input, retained-input, reviewed-head-only, and unreadable-ledger cases: all returned nil error and a skip warning. This is the sole production guard in this window, consumed by cmd/sdlc/actualexclusion.go:323. ARCH-PURPOSE: validate authority at the terminal and pinned source snapshots as well as the selected close; add regressions covering both timings and all four evidence forms.'
          family: legacy-bound-must-check-source-authority
          round: 2
        - kind: defect
          id: BR-2
          severity: Minor
          title: The old-import fixture comment contradicts the new partial-ledger branch
          detail: cmd/sdlc/actualexclusion_test.go:625-626 says this generation never had a review-bearing close or archive transition, but mode.damaged immediately creates both through legacyCapturedClose and f.terminal. This is the only contradictory fixture-description instance found in the window. Move the comment into the born-archived branch or describe both modes.
          family: comments-must-match-covered-branches
          round: 2
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#569 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T02:07:20-07:00 (codex) — passed

## Round 2 — 2026-10-23T02:23:01-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `legacy-bound-must-check-source-authority` (kind: defect) Modern evidence saved after the selected close escapes the archive-bound guard
  cmd/sdlc/legacycompletion.go:187-192 checks only selected.CloseCommit when present. A modern or unreadable ledger saved at archival, or in a later ledger-only commit before cutover, is therefore ignored and loadAdoptedExclusions permits the disjoint skip. Both timings were reproduced with independent damaged-input, retained-input, reviewed-head-only, and unreadable-ledger cases: all returned nil error and a skip warning. This is the sole production guard in this window, consumed by cmd/sdlc/actualexclusion.go:323. ARCH-PURPOSE: validate authority at the terminal and pinned source snapshots as well as the selected close; add regressions covering both timings and all four evidence forms.
- **BR-2** [Minor] `comments-must-match-covered-branches` (kind: defect) The old-import fixture comment contradicts the new partial-ledger branch
  cmd/sdlc/actualexclusion_test.go:625-626 says this generation never had a review-bearing close or archive transition, but mode.damaged immediately creates both through legacyCapturedClose and f.terminal. This is the only contradictory fixture-description instance found in the window. Move the comment into the born-archived branch or describe both modes.

## Open findings

- **BR-1** [Critical] `legacy-bound-must-check-source-authority` (kind: defect) Modern evidence saved after the selected close escapes the archive-bound guard
- **BR-2** [Minor] `comments-must-match-covered-branches` (kind: defect) The old-import fixture comment contradicts the new partial-ledger branch
