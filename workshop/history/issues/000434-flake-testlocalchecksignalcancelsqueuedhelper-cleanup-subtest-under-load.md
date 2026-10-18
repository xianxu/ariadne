---
id: 000434
status: done
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '793f4294507e274e2ee827f20d0e6c4f50099fc2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T21:15:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c8f878ae", done: "88084f14"}
actual_hours: 0.79
---

# flake: TestLocalCheckSignalCancelsQueuedHelper cleanup subtest under load

## Problem

`TestLocalCheckSignalCancelsQueuedHelper` (cmd/sdlc/localcheck_signal_unix_test.go) fails intermittently under load and passes alone:

- 2026-10-17, ariadne:2 on the #431 branch (a broad `go test ./cmd/sdlc/ -run 'LocalCheck|CIGate|Suite|Landing|Merge'` run alongside other slots' suites): subtest `terminated/cleanup` failed at localcheck_signal_unix_test.go:171, "cleanup never reached its blocked Git seam". Rerun alone: 3/3 pass.
- Earlier (during #424, ariadne:2): the same test failed once under load, passed 3/3 alone and in the full suite.

The cleanup phase waits for the fixture's blocked `git worktree remove --force` to write its pid; under load the run's 50 ms `localCheckCleanupTimeout` (set by SDLC_BLOCK_PHASE=cleanup) likely expires before the shimmed git starts, so the seam is never reached.

## Spec

**Measured cause (Log has the runs):** the child sets `localCheckCleanupTimeout = 50 ms`, and that budget applies
per call to *every* cleanup git call, not only the blocked one. Two failures follow under load: the deadline kills the
shim before its shell writes the readiness pid ("cleanup never reached its blocked Git seam"), and it kills the
ordinary `git worktree prune` that follows, leaving the registry entry ("cancelled run retained its worktree").
The class is **a deadline that starts before the fixture's readiness signal**; swept, it has two more members:
`TestBoundedLandingGitTerminatesDescendants/deadline` (500 ms against an alias that must report its child pid) and
`TestMakeProbeBoundsDescendantPipes` (500 ms against make + `$(shell)` reporting its child pid). The
boundary-push and integrate-gate timeout tests assert only that the deadline bounds the call, so they are safe.

**Fix: the deadline fires on the test's cue, after readiness, instead of racing it (no sleeps, no wider timeouts).**

- `testfix.DeadlineOnCue() (ctx, expire)`: a context whose `Done` closes, with `Err() == DeadlineExceeded`, only
  when the test calls `expire`. The production code derives its own `WithTimeout` from the ctx it is handed, so the
  cue reaches it as a real deadline (`DeadlineExceeded` all the way down).
- `TestBoundedLandingGitTerminatesDescendants/deadline` and `TestMakeProbeBoundsDescendantPipes`: keep a long
  limit, hand the runner or the probe a cue ctx, wait for the child pid (the readiness signal), then expire, and
  assert `DeadlineExceeded` and no live descendant as today.
- The signal test's cleanup phase (the deadline lives in a child process, so a cue ctx cannot reach it): drop the
  global 50 ms budget, so ordinary cleanup git runs under the production recovery budget. The shim blocks
  `worktree remove --force` on a FIFO after writing its pid. The test waits for the pid, asserts the cancelled child
  is still in cleanup (cleanup really runs after cancellation, which is what the phase proves), then releases the
  FIFO, and asserts the child exits and the worktree is retired (dir and registry entry).
- The cleanup's own boundedness gets a direct unit test instead: `localCheckCleanupGit` with a cancelled command
  ctx still runs git (WithoutCancel), and against a hung alias returns `DeadlineExceeded` within its (shrunk)
  budget. That assertion needs no readiness signal, so it cannot race.

## Done when

- The cleanup subtest doesn't depend on process-start latency under load (e.g. the fixture waits for the shim's readiness, or the assertion accepts a deadline that fired before the seam), and a loaded run (`-count=20 -parallel` alongside a suite) passes.

## Plan

- [x] `testfix.DeadlineOnCue` + its unit test
- [x] Bounded-git deadline test and make-probe test fire the deadline on cue after readiness
- [x] Signal test cleanup phase: FIFO-gated shim, no global 50 ms; `localCheckCleanupGit` unit test for the bound
- [x] Loaded verification: six concurrent loops of all three tests, plus the shrunk-deadline repro now passing

## Log

### 2026-10-17
- 2026-10-17: closed — Round-2 advisory swept: localcheck_run_process_test and cli_signal_unix_test now wait via testfix.AwaitPIDFile; TestCLISignalCancelsOwnedReviewer, TestCLISignalOwnershipKeepsLockUntilUnwind, TestCLICancelledCloseBypassDoesNotWrite, TestRunLocalCheckTerminatesDescendants pass with -count=2. (fixed-then-ship); review verdict: SHIP
- 2026-10-17: closed — Round-1 advisories fixed: make-probe hanging-shell asserts DeadlineExceeded; pid-file waits share testfix.AwaitPIDFile; bounded-git cue derives from its ctx. The four affected tests, TestDeadlineOnCue and the context guard pass with -count=3. (fixed-then-ship); review verdict: SHIP
- 2026-10-17: closed — sdlc-local/suite green on 90a771d1 (exit 0, 9m23s; later commit atlas-only). Mechanism measured: the child cleanup budget at 2 ms reproduced "cleanup never reached its blocked Git seam" 10/10; under load the 50 ms per-call budget also killed worktree prune ("retained its worktree"). Fix: deadlines fire on the test cue after readiness (testfix.DeadlineOnCue) in TestBoundedLandingGitTerminatesDescendants/deadline and TestMakeProbeBoundsDescendantPipes; the signal test cleanup shim waits on a FIFO the test releases; TestLocalCheckCleanupGitIsIndependentAndBounded covers the bound directly. Six concurrent loops x10 of the four tests at load average 60-72: 6/6 PASS.; review verdict: SHIP

- Measured on 2026-10-17 (ariadne:2). Mechanism: with the child's cleanup budget shrunk from 50 ms to 2 ms,
  `terminated/cleanup` fails 10/10 with exactly "cleanup never reached its blocked Git seam". Natural rate: 0/40 at
  load average 17. Six concurrent loops of 15 (`-test.count`, a compiled test binary) hit a different failure once,
  "cancelled run retained its worktree" (a registry entry with HEAD 0000 left behind), which is the same 50 ms
  budget killing the ordinary `git worktree prune`: `boundedLandingGitRunner.run` applies the timeout per call.
- Implemented. `testfix.DeadlineOnCue` (+ `TestDeadlineOnCue`: a derived WithTimeout stays live until the cue, then
  reports DeadlineExceeded). The bounded-git `deadline` mode and both make-probe modes now expire on cue after the
  child pid appears, with production-length limits. The signal test's cleanup phase has no global 50 ms: its shim
  reports in, then waits on a FIFO; the test asserts the cancelled run is still waiting, releases the FIFO (a
  blocking writer open in a goroutine, since a non-blocking one fails when the shim has not opened the read end yet:
  the first attempt hung 3/3 for exactly that reason), and asserts the run exits, its Git is gone and the worktree is
  retired. `TestLocalCheckCleanupGitIsIndependentAndBounded` covers what the 50 ms used to: cleanup Git runs after
  cancellation and a hung one returns DeadlineExceeded within its budget.
- Loaded verification: six concurrent loops of the four affected tests, 10 iterations each (60 per test), at load
  average 60–72: 6/6 PASS.
- Recorded suite on 0389da34 failed on one test only: `TestVerbContextsReachTrackerReads` (the context guard)
  refused `context.Background` in `testfix/deadline.go`. `DeadlineOnCue` now takes the caller's context (it carries
  its values; only the cue ends it). Failed attempt kept: test-stats/failed/fd3257c5944743de8e7ffe2885a526aa.log.
- Close review round 1: SHIP, two advisory Minors fixed in the round. The make-probe test asserts
  `DeadlineExceeded` in hanging-shell mode (exited-parent ends on the pipe-drain bound, not the deadline, so it keeps
  `err != nil`); the three pid-file wait loops now share `testfix.AwaitPIDFile` (ARCH-DRY), and the bounded-git test
  derives its cue context from its own ctx.
- Close review round 2: SHIP; its advisory (one pid-file loop left) swept as a class: `localcheck_run_process_test`
  and `cli_signal_unix_test` (the reviewer-pid wait, with its exit check as the `stop` hook) now use
  `testfix.AwaitPIDFile` too. Left alone on purpose: the signal test's helper/owner poll (a pid file plus a ticket
  JSON) and the two-checkout lease test's wait (two pids plus two leases), which wait on more than one signal.
