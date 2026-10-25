# Boundary Review — ariadne#599 (whole-issue close)

| field | value |
|-------|-------|
| issue | 599 — Suite-runner tests report TimedOut after milliseconds under load |
| repo | ariadne |
| issue file | workshop/issues/000599-suite-runner-tests-report-timedout-after-milliseconds-under-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6afd91697c6911c521b5c2ef83ce60097c17b0da..ee2923e1bae549470d7de4e33f749b8942424b35 |
| command | sdlc close --issue 599 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T12:32:42-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Targeted tests pass. Now I'm checking that the new test fails when the fix is reverted.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core fix is correct, and it is proven by the tests. In `runUnderResourceBatch`, the deadline now starts after setup. The timeout verdict is now read as soon as `runManagedCommandBatch` returns, before the slot release, lease release and `q.finish` bookkeeping. `Duration` also ends at that point. I ran the new `TestLocalCheckDeadlineBoundsTheCommandsOnly` and `TestDeadlineStopped`, together with the light-lane, terminate-descendants, `TestHelperRefresh*` and graph tests, and they all pass. I then reverted each fix half separately in this checkout, and the matching subtest went red each time:
- Moving the verdict read back to after the bookkeeping fails "slow bookkeeping" with TimedOut after 6.7 ms.
- Moving the deadline back to before setup fails "slow setup" with TimedOut after 2 µs.

Two things are still open, and both are cheap to fix. First, two Plan items are ticked but not delivered at the pinned head: the `prepareAffectedGraph` regression test and the #596 Log note. Second, the `prepareAffectedGraph` behaviour change has no test. The Log says no seam could add one, but a seam would.

1. **Strengths**
   - `cmd/sdlc/localcheck_run.go:669-678`: the deadline now covers only the command batch. Both the verdict (`timedOut`) and the end time (`ended`) are captured at the same moment, right after the batch returns. This removes both failure modes the Spec named.
   - `deadlineStopped` (`localcheck_run.go:731`) is a small pure function, and `TestDeadlineStopped` tests it directly with no IO, including the wrapped-error case. This satisfies ARCH-PURE.
   - The `phaseHook` seam makes the test independent of machine load, which the Done-when requires. The "real overrun" case also checks that a genuine timeout is still recorded as TimedOut and not stopped early.
   - The atlas is updated in both places that describe this behaviour (`ci-merge-check.md:209-214` and `:374`), and both claims match the code.

2. **Critical:** none.

3. **Important**
   - **Two Plan items are ticked but not delivered.**
     - Plan item 1, sub-bullet "`prepareAffectedGraph`: a successful command is not turned into a deadline error", is ticked `[x]`, but no such test exists. The Log admits this.
     - Plan item 4 is ticked, and commit ee2923e1b says "#596's Log now names #599". At head, `workshop/issues/000596-flake-testhelperrefreshrunnerresourceexcluded.md` contains no "599", and the file is not touched anywhere in the window.
     - Fix: write the #596 note, add the test (next finding), or untick the items and add a Revision entry.
   - **`prepareAffectedGraph` has no regression test** (`localcheck_graphprepare.go:34-35`).
     - The Log says "no seam can delay its teardown without changing `Run`'s result". That is wrong: a `d.runPhase("ran")` call placed between `cmd.Run()` and the `ctx.Err()` read delays exactly the window that matters.
     - Suggested test: set a small `d.minute` and `TimeoutMinutes: 1`, use command `true`, and have the hook sleep past the deadline. Assert that the result is nil. The old condition `if ctx.Err() != nil` would make this test fail.

4. **Minor**
   - The same class of bug survives in `cmd/sdlc/localchecks_make.go:105`: there, `ctx.Err()` is read after `Output`/`Terminate`, so it can override a make probe that succeeded. The window is only the `Terminate` call, so the risk is low. The Spec named this class (ARCH-PURPOSE). Either apply the same `err != nil &&` guard or record why this site is exempt.
     - I checked the other similar site, `helper_refresh.go:69`: there a late `ctx.Err()` only downgrades an optional refresh to a warning, so it is outside the class.
   - `prepareAffectedGraph` now also ignores a parent cancellation that arrives after a successful preparation. This looks harmless, since the affected run that follows still sees the cancelled ctx. A one-line comment saying so would help.

5. **Test coverage notes**
   - The new tests in `cmd/sdlc/localcheck_deadline_test.go` cover the run path in both directions:
     - The command finishes but the slow phase runs past the deadline: setup and bookkeeping are each covered and are not recorded as TimedOut.
     - The command itself overruns: it is recorded as TimedOut.
   - The "light run is still a LaneOverrun" case is covered only by the existing `TestRunLocalCheckLightOverrunRetriesInNormalLine` (`suitelane_run_test.go:109`), which passed.
   - The `runErr != nil` guard inside `deadlineStopped` is tested only by the unit test, because the verdict is now read at once and the context is rarely expired by then. That is acceptable.

6. **Architecture**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. The verdict is a pure function, and the IO shell around it is thin.
   - ARCH-PURPOSE: flagged as Minor for the unswept `localchecks_make.go:105`.
   - ARCH-EVOLVE: pass. No persisted format changes; `Start`/`Duration` keep their meaning, and `Start` already marked the beginning of the commands before this change.

7. **Plan revisions**
   - If the `prepareAffectedGraph` test is not added: "## Revisions — Plan item 1's `prepareAffectedGraph` sub-bullet was not delivered as a test; the change is one condition, and it is covered only by review."
   - If the #596 note is written elsewhere: name where it was written.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: plan-tick-without-delivery
    title: |
      Plan items ticked [x] that the pinned head does not deliver
    detail: |
      Every instance in this window: (1) Plan item 4 is ticked, and commit ee2923e1b claims the issue 596 Log names 599, but workshop/issues/000596-flake-testhelperrefreshrunnerresourceexcluded.md has no mention of 599 and is not touched in the window. (2) Plan item 1's sub-bullet promising a prepareAffectedGraph test is ticked, but no such test exists (the Log admits it). Write the note and add the test, or untick both and add a Revisions entry.
  - id: new
    severity: Important
    kind: defect
    family: behavior-fix-needs-failing-test
    title: |
      prepareAffectedGraph success-stands change has no regression test
    detail: |
      localcheck_graphprepare.go:35 changes behavior, but no test fails without the change. The Log's claim that no seam can delay teardown without changing Run's result is wrong. A d.runPhase("ran") call between cmd.Run() and the ctx.Err() read would allow a test: small d.minute, command true, hook sleeps past the deadline, assert a nil error.
  - id: new
    severity: Minor
    kind: defect
    family: deadline-verdict-from-command-end
    title: |
      localchecks_make.go:105 still lets a post-hoc ctx.Err() override a successful make probe
    detail: |
      This is a sibling of the class the Spec named (ARCH-PURPOSE). After Output and Terminate, ctx.Err() is checked before err, so a probe that succeeded is reported as interrupted if the deadline passes during Terminate. The window is small. Apply the err != nil guard or record why this site is exempt. helper_refresh.go:69 is outside the class, because it only downgrades an optional refresh to a warning.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"bd3ae620-4e11-445e-83bd-6773c8439031","repository":"github.com/xianxu/ariadne","issue_id":"000599","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/5603fb7040ed387a","reviewed_head":"ee2923e1bae549470d7de4e33f749b8942424b35","slot":"ariadne:2","requested_at":"2026-10-24T19:29:43.006158Z"},"phase":"finished","pid":35705,"started_at":"2026-10-24T19:29:43.107096Z","finished_at":"2026-10-24T19:32:42.172399Z","session_id":"33f2636a-06fa-4f96-bf50-7c1e6e01b739","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T12:38:35-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 599 — Suite-runner tests report TimedOut after milliseconds under load |
| repo | ariadne |
| issue file | workshop/issues/000599-suite-runner-tests-report-timedout-after-milliseconds-under-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6afd91697c6911c521b5c2ef83ce60097c17b0da..8d3244d30f21806192c0b77d2c6f1852dbb0ed14 |
| command | sdlc close --issue 599 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T12:38:35-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Now mutation: revert the graph-prepare guard.

I'm finishing the review. All three prior findings have been checked and two mutation tests went red as expected. The verdict is below.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix does what the Spec asks. In `runUnderResourceBatch`, the deadline now starts right before `runManagedCommandBatch`. The timeout verdict is read as soon as the batch returns, and `Duration` ends at that same point. `prepareAffectedGraph` and the make probe no longer let a later `ctx.Err()` override a command that succeeded. I ran two mutations to confirm the tests catch the bug:
- **Graph-prepare guard removed:** `TestAffectedGraphPreparationSuccessOutlivesItsDeadline` fails with "a successful preparation became graph preparation: context deadline exceeded".
- **Verdict read moved back after bookkeeping:** `TestLocalCheckDeadlineBoundsTheCommandsOnly/slow_bookkeeping` fails.

The nine tests listed in the issue, plus `TestHelperRefresh*` and the new tests, pass with `-count=3`. `go vet` is clean. The only thing still open is BR-3. The make-probe guard is correct, but no test fails without it. It is Minor and does not block.

Done-when: current. The revision added the make probe to the class sweep, but the Done-when clauses (root cause fixed and pinned without depending on load, production exposure stated, light-lane test deterministic, listed tests green) still describe what this diff builds.

1. **Strengths**
   - `cmd/sdlc/localcheck_run.go:676-678`: the deadline, the batch and the verdict read sit in one short block. `deadlineStopped` (`:733`) is a pure function, and `TestDeadlineStopped` tests it directly, including a wrapped deadline error.
   - The `phaseHook` seam (`localcheck_run.go:72`, `runPhase` `:738`) stands in for machine load, so the slow-setup and slow-bookkeeping tests don't depend on how fast the machine is. Both mutations above go red.
   - `TestLocalCheckDeadlineBoundsTheCommandsOnly` also covers the opposite case: a real overrun is still `TimedOut`, and `Duration` is at least the deadline.
   - The atlas change (`atlas/workflow/ci-merge-check.md:209-214, 374`) describes the new deadline behaviour and the graph-preparation rule accurately.

2. **Critical:** none.

3. **Important:** none new.

4. **Minor**
   - BR-3 is still open (see dispositions). The make-probe guard at `localchecks_make.go:106` has no regression test.
   - ARCH-DRY: "a context verdict stands only if the command failed" now appears three times. `deadlineStopped` handles it at `localcheck_run.go:733`, and the same check is written inline at `localcheck_graphprepare.go:38` and `localchecks_make.go:106`. The inline versions also treat cancellation this way, not only deadlines. One shared pure helper (for example `stoppedBy(runErr, ctxErr) error`) used at all three sites would let one pure test cover all of them, which would also close BR-3. This is part of the BR-3 fix, not a new finding.

5. **Test coverage notes**
   - I verified the regression tests for the run-path verdict and for graph prepare by mutation. The slow-setup mutation is claimed in the Log; I did not repeat it.
   - No test covers a light-lane overrun with slow bookkeeping. `TestRunLocalCheckLightOverrunRetriesInNormalLine` passes repeatedly, and the code sets `LaneOverrun` in the same branch as `TimedOut`, so the risk is low.
   - I did not run the full suite: it takes about 1081s, longer than this review's limit. The repeated targeted runs pass.

6. **Architectural notes**
   - ARCH-DRY: flagged, Minor, as above.
   - ARCH-PURE: pass. The verdict is a pure function, and the seam is injected through `localRunDeps`.
   - ARCH-PURPOSE: pass for the class the Spec scoped (deadline verdicts on the local-check, suite and graph path). The same post-hoc `ctx.Err()` override still exists in places outside the Spec, for example `internal/affected/golist.go:153`, `weeklystats_inputs.go:163`, `weeklyresource_machine.go:41`, `internal/netcheck/probe.go:64`, `internal/judge/dispatch.go:186` and `cigate.go:341`. `golist.go:153` only falls back to the full suite, a safe outcome, like the `helper_refresh` exemption. These are worth a follow-up audit issue, not a gate on this one.
   - Also outside this issue's scope: `interrupted := ownedCtx.Err()` (`localcheck_run.go:694`) is still read after bookkeeping. So a signal that arrives during bookkeeping turns a passing run into `queueInterruptedError`. That is a cancellation, not a deadline, so arguably it is intended.
   - ARCH-EVOLVE: pass. No persisted format changes. `TimedOut` and `Duration` keep their meaning; they are just measured more accurately.

7. **Plan revision recommendations**
   - None needed for BR-1: the #596 note is real (commit `5ee14b6b8` on `origin/main` names #599), and the Log says why the branch head doesn't carry it.
   - If BR-3 is answered with a shared helper, add a Revisions line for it.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The 596 note exists in 5ee14b6b8 on origin/main (lines naming ariadne 599 as the cause), and the Log explains why the branch head lacks it. The graph-prepare test now exists (TestAffectedGraphPreparationSuccessOutlivesItsDeadline).
  - id: BR-2
    disposition: addressed
    note: |
      Verified by mutation: dropping the err != nil guard at localcheck_graphprepare.go:38 makes TestAffectedGraphPreparationSuccessOutlivesItsDeadline fail with "a successful preparation became graph preparation: context deadline exceeded".
  - id: BR-3
    disposition: not-addressed
    note: |
      The guard at localchecks_make.go:106 is correct, but nothing fails without it. Cheapest fix: route all three sites (localcheck_run.go:733, localcheck_graphprepare.go:38, localchecks_make.go:106) through one pure helper tested once (ARCH-DRY), or add a phase seam to the probe.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7ea671fa-5179-407e-be43-4a04ca337bfc","repository":"github.com/xianxu/ariadne","issue_id":"000599","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/a6f4fc35d24621a0","reviewed_head":"8d3244d30f21806192c0b77d2c6f1852dbb0ed14","slot":"ariadne:2","requested_at":"2026-10-24T19:34:50.231883Z"},"phase":"finished","pid":99379,"started_at":"2026-10-24T19:34:50.328799Z","finished_at":"2026-10-24T19:38:34.954173Z","session_id":"35178aa6-8ab8-4399-bb33-526ecc47453c","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T12:42:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 599 — Suite-runner tests report TimedOut after milliseconds under load |
| repo | ariadne |
| issue file | workshop/issues/000599-suite-runner-tests-report-timedout-after-milliseconds-under-load.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5ddb607d74c60e995dc47f9b8fd03149a0c4041a..a2f195e26057396982fdc9666f165d5ae6a8a261 |
| command | sdlc close --issue 599 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T12:42:25-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This window fixes BR-3 the right way. It doesn't patch only the make probe. All three sites that decide whether a deadline stopped a command now call one pure helper, `stoppedBy(runErr, ctxErr) (cause, stopped)`, at `cmd/sdlc/localcheck_run.go:737`. The sites are the run verdict, graph preparation and the make probe. This covers the rule the `deadline-verdict-from-command-end` family names: a command that succeeded stands, even if its context ended afterwards. I checked each site against the code it replaced and behavior is unchanged:
- **Run verdict:** `stopped && errors.Is(cause, DeadlineExceeded)` is the same test as the old `runErr != nil && errors.Is(ctxErr, DeadlineExceeded)`, because `cause` is `ctxErr` whenever `stopped` is true.
- **Graph preparation:** it replaces `err` with `ctx.Err()` only when the command failed, as before.
- **Make probe:** it reports "interrupted" only when `err != nil`.

`go vet` is clean, and the targeted tests pass (`TestStoppedBy`, `TestLocalCheckDeadline*`, graph-preparation and probe tests). The new `## Revisions` entry only records the fix and moves the out-of-scope sites to a separate audit for the TL. It changes no scope.

Done-when: current — the diff is a refactor that keeps behavior the same. It neither widens nor narrows the root-cause fix that `## Done when` covers.

1. **Strengths**
   - `stoppedBy` (`localcheck_run.go:733-742`) is the single shared rule (ARCH-DRY). Three copies of the logic, slightly different from each other, became one.
   - The table in `TestStoppedBy` adds the case of a command that finished before the caller cancelled (`{nil, context.Canceled}`). The old helper only covered deadlines, so this is new coverage.
   - The helper is pure: it takes two errors and returns a result, and its test does no IO (ARCH-PURE).
2. **Critical:** none.
3. **Important:** none.
4. **Minor:**
   - The old test checked that a wrapped deadline still counts as a deadline (`errors.Join(context.DeadlineExceeded)`). That check was removed. The `errors.Is` call it covered still exists at `localcheck_run.go:679`, but nothing tests it now. Very low risk.
   - The close entry in `## Log` names `TestDeadlineStopped`, which this window renames to `TestStoppedBy`. The entry is a historical record, so leaving it is acceptable.
5. **Test coverage:** `TestStoppedBy` pins the shared rule. The make probe's call site has no test of its own, but the rule it delegates to is tested. For a Minor finding that is enough.
6. **Architecture:**
   - **ARCH-DRY:** pass.
   - **ARCH-PURE:** pass.
   - **ARCH-PURPOSE:** pass. The rule's family is swept within the local-check path, and other packages that read `ctx.Err()` after a command returns are reported to the TL as a possible audit.
   - **ARCH-EVOLVE:** not applicable; no persisted format changes.
7. **Plan revisions:** none needed. The round-2 entry in `## Revisions` matches the code.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      All three sites, make probe included (localchecks_make.go:105), now call the pure stoppedBy (localcheck_run.go:737); TestStoppedBy pins the rule; behavior is the same at every site; vet and targeted tests pass.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"38a20994-4d09-4ab0-8447-20c8782c40a8","repository":"github.com/xianxu/ariadne","issue_id":"000599","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/5239bc663be4d362","reviewed_head":"a2f195e26057396982fdc9666f165d5ae6a8a261","slot":"ariadne:2","requested_at":"2026-10-24T19:41:18.470318Z"},"phase":"finished","pid":57115,"started_at":"2026-10-24T19:41:18.594469Z","finished_at":"2026-10-24T19:42:25.270608Z","session_id":"e4a8ffa0-eb4b-429f-861d-983dd81f73ce","outcome":"returned"}] -->
