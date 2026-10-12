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
