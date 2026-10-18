# Boundary Review — ariadne#434 (whole-issue close)

| field | value |
|-------|-------|
| issue | 434 — flake: TestLocalCheckSignalCancelsQueuedHelper cleanup subtest under load |
| repo | ariadne |
| issue file | workshop/issues/000434-flake-testlocalchecksignalcancelsqueuedhelper-cleanup-subtest-under-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | d6be01d5c70701215c16224ead1834f66e587576..0e15fbe282dc936b12296577e873e4ab353603b0 |
| command | sdlc close --issue 434 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T22:16:22-07:00 |
| verdict | SHIP |

## Review

The diff is reviewed and the affected tests pass, so I'm writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

This change replaces wall-clock deadlines that could fire before a test fixture was ready. Each deadline now fires when the test signals it, after readiness, and there are no added sleeps or longer timeouts. The fix covers the whole class. The Spec names three affected tests, and all three changed: the signal cleanup subtest, `TestBoundedLandingGitTerminatesDescendants/deadline` and `TestMakeProbeBoundsDescendantPipes`. The boundedness check the signal test used to make with its 50 ms budget now has its own unit test. I checked the production paths each test drives:
- `probeMakeTestFactsContext` builds its timeout from the ctx it is given (`localchecks_make.go:93`).
- `newBoundedLandingGit` keeps the ctx, so a signalled deadline reaches the code as `DeadlineExceeded`.
- `localCheckCleanupGit` wraps the run's context in `context.WithoutCancel` (`localcheck_run.go:258`), which the new unit test exercises.

`go vet` is clean. Running the four affected tests, `TestDeadlineOnCue` and the context guard test with `-count=3` passes. Nothing blocks SHIP.

1. **Strengths**
   - `internal/testfix/deadline.go:32`: `DeadlineOnCue` is small and correct. A custom `Done`/`Err` makes `WithTimeout` set up a watcher goroutine, so a derived context reports `DeadlineExceeded`, and the `sync.Once` makes `expire` safe to call twice. The unit test checks both the derived context and the cue context, before and after `expire`.
   - `localcheck_signal_unix_test.go:173-200`: the cleanup phase now proves something stronger than before. The cancelled run is shown to still be waiting on its cleanup Git. After the gate is released, the shim runs the real `git worktree remove --force`, so the "worktree retired" check tests real removal.
   - FIFO release: the release uses a blocking open in a goroutine, so it works whichever side opens the pipe first. The `t.Cleanup` uses a non-blocking open, so it can't hang if the shim never opened its end.
   - `landinggit_process_test.go:153`: `TestLocalCheckCleanupGitIsIndependentAndBounded` checks both properties directly. Cleanup Git still runs after the run is cancelled, and a hung cleanup Git returns `DeadlineExceeded` within its budget. Neither check depends on how fast the fixture starts.
   - The atlas has a matching note in `atlas/workflow/sdlc-binary.md:1890`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `localchecks_make_process_test.go:64`: the make-probe test still checks only `err != nil`. The Spec says both changed tests "assert `DeadlineExceeded`", but this one doesn't, so the Spec says more than the code does. Fix: add `errors.Is(err, context.DeadlineExceeded)`. The landing-git test (`landinggit_process_test.go:67`) already checks this; it is the only other member of this family.
   - `landinggit_process_test.go:50`: in deadline mode, `ctx` is replaced by `DeadlineOnCue(context.Background())`, so the deferred `cancel` applies to an unused context. Passing the existing `ctx` as the parent would be tidier.
   - `localcheck_signal_unix_test.go:177-191`: this hand-written pid-polling loop has the same shape as the helper/owner poll earlier in the test and as `awaitLandingGitChild`. A shared "await pid file" helper in `testfix` could replace them later. This is ARCH-DRY, Minor.
   - The release goroutine (`localcheck_signal_unix_test.go:66`) can stay blocked for the rest of the test process if the shim is killed between writing its pid and opening the FIFO. The test still finishes, so the leak doesn't matter in practice.

5. **Test coverage notes:** the loaded run in Done when is recorded in the Log: six concurrent loops, 60 runs per test, at load average 60–72. The Log also records a repro that fails 10/10 with a 2 ms budget. That repro no longer applies, because the global 50 ms budget was removed. I did not re-run under load; my own runs were unloaded and passed.

6. **Architectural notes**
   - ARCH-DRY: pass. The one possible shared helper is listed under Minor.
   - ARCH-PURE: pass. This is test-only code, and `DeadlineOnCue` is pure.
   - ARCH-PURPOSE: pass. All three tests in the class were changed, not only the one named in the issue, and the boundedness check now has its own unit test.
   - `DeadlineOnCue` is the pattern future deadline tests should use.

7. **Plan revision recommendations:** add a `## Revisions` entry saying the make-probe test asserts a non-nil error, not `DeadlineExceeded`. The alternative is to add the stricter check.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: deadline-test-asserts-error-kind
    title: |
      Make-probe deadline test asserts only err != nil; Spec claims DeadlineExceeded
    detail: |
      localchecks_make_process_test.go:64 accepts any error, but the Spec says both changed tests assert DeadlineExceeded. Add errors.Is(err, context.DeadlineExceeded). The only other family member, landinggit_process_test.go:67, already checks it.
  - id: new
    severity: Minor
    kind: defect
    family: test-poll-loop-duplication
    title: |
      Hand-written pid-file polling loop repeated (ARCH-DRY)
    detail: |
      localcheck_signal_unix_test.go:177-191 has the same shape as the helper/owner poll at :136-163 and as awaitLandingGitChild in landinggit_process_test.go. A testfix "await pid file" helper could replace all three.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e7bced6c-dc9e-4ec5-b574-7d403d111d65","repository":"github.com/xianxu/ariadne","issue_id":"000434","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/bd1141e8e5e31227","reviewed_head":"0e15fbe282dc936b12296577e873e4ab353603b0","slot":"ariadne:2","requested_at":"2026-10-18T05:14:47.23119Z"},"phase":"finished","pid":11732,"started_at":"2026-10-18T05:14:47.532613Z","finished_at":"2026-10-18T05:16:22.414344Z","session_id":"ee20cce2-33c3-4017-a1d0-53c058a0077c","outcome":"returned"}] -->
