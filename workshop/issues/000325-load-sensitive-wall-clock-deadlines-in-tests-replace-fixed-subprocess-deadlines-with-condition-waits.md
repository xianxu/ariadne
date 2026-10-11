---
id: 000325
status: open
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-11
estimate_hours:
card_mirror: 'f28b62545ea68278a2e215bba6857f3cd501f356' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T15:37:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:0
    worktree: /Users/xianxu/workspace/ariadne
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

- [ ]

## Log

### 2026-10-10
