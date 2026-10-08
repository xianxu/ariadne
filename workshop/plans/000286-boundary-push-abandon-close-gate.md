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
    - "n": 3
      timestamp: "2026-10-08T16:23:27-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: Plan still cites landing.go:558, ghclient.go:149, landing.go:310, handoff.go:205 (pushIssueBranch has since moved to boundarypush.go).
          round: 3
        - id: BR-2
          disposition: addressed
          note: TestBoundaryVerbsPushTheIssueBranch now runs a judged (stubJudge SHIP) M2 milestone-close and asserts origin == HEAD; removing close.go:1410 fails it.
          round: 3
        - id: BR-3
          disposition: addressed
          note: TestLandingDeletesTheRemoteBranch now runs a second runMerge with --branch after the remote branch is gone.
          round: 3
        - id: BR-4
          disposition: addressed
          note: landing.go:491 warning names `sdlc merge --branch B --yes`; the resume path reaches deleteRemoteBranch (exercised by the BR-3 test).
          round: 3
        - id: BR-5
          disposition: addressed
          note: boundaryPushTimeout plus process-group cancel in gitRaw; TestBoundaryPushIsBounded uses a sleeping pre-push hook.
          round: 3
      findings:
        - id: BR-6
          severity: Important
          title: Reconcile's close-completion push (closetracker.go:149) and start-plan's rerun push have no regression test
          detail: '2nd in family. Rule: every boundary-push call site needs an origin assertion that fails when that call is removed. Covered: startplan first run, milestoneclose.go:187, close.go:1410, closetracker.go:427, landing pr and merge. Uncovered: closetracker.go:149 (no proof row in the reconcile contract) and the plan''s start-plan-rerun test. Add both and list them in catalog Proofs.'
          family: boundary-push-path-coverage
          round: 3
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 4
      timestamp: "2026-10-08T16:26:06-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: 'Verified red-without: TestReconciledClosePushes fails with closetracker.go:149 removed; TestStartPlanRerunPushes fails with startplan.go:344 removed; both listed in catalog Proofs.'
          round: 4
        - id: BR-1
          disposition: not-addressed
          note: 'Plan still carries line anchors (landing.go:558/:310, ghclient.go:149 now :154, handoff.go:205 for a function now in boundarypush.go); Integration-points row says deleteLandingBranch gains deleteRemoteBranch but the call is in runDurableMerge. Rule: name functions, never lines; add a Revisions entry.'
          round: 4
      boundary: M1
      recipe: milestone-review
      blocked: false
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

## Round 3 — 2026-10-08T16:23:27-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — not-addressed — Plan still cites landing.go:558, ghclient.go:149, landing.go:310, handoff.go:205 (pushIssueBranch has since moved to boundarypush.go).
- BR-2 — addressed — TestBoundaryVerbsPushTheIssueBranch now runs a judged (stubJudge SHIP) M2 milestone-close and asserts origin == HEAD; removing close.go:1410 fails it.
- BR-3 — addressed — TestLandingDeletesTheRemoteBranch now runs a second runMerge with --branch after the remote branch is gone.
- BR-4 — addressed — landing.go:491 warning names `sdlc merge --branch B --yes`; the resume path reaches deleteRemoteBranch (exercised by the BR-3 test).
- BR-5 — addressed — boundaryPushTimeout plus process-group cancel in gitRaw; TestBoundaryPushIsBounded uses a sleeping pre-push hook.

### Raised

- **BR-6** [Important] `boundary-push-path-coverage` Reconcile's close-completion push (closetracker.go:149) and start-plan's rerun push have no regression test
  2nd in family. Rule: every boundary-push call site needs an origin assertion that fails when that call is removed. Covered: startplan first run, milestoneclose.go:187, close.go:1410, closetracker.go:427, landing pr and merge. Uncovered: closetracker.go:149 (no proof row in the reconcile contract) and the plan's start-plan-rerun test. Add both and list them in catalog Proofs.

## Round 4 — 2026-10-08T16:26:06-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — Verified red-without: TestReconciledClosePushes fails with closetracker.go:149 removed; TestStartPlanRerunPushes fails with startplan.go:344 removed; both listed in catalog Proofs.
- BR-1 — not-addressed — Plan still carries line anchors (landing.go:558/:310, ghclient.go:149 now :154, handoff.go:205 for a function now in boundarypush.go); Integration-points row says deleteLandingBranch gains deleteRemoteBranch but the call is in runDurableMerge. Rule: name functions, never lines; add a Revisions entry.

## Open findings

- **BR-1** [Minor] `stale-line-anchor` runDurablePR is at landing.go:492 (plan says ~:558) and legacy --delete-branch at ghclient.go:154 (plan says :149)
