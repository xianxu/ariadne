# Boundary Review — ariadne#620 (whole-issue close)

| field | value |
|-------|-------|
| issue | 620 — flake: TestKnownFlakeGate |
| repo | ariadne |
| issue file | workshop/issues/000620-flake-testknownflakegate.md |
| boundary | whole-issue close |
| milestone | — |
| window | ce04fa47f783aa2c8e42e0cebc4f476ff7f9ed26..71f220cc1ddb633eda1e9f80722d80acdfbe93e1 |
| command | sdlc close --issue 620 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T01:07:59-07:00 |
| verdict | SHIP |

## Review

The test passes (9.8 s, which would have failed the old 8 s bound). Next I'm running the mutation check: I'll remove the check-timeout bound and confirm the test goes red.

```verdict
verdict: SHIP
confidence: high
```

This boundary is ready to ship. The diff touches one test file, `cmd/sdlc/localcheck_flakes_test.go`. The old BR-16 subtest proved the bound by checking that `runLocalCheck` finished in under 8 s, which made it a timing flake. It now proves the bound with the rerun's saved `TimedOut` record and its `RerunInconclusive` outcome. That follows the Spec's root cause: the runner already records when the deadline fires. No retry loop or longer timeout was added, as Done-when requires. I checked the main claims myself:

- **Passes on this machine:** the subtest passed in 9.84 s, so the old 8 s check would have failed here.
- **Mutation check:** I changed `rerunEnvelope` to `return 1, knownFlakeRerunTimeout`, which drops the check-timeout bound. The subtest failed after 124.76 s: the outcome was `known-flake-rerun` instead of inconclusive. I then restored the file, and the tree is clean.

The only finding is a Minor comment imprecision.

1. **Strengths**
   - `localcheck_flakes_test.go:287-298`: the test now checks the runner's own deadline event, with `TimedOut` on the attempt whose `RerunOf` is set and whose `Check` is `suiteRerunCheck`. This holds however loaded the machine is.
   - The 120 s sleep is longer than any real run, so a rerun without the bound finishes and passes, and the outcome check catches it (confirmed by the mutation).
   - The `TimedOut` check is meaningful: `knownFlakeRerunTimeout = time.Hour` means the only deadline that can fire is the check's declared one (`min(knownFlakeRerunTimeout, declared)` in `localcheck_affected_rerun.go:39`).
   - `rerun = &a` inside the loop is safe because go.mod declares Go 1.26, which gives each loop iteration its own variable.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `localcheck_flakes_test.go:277-278`: the comment says the failing suite command "shares the check's budget". In fact the parent and the rerun each get the full declared timeout separately (`rerunEnvelope` uses `TimeoutMinutes * minute` on its own). The 5 s minute gives margin to each of them, not to a shared budget. I found only this one instance in the window.
   - The subtest's minimum run time goes from about 1 s to about 5 s, since it now waits for a real 5 s deadline. That is acceptable for a deterministic check.

5. **Test coverage**
   - Both cases of the Spec's claim are exercised: the bounded case runs as the test, and the unbounded case is the mutation, which goes red.
   - "Closing this issue unregisters the tests" is part of the close process, not code in this window.

6. **Architecture**
   - ARCH-DRY: pass. It reuses the existing `gateAttempts`, `scriptRerun` and `suiteFailing` helpers.
   - ARCH-PURE: pass. Only test code changed.
   - ARCH-PURPOSE: pass. The root cause is fixed. The Log names sibling wall-clock bounds in other tests and reports them to the hub instead of fixing them here. That is fair: they are separate tests, and this issue is about this flake.
   - ARCH-EVOLVE: pass. No persisted format changed.

7. **Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: comment-claim-accuracy
    title: |
      BR-16 comment says the parent command "shares the check's budget"; each gets its own declared timeout
    detail: |
      At cmd/sdlc/localcheck_flakes_test.go:277, rerunEnvelope gives the rerun the full TimeoutMinutes*minute on its own,
      separate from the parent's run. Suggested wording: "each run is bounded by the check's declared timeout".
      This is the only instance in the window.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bb6bf55c-ed53-4601-831c-bfc6b5f0c54c","repository":"github.com/xianxu/ariadne","issue_id":"000620","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/fdb7fcb6508fa344","reviewed_head":"71f220cc1ddb633eda1e9f80722d80acdfbe93e1","slot":"ariadne:2","requested_at":"2026-10-25T08:03:56.842382Z"},"phase":"finished","pid":26444,"started_at":"2026-10-25T08:03:56.939021Z","finished_at":"2026-10-25T08:07:59.049472Z","session_id":"80838a87-6a68-4498-8f0b-01afddf3532c","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T01:12:44-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 620 — flake: TestKnownFlakeGate |
| repo | ariadne |
| issue file | workshop/issues/000620-flake-testknownflakegate.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7c2ac31c71c68fd65424412323b0d54896b255f8..a6ade96740168b1024144219f8504a328752b5bc |
| command | sdlc close --issue 620 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T01:12:44-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The only code change in this window is a three-line comment rewrite in `cmd/sdlc/localcheck_flakes_test.go:277-279`, and it fixes BR-1. The old comment said the failing suite command "shares the check's budget" with the rerun. The new one says each run, the suite command and the rerun, is bounded by the check's declared timeout. The code confirms the new wording:
- **Suite command:** gets `time.Duration(c.TimeoutMinutes)*d.minute` (`cmd/sdlc/localcheck_run.go:317`).
- **Rerun:** `rerunEnvelope` returns `min(knownFlakeRerunTimeout, declared)` (`cmd/sdlc/localcheck_affected_rerun.go:37-39`). The test raises `knownFlakeRerunTimeout` to 1h, so `declared` (5s) wins.

The rest of the window is issue-close bookkeeping and gate artifacts: the status moves to `codecomplete`, `card_mirror` and `actual_hours` are updated, a close log entry is added, and the plan files `-close-gate.md` and `-close-review.md` are new. Nothing changes `## Spec`, `## Revisions` or `## Done when`, so scope is unchanged and no behavior was added.

1. **Strengths**
   - The new comment matches the code path exactly. It names both runs and the right bound.
   - The fix stays within the finding. Nothing else was touched.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new. BR-1 is in family `comment-claim-accuracy`. The prior round said it was the only instance in the window, and this interdiff adds no other comments, so the family does not escalate.

5. **Test coverage notes:** BR-1 only asked for a comment change, so it needs no regression test. The existing BR-16 subtest checks the rerun's `TimedOut` record, and that remains the evidence for the behavior the comment describes.

6. **Architectural notes**
   - **ARCH-DRY:** pass. No logic was added.
   - **ARCH-PURE:** pass. No IO or logic changed.
   - **ARCH-PURPOSE:** pass. The one instance of the class was fixed, and the window has no siblings.
   - **ARCH-EVOLVE:** pass. The only persisted-format change is the new `actual_hours` frontmatter key in the issue file, which is additive.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Comment at localcheck_flakes_test.go:277-279 now says each run is bounded by the declared timeout; matches localcheck_run.go:317 and rerunEnvelope min(knownFlakeRerunTimeout, declared).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8f7cab43-307c-47ba-b490-bfd5464c06b8","repository":"github.com/xianxu/ariadne","issue_id":"000620","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/d99994af5a7b6b3d","reviewed_head":"a6ade96740168b1024144219f8504a328752b5bc","slot":"ariadne:2","requested_at":"2026-10-25T08:12:19.47012Z"},"phase":"finished","pid":33427,"started_at":"2026-10-25T08:12:19.570247Z","finished_at":"2026-10-25T08:12:44.155159Z","session_id":"2192a26d-9b2e-4274-8fd0-fb1bc0216570","outcome":"returned"}] -->
