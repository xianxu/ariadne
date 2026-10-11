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

---

## Re-review — 2026-10-10T17:37:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 321 — active-time: off the issue branch, every slot's tracker card commits bound and claim runs, so sdlc actual after landing hands an issue's run to an unrelated card write |
| repo | ariadne |
| issue file | workshop/issues/000321-active-time-off-the-issue-branch-every-slot-s-tracker-card-commits-bound-and-claim-runs-so-sdlc-actual-after-landing-hands-an-issue-s-run-to-an-unrelated-card-write.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0f4aaae841e56e941c2b707e612cda74745165d8..b2d2fdee2d724bb31d17f4cba1174dcdfa07ba29 |
| command | sdlc close --issue 321 |
| reviewer | claude |
| timestamp | 2026-10-10T17:37:04-07:00 |
| verdict | SHIP |

## Review

I'm checking that the scratch revert test I just tried left the repo untouched. Its temp directory couldn't be created, so the test never ran.

The failed `mktemp` left `$d` empty, so git created a scratch worktree at `./b2d2fdee`. I'll run the revert test there and then remove it.

```verdict
verdict: SHIP
confidence: high
```

The change is small and correct, and the regression test proves it. `besideHead` uses `rev-list <extraRefs> ^HEAD` to collect the commits that only the tracker reaches. `loadWindowCommits` then drops each of those unless its lead names `Scope.Issue`, and it does this in every scope. That is the root cause the Spec names. The fix changes nothing on an issue branch, where the `own`/`names` rule already filtered these commits. When `Scope.Issue` is empty, `besideHead` returns nil, so the zero Scope still keeps every commit.

I checked that the test goes red. In a scratch worktree I disabled the new guard (`if false && beside[...]`), and `TestTrackerCommitsOfOtherIssuesDoNotClaim` failed: #5's 00:09:58 card became a boundary and `PerIssue = map[5:40]`. On the real tree it passes. The scratch worktree is removed and the tree is clean.

1. **Strengths**
   - `commit.go:160-176`: the "beside HEAD" set comes from git's own reachability, not from guessing at subject text. Tracker commits that HEAD can reach keep the normal rules.
   - `commit.go:117-121`: `names` is computed once and shared by both filters. That keeps a single lead rule, the same one #270 and #317 use.
   - The test pins both the boundary set and the per-issue split that `Compute` returns. The second is the symptom a user would see.
   - The Scope comment and the atlas now both say the lead-less case is dropped, which matches the code.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - The `#321` sentences added to the Scope comment (`commit.go:53`) and the atlas paragraph (`atlas/workflow/sdlc-binary.md:1031-1032`) aren't re-wrapped to the width of the surrounding lines. Style only.

5. **Test coverage**
   - The unscoped case where another issue's card claims a run is pinned, and the test fails without the fix.
   - No test pins the lead-less tracker card being dropped when the measurement is unscoped. That behaviour is now documented, and it follows from the same `names` predicate. Adding a lead-less card to the existing fixture would pin it cheaply.
   - The Done-when replay for #317 (0.39h) depends on live transcripts. The only evidence for it is the Log entry, and I did not reproduce it.

6. **Architecture**
   - ARCH-DRY: pass. It reuses `LeadLocalNums` and the existing `names` rule.
   - ARCH-PURE: pass. The new git IO is one thin helper next to `branchCommits`, and the filtering is a pure set lookup.
   - ARCH-PURPOSE: pass. The fix applies in every scope, as the issue requires. Landed-branch scoping and the `selectClaimant` change are separate extensions, and the Log records the TL's decision to leave them out.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Scope comment (commit.go:53, "one with no lead is dropped too") and atlas ("or a card with no lead, is not a boundary at all") now match the code: names is false for an empty lead, so beside-only lead-less commits are skipped.
  - id: BR-2
    disposition: withdrawn
    note: |
      git log over the same refs already reads the whole tracker history; Log records about 50 ms on ariadne. Not worth window-bounding now.
```
