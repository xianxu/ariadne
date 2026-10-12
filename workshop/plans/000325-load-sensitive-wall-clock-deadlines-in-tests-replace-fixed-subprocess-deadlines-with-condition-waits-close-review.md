# Boundary Review — ariadne#325 (whole-issue close)

| field | value |
|-------|-------|
| issue | 325 — Load-sensitive wall-clock deadlines in tests: replace fixed subprocess deadlines with condition waits |
| repo | ariadne |
| issue file | workshop/issues/000325-load-sensitive-wall-clock-deadlines-in-tests-replace-fixed-subprocess-deadlines-with-condition-waits.md |
| boundary | whole-issue close |
| milestone | — |
| window | ab362ac4c35f137125f25e36a45f7b994d7493ad..9933e8dbbb8aabd0ef85c067501b3169bc73c8c7 |
| command | sdlc close --issue 325 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T08:15:34-07:00 |
| verdict | SHIP |

## Review

Still reviewing: checking the testgate timeout formatting and the judge cancellation timing, then I'll run the changed tests.

Running the mutation check: making the planning review hold the lock should fail the test.

```verdict
verdict: SHIP
confidence: high
```

The issue's purpose is met: both concurrency tests now prove their claim with an event instead of a deadline. The event is the repo lock's own wait line, single-sourced as `repolock.WaitMessagePrefix` and caught by `testfix.Watch`. Every other short deadline became one shared hang ceiling. Both `#325` quarantines are gone. I checked the Done-when clauses myself on the pinned head instead of relying on the Log:
- **Under load:** at a load average of about 105, `TestPlanningReviewConcurrencySchedules` passed in 30 s and `TestReviewConcurrencySchedules` passed all 14 subtests unquarantined. The other changed packages pass too: testfix, testgate, judge, repolock and processgroup.
- **Mutation:** I made `dispatchPlanningReview` keep the lock through `judge.Dispatch`. Both subtests then failed in about 4 s with `unrelated mutation waited for the repo lock … waiting for sdlc repo lock held by pid …: sdlc change-code`. I restored the file afterwards and the tree is clean.

Nothing blocks SHIP. What's left is two Minor DRY items.

1. **Strengths**
   - `cmd/sdlc/internal/repolock/repolock.go:28,217`: the wait line's text is single-sourced, and both the production print and the tests use it (ARCH-DRY pass).
   - `cmd/sdlc/planningreview_test.go:136-153` and `cmd/sdlc/review_concurrency_unix_test.go:166-189`: the tests fail the moment the wait line appears, which keeps the intent "runs concurrently" without timing anything. The wiring is live: the CLI hands the lock `cmd.ErrOrStderr()` (`cmd/sdlc/repolock.go:136`).
   - `cmd/sdlc/closereview_test.go:184`: `hangGuard` now aliases `testfix.HangCeiling`, so the existing ceiling was reused rather than a second one added.
   - `cmd/sdlc/internal/judge/cancellation_unix_test.go:97`: the elapsed check moved to "less than the production grace". It still proves the injected 50 ms grace was the one that applied, but no longer measures machine load.
   - `pkg/testgate/testgate_test.go:203-230`: the slot is released when the helper prints its `waiting` line, not after a fixed sleep. That race is gone, and the gate's 1 s re-poll keeps the "shrunk timeout" assertion meaningful.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **ARCH-DRY, hang-ceiling literals:** `cmd/sdlc/internal/repolock/repolock_test.go:315,348,361,397` hard-code `time.Minute` and call it a "hang ceiling". repolock tests can import `testfix` (it has no ariadne dependencies, so there's no import cycle), so these should use `testfix.HangCeiling`. Other instances of the same pattern:
     - The weave tests use `time.Minute` at 9 sites: `generator_lifetime_test.go` ×3, `release/main_test.go`, `staging/stage_test.go`, `plan/publication_test.go`, `weavefs/runner_owned_test.go` ×2 and `runner_setup_test.go`.
     - `acquire/git_refresh_test.go` uses it twice as an `ExecGit.Timeout`.

     Weave can't import `cmd/sdlc/internal/testfix`, which the Log already justifies, so the weave sites are a separate, lower-priority case. A shared test constant would also make the 1 min vs 2 min ceilings visible.
   - **ARCH-DRY:** `cmd/sdlc/review_concurrency_unix_test.go:170-177` copies the body of `executeSDLCTestCommand` (`cmd/sdlc/repolock_test.go:504`) so it can pass a context and a stderr writer. A variant like `executeSDLCTestCommandWith(ctx, errW, args...)` would keep one helper.
   - **ARCH-DRY, accepted:** `pkg/testgate/testgate_test.go:209-222` hand-rolls a needle-watcher that does the same job as `testfix.Watch`. `pkg/` can't import `cmd/sdlc/internal`, so this is acceptable and I'm not raising it as a finding.

5. **Test coverage notes**
   - `TestWatchSignalsOnNeedleAcrossWrites` covers a needle split across writes, a second match not closing the channel twice, and output being kept.
   - The mutation check above shows the planning test goes red when concurrency breaks. The Log says the close-path mutation was also run; I didn't re-run that one, but its wiring through `root.SetErr(waits)` → `cmd.ErrOrStderr()` checks out.
   - `processgroup`'s 5 s elapsed bound is still there; the comment explains it is the shutdown contract, not a hang guard, which is reasonable.

6. **Architectural notes**
   - ARCH-PURE pass: `Watch` is a small writer you can test without IO, and the IO stays in the test glue.
   - ARCH-PURPOSE pass: both named tests were converted to event evidence rather than just given bigger deadlines, both quarantines are deleted, and the sweep table records each site as fixed or justified.
   - Future tests that need to prove "A did not wait for B" should use `Watch` with `WaitMessagePrefix`; the atlas paragraph (`atlas/workflow/sdlc-binary.md:1565`) documents that.

7. **Plan revision recommendations:** none; the plan matches the code.

```findings
findings:
  - id: new
    severity: Minor
    family: hang-ceiling-single-source
    title: |
      repolock_test.go hard-codes time.Minute hang ceilings instead of testfix.HangCeiling
    detail: |
      cmd/sdlc/internal/repolock/repolock_test.go:315,348,361,397 can import testfix (no cycle) but restate the ceiling as a literal. Same family, separate case: 9 weave sites use time.Minute because they cannot import cmd/sdlc/internal/testfix (generator_lifetime x3, release/main, staging/stage, plan/publication, weavefs/runner_owned x2, runner_setup), plus 2 ExecGit timeouts in acquire/git_refresh_test.go. A shared weave test constant would make the 1 min vs 2 min ceilings explicit.
  - id: new
    severity: Minor
    family: test-command-helper-reuse
    title: |
      review_concurrency_unix_test.go copies executeSDLCTestCommand's body to pass ctx and stderr
    detail: |
      cmd/sdlc/review_concurrency_unix_test.go:170-177 rebuilds the root command by hand. Extending cmd/sdlc/repolock_test.go:504 with a variant that takes a context and a stderr writer would keep one helper. This is the only such copy in the window.
```
