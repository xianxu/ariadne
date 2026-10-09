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
    - "n": 5
      timestamp: "2026-10-08T17:03:42-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: The plan has no file.go:NNN anchors left (grep found none); functions are named and a Revisions entry records it.
          round: 5
      findings:
        - id: BR-7
          severity: Important
          title: abandon archives plan files from main, not the kept tip, so a reopen replaces the branch's plan edits
          detail: 'abandon.go:281-289 copies p.Content from main''s view, while the details come from rec.Head. Reopen''s resolveArchiveConflicts takes main''s side with git rm, and step 4 at abandon.go:490 removes the live copy and moves the stale archived plan back. Rule: for started work, everything abandon archives comes from the kept tip. No test covers plan files in abandon or reopen.'
          family: archive-source-of-truth
          round: 5
        - id: BR-8
          severity: Important
          title: README update appears missing for sdlc abandon, the archive ref and the set-status redirect
          detail: README's Concurrent issue work section lists every ownership and lifecycle verb (claim, unclaim, reclaim, move, issue show) but not abandon or the reopen restore.
          family: readme-verb-surface
          round: 5
        - id: BR-9
          severity: Minor
          title: the abandon resume path skips the ownership check that plan D7 keeps for reruns
          detail: abandon.go:128-162 checks only a clean tree on resume; a foreign workspace can run dropAbandonedBranch. The containment lease limits the harm.
          family: rerun-skips-precondition
          round: 5
        - id: BR-10
          severity: Minor
          title: restoreAbandoned reads WF_HISTORY_DIR/WF_PLANS_DIR from the environment while abandon takes --history-dir/--plans-dir flags
          family: config-source-divergence
          round: 5
        - id: BR-11
          severity: Minor
          title: the Core concepts table names ClearCardAbandoned, a 5-argument abandonDecision and a {Ref, Head} record, none of which match the code
          family: plan-table-drift
          round: 5
        - id: BR-12
          severity: Minor
          title: reopen rerun variants after step 2 and after the card write are untested, and the after-card-write rerun never runs finishReopen
          detail: 'This is the 2nd finding in this family. Rule: every test variant the plan lists maps to a named test in the contract''s Proofs or to a Revisions entry that drops it. After a lost response on the card write, the rerun returns "already has that" and skips the mirror commit and the archive-ref deletion.'
          family: test-claim-exceeds-test
          round: 5
        - id: BR-13
          severity: Minor
          title: a forced set-status from punt to open or blocked clears the abandoned record without restoring, leaving an archive ref nothing points to
          family: record-cleared-without-effect
          round: 5
        - id: BR-14
          severity: Minor
          title: reading a ref's tip from the remote via ls-remote is duplicated at abandon.go:216, abandon.go:355, boundarypush.go:81 and handoff.go:284
          family: shared-remote-ref-read
          round: 5
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 6
      timestamp: "2026-10-08T17:18:24-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: keptPlans archives plans from rec.Head; TestAbandonKeepsTheBranchsPlans goes red with the tip overlay removed (archived plan v1).
          round: 6
        - id: BR-8
          disposition: addressed
          note: README.md:42-50 now covers boundary pushes, sdlc abandon, refs/ariadne/abandoned/NNNNNN and the set-status working restore.
          round: 6
        - id: BR-9
          disposition: not-addressed
          note: The ownership check exists at abandon.go:165-169, but no test covers a foreign resume; with the check disabled every TestAbandon test still passes.
          round: 6
        - id: BR-10
          disposition: addressed
          note: The flags are removed; abandon and reopen both read plansDir()/historyDir() from WF_PLANS_DIR/WF_HISTORY_DIR, and helptext has no stale flags.
          round: 6
        - id: BR-11
          disposition: addressed
          note: The table now reads Abandoned {Ref, Branch, Head}, SetCardAbandoned (nil clears) and abandonDecision(card, as, today, rec), matching abandoned.go:16 and abandon.go:63.
          round: 6
        - id: BR-12
          disposition: addressed
          note: TestReopenRerunVariants covers both variants; disabling the working-status branch in reopenAbandoned turns "after the card write" red.
          round: 6
        - id: BR-13
          disposition: addressed
          note: statusDecision refuses leaving terminal except by working when the record is started; TestAbandonedWorkLeavesOnlyByReopen goes red without the guard.
          round: 6
        - id: BR-14
          disposition: addressed
          note: The four named sites use remoteRefTip; the pre-existing landing.go:392 raw ls-remote (outside the window) is a sibling to fold in when that file is next touched.
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: false
    - "n": 7
      timestamp: "2026-10-08T17:22:48-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: not-addressed
          note: The ownership guard on the resume path exists (abandon.go:164-169), but no test refuses a foreign resume, so nothing fails if the guard is removed.
          round: 7
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

## Round 5 — 2026-10-08T17:03:42-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — The plan has no file.go:NNN anchors left (grep found none); functions are named and a Revisions entry records it.

### Raised

- **BR-7** [Important] `archive-source-of-truth` abandon archives plan files from main, not the kept tip, so a reopen replaces the branch's plan edits
  abandon.go:281-289 copies p.Content from main's view, while the details come from rec.Head. Reopen's resolveArchiveConflicts takes main's side with git rm, and step 4 at abandon.go:490 removes the live copy and moves the stale archived plan back. Rule: for started work, everything abandon archives comes from the kept tip. No test covers plan files in abandon or reopen.
- **BR-8** [Important] `readme-verb-surface` README update appears missing for sdlc abandon, the archive ref and the set-status redirect
  README's Concurrent issue work section lists every ownership and lifecycle verb (claim, unclaim, reclaim, move, issue show) but not abandon or the reopen restore.
- **BR-9** [Minor] `rerun-skips-precondition` the abandon resume path skips the ownership check that plan D7 keeps for reruns
  abandon.go:128-162 checks only a clean tree on resume; a foreign workspace can run dropAbandonedBranch. The containment lease limits the harm.
- **BR-10** [Minor] `config-source-divergence` restoreAbandoned reads WF_HISTORY_DIR/WF_PLANS_DIR from the environment while abandon takes --history-dir/--plans-dir flags
- **BR-11** [Minor] `plan-table-drift` the Core concepts table names ClearCardAbandoned, a 5-argument abandonDecision and a {Ref, Head} record, none of which match the code
- **BR-12** [Minor] `test-claim-exceeds-test` reopen rerun variants after step 2 and after the card write are untested, and the after-card-write rerun never runs finishReopen
  This is the 2nd finding in this family. Rule: every test variant the plan lists maps to a named test in the contract's Proofs or to a Revisions entry that drops it. After a lost response on the card write, the rerun returns "already has that" and skips the mirror commit and the archive-ref deletion.
- **BR-13** [Minor] `record-cleared-without-effect` a forced set-status from punt to open or blocked clears the abandoned record without restoring, leaving an archive ref nothing points to
- **BR-14** [Minor] `shared-remote-ref-read` reading a ref's tip from the remote via ls-remote is duplicated at abandon.go:216, abandon.go:355, boundarypush.go:81 and handoff.go:284

## Round 6 — 2026-10-08T17:18:24-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — keptPlans archives plans from rec.Head; TestAbandonKeepsTheBranchsPlans goes red with the tip overlay removed (archived plan v1).
- BR-8 — addressed — README.md:42-50 now covers boundary pushes, sdlc abandon, refs/ariadne/abandoned/NNNNNN and the set-status working restore.
- BR-9 — not-addressed — The ownership check exists at abandon.go:165-169, but no test covers a foreign resume; with the check disabled every TestAbandon test still passes.
- BR-10 — addressed — The flags are removed; abandon and reopen both read plansDir()/historyDir() from WF_PLANS_DIR/WF_HISTORY_DIR, and helptext has no stale flags.
- BR-11 — addressed — The table now reads Abandoned {Ref, Branch, Head}, SetCardAbandoned (nil clears) and abandonDecision(card, as, today, rec), matching abandoned.go:16 and abandon.go:63.
- BR-12 — addressed — TestReopenRerunVariants covers both variants; disabling the working-status branch in reopenAbandoned turns "after the card write" red.
- BR-13 — addressed — statusDecision refuses leaving terminal except by working when the record is started; TestAbandonedWorkLeavesOnlyByReopen goes red without the guard.
- BR-14 — addressed — The four named sites use remoteRefTip; the pre-existing landing.go:392 raw ls-remote (outside the window) is a sibling to fold in when that file is next touched.

## Round 7 — 2026-10-08T17:22:48-07:00 (claude) — passed

### Disposed

- BR-9 — not-addressed — The ownership guard on the resume path exists (abandon.go:164-169), but no test refuses a foreign resume, so nothing fails if the guard is removed.

## Open findings

- **BR-9** [Minor] `rerun-skips-precondition` the abandon resume path skips the ownership check that plan D7 keeps for reruns
