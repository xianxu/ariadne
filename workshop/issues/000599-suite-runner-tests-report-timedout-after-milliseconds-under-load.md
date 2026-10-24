---
id: 000599
status: codecomplete
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '21ed23a9c5ba9bf493ae9321eb6e4ea7815d6bd4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T10:45:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "a1af2ce4", done: "4447d243"}
actual_hours: N/A
---

# Suite-runner tests report TimedOut after milliseconds under load

## Problem

Under machine load, fake suites in the local-check and suite-runner tests
report `TimedOut` after milliseconds on a 1-minute or 30-minute timeout. Seen
on ariadne#593's head b8b2e7f9 in two full suites (2026-10-24):

- Attempt b60aa6b12878755b709c1f56faaa8c93: `TestHelperRefreshPrecedesSuitePreflight/recorded`
  reported TimedOut after 0.66 ms (`command: true`, TimeoutMinutes 1).
- Attempt 049e81aac7f5dc327189075c6bb967c0 (wall 1761s against a usual 1081s):
  nine failures, several the same shape (TimedOut after 19 to 33 ms):
  `TestRunLocalCheckLightOverrunRetriesInNormalLine`,
  `TestRunLocalCheckTerminatesDescendants/timeout`,
  `TestSuiteCapabilityResumeRetryAfterRepair`,
  `TestSuiteEnvironmentBaselineIsNotRedMain`,
  `TestSuiteEnvironmentRetentionFailureKeepsRawEvidence/false` and `/true`,
  `TestSuiteInfrastructureBaselineCannotBlameCars`,
  `TestSuiteInfrastructureFailureHasNoCodeAuthority/affected`,
  `TestSuiteInfrastructureMissingGoNeverStarts`.

Run in isolation, eight of the nine pass.
`TestRunLocalCheckLightOverrunRetriesInNormalLine` fails deterministically
outside the suite, identically on main e5ba9973f ("the retry does not carry the
overrun"), though main's baseline inside `make test` passed it earlier the same
day.

A TimedOut after under 1 ms on a 1-minute timeout is not plain load; it looks
like a deadline computed from a stale or wrong start time (TL observation).
If production suite runs share that computation, a real suite could be killed
early, which is serious. `TestHelperRefreshPrecedesSuitePreflight` is
registered under #596 meanwhile; the others are deliberately not registered as
flakes, since they are one bug.

## Spec

Root cause (measured): `runUnderResourceBatch` (`cmd/sdlc/localcheck_run.go`)
starts the check's deadline, `context.WithTimeout(ownedCtx, timeout)`, before
the pre-command setup (`prepareScratch`, runner-owner attach and directory
syncs, the SSD temp listing), and reads `runCtx.Err() == DeadlineExceeded`
only after the post-run bookkeeping (slot release, lease release, `q.finish`).
Two consequences:

1. Setup time is charged to the command's budget. Tests set `minute` to 1s
   (`localRunRepo`), so under load setup can spend the whole budget: the batch
   sees an expired context, runs nothing, and the run is TimedOut with a
   Duration of microseconds (the 0.66 ms and 19 to 33 ms failures).
2. A command that exits before its deadline is still recorded TimedOut when the
   deadline passes during the bookkeeping after it.

The light-lane test is the same bug: its 1s lane bound starts before setup, so
the measured stopped Duration is the bound minus the setup time. Instrumented
on this machine it read 0.493s, under the test's 0.5s floor, so it fails
deterministically here and passes where setup is fast.

Production exposure: same code path, with `minute` = 1 minute and timeouts of
30 to 120 minutes. A real suite is never killed early by more than the setup
time (seconds), but (2) can record a suite that exited, even with exit 0, as
TimedOut when it finishes within the bookkeeping window of its deadline. The
same post-hoc read exists in `prepareAffectedGraph`
(`localcheck_graphprepare.go`): `ctx.Err()` after `Terminate` overrides a
successful exit.

Fix (ARCH-PURPOSE, the class: a deadline must measure the command, and a
timeout verdict must be taken from the command's own end):
- Start the deadline immediately before `runManagedCommandBatch`, after setup.
- Have the batch report whether its deadline stopped it, read at the moment
  the batch returns, and set `TimedOut` and `LaneOverrun` from that, not from
  `runCtx.Err()` after bookkeeping.
- In `prepareAffectedGraph`, take the deadline verdict only when `cmd.Run`
  failed, so a successful exit stands.

## Done when

- The root cause of the early TimedOut is found and fixed, and a test pins it
  without depending on machine load.
- It is stated, with evidence, whether production suite runs were exposed.
- `TestRunLocalCheckLightOverrunRetriesInNormalLine` passes deterministically
  in and outside `make test`.
- The listed tests pass in a full suite.

## Plan

- [x] Tests first, load-independent through seams, not sleeps:
  - setup that outlasts the budget (a slow `normalTempDir`/scratch seam) still
    runs the command with its full timeout;
  - a command that exits 0 before its deadline, with bookkeeping after it
    slower than the remaining budget, is recorded passed, not TimedOut;
  - a command that really outlasts its timeout is still TimedOut (and a light
    run still a LaneOverrun);
  - `prepareAffectedGraph`: a successful command is not turned into a deadline
    error.
- [x] Move the deadline in `runUnderResourceBatch`; return the timeout verdict
  from the batch; same fix in `prepareAffectedGraph`.
- [x] Run the nine listed tests plus `TestHelperRefresh*` in and outside the
  suite environment, repeated; then the full suite at pre-close.
- [x] Unregister nothing: #596's tests are its own issue to close, but note in
  #596's Log that #599 explains this failure shape.

## Log

### 2026-10-24
- 2026-10-24: closed — Full suite passed on ef778af0 under load (sdlc-local/suite, wall 1243s, 0 failed); round-1 and round-2 fixes (a2f195e2: shared stoppedBy for run, graph-prep and make-probe sites) pass targeted runs 2x (TestAffectedGraphPreparation*, TestLocalCheckDeadline*, TestStoppedBy, light-lane, terminate-descendants, make probe tests); mutations (deadline before setup, verdict after bookkeeping, graph-prep guard dropped) fail the tests; the nine tests that failed on #593 head b8b2e7f9 passed 3x unsandboxed. Actuals: --no-actual under the TL standing authorization for the ariadne#593 64 KiB refusal; fleet binary refused with: measurement branch proof: measurement snapshot does not serialize: json: error calling MarshalJSON for type *gatestate.MeasurementInputs: measurement exceeds 65536 bytes. The #593 branch build (b8b2e7f9) measures #599 at 0.29h, window 2026-10-24T10:45:56-07:00 to 11:03:27-07:00, code pin ef778af0. No typed hours. (fixed-then-ship) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-24: closed — Full suite passed on ef778af0 under load (sdlc-local/suite, wall 1243s, 0 failed); round-1 fixes on 8d3244d3 pass targeted runs (TestAffectedGraphPreparation*, TestLocalCheckDeadline*, TestDeadlineStopped, make probe tests); TestLocalCheckDeadlineBoundsTheCommandsOnly, TestDeadlineStopped and TestAffectedGraphPreparationSuccessOutlivesItsDeadline pin the fix, and each mutation (deadline before setup, verdict after bookkeeping, graph-prep guard dropped) fails them; the nine tests that failed on #593 head b8b2e7f9 passed 3x unsandboxed. Actuals: --no-actual under the TL standing authorization for the ariadne#593 64 KiB refusal; fleet binary refused with: measurement branch proof: measurement snapshot does not serialize: json: error calling MarshalJSON for type *gatestate.MeasurementInputs: measurement exceeds 65536 bytes. The #593 branch build (b8b2e7f9) measures #599 at 0.29h, window 2026-10-24T10:45:56-07:00 to 11:03:27-07:00, code pin ef778af0. No typed hours. (done-when: contract moved since change-code; judged by the close review); review verdict: FIX-THEN-SHIP

- Plan approved by the TL. Instrumented the light-lane test: the stopped run
  measured 0.493s on a 1s bound, so setup took about 0.5s of the deadline
  (`prepareScratch` sweeps stale temp folders under every scratch root, which
  grows with the machine's load and scratch contents).
- Production exposure: real suites (timeouts 30 to 120 min) were never killed
  more than the setup time early, but a suite that exited, even passing,
  within the bookkeeping window after its commands (slot release, lease
  release, `q.finish`) could be recorded `TimedOut`. `prepareAffectedGraph`
  could likewise turn a successful preparation into a deadline error.
- Fix: the deadline starts right before `runManagedCommandBatch`; the verdict
  is `deadlineStopped(runErr, runCtx.Err())` read on return; `Duration` ends
  there too. A `phaseHook` test seam ("setup", "ran") stands in for load.
  `TestLocalCheckDeadlineBoundsTheCommandsOnly` (slow setup, slow bookkeeping,
  real overrun) and `TestDeadlineStopped`. Mutations: the deadline before setup
  fails "slow setup"; the verdict read after bookkeeping fails "slow
  bookkeeping". The graph-prepare change has no load-independent test (no seam
  can delay its teardown without changing `Run`'s result); it is one condition.
- Close review round 1 (FIX-THEN-SHIP), all three fixed:
  - BR-2: the "no seam" claim above was wrong. `prepareAffectedGraph` now calls
    `d.runPhase("ran")` after teardown, and
    `TestAffectedGraphPreparationSuccessOutlivesItsDeadline` pins a success
    that outlives its deadline and a real overrun; dropping the
    `err != nil &&` guard fails it.
  - BR-1: the #596 note is on main in 5ee14b6b8 (`sdlc issue publish`, which
    takes this branch's copy back, so the branch head does not carry it; it
    arrives with the next main merge).
  - Minor (class sweep): `localchecks_make.go`'s make probe had the same
    post-hoc `ctx.Err()` override and gets the same guard. `helper_refresh.go`
    only downgrades an optional refresh to a warning, so it is outside the class.

## Revisions

- 2026-10-24 (close review round 1): the class sweep adds the make probe
  (`localchecks_make.go`), and the graph-preparation test is delivered through
  the phase seam rather than left to review.
- 2026-10-24 (close review round 2, Minor BR-3, fixed in round): the three
  sites share one pure `stoppedBy(runErr, ctxErr) (cause, stopped)`
  (ARCH-DRY), replacing `deadlineStopped` and the two inline guards;
  `TestStoppedBy` covers it once, so the make probe's guard is now tested. The
  reviewer's out-of-scope list (post-hoc `ctx.Err()` reads in `golist.go`,
  `weeklystats_inputs.go`, `weeklyresource_machine.go`, `netcheck/probe.go`,
  `judge/dispatch.go`, `cigate.go`, and the `interrupted` read after
  bookkeeping) is outside this issue's local-check path; reported to the TL as
  a possible audit.
