# Boundary Review — ariadne#304 (milestone M1)

| field | value |
|-------|-------|
| issue | 304 — Milestone review window absorbs integrated main; close's printed window misleads |
| repo | ariadne |
| issue file | workshop/issues/000304-whole-issue-close-review-window-collapses-after-integrating-main.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | b77e971865f485b0b0fca59a03130d7febb04154..7dfde2644273c09dcf3e86465202672f9e98228a |
| command | sdlc milestone-close --issue 304 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-09T21:18:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 does what the plan says. The milestone and whole-issue windows are planned on the branch patch. `gitx.RebasedReviewedBase` replays the reviewed head onto today's main using `merge-tree --merge-base`, and the replay is deterministic. A head is stamped as reviewed only when the round finalizes. Milestone-close now commits its own evidence and pins the reviewed head. Both Done-when items hold. The `TestMilestoneWindow_ExcludesIntegratedMain` sub-tests pass (merge, rebase and conflict, run locally), and #269's spec line is on `origin/main` (`1b2f56b1`, line 73).

The targeted suites are green. I ran the `gitx`, `gatestate` and `cmd/sdlc` window, ledger, pin, evidence and pathspec tests.

Two contract gaps keep this from SHIP; both are cheap to fix:
1. A post-#304 issue whose first milestone was closed with `--no-judge` still takes its next window from the legacy trailer. That skips the unreviewed work, which breaks D5.
2. When `merge-base` finds more than one base (criss-cross), or fails, `reviewMainBase` returns nothing silently. The window is then labelled "interdiff" although it covers everything main merged in. D2 says this case should produce a named fallback.

**Strengths**
- `planReviewWindow` (`cmd/sdlc/reviewwindowplan.go`) is a pure, total decision that gets its facts from a thin IO shell. One `reviewWindow` value feeds the atlas gate, the review manifest, the trailer and the printed line, which keeps the #58 "same window" property.
- `RebasedReviewedBase` tests run against real git and cover the cases that matter: no integration, merge, rebase, new work, conflict resolution, determinism, criss-cross, and explicit base vs implicit.
- `roundAdvancesBoundary` (`cmd/sdlc/boundaryledger.go:256`) is one predicate that matches the finalize conditions. It has a table test with seven rows, including a waived ledger block and a literal `HEAD`.
- `commitMilestoneEvidence` shares `indexEvidenceEntry` and `applyEvidenceCommit` with the whole-issue path rather than copying them (ARCH-DRY). The compare-and-swap refusal has its own test, and the e2e test checks that staged unrelated work stays staged and that the pushed tip is the evidence commit.
- `reviewTrailers` names the reviewed head rather than the synthetic commit S, and there is a test for it.

**Critical:** none.

**Important**
1. **`cmd/sdlc/reviewwindowplan.go`, `gatherWindowFacts`, and `cmd/sdlc/milestoneclose.go:343` (`previousReviewBoundary`).**
   - **What happens:** the legacy fallback runs whenever no ledger round carries a reviewed head, not only for pre-#304 issues. Since commit `7dfde264`, a `--no-judge` milestone-close commits `#N M1: close` with `Review-Verdict: not-run` itself.
   - **Failure case:** M1 closed with `--no-judge`, then M2 is closed. There is no stamped round, so `previousReviewBoundary` returns M1's not-run commit. The M2 window starts there, and M1's work is never reviewed. That violates D5: "a `--no-judge` skip never stamps it, so the next window still covers everything unreviewed."
   - **Fix:** exclude `Review-Verdict: not-run` commits from `previousReviewBoundary`, or skip the legacy fallback when the ledger exists and its last round post-dates #304. Add a fixture: M1 `--no-judge`, then M2, and the window must include M1's files.
2. **`cmd/sdlc/reviewwindowplan.go`, `reviewMainBase`.**
   - **What happens:** a criss-cross history (`len(bases) != 1`) or a `merge-base` error returns `""`. The milestone path treats that as "on main" and sets `Rebased = H_r`, then prints "interdiff since M1 review". In fact it diffs `H_r..HEAD`, which includes everything main merged in.
   - **Why it matters:** the code comment says "RebasedReviewedBase names it", but `RebasedReviewedBase` is never called on this path. D2 promises an error that names the bases and a `branchPatchFallback`. This is the same misleading-output class the issue exists to fix.
   - **Fix:** use one exported `gitx.SoleMergeBase` (which also covers ARCH-DRY with `soleMergeBase` in `rebasedpatch.go`). Have `reviewMainBase` return its error, and set `RebaseErr` from it so the planner picks `windowBranchFallback`. Add a `TestPlanReviewWindow` row for this case, plus a fixture.

**Minor**
- `reviewMainRef` silently falls back to the possibly stale `gitx.MergeBaseWithMain` when `openTracker` or `Snapshot()` fails. Consider a `cwarn` so a failed fetch doesn't quietly bring back the stale-ref widening.
- The `reviewed:` value read from the ledger is passed to git without a hex check. Parse it with `isResolvedSHA` in `latestReviewedFor` (ARCH-SECURE: the ledger is a file someone can edit by hand).
- `milestone-close.md` help says only pre-#304 milestones use the trailer boundary, and doesn't mention that `--no-judge` now commits not-run evidence. Update it together with Important #1.
- The pins (`refs/sdlc/reviewed/<id>/*`) are created in M1, but nothing removes them until M2 Task 11. That is acceptable at this boundary; just make sure M2 delivers it.

**Test coverage notes**
- Missing: a `--no-judge` first milestone followed by M2 (Important #1).
- Missing: a criss-cross or `merge-base`-failure path through `planBoundaryWindow` (Important #2). `TestRebasedReviewedBase_CrissCrossRefuses` only exercises the gitx layer, which the planner skips in this case.
- Plan Task 6 (c), FIX-THEN-SHIP at a milestone, and (d), legacy mode with no evidence commit and a stamped ledger, have no dedicated tests that I found. `TestPersistBoundaryRound_*` covers the stamping part only.

**Architecture**
- ARCH-DRY: flagged. `reviewMainBase` re-parses `merge-base --all` beside `gitx.soleMergeBase` (Important #2). Otherwise it passes, because the evidence helpers were extracted.
- ARCH-PURE: passes. The planner, formatter, `LatestReviewed` and `roundAdvancesBoundary` are pure; IO stays in `gatherWindowFacts` and `planBoundaryWindow`.
- ARCH-PURPOSE: passes. The boundary now comes from the ledger, the window from the branch patch, and the printed line says what it is. The publish gate and close binding are M2 scope, as planned.
- ARCH-MOCK: passes. git is exercised for real in hermetic repos, and the tracker fixtures already in the suite are reused.
- ARCH-CONSTRAINTS: passes. It adds one `merge-tree` plus one `commit-tree` per boundary invocation, and fetches main at most once per process thanks to the cache.
- ARCH-SECURE: minor flag on the unvalidated `reviewed:` value (Minor list).
- ARCH-ORDER: passes. Evidence and pin are written before the push. A failed evidence commit falls back to the paste protocol, and the branch move is a compare-and-swap, so it never overwrites a branch that moved.
- ARCH-FUNERAL: passes for this milestone. S is collected by gc, and pin removal is scheduled in M2 Task 11.

**Atlas / README:** the atlas is updated (`gate-state.md`, `sdlc-binary.md`). No new user-typed surface was added, so README doesn't need a change.

**Plan revision recommendations**
- Add a `## Revisions` entry recording that D6's legacy fallback must ignore `not-run` trailers, because the binary now commits them for `--no-judge` milestones (`7dfde264`).

```findings
findings:
  - id: new
    severity: Important
    family: unreviewed-work-escapes-window
    title: |
      Legacy trailer fallback treats a binary-committed not-run milestone as a review boundary
    detail: |
      With no stamped round, gatherWindowFacts uses previousReviewBoundary, which matches the not-run evidence commit that milestone-close --no-judge now writes itself (7dfde264). M1 closed with --no-judge, then M2, means M2's window starts after M1's unreviewed work, which breaks D5. Exclude not-run trailers (or gate the fallback to pre-#304 ledgers) and add a fixture.
  - id: new
    severity: Important
    family: window-fallback-must-be-named
    title: |
      reviewMainBase silently returns empty on criss-cross or merge-base error, mislabelling an over-covering window as interdiff
    detail: |
      len(bases) != 1 or a git error returns empty; the milestone path then sets Rebased = H_r and prints "interdiff since Mx review" while diffing H_r..HEAD including merged main. D2 requires an error naming the bases and branchPatchFallback; the comment claiming RebasedReviewedBase names it is false on this path. Reuse an exported gitx.SoleMergeBase and route its error into RebaseErr.
  - id: new
    severity: Minor
    family: silent-degrade-to-stale-main
    title: |
      reviewMainRef falls back to the stale MergeBaseWithMain without a warning when the tracker or snapshot fails
  - id: new
    severity: Minor
    family: untrusted-persisted-input-parse
    title: |
      Ledger reviewed value is passed to git without a hex SHA check
    detail: |
      Validate with isResolvedSHA in latestReviewedFor so a hand-edited ledger cannot hand git an option-like or ref-like argument.
  - id: new
    severity: Minor
    family: help-text-drift
    title: |
      milestone-close help says only pre-#304 milestones use the trailer boundary; it omits that --no-judge now commits not-run evidence
```

---

## Re-review — 2026-10-09T21:33:52-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 304 — Milestone review window absorbs integrated main; close's printed window misleads |
| repo | ariadne |
| issue file | workshop/issues/000304-whole-issue-close-review-window-collapses-after-integrating-main.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | b77e971865f485b0b0fca59a03130d7febb04154..1f37e2eab9fd366ca11908bfd897f3154287f18f |
| command | sdlc milestone-close --issue 304 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-09T21:33:52-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both Important findings from round 1 are fixed, and each fix is guarded by a regression test that fails without it. I checked this by reverting each fix in a scratch copy made with `git archive`, outside the worktree:
- **BR-2:** putting back the old unanchored `--grep=^Review-Verdict:` makes `TestMilestoneWindow_NotRunEvidenceIsNotABoundary` fail with "a not-run trailer became a boundary".
- **BR-3:** removing the `MainBaseErr` short-circuit makes `TestReviewWindow_CrissCrossIsANamedFallback` fail. The fixture really does produce a criss-cross history in this environment; the test passed rather than skipping.

All targeted window, milestone-close and pin tests pass at HEAD, including the new FIX-THEN-SHIP evidence subtest and the legacy no-commit stamp test. The remaining open items are Minor. BR-4 and BR-5 have the right code but no test that fails without them. BR-1 was declined on purpose. None of them blocks the boundary.

1. **Strengths**
   - `milestoneclose.go:357-363`: the set of verdicts that count as a boundary comes from `vocab.Verdict().Categories["finalizing"]` and is passed through `regexp.QuoteMeta`, so no hand-written copy of the verdict model was added (ARCH-DRY, ARCH-PURPOSE).
   - `gitx.SoleMergeBase` is now the only place that parses merge bases. `reviewMainBase` no longer has its own copy, and the error names the bases (`criss-cross history: merge bases …`).
   - The branch-patch fallback in `planReviewWindow` stays pure. The stale-main and criss-cross conditions arrive as facts (`MainStale`, `MainBaseErr`) and come out as notes, so the IO stays in `gatherWindowFacts` (ARCH-PURE).
   - `gatherWindowFacts` returns early on `MainBaseErr` for milestone windows. That closes the path where an empty `MainBase` used to give `Rebased = H_r` labelled as an interdiff.
   - The lesson added to `workshop/lessons.md` names the general rule: when a verb starts producing an artifact, re-read every reader of it.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `milestoneclose.go:362`: the `$` anchor no longer matches 2 historical trailers that carry commentary, e.g. `Review-Verdict: SHIP (follow-up on Minor findings)`. The window then over-covers, which is the safe direction, so this is not raised as a finding.
   - `reviewMainBase` checks for criss-cross only against `TrunkRef()`, but `BranchPoint` may pick local `main`'s merge base when it differs from the trunk's. This is an edge case only.

5. **Test coverage**
   - BR-2 and BR-3 are guarded as shown by the reversions above.
   - The stale-main note (BR-4) and the `latestReviewedFor` SHA guard (BR-5) have no test that fails without them. `isResolvedSHA` itself is tested in `reviewanchor_test.go:119`, but the call site is not.

6. **Architecture**
   - **ARCH-DRY:** pass. The merge-base parser and the verdict set each have one source.
   - **ARCH-PURE:** pass.
   - **ARCH-PURPOSE:** pass. The not-run exclusion applies to every reader of the trailer boundary that I found (`previousReviewBoundary` is the only one).
   - **ARCH-MOCK:** pass. Tests use real git fixture repositories behind the existing `gitx` seam.
   - **ARCH-CONSTRAINTS:** pass. The fetch happens at most once per process and checkout.
   - **ARCH-SECURE:** pass, with BR-5 now validating the ledger value at the code level.
   - **ARCH-ORDER:** pass. The window plan is a pure function of facts and holds no state between events.
   - **ARCH-FUNERAL:** pass. No new durable artifacts in this round; the pins were reviewed in earlier rounds.

7. **Plan revisions:** none needed. The 2026-10-09 Revisions entry matches the code.

```findings
dispose:
  - id: BR-1
    disposition: withdrawn
    note: |
      Declined on purpose in Revisions (TL-approved contract); the named tests exist and are mutation-checked, so reshaping the plan prose after implementation adds nothing.
  - id: BR-2
    disposition: addressed
    note: |
      The grep is anchored to the finalizing set from the verdict model (milestoneclose.go:357-363); TestMilestoneWindow_NotRunEvidenceIsNotABoundary fails when the old grep is restored.
  - id: BR-3
    disposition: addressed
    note: |
      SoleMergeBase is exported and its error goes to MainBaseErr, giving a named windowBranchFallback; TestReviewWindow_CrissCrossIsANamedFallback fails with the branch removed.
  - id: BR-4
    disposition: not-addressed
    note: |
      Code fixed (reviewMainRef returns a stale note that joinNotes adds to the window note), but no test fails without it; add a fixture whose openTracker or Snapshot fails.
  - id: BR-5
    disposition: not-addressed
    note: |
      Code fixed (isResolvedSHA guard in latestReviewedFor), but no test covers the call site; add a ledger with a hand-edited reviewed value and assert the branch-patch fallback.
  - id: BR-6
    disposition: addressed
    note: |
      helptext/milestone-close.md now says a --no-judge skip commits its not-run evidence, pins nothing, and never bounds a window; it matches previousReviewBoundary.
```
