---
id: 000434
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7af639cd9f9bc60256b64081218c982a37fb5d18' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T21:15:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c8f878ae", done: "88084f14"}
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

- [ ] `testfix.DeadlineOnCue` + its unit test
- [ ] Bounded-git deadline test and make-probe test fire the deadline on cue after readiness
- [ ] Signal test cleanup phase: FIFO-gated shim, no global 50 ms; `localCheckCleanupGit` unit test for the bound
- [ ] Loaded verification: six concurrent loops of all three tests, plus the shrunk-deadline repro now passing

## Log

### 2026-10-17

- Measured on 2026-10-17 (ariadne:2). Mechanism: with the child's cleanup budget shrunk from 50 ms to 2 ms,
  `terminated/cleanup` fails 10/10 with exactly "cleanup never reached its blocked Git seam". Natural rate: 0/40 at
  load average 17. Six concurrent loops of 15 (`-test.count`, a compiled test binary) hit a different failure once,
  "cancelled run retained its worktree" (a registry entry with HEAD 0000 left behind), which is the same 50 ms
  budget killing the ordinary `git worktree prune`: `boundedLandingGitRunner.run` applies the timeout per call.
