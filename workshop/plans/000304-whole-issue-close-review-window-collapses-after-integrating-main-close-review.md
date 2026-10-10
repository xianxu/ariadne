# Boundary Review — ariadne#304 (whole-issue close)

| field | value |
|-------|-------|
| issue | 304 — Milestone review window absorbs integrated main; close's printed window misleads |
| repo | ariadne |
| issue file | workshop/issues/000304-whole-issue-close-review-window-collapses-after-integrating-main.md |
| boundary | whole-issue close |
| milestone | — |
| window | b77e971865f485b0b0fca59a03130d7febb04154..cd2ad8c07bf609e362a308cc6be7b799a3a41fd5 |
| command | sdlc close --issue 304 |
| reviewer | claude |
| timestamp | 2026-10-09T22:04:18-07:00 |
| verdict | SHIP |

## Review

I reviewed #304 at the whole-issue close, window `b77e9718..cd2ad8c0`. The verdict is SHIP: nothing blocks the close, and the three Minor findings below don't block a gate.

```verdict
verdict: SHIP
confidence: high
```

The issue's purpose is delivered. Every review window is now worked out from the branch's own changes against main (the "branch patch"), not from a commit range. So merging or rebasing main no longer collapses a window or widens it. M2 extends the same idea to publishing: the publish gate replays each close's reviewed changes onto today's main and compares that with HEAD. A close's ID (its `Close-Token`) now rides in the close commit's message, so the close stays recognised after a rebase and after it lands. Saved references to reviewed commits ("pins") are removed when the issue is done, abandoned or archived. I ran the stat and name-status commands, then read the full M2 code diff, the window planner, the pin lifecycle and the new tests. The targeted tests pass (`TestMilestoneWindow_RejectsANonSHAReviewedValue`, `TestReviewWindow_NamesAStaleMainWhenTheFetchFails`, the publish-gate, pin and tracker-rebase suites, and the rebased-patch tests: `ok … 89.7s`). Both open findings, BR-4 and BR-5, are fixed and each has a test that fails without the fix.

## 1. Strengths
- **The window decision is a pure function.** `planReviewWindow` (`cmd/sdlc/reviewwindowplan.go:76`) is a total, side-effect-free decision over facts gathered separately in `gatherWindowFacts`. The close output, atlas gate, manifest and trailer all use the one window it returns (ARCH-PURE, ARCH-DRY).
- **The publish gate's decision is also pure, and each refusal states its own cause.** `classifyPublishDelta` (`publishgate.go`) is pure and tested row by row. The "newest anchor" rule is generalised by `publishDeltaRank` with no commit counting, and the tests cover the merge case, the rebase case, a post-merge code change (naming only `late.go`), a conflict, and several closes on one branch (`publishgate_test.go:421`).
- **A dropped close now fails loudly.** `refuseUnownedCompletions` turns a squash or amend that loses the close into an explicit refusal; before, the gate just reported "nothing to verify". `TestSquashedCloseRefusesLoudly` covers it and checks how the error is attributed.
- **A reopened issue can't be claimed by its old close.** Matching by close generation is exact, and `TestOlderCloseTokenNeverClaimsANewerClose` pins the #283/#301 supersession rule.
- **Each pin-removal site is tested on its own.** `TestCompleteOnCardEndsPins` tests the done-site removal without the later sweep that would hide a missing removal. The landing tests cover merge-commit, squash and settle.

## 2. Critical findings
None.

## 3. Important findings
None.

## 4. Minor findings
- **A rejected `reviewed:` value is mislabelled** (`reviewwindowplan.go:227`, family `window-fallback-must-be-named`, 2nd finding).
  - **What happens:** when `latestReviewedFor` rejects a hand-edited value, it returns `ok=false`. The caller then takes the old trailer-based fallback whenever a `Review-Verdict` trailer exists.
  - **Why it's wrong:** the window gets labelled "pre-#304 boundary from the Review-Verdict trailer", and the code comment says "branch-patch fallback". Both are wrong for this path.
  - **Rule:** a ledger fact that is discarded must reach `planReviewWindow` as a reason, not as absence.
  - **Fix:** on a bad SHA, set `Reviewed` and `RebaseErr = "ledger reviewed value is not a commit id"`. The window then becomes the named fallback.
- **Pin sweeps can delete another slot's pins** (`reviewpin.go:105` `sweepLegacyPins`, `push.go:681`).
  - **What happens:** pins live under `refs/sdlc/reviewed/`, which every linked worktree shares. The decision about which issues are still live comes from the local checkout's `issuesDir` only.
  - **Effect:** a push in one slot can delete the pins of an issue that is live only on another slot's branch. That only degrades things: the window over-covers and the publish gate refuses until a re-close, so nothing unreviewed gets through.
- **Two pin cleanups throw away their warnings** (`trackercompletion.go:188`, `:220`).
  - **What happens:** `_ = unpinReviewed(...)` and `_ = sweepTrackedPins(...)` discard the warning they return.
  - **Why it's wrong:** the header in `reviewpin.go` promises "every failure is a warning". The other call sites (abandon, push, merge) do print theirs with `cwarn`.

## 5. Test coverage notes
- **BR-5:** without the `isResolvedSHA` check, `--output=/tmp/x` makes `rev-parse` fail, which gives `windowBranchFallback`. The test expects `windowBranchPatch`, so it fails without the fix.
- **BR-4:** the test asserts the "last fetched" note, so it fails if the note is removed.
- **Not covered:**
  - `closeTokenSince`'s time-bounded scan has no direct test for clock skew. It's low risk because a rebase refreshes committer dates.
  - `refuseUnownedCompletions` has no negative test for a branch that legitimately edits another codecomplete issue's details file. That case would also be refused.

## 6. Architectural notes
| Principle | Result | Note |
|---|---|---|
| ARCH-DRY | pass | `indexEvidenceEntry` and `applyEvidenceCommit` were extracted and shared by the milestone and whole-issue evidence paths. |
| ARCH-PURE | pass | — |
| ARCH-PURPOSE | pass | Every consumer of a window (close, milestone-close, publish gate, landing, settle) now works from the branch patch or the close token. |
| ARCH-MOCK | pass | Tests use real git repos and the tracker fake. |
| ARCH-CONSTRAINTS | pass | The main fetch is cached once per process per checkout. |
| ARCH-SECURE | pass | The ledger SHA is validated; `CommitsWithCloseToken` matches the trailer exactly instead of searching message bodies. |
| ARCH-ORDER | pass | The close lifecycle stays inside the tracker's step driver. |
| ARCH-FUNERAL | pass with a Minor | Every pin family has a named end. The cross-slot sweep scope is the Minor above. |

## 7. Plan revision recommendations
None. The plan still matches the code.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      reviewMainRef sets a "main as last fetched" note that joinNotes threads into every window kind, and the publish gate cwarns it; TestReviewWindow_NamesAStaleMainWhenTheFetchFails fails without it.
  - id: BR-5
    disposition: addressed
    note: |
      latestReviewedFor rejects non-SHA values via isResolvedSHA; TestMilestoneWindow_RejectsANonSHAReviewedValue goes red without it (rev-parse fails, giving BranchFallback, not BranchPatch).
findings:
  - id: new
    severity: Minor
    family: window-fallback-must-be-named
    title: |
      A rejected ledger reviewed value falls through to the trailer fallback, mislabelled as a pre-#304 boundary
    detail: |
      2nd in family. Rule: a ledger fact that is discarded must reach planReviewWindow as a reason, never as absence. Fix: on a bad SHA, set Reviewed and RebaseErr so the window is the named windowBranchFallback; also fix the code comment at reviewwindowplan.go:227.
  - id: new
    severity: Minor
    family: shared-ref-swept-by-local-view
    title: |
      sweepLegacyPins decides liveness from the local checkout's issuesDir but deletes worktree-shared refs/sdlc/reviewed pins
    detail: |
      A push in one slot can delete pins of an issue live only on another slot's branch. It degrades to over-cover or a publish refusal that asks for a re-close, never to unreviewed code passing.
  - id: new
    severity: Minor
    family: pin-failure-must-warn
    title: |
      completeOnCard and settleLandedCompletions discard unpin and sweep warnings with a blank assignment
    detail: |
      Contradicts the reviewpin.go header ("every failure is a warning"); the abandon, push and merge sites cwarn theirs.
```

---

## Re-review — 2026-10-09T22:12:50-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 304 — Milestone review window absorbs integrated main; close's printed window misleads |
| repo | ariadne |
| issue file | workshop/issues/000304-whole-issue-close-review-window-collapses-after-integrating-main.md |
| boundary | whole-issue close |
| milestone | — |
| window | b77e971865f485b0b0fca59a03130d7febb04154..0f4e1bf53e5a94705da8e339df20d16aab0c7912 |
| command | sdlc close --issue 304 |
| reviewer | claude |
| timestamp | 2026-10-09T22:12:50-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The review covered the full window `b77e9718..0f4e1bf5`: I read the stat and name-status, then inspected the last commit `0f4e1bf5` in detail, because that is the one claiming to fix the open findings. All three open Minors (BR-7, BR-8, BR-9) are fixed. The two behavior fixes, BR-7 and BR-8, each have a regression test that would fail if the fix were reverted. BR-9 is a warning-only change; its class is now uniform across the tree, but no test covers it. The build is clean, and the targeted tests pass (`TestPlanReviewWindow|TestMilestoneWindow|TestLegacyArchivesEndPins|Pin`, 21s). One new Minor: removing the legacy sweep leaves no removal path for pins in a specific legacy-repository case (detailed below).

1. **Strengths**
   - **The planner stays pure.** `cmd/sdlc/reviewwindowplan.go:98` turns `LedgerErr` into a named `windowBranchFallback`, ahead of the trailer case. The table test row "unusable ledger: named fallback" pins this with `LegacyTrailerBase` set, so a regression to the trailer path would fail the test.
   - **A missing ledger is not treated as unreadable.** `readGateLedger` (`planreview.go:62`) returns an empty ledger when the file doesn't exist. So a first milestone never wrongly becomes a fallback.
   - **Unpinning follows what the archive moved.** `unpinArchived` (`reviewpin.go:106`) ends pins only for the issues this archive actually moved. It skips duplicates and non-issue moves; plan and sidecar paths resolve to the same id and are deduped. This removes the problem of judging worktree-shared refs from one checkout's view.
   - **Every pin call site now surfaces failures.** `grep '_ = (unpin|sweep|pin)'` finds nothing. All `pinReviewed`, `unpinReviewed` and `sweep*` call sites pass their warning to `cwarn`, so the `reviewpin.go` header's claim that every failure is a warning now holds.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **Legacy pins can outlive their issue (ARCH-FUNERAL).** With `sweepLegacyPins` removed, a legacy (non-tracker) repository has no removal path for pins of an issue that was archived by hand or abandoned on another slot's branch. Only `sdlc push`/`merge` archives remove them.
     - Bound: at most milestones + 1 refs per such issue, each keeping old commits reachable.
     - This was the right trade for safety, but the end of that residue is now unnamed.
   - **Warnings go straight to `os.Stderr`.** `trackercompletion.go:190,224` writes them there rather than to an injected writer. `trackerEnv` carries no stderr, so this is acceptable, but the lines can't be observed in tests.

5. **Test coverage**
   - **BR-7:** `TestMilestoneWindow_RejectsANonSHAReviewedValue` now asserts a named fallback containing "not a commit id". Before the fix the window was a branch patch with no note, so the test would fail without the fix.
   - **BR-8:** `TestLegacyArchivesEndPins` adds `000777`, an issue live only on another slot. The old sweep would have deleted its pin.
   - **Not tested:** the unreadable-ledger path in `gatherWindowFacts` is only covered by the pure planner table, and the BR-9 warnings are not covered at all.

6. **Architecture**
   - **Pass:**
     - ARCH-DRY: one `unpinArchived` helper serves both archive sites.
     - ARCH-PURE: the planner stays pure and the facts-gathering function is a thin shell.
     - ARCH-PURPOSE: the BR-9 class is swept, not just the instances the finding named.
     - ARCH-SECURE: non-SHA values still never reach git, and the failure now shows up in the window note.
   - **N/A:**
     - ARCH-MOCK: no new external seam.
     - ARCH-CONSTRAINTS: no new hot path.
     - ARCH-ORDER: this commit adds no state between events. The pins are idempotent ref writes.
   - **Flag:** ARCH-FUNERAL, as in the first Minor above.

7. **Plan revisions:** none required. The atlas (`pre-merge-checks.md:55`, `sdlc-binary.md:1259`) already says pins end "at done, at abandon and by both legacy archives, and swept by settle", which matches the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      LedgerErr routes to a named windowBranchFallback before the trailer case; table row plus RejectsANonSHAReviewedValue fail without the fix.
  - id: BR-8
    disposition: addressed
    note: |
      unpinArchived unpins exactly the moved issue ids; TestLegacyArchivesEndPins pins 000777 (live elsewhere) and would go red under the old sweep.
  - id: BR-9
    disposition: addressed
    note: |
      Both discards now cwarn; grep shows no remaining blank-assigned pin/unpin/sweep result. Warn-only change, no dedicated test.
findings:
  - id: new
    severity: Minor
    family: residue-without-removal-path
    title: |
      Legacy repos now have no removal path for pins of issues archived by hand or abandoned elsewhere
    detail: |
      Dropping sweepLegacyPins (correctly, for worktree-shared refs) leaves only push/merge archives to end legacy pins; a hand-archived issue keeps up to milestones+1 refs forever. Name the end, e.g. a sweep keyed on refs older than N days whose id has no live file in ANY worktree (git worktree list).
```
