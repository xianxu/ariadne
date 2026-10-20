---
gate: boundary-review
issue: 479
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0017bec4-9149-4a4a-a710-8cb74c288941
            repository: github.com/xianxu/ariadne
            issue_id: "000479"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/007572c22853e62d
            reviewed_head: 04bf0c5e603d97658e0e3827ef8364aa49664498
            slot: ariadne:2
            requested_at: "2026-10-20T04:08:55.844434Z"
          phase: finished
          pid: 76336
          started_at: "2026-10-20T04:08:56.022127Z"
          finished_at: "2026-10-20T04:12:25.454834Z"
          session_id: 1bae3aab-a324-47d8-8169-344b0b0a34ad
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 04bf0c5e603d97658e0e3827ef8364aa49664498
        tracker_head: 95496a0a11396eab41e4d4e1625f370a8cbacce3
        branch_point: 4ce53fe9728bee39a0beb2d57cd51b678a13146a
        since: "2026-10-19T20:21:13-07:00"
        until: "2026-10-19T20:26:20-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:21:13-07:00"
              end: "2026-10-19T20:26:20-07:00"
              began: "2026-10-19T20:21:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 4ecd10dd3ceeef0b5f04a47472705a14e51407efc3ddcbb04b5f2c634ed0742c
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 1219043
              windows:
                - start: "2026-10-19T20:21:13-07:00"
                  end: "2026-10-19T20:26:20-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 4.2722999999999995
      "n": 1
      timestamp: "2026-10-19T21:12:26-07:00"
      dispatched: "2026-10-19T21:08:55-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: TestNoParallelFlakeCoverageMatcher uses go/parser.ParseDir, deprecated in Go 1.26
          detail: flake_fixed_test.go:93. It works for this name-only scan, but staticcheck SA1019 would flag it. Use go/packages, or os.ReadDir plus parser.ParseFile. This is the only instance in the window.
          family: deprecated-stdlib-api
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The no-parallel-matcher lint only catches functions whose names end in Covers, not inline exact-match coverage
          detail: Done-when 2 actually rests on the Log's manual audit of the remaining == sites (removeFlakeTests, runFlakeRegister, Registry.Lookup). The comment on the AST test could say it guards only that one regression shape.
          family: lint-proxy-narrower-than-claim
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: 'ARCH-DRY: fixedCandidates repeats Registry.Cover''s nearest-covering-name loop'
          detail: flake_fixed.go:107 and internal/flakes/registry.go:183 both pick the longest name that Covers the leaf. Fixed issues need ties to be ambiguous, so the loops can't simply be merged. A shared flakes helper that returns the nearest covering name could feed both. Only these two instances exist.
          family: nearest-cover-duplicated
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 04bf0c5e603d97658e0e3827ef8364aa49664498
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 734fd531-3145-4f41-b30a-f2e9954ed44f
            repository: github.com/xianxu/ariadne
            issue_id: "000479"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/d935271805f5f34c
            reviewed_head: 3bea1fc6f744779005297def5e253cff45e17aa2
            slot: ariadne:2
            requested_at: "2026-10-20T07:17:45.215755Z"
          phase: finished
          pid: 63916
          started_at: "2026-10-20T07:17:45.309992Z"
          finished_at: "2026-10-20T07:19:41.953938Z"
          session_id: ebe6584a-917b-4080-8658-60ca76740526
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 3bea1fc6f744779005297def5e253cff45e17aa2
        tracker_head: be274dd1493186715136b5c7f828cff04f1ccbe0
        branch_point: 4ce53fe9728bee39a0beb2d57cd51b678a13146a
        since: "2026-10-19T20:21:13-07:00"
        until: "2026-10-19T21:15:26-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T20:21:13-07:00"
              end: "2026-10-19T21:15:26-07:00"
              began: "2026-10-19T20:21:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 1bdd15d0a2ff3b4d4af5a42c708dbfa9d4fa1e3adcf333758e0ce3c35fcde98a
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 2971724
              windows:
                - start: "2026-10-19T20:21:13-07:00"
                  end: "2026-10-19T21:15:26-07:00"
        queue_waits:
            - attempt_id: bcd9b7407517e24367c744af82d5db34
              slot: ariadne:2
              record_digest: 38cc6a62061e1a988f1d7f0f855462befe0fbd9fd08ff6e8d5e7da15ab214da4
              start: "2026-10-20T03:26:41.421592Z"
              end: "2026-10-20T03:54:30.635003Z"
              queue_seconds: 1669.213411
              raw_seconds: 1669.213411
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 31.87565
        cuts:
            - dispatched: "2026-10-19T21:08:55-07:00"
              finalized: "2026-10-19T21:12:26-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-20T00:19:42-07:00"
      dispatched: "2026-10-20T00:17:45-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: flake_fixed_test.go now uses os.ReadDir plus parser.ParseFile; ParseDir and the io/fs import are gone.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The lint comment now says it guards only the parallel *Covers shape and that inline identity comparisons are not coverage.
          round: 2
        - id: BR-3
          disposition: addressed
          note: flakes.Nearest backs both Registry.Cover and fixedCandidates; making it exact-match fails tests in both packages.
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 3bea1fc6f744779005297def5e253cff45e17aa2
      blocked: false
---

# Gate ledger — ariadne#479 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T21:12:26-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `deprecated-stdlib-api` (kind: defect) TestNoParallelFlakeCoverageMatcher uses go/parser.ParseDir, deprecated in Go 1.26
  flake_fixed_test.go:93. It works for this name-only scan, but staticcheck SA1019 would flag it. Use go/packages, or os.ReadDir plus parser.ParseFile. This is the only instance in the window.
- **BR-2** [Minor] `lint-proxy-narrower-than-claim` (kind: defect) The no-parallel-matcher lint only catches functions whose names end in Covers, not inline exact-match coverage
  Done-when 2 actually rests on the Log's manual audit of the remaining == sites (removeFlakeTests, runFlakeRegister, Registry.Lookup). The comment on the AST test could say it guards only that one regression shape.
- **BR-3** [Minor] `nearest-cover-duplicated` (kind: defect) ARCH-DRY: fixedCandidates repeats Registry.Cover's nearest-covering-name loop
  flake_fixed.go:107 and internal/flakes/registry.go:183 both pick the longest name that Covers the leaf. Fixed issues need ties to be ambiguous, so the loops can't simply be merged. A shared flakes helper that returns the nearest covering name could feed both. Only these two instances exist.

## Round 2 — 2026-10-20T00:19:42-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — flake_fixed_test.go now uses os.ReadDir plus parser.ParseFile; ParseDir and the io/fs import are gone.
- BR-2 — addressed — The lint comment now says it guards only the parallel *Covers shape and that inline identity comparisons are not coverage.
- BR-3 — addressed — flakes.Nearest backs both Registry.Cover and fixedCandidates; making it exact-match fails tests in both packages.

## Open findings

(none — every finding has been disposed)
