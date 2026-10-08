---
gate: boundary-review
issue: 286
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-08T16:15:52-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
          detail: |-
            The behavior claims are correct and only the line anchors have moved. Name the functions and drop the line numbers.
            (carried from plan-quality PQ-9, deferred to the boundary review)
          family: stale-line-anchor
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-10-08T16:15:52-07:00"
      agent: claude
      findings:
        - id: BR-2
          severity: Important
          title: The judged milestone-close push (close.go:1409) has no regression test; only the --no-judge path is exercised
          detail: TestBoundaryVerbsPushTheIssueBranch runs milestone-close with --no-judge, and the close that follows pushes anyway, so removing the finalizeBoundaryReview milestonePush call fails no test. Add a stubJudge SHIP milestone-close step and assert that origin equals HEAD right after it.
          family: boundary-push-path-coverage
          round: 2
        - id: BR-3
          severity: Minor
          title: TestLandingDeletesTheRemoteBranch says it covers a resumed landing, but it runs only one merge
          detail: Add a second runMerge from rest after the branch is gone, or drop the resume claim from the comment.
          family: test-claim-exceeds-test
          round: 2
        - id: BR-4
          severity: Minor
          title: Merge's remote-delete warning names no recovery command
          detail: Suggest naming `sdlc merge --branch B --yes`, which observation-resumes into deleteRemoteBranch.
          family: warning-names-recovery
          round: 2
        - id: BR-5
          severity: Minor
          title: A boundary push has no timeout, so a hanging network blocks the verb despite D2's warn-only intent
          family: operating-envelope-unbounded-io
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#286 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-08T16:15:52-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `stale-line-anchor` runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
  The behavior claims are correct and only the line anchors have moved. Name the functions and drop the line numbers.
  (carried from plan-quality PQ-9, deferred to the boundary review)

## Round 2 — 2026-10-08T16:15:52-07:00 (claude) — BLOCKED

### Raised

- **BR-2** [Important] `boundary-push-path-coverage` The judged milestone-close push (close.go:1409) has no regression test; only the --no-judge path is exercised
  TestBoundaryVerbsPushTheIssueBranch runs milestone-close with --no-judge, and the close that follows pushes anyway, so removing the finalizeBoundaryReview milestonePush call fails no test. Add a stubJudge SHIP milestone-close step and assert that origin equals HEAD right after it.
- **BR-3** [Minor] `test-claim-exceeds-test` TestLandingDeletesTheRemoteBranch says it covers a resumed landing, but it runs only one merge
  Add a second runMerge from rest after the branch is gone, or drop the resume claim from the comment.
- **BR-4** [Minor] `warning-names-recovery` Merge's remote-delete warning names no recovery command
  Suggest naming `sdlc merge --branch B --yes`, which observation-resumes into deleteRemoteBranch.
- **BR-5** [Minor] `operating-envelope-unbounded-io` A boundary push has no timeout, so a hanging network blocks the verb despite D2's warn-only intent

## Open findings

- **BR-1** [Minor] `stale-line-anchor` runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
- **BR-2** [Important] `boundary-push-path-coverage` The judged milestone-close push (close.go:1409) has no regression test; only the --no-judge path is exercised
- **BR-3** [Minor] `test-claim-exceeds-test` TestLandingDeletesTheRemoteBranch says it covers a resumed landing, but it runs only one merge
- **BR-4** [Minor] `warning-names-recovery` Merge's remote-delete warning names no recovery command
- **BR-5** [Minor] `operating-envelope-unbounded-io` A boundary push has no timeout, so a hanging network blocks the verb despite D2's warn-only intent
