---
id: 000253
status: working
deps: []
github_issue:
created: 2026-09-27
updated: 2026-09-28
estimate_hours:
card_mirror: '1c3ce1093c13414aedd4a527a73e5cac3533fdc0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T11:24:11-07:00
---

# Keep the sdlc test suite fast: tiers, shared binary, parallel-safe e2e

## Problem

The full Go suite (`go test ./cmd/sdlc/...`) takes about 20 minutes, which is
too slow for the inner development loop and slows every review boundary. Of
about 1,660 test functions in 18 packages, almost all of the time is one
package:

| package | test funcs | wall time |
|---|---|---|
| `cmd/sdlc` | 830 | ~1,140 s |
| `internal/gitx` | 96 | ~64 s |
| `internal/tracker` | 44 | ~23 s |
| the other 15 | ~690 | < 10 s each |

Causes:

- `cmd/sdlc` tests run **serially**. Go parallelizes packages, not tests
  within a package, unless they call `t.Parallel()`. Only 1 of the 127 test
  files does, because many tests change process-global state: the working
  directory (`t.Chdir`/`os.Chdir`), or package-level seams such as `ghClient`,
  `mergeRunner`, `prRunner`, `judge.Run`, `resolveLandingWorkspace` and
  `validateChangedInstancesFn`.
- About 30 test files are **real-Git end-to-end** tests. They build bare
  remotes, clones and slots and drive `sdlc` through real `git`: hundreds of
  process spawns per test, and each spawn is expensive on macOS (more so in
  the sandbox).
- 4 test files **compile the `sdlc` binary** themselves (several seconds each).
- #252 added the largest e2e tests (slot cycle, leftover-branch matrix,
  migration), and the suite grew from about 16 to about 20 minutes.

Fast tests are development hygiene. A slow suite gets run less often, bugs
land later, and reviews wait.

## Spec

1. **Measure first.** Run the full suite with `go test -json` and produce a
   per-test timing report (top N by wall time, grouped by kind: pure,
   real-Git, builds-binary). Keep the report as a script so it can be re-run.
2. **Two tiers.**
   - `go test -short ./cmd/sdlc/...` skips the real-Git e2e tests, through
     one helper (e.g. `testfix.E2E(t)`). This is the inner loop, with a
     budget of about 60 s.
   - The full suite still runs at the gates (`change-code`, `close`,
     `milestone-close`, merge checks) and in CI.
   - A test marked e2e must actually be one: a source guard rejects the
     marker on tests that never touch git.
3. **Shared binary.** Build `sdlc` once in `TestMain`, into a temp dir, and
   share the path with the tests that need an exit-code-level binary.
4. **Parallel-safe e2e.**
   - Thread the working directory and the package-level seams through
     explicit parameters or a per-test environment instead of globals.
   - Then give the real-Git tests `t.Parallel()`.
   - Keep a guard that fails a parallel test that mutates process-global
     state.
5. **A standing budget.**
   - A check (in the suite or the merge checks) fails when the `-short` tier
     exceeds its budget, or when a single non-e2e test exceeds a per-test cap
     (e.g. 2 s).
   - Record the rule in `AGENTS.base.md` / the lessons: new real-Git tests
     are e2e-tier and parallel-safe; unit tests stay fast.

## Done when

- A timing report script exists, and its first output is recorded in the Log.
- `go test -short ./cmd/sdlc/...` finishes in about 60 s or less on the
  development machine.
- The full suite finishes in at most 5 minutes on the same machine (from
  about 20).
- A guard enforces the budget (short-tier total and per-test cap for non-e2e
  tests), and the e2e marker is checked.
- The sdlc binary is built once per test process.
- Development guidance names the tiers and the budget.

## Plan

- [ ] Timing report (`go test -json`), recorded in the Log
- [ ] `-short` tier via one e2e helper, plus its guard
- [ ] Shared binary in `TestMain`
- [ ] Parallel-safe seams for the real-Git tests (cwd and package-level stubs), then `t.Parallel()`
- [ ] Budget guard and development guidance

## Log

### 2026-09-27

- Filed from #252 (issue-cards tracker), whose e2e tests pushed the full
  suite to about 20 minutes. Numbers above are from #252's full runs at
  4f7260ae (18 packages ok; cmd/sdlc 1,140 s).

### 2026-09-28

- Claimed. Timing report script: `scripts/test-timing.py` over
  `go test -json` output (packages, totals by kind, top N).
- Baseline at 8f27d185, M-series Mac with 12 cores, go 1.27.1, Apple Git 2.50.1, in
  the agent sandbox:
  - `go test ./cmd/sdlc/...` **cannot pass as-is**: `cmd/sdlc` hits go's
    default 10-minute `-timeout` about 360 tests in. Any gate/CI run needs an
    explicit `-timeout`.
  - `cmd/sdlc` alone with `-timeout 60m`: **1,531 s** wall, 838 top-level
    tests. By kind (serial sums): real-git 423 tests / 1,226 s;
    builds-binary 74 / 248 s; pure 341 / 57 s.
  - Other packages: gitx 79 s, tracker 32 s, fleet 13 s, the rest < 5 s.
  - Top tests: TransferGuardOverBranchShapes 75 s, LandingRetainsWorkspace
    59 s, LeftoverBranchIsLockedUntilCaughtUpThenWorks 48 s,
    TrackerFullSlotCycle 48 s, DurableRunMergeCompletesTrackedClose… 42 s,
    CLISignalCancelsOwnedReviewer 41 s.
  - Known failures: #210 (TestFleetPlanHas… reads an archived plan);
    processgroup's TestCancellationKillsDescendants fails only in the sandbox
    (`/bin/ps` blocked).
- Tests are serial within a package (no `t.Parallel()` anywhere), so
  `cmd/sdlc` runs on about one core while the other packages finish in 80 s.
- Constant factor found: `/usr/bin/git` on macOS is the xcrun shim, about
  16 ms per spawn against about 5 ms for the real binary
  (`$(git --exec-path)/git`). Measuring `cmd/sdlc` with that directory first
  on PATH before deciding whether parallelism is needed at all.
- Scope: the stateful git fake (ARCH-MOCK, run tests on both backends) is
  split out to #261 at the operator's direction; #253 makes the real-git
  tests fast.

