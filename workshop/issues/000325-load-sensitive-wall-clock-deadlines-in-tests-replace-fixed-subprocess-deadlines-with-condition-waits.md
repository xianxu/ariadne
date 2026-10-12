---
id: 000325
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-12
estimate_hours:
card_mirror: '74c890a417c41f4d4fc1c6577a2014a106b655c2' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T15:37:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "4e38862a", done: "70bbe13f"}
---

# Load-sensitive wall-clock deadlines in tests: replace fixed subprocess deadlines with condition waits

## Problem

Tests with fixed wall-clock deadlines fail under machine contention. `TestPlanningReviewConcurrencySchedules` fails its 2s subprocess deadline at load ~50, 4/4 on main and 4/4 on rerun (ariadne:2, 10-10). With several slots running suites in parallel that load is routine, so these failures are noise that agents learn to ignore, and real regressions hide behind them.

## Spec

Find tests whose pass/fail depends on a short wall-clock deadline (subprocess timeouts, sleeps, `time.After` races). Replace them with condition waits (wait for the event with a generous ceiling), or scale the deadline by an explicit test-time factor. Keep the assertion's intent (e.g. "runs concurrently" is checked by ordering or overlap evidence, not by elapsed time under 2s).

## Done when

- `TestPlanningReviewConcurrencySchedules` passes 10/10 with the load near 50 (e.g. a CPU burner running alongside it), and still fails if concurrency is broken (mutation check).
- A sweep lists the other deadline-sensitive tests in `## Log`, each fixed or justified.
- Both `testfix.Quarantine(t, "#325", …)` calls added by ariadne#330 are deleted (`TestPlanningReviewConcurrencySchedules`, and the 20 s start-deadline step of `TestReviewConcurrencySchedules`), and both tests run unquarantined (asked by #330's close review).

## Plan

Design: the short deadlines in both tests are hang guards standing in for concurrency evidence. The evidence that an unrelated mutation does not wait behind a review is the repo lock's own wait line (`waiting for sdlc repo lock held by …`, printed the moment a second acquirer finds the lock held), so the tests watch for that line and fail on it at once; every deadline left becomes one shared hang ceiling that catches a hang, not a slow machine.

- [ ] `repolock.WaitMessagePrefix`: the wait line's text, single-sourced (ARCH-DRY) so tests key on it
- [ ] `testfix.HangCeiling` + `testfix.Watch` (an io.Writer that signals when its output contains a needle)
- [ ] `TestPlanningReviewConcurrencySchedules`: unrelated `set-status` fails on the lock-wait line, not on a 2 s deadline; barrier/CLI/reap bounds use the ceiling; quarantine deleted
- [ ] `TestReviewConcurrencySchedules`: 20 s start deadline, 3 s unrelated-sync deadline and 7 s termination waits → lock-wait evidence + ceiling; quarantine deleted
- [ ] Load check: `TestPlanningReviewConcurrencySchedules` 10/10 under a CPU burner (load ~50); mutation check (review holds the lock) fails it
- [ ] Sweep of other deadline-sensitive tests in `## Log`, each fixed or justified

## Log

### 2026-10-10

### 2026-10-12 — ariadne:3

Shared pieces: `testfix.HangCeiling` (2 min) is the one hang bound for cmd/sdlc tests, and `closereview_test.go`'s `hangGuard` now aliases it (ARCH-DRY). `testfix.Watch` signals when output contains a needle. `repolock.WaitMessagePrefix` single-sources the lock's wait line so tests key on it.

Sweep (grep of `time.After`/`WithTimeout`/`Now().Add`/`Since`/`Sleep` across `*_test.go`, classified A = verdict rides on how long real work takes, B = hang guard, C = fine):

| Site | Was | Now |
|---|---|---|
| `planningreview_test.go` TestPlanningReviewConcurrencySchedules | 2 s for a built-binary `set-status`; 15 s barrier; 30 s CLI; ~20 s reviewer spin | fails on the lock-wait line; everything else `HangCeiling`; quarantine deleted |
| `review_concurrency_unix_test.go` TestReviewConcurrencySchedules | 20 s reviewer start; 3 s unrelated sync; 7 s termination | sync fails on the lock-wait line; ceiling elsewhere; quarantine deleted |
| `close_finalize_test.go` (6 sites) | 2 s / 5 s for an in-process close after release (real git) | `hangGuard` |
| `judge/cancellation_unix_test.go` | elapsed < 1 s including two re-exec starts; 2–3 s waits | cancel test: elapsed (from cancel, both processes up) < production grace, which still proves the injected 50 ms governed; drain test: `ErrWaitDelay` is the evidence, no elapsed bound; waits at ceiling |
| `processgroup/process_unix_test.go` | 3 s ready, 4 s select, 2 s cleanup/AwaitExit | ceiling; the 5 s shutdown check is kept: it is the product contract, not a guard |
| `cli_signal_unix_test.go` (5 sites) | 3/5/8/20 s around a spawned CLI | `hangGuard` |
| `weave/acquire/git_refresh_test.go` | elapsed < 3 s | overflow: the error names the limit that fired, timeout raised to 1 min, elapsed check dropped; cancel: < 15 s against the fixture's 30 s sleep |
| `ghlanding_test.go` cancel | < 3 s | < 15 s against the fake gh's 30 s sleep |
| `weavefs/runner_owned_test.go` | 5 s ctx (normal mode fails if it fires), 3 s | 1 min hang ceilings (weave cannot import cmd/sdlc's testfix) |
| `weave/acquire` TestRestoreWithRawGitPaths (found failing under load during this work: `git clone … context deadline exceeded`) | `ExecGit.Timeout: 1 s` per git op | 1 min: the test is about raw paths, not timeouts |
| `repolock_test.go` real-clock concurrency tests | `WaitTimeout: 1 s` (not under test) and a 1 s wait for the second acquirer to queue | 1 min hang ceilings |
| `testgate_test.go` TestDirectRunShrinksTimeoutByTheWait | release after `sleep 1.5s` (helper might not have queued yet) | release on the helper's `[test-gate] waiting` line |

B-class hang guards of 10–15 s on subprocess work became 1 min (weave: generator_lifetime ×3, release, staging, runner_setup, plan/publication) or `hangGuard` (fleet_e2e's git call); testgate's nested re-exec too (its `nested err=<nil>` is asserted). Justified, left as is: fleet_e2e's 30 s binary run; fleet/recordsdeadline (< 5 s around a 500 ms injected deadline) and boundarypush (< 10 s) assert the injected deadline bounded git, so their slack is the margin, not the guard. Late-writer negative checks (sleep 150 ms then assert absent) can only pass falsely under load, never fail. Fake-clock and injected-sleep tests (integrategate, catchup, estimatesource, processmanual, reviewbox, plan/apply, heartbeat, repolock) do not read the wall clock. If one of the B rows flakes it gets the same treatment.

