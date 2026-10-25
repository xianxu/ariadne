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
