---
gate: boundary-review
issue: 583
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3b13a1d9-17ac-4ee7-bbfa-949bce702227
            repository: github.com/xianxu/ariadne
            issue_id: "000583"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/6fa214c240f6e3a9
            reviewed_head: 8f64a172590ab44dcb07a1f059e4183c80571608
            slot: ariadne:2
            requested_at: "2026-10-24T03:31:13.978277Z"
          phase: finished
          pid: 53625
          started_at: "2026-10-24T03:31:14.117277Z"
          finished_at: "2026-10-24T03:32:42.918145Z"
          session_id: 07a3ef5e-5397-419d-8683-8b623e5bb8cb
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8f64a172590ab44dcb07a1f059e4183c80571608
        tracker_head: cbcee59d5d52fcfda4ed01b5942d31cd9b735558
        branch_point: fa9b49183e8924fb5b034678c696dd8a1ddedd26
        since: "2026-10-23T20:24:51-07:00"
        until: "2026-10-23T20:30:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T20:24:51-07:00"
              end: "2026-10-23T20:30:42-07:00"
              began: "2026-10-23T20:24:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 15c684797757310fca6ca827c2a280ec07445b5cc5b8e2c9e236f2a9665ee45f
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 16611805
              windows:
                - start: "2026-10-23T20:24:51-07:00"
                  end: "2026-10-23T20:30:42-07:00"
      "n": 1
      timestamp: "2026-10-23T20:32:44-07:00"
      dispatched: "2026-10-23T20:31:13-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: 'Reuse condition branches untested: pre-close resolving to suite, plain --name suite rerun, boarding reuse'
          detail: 'cmd/sdlc/localcheck_cmd.go:240 reuses when f.PreClose or (wantAffected and no other names). The fixture always turns trains on and declares affected.json, so pre-close always resolves to affected. Never run in this family: (a) pre-close resolving to suite, where only the f.PreClose disjunct allows reuse; (b) plain --name suite with a same-head pass must rerun (claimed at local-check.md:267 and in the code comment); (c) --boarding --name affected reuse (named in the Plan); (d) reuse of a known-flake-qualified pass (named in the Spec). Add a fixture without trains asserting pre-close reuse, assert 2 suite attempts after a second --name suite, add the boarding flags to the reuse loop, and cite or add known-flake reuse coverage.'
          family: done-when-mode-coverage
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Pre-close reuse prints "running affected" and then "nothing to run"
          detail: localcheck_cmd.go:171 announces running the resolved check before the reuse short-circuit at :240. The message at :246 also names f.Names rather than --pre-close. Reword the first line as "resolved to", or print it after the reuse check.
          family: reuse-output-wording
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 481e473a-78d1-4206-9637-b74f56dc2a55
            repository: github.com/xianxu/ariadne
            issue_id: "000583"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/96233add20fe1add
            reviewed_head: d4a79fa2665c5d708e3807f554b27fa0202c32b7
            slot: ariadne:2
            requested_at: "2026-10-24T03:34:30.408878Z"
          phase: finished
          pid: 7675
          started_at: "2026-10-24T03:34:30.512238Z"
          finished_at: "2026-10-24T03:35:50.87265Z"
          session_id: 78f7ae90-86b1-44a0-aa89-bb5b0c7adf61
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d4a79fa2665c5d708e3807f554b27fa0202c32b7
        tracker_head: cbcee59d5d52fcfda4ed01b5942d31cd9b735558
        branch_point: fa9b49183e8924fb5b034678c696dd8a1ddedd26
        since: "2026-10-23T20:24:51-07:00"
        until: "2026-10-23T20:33:54-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T20:24:51-07:00"
              end: "2026-10-23T20:33:54-07:00"
              began: "2026-10-23T20:24:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 78cc8f4fd293b22db5bb618b12f2d0c1771ac117d7d020835d24066da332785d
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 16737017
              windows:
                - start: "2026-10-23T20:24:51-07:00"
                  end: "2026-10-23T20:33:54-07:00"
        cuts:
            - dispatched: "2026-10-23T20:31:13-07:00"
              finalized: "2026-10-23T20:32:44-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T20:35:51-07:00"
      dispatched: "2026-10-23T20:34:30-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: precloseck_test.go now covers pre-close resolving to suite (no trains), the plain --name suite rerun (2 attempts) and boarding reuse; known-flake reuse is cited to localcheck_flakes_test.go:255. Removing the f.PreClose disjunct, or widening the condition to bare ok, turns the test red.
          round: 2
        - id: BR-2
          disposition: addressed
          note: localcheck_cmd.go:171 now prints "resolved to X" (pinned at precloseck_test.go:108), and the reuse line at :246 names --pre-close when it applies.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: d4a79fa2665c5d708e3807f554b27fa0202c32b7
      blocked: false
---

# Gate ledger — ariadne#583 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T20:32:44-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `done-when-mode-coverage` (kind: defect) Reuse condition branches untested: pre-close resolving to suite, plain --name suite rerun, boarding reuse
  cmd/sdlc/localcheck_cmd.go:240 reuses when f.PreClose or (wantAffected and no other names). The fixture always turns trains on and declares affected.json, so pre-close always resolves to affected. Never run in this family: (a) pre-close resolving to suite, where only the f.PreClose disjunct allows reuse; (b) plain --name suite with a same-head pass must rerun (claimed at local-check.md:267 and in the code comment); (c) --boarding --name affected reuse (named in the Plan); (d) reuse of a known-flake-qualified pass (named in the Spec). Add a fixture without trains asserting pre-close reuse, assert 2 suite attempts after a second --name suite, add the boarding flags to the reuse loop, and cite or add known-flake reuse coverage.
- **BR-2** [Minor] `reuse-output-wording` (kind: defect) Pre-close reuse prints "running affected" and then "nothing to run"
  localcheck_cmd.go:171 announces running the resolved check before the reuse short-circuit at :240. The message at :246 also names f.Names rather than --pre-close. Reword the first line as "resolved to", or print it after the reuse check.

## Round 2 — 2026-10-23T20:35:51-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — precloseck_test.go now covers pre-close resolving to suite (no trains), the plain --name suite rerun (2 attempts) and boarding reuse; known-flake reuse is cited to localcheck_flakes_test.go:255. Removing the f.PreClose disjunct, or widening the condition to bare ok, turns the test red.
- BR-2 — addressed — localcheck_cmd.go:171 now prints "resolved to X" (pinned at precloseck_test.go:108), and the reuse line at :246 names --pre-close when it applies.

## Open findings

(none — every finding has been disposed)
