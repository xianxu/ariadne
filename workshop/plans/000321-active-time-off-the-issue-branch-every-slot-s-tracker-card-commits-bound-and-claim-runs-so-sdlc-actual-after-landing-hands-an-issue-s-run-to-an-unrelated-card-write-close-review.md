# Boundary Review — ariadne#321 (whole-issue close)

| field | value |
|-------|-------|
| issue | 321 — active-time: off the issue branch, every slot's tracker card commits bound and claim runs, so sdlc actual after landing hands an issue's run to an unrelated card write |
| repo | ariadne |
| issue file | workshop/issues/000321-active-time-off-the-issue-branch-every-slot-s-tracker-card-commits-bound-and-claim-runs-so-sdlc-actual-after-landing-hands-an-issue-s-run-to-an-unrelated-card-write.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0f4aaae841e56e941c2b707e612cda74745165d8..a975fa97db1c2aab24ad6a5bd31181ad87d7b431 |
| command | sdlc close --issue 321 |
| reviewer | claude |
| timestamp | 2026-10-10T17:34:14-07:00 |
| verdict | SHIP |

## Review

Tests pass; checking whether the regression goes red without the fix, then wrapping up.

```verdict
verdict: SHIP
confidence: high
```

This change is small and correct. `loadWindowCommits` now builds a `besideHead` set: the commits that are reachable from the extra refs (the tracker) but not from HEAD. It drops any of those whose lead doesn't name `Scope.Issue`, whether or not the measurement is scoped. That fixes the root cause the Problem section names: off the issue branch, another slot's card write could claim a whole run. On an issue branch, the existing `own`/`names` rule already filtered these commits, so the new filter changes nothing there. When there is no measured issue (the standalone `active-time` CLI, which passes no extra refs), the filter is skipped, which matches the zero-Scope contract. The activetime package tests pass (`go test ./cmd/sdlc/internal/activetime/` → ok). I couldn't revert the fix in a scratch worktree because the sandbox blocked creating a temp dir. Reading the code, though, the new test has to fail without the fix: the `#5` card at 00:09:58 would show up in `boundaryShape` and claim the 00:10–00:50 run.

1. **Strengths**
   - `commit.go:160-176` `besideHead` gets the "tracker-only" set from git (`rev-list <refs> ^HEAD`) instead of guessing from subject patterns. Card commits reachable from HEAD stay under the normal rules.
   - `commit.go:117-121` computes `names` once and reuses it in both filters. This follows the same lead rule as #270/#317 (ARCH-DRY: pass).
   - The test pins both the boundary set and the per-issue result from `Compute`. That second check is the user-visible symptom (#9 = 40 min, #5 = 0), so the test covers more than internal state.
   - The Scope doc comment and the atlas paragraph were updated together and describe the code accurately.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `besideHead` runs `rev-list` on the tracker's whole history on every measurement, with no window bound. The cost is negligible today, and `git log` already reads all of it. Only worth noting if the tracker grows large.
   - Unscoped, the tracker's lead-less "neutral" boundary commits are now dropped as well, because `names` is false for them. That's the right call for another slot's writes, but the atlas text says "a card commit counts only when its lead names the measured issue", and nothing in it calls out the neutral case. Optional wording tweak.
   - Spec bullets 2 and 3 (landed-branch scoping and `selectClaimant` preferring inside claimants) are explicitly out of scope, with the TL decision recorded in the Log. Main's own other-issue commits can still bound runs in unscoped mode. Not a finding against this issue, but a filed follow-up would keep it visible.

5. **Test coverage**
   - The new test covers the unscoped case. The scoped case with a tracker ref is already covered by the existing `own`/`names` path, and the branch-scope tests stay green.
   - The Done-when replay for #317 (0.39h) depends on the live transcripts, so the Log entry is the only evidence for it. I didn't reproduce it.

6. **Architecture**
   - ARCH-DRY: pass. It reuses `LeadLocalNums` and the same `names` predicate.
   - ARCH-PURE: pass. The new git IO is one thin helper next to the existing `branchCommits`, and the filtering itself is a pure map lookup.
   - ARCH-PURPOSE: pass. "In every scope" is delivered, and the deferred items are separable extensions, not the point of this issue.

7. **Plan revisions:** none needed.

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-precision
    title: |
      Atlas wording omits that lead-less tracker commits are also dropped unscoped
    detail: |
      With Scope.Issue set, a tracker-only commit with no lead (a neutral boundary) is now filtered too, since names is false. The atlas and Scope comment describe only the other-issue case.
  - id: new
    severity: Minor
    family: unbounded-history-read
    title: |
      besideHead rev-lists the entire tracker history on every measurement
    detail: |
      Not window-bounded; negligible now and git log already reads it all, but worth noting if the tracker grows.
```
