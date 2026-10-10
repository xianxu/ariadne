---
gate: boundary-review
issue: 304
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-09T21:18:02-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Tasks 1, 2, 3, 6, 9 and 10 list test cases in prose; compress each to one strategy line plus its mutation guard
          detail: (carried from plan-quality PQ-3, deferred to the boundary review)
          family: prose-test-enumeration
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-10-09T21:18:02-07:00"
      agent: claude
      findings:
        - id: BR-2
          severity: Important
          title: Legacy trailer fallback treats a binary-committed not-run milestone as a review boundary
          detail: With no stamped round, gatherWindowFacts uses previousReviewBoundary, which matches the not-run evidence commit that milestone-close --no-judge now writes itself (7dfde264). M1 closed with --no-judge, then M2, means M2's window starts after M1's unreviewed work, which breaks D5. Exclude not-run trailers (or gate the fallback to pre-#304 ledgers) and add a fixture.
          family: unreviewed-work-escapes-window
          round: 2
        - id: BR-3
          severity: Important
          title: reviewMainBase silently returns empty on criss-cross or merge-base error, mislabelling an over-covering window as interdiff
          detail: len(bases) != 1 or a git error returns empty; the milestone path then sets Rebased = H_r and prints "interdiff since Mx review" while diffing H_r..HEAD including merged main. D2 requires an error naming the bases and branchPatchFallback; the comment claiming RebasedReviewedBase names it is false on this path. Reuse an exported gitx.SoleMergeBase and route its error into RebaseErr.
          family: window-fallback-must-be-named
          round: 2
        - id: BR-4
          severity: Minor
          title: reviewMainRef falls back to the stale MergeBaseWithMain without a warning when the tracker or snapshot fails
          family: silent-degrade-to-stale-main
          round: 2
        - id: BR-5
          severity: Minor
          title: Ledger reviewed value is passed to git without a hex SHA check
          detail: Validate with isResolvedSHA in latestReviewedFor so a hand-edited ledger cannot hand git an option-like or ref-like argument.
          family: untrusted-persisted-input-parse
          round: 2
        - id: BR-6
          severity: Minor
          title: milestone-close help says only pre-#304 milestones use the trailer boundary; it omits that --no-judge now commits not-run evidence
          family: help-text-drift
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-09T21:33:52-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: withdrawn
          note: Declined on purpose in Revisions (TL-approved contract); the named tests exist and are mutation-checked, so reshaping the plan prose after implementation adds nothing.
          round: 3
        - id: BR-2
          disposition: addressed
          note: The grep is anchored to the finalizing set from the verdict model (milestoneclose.go:357-363); TestMilestoneWindow_NotRunEvidenceIsNotABoundary fails when the old grep is restored.
          round: 3
        - id: BR-3
          disposition: addressed
          note: SoleMergeBase is exported and its error goes to MainBaseErr, giving a named windowBranchFallback; TestReviewWindow_CrissCrossIsANamedFallback fails with the branch removed.
          round: 3
        - id: BR-4
          disposition: not-addressed
          note: Code fixed (reviewMainRef returns a stale note that joinNotes adds to the window note), but no test fails without it; add a fixture whose openTracker or Snapshot fails.
          round: 3
        - id: BR-5
          disposition: not-addressed
          note: Code fixed (isResolvedSHA guard in latestReviewedFor), but no test covers the call site; add a ledger with a hand-edited reviewed value and assert the branch-patch fallback.
          round: 3
        - id: BR-6
          disposition: addressed
          note: helptext/milestone-close.md now says a --no-judge skip commits its not-run evidence, pins nothing, and never bounds a window; it matches previousReviewBoundary.
          round: 3
      boundary: M1
      recipe: milestone-review
      reviewed: 1f37e2eab9fd366ca11908bfd897f3154287f18f
      blocked: false
    - "n": 4
      timestamp: "2026-10-09T22:04:19-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: reviewMainRef sets a "main as last fetched" note that joinNotes threads into every window kind, and the publish gate cwarns it; TestReviewWindow_NamesAStaleMainWhenTheFetchFails fails without it.
          round: 4
        - id: BR-5
          disposition: addressed
          note: latestReviewedFor rejects non-SHA values via isResolvedSHA; TestMilestoneWindow_RejectsANonSHAReviewedValue goes red without it (rev-parse fails, giving BranchFallback, not BranchPatch).
          round: 4
      findings:
        - id: BR-7
          severity: Minor
          title: A rejected ledger reviewed value falls through to the trailer fallback, mislabelled as a pre-#304 boundary
          detail: '2nd in family. Rule: a ledger fact that is discarded must reach planReviewWindow as a reason, never as absence. Fix: on a bad SHA, set Reviewed and RebaseErr so the window is the named windowBranchFallback; also fix the code comment at reviewwindowplan.go:227.'
          family: window-fallback-must-be-named
          round: 4
        - id: BR-8
          severity: Minor
          title: sweepLegacyPins decides liveness from the local checkout's issuesDir but deletes worktree-shared refs/sdlc/reviewed pins
          detail: A push in one slot can delete pins of an issue live only on another slot's branch. It degrades to over-cover or a publish refusal that asks for a re-close, never to unreviewed code passing.
          family: shared-ref-swept-by-local-view
          round: 4
        - id: BR-9
          severity: Minor
          title: completeOnCard and settleLandedCompletions discard unpin and sweep warnings with a blank assignment
          detail: Contradicts the reviewpin.go header ("every failure is a warning"); the abandon, push and merge sites cwarn theirs.
          family: pin-failure-must-warn
          round: 4
      recipe: milestone-review
      reviewed: cd2ad8c07bf609e362a308cc6be7b799a3a41fd5
      blocked: false
    - "n": 5
      timestamp: "2026-10-09T22:12:50-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: LedgerErr routes to a named windowBranchFallback before the trailer case; table row plus RejectsANonSHAReviewedValue fail without the fix.
          round: 5
        - id: BR-8
          disposition: addressed
          note: unpinArchived unpins exactly the moved issue ids; TestLegacyArchivesEndPins pins 000777 (live elsewhere) and would go red under the old sweep.
          round: 5
        - id: BR-9
          disposition: addressed
          note: Both discards now cwarn; grep shows no remaining blank-assigned pin/unpin/sweep result. Warn-only change, no dedicated test.
          round: 5
      findings:
        - id: BR-10
          severity: Minor
          title: Legacy repos now have no removal path for pins of issues archived by hand or abandoned elsewhere
          detail: Dropping sweepLegacyPins (correctly, for worktree-shared refs) leaves only push/merge archives to end legacy pins; a hand-archived issue keeps up to milestones+1 refs forever. Name the end, e.g. a sweep keyed on refs older than N days whose id has no live file in ANY worktree (git worktree list).
          family: residue-without-removal-path
          round: 5
      recipe: milestone-review
      reviewed: 0f4e1bf53e5a94705da8e339df20d16aab0c7912
      blocked: false
---

# Gate ledger — ariadne#304 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T21:18:02-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `prose-test-enumeration` Tasks 1, 2, 3, 6, 9 and 10 list test cases in prose; compress each to one strategy line plus its mutation guard
  (carried from plan-quality PQ-3, deferred to the boundary review)

## Round 2 — 2026-10-09T21:18:02-07:00 (claude) — BLOCKED

### Raised

- **BR-2** [Important] `unreviewed-work-escapes-window` Legacy trailer fallback treats a binary-committed not-run milestone as a review boundary
  With no stamped round, gatherWindowFacts uses previousReviewBoundary, which matches the not-run evidence commit that milestone-close --no-judge now writes itself (7dfde264). M1 closed with --no-judge, then M2, means M2's window starts after M1's unreviewed work, which breaks D5. Exclude not-run trailers (or gate the fallback to pre-#304 ledgers) and add a fixture.
- **BR-3** [Important] `window-fallback-must-be-named` reviewMainBase silently returns empty on criss-cross or merge-base error, mislabelling an over-covering window as interdiff
  len(bases) != 1 or a git error returns empty; the milestone path then sets Rebased = H_r and prints "interdiff since Mx review" while diffing H_r..HEAD including merged main. D2 requires an error naming the bases and branchPatchFallback; the comment claiming RebasedReviewedBase names it is false on this path. Reuse an exported gitx.SoleMergeBase and route its error into RebaseErr.
- **BR-4** [Minor] `silent-degrade-to-stale-main` reviewMainRef falls back to the stale MergeBaseWithMain without a warning when the tracker or snapshot fails
- **BR-5** [Minor] `untrusted-persisted-input-parse` Ledger reviewed value is passed to git without a hex SHA check
  Validate with isResolvedSHA in latestReviewedFor so a hand-edited ledger cannot hand git an option-like or ref-like argument.
- **BR-6** [Minor] `help-text-drift` milestone-close help says only pre-#304 milestones use the trailer boundary; it omits that --no-judge now commits not-run evidence

## Round 3 — 2026-10-09T21:33:52-07:00 (claude) — passed

### Disposed

- BR-1 — withdrawn — Declined on purpose in Revisions (TL-approved contract); the named tests exist and are mutation-checked, so reshaping the plan prose after implementation adds nothing.
- BR-2 — addressed — The grep is anchored to the finalizing set from the verdict model (milestoneclose.go:357-363); TestMilestoneWindow_NotRunEvidenceIsNotABoundary fails when the old grep is restored.
- BR-3 — addressed — SoleMergeBase is exported and its error goes to MainBaseErr, giving a named windowBranchFallback; TestReviewWindow_CrissCrossIsANamedFallback fails with the branch removed.
- BR-4 — not-addressed — Code fixed (reviewMainRef returns a stale note that joinNotes adds to the window note), but no test fails without it; add a fixture whose openTracker or Snapshot fails.
- BR-5 — not-addressed — Code fixed (isResolvedSHA guard in latestReviewedFor), but no test covers the call site; add a ledger with a hand-edited reviewed value and assert the branch-patch fallback.
- BR-6 — addressed — helptext/milestone-close.md now says a --no-judge skip commits its not-run evidence, pins nothing, and never bounds a window; it matches previousReviewBoundary.

## Round 4 — 2026-10-09T22:04:19-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — reviewMainRef sets a "main as last fetched" note that joinNotes threads into every window kind, and the publish gate cwarns it; TestReviewWindow_NamesAStaleMainWhenTheFetchFails fails without it.
- BR-5 — addressed — latestReviewedFor rejects non-SHA values via isResolvedSHA; TestMilestoneWindow_RejectsANonSHAReviewedValue goes red without it (rev-parse fails, giving BranchFallback, not BranchPatch).

### Raised

- **BR-7** [Minor] `window-fallback-must-be-named` A rejected ledger reviewed value falls through to the trailer fallback, mislabelled as a pre-#304 boundary
  2nd in family. Rule: a ledger fact that is discarded must reach planReviewWindow as a reason, never as absence. Fix: on a bad SHA, set Reviewed and RebaseErr so the window is the named windowBranchFallback; also fix the code comment at reviewwindowplan.go:227.
- **BR-8** [Minor] `shared-ref-swept-by-local-view` sweepLegacyPins decides liveness from the local checkout's issuesDir but deletes worktree-shared refs/sdlc/reviewed pins
  A push in one slot can delete pins of an issue live only on another slot's branch. It degrades to over-cover or a publish refusal that asks for a re-close, never to unreviewed code passing.
- **BR-9** [Minor] `pin-failure-must-warn` completeOnCard and settleLandedCompletions discard unpin and sweep warnings with a blank assignment
  Contradicts the reviewpin.go header ("every failure is a warning"); the abandon, push and merge sites cwarn theirs.

## Round 5 — 2026-10-09T22:12:50-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — LedgerErr routes to a named windowBranchFallback before the trailer case; table row plus RejectsANonSHAReviewedValue fail without the fix.
- BR-8 — addressed — unpinArchived unpins exactly the moved issue ids; TestLegacyArchivesEndPins pins 000777 (live elsewhere) and would go red under the old sweep.
- BR-9 — addressed — Both discards now cwarn; grep shows no remaining blank-assigned pin/unpin/sweep result. Warn-only change, no dedicated test.

### Raised

- **BR-10** [Minor] `residue-without-removal-path` Legacy repos now have no removal path for pins of issues archived by hand or abandoned elsewhere
  Dropping sweepLegacyPins (correctly, for worktree-shared refs) leaves only push/merge archives to end legacy pins; a hand-archived issue keeps up to milestones+1 refs forever. Name the end, e.g. a sweep keyed on refs older than N days whose id has no live file in ANY worktree (git worktree list).

## Open findings

- **BR-10** [Minor] `residue-without-removal-path` Legacy repos now have no removal path for pins of issues archived by hand or abandoned elsewhere
