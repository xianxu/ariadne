---
gate: boundary-review
issue: 599
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: bd3ae620-4e11-445e-83bd-6773c8439031
            repository: github.com/xianxu/ariadne
            issue_id: "000599"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/5603fb7040ed387a
            reviewed_head: ee2923e1bae549470d7de4e33f749b8942424b35
            slot: ariadne:2
            requested_at: "2026-10-24T19:29:43.006158Z"
          phase: finished
          pid: 35705
          started_at: "2026-10-24T19:29:43.107096Z"
          finished_at: "2026-10-24T19:32:42.172399Z"
          session_id: 33f2636a-06fa-4f96-bf50-7c1e6e01b739
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T12:32:42-07:00"
      dispatched: "2026-10-24T12:29:43-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Plan items ticked [x] that the pinned head does not deliver
          detail: 'Every instance in this window: (1) Plan item 4 is ticked, and commit ee2923e1b claims the issue 596 Log names 599, but workshop/issues/000596-flake-testhelperrefreshrunnerresourceexcluded.md has no mention of 599 and is not touched in the window. (2) Plan item 1''s sub-bullet promising a prepareAffectedGraph test is ticked, but no such test exists (the Log admits it). Write the note and add the test, or untick both and add a Revisions entry.'
          family: plan-tick-without-delivery
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: prepareAffectedGraph success-stands change has no regression test
          detail: 'localcheck_graphprepare.go:35 changes behavior, but no test fails without the change. The Log''s claim that no seam can delay teardown without changing Run''s result is wrong. A d.runPhase("ran") call between cmd.Run() and the ctx.Err() read would allow a test: small d.minute, command true, hook sleeps past the deadline, assert a nil error.'
          family: behavior-fix-needs-failing-test
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: localchecks_make.go:105 still lets a post-hoc ctx.Err() override a successful make probe
          detail: This is a sibling of the class the Spec named (ARCH-PURPOSE). After Output and Terminate, ctx.Err() is checked before err, so a probe that succeeded is reported as interrupted if the deadline passes during Terminate. The window is small. Apply the err != nil guard or record why this site is exempt. helper_refresh.go:69 is outside the class, because it only downgrades an optional refresh to a warning.
          family: deadline-verdict-from-command-end
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7ea671fa-5179-407e-be43-4a04ca337bfc
            repository: github.com/xianxu/ariadne
            issue_id: "000599"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/a6f4fc35d24621a0
            reviewed_head: 8d3244d30f21806192c0b77d2c6f1852dbb0ed14
            slot: ariadne:2
            requested_at: "2026-10-24T19:34:50.231883Z"
          phase: finished
          pid: 99379
          started_at: "2026-10-24T19:34:50.328799Z"
          finished_at: "2026-10-24T19:38:34.954173Z"
          session_id: 35178aa6-8ab8-4399-bb33-526ecc47453c
          outcome: returned
      "n": 2
      timestamp: "2026-10-24T12:38:35-07:00"
      dispatched: "2026-10-24T12:34:50-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: The 596 note exists in 5ee14b6b8 on origin/main (lines naming ariadne 599 as the cause), and the Log explains why the branch head lacks it. The graph-prepare test now exists (TestAffectedGraphPreparationSuccessOutlivesItsDeadline).
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'Verified by mutation: dropping the err != nil guard at localcheck_graphprepare.go:38 makes TestAffectedGraphPreparationSuccessOutlivesItsDeadline fail with "a successful preparation became graph preparation: context deadline exceeded".'
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: 'The guard at localchecks_make.go:106 is correct, but nothing fails without it. Cheapest fix: route all three sites (localcheck_run.go:733, localcheck_graphprepare.go:38, localchecks_make.go:106) through one pure helper tested once (ARCH-DRY), or add a phase seam to the probe.'
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 8d3244d30f21806192c0b77d2c6f1852dbb0ed14
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 38a20994-4d09-4ab0-8447-20c8782c40a8
            repository: github.com/xianxu/ariadne
            issue_id: "000599"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/5239bc663be4d362
            reviewed_head: a2f195e26057396982fdc9666f165d5ae6a8a261
            slot: ariadne:2
            requested_at: "2026-10-24T19:41:18.470318Z"
          phase: finished
          pid: 57115
          started_at: "2026-10-24T19:41:18.594469Z"
          finished_at: "2026-10-24T19:42:25.270608Z"
          session_id: e4a8ffa0-eb4b-429f-861d-983dd81f73ce
          outcome: returned
      "n": 3
      timestamp: "2026-10-24T12:42:25-07:00"
      dispatched: "2026-10-24T12:41:18-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-3
          disposition: addressed
          note: All three sites, make probe included (localchecks_make.go:105), now call the pure stoppedBy (localcheck_run.go:737); TestStoppedBy pins the rule; behavior is the same at every site; vet and targeted tests pass.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: a2f195e26057396982fdc9666f165d5ae6a8a261
      blocked: false
---

# Gate ledger — ariadne#599 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T12:32:42-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `plan-tick-without-delivery` (kind: defect) Plan items ticked [x] that the pinned head does not deliver
  Every instance in this window: (1) Plan item 4 is ticked, and commit ee2923e1b claims the issue 596 Log names 599, but workshop/issues/000596-flake-testhelperrefreshrunnerresourceexcluded.md has no mention of 599 and is not touched in the window. (2) Plan item 1's sub-bullet promising a prepareAffectedGraph test is ticked, but no such test exists (the Log admits it). Write the note and add the test, or untick both and add a Revisions entry.
- **BR-2** [Important] `behavior-fix-needs-failing-test` (kind: defect) prepareAffectedGraph success-stands change has no regression test
  localcheck_graphprepare.go:35 changes behavior, but no test fails without the change. The Log's claim that no seam can delay teardown without changing Run's result is wrong. A d.runPhase("ran") call between cmd.Run() and the ctx.Err() read would allow a test: small d.minute, command true, hook sleeps past the deadline, assert a nil error.
- **BR-3** [Minor] `deadline-verdict-from-command-end` (kind: defect) localchecks_make.go:105 still lets a post-hoc ctx.Err() override a successful make probe
  This is a sibling of the class the Spec named (ARCH-PURPOSE). After Output and Terminate, ctx.Err() is checked before err, so a probe that succeeded is reported as interrupted if the deadline passes during Terminate. The window is small. Apply the err != nil guard or record why this site is exempt. helper_refresh.go:69 is outside the class, because it only downgrades an optional refresh to a warning.

## Round 2 — 2026-10-24T12:38:35-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — The 596 note exists in 5ee14b6b8 on origin/main (lines naming ariadne 599 as the cause), and the Log explains why the branch head lacks it. The graph-prepare test now exists (TestAffectedGraphPreparationSuccessOutlivesItsDeadline).
- BR-2 — addressed — Verified by mutation: dropping the err != nil guard at localcheck_graphprepare.go:38 makes TestAffectedGraphPreparationSuccessOutlivesItsDeadline fail with "a successful preparation became graph preparation: context deadline exceeded".
- BR-3 — not-addressed — The guard at localchecks_make.go:106 is correct, but nothing fails without it. Cheapest fix: route all three sites (localcheck_run.go:733, localcheck_graphprepare.go:38, localchecks_make.go:106) through one pure helper tested once (ARCH-DRY), or add a phase seam to the probe.

## Round 3 — 2026-10-24T12:42:25-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — All three sites, make probe included (localchecks_make.go:105), now call the pure stoppedBy (localcheck_run.go:737); TestStoppedBy pins the rule; behavior is the same at every site; vet and targeted tests pass.

## Open findings

(none — every finding has been disposed)
