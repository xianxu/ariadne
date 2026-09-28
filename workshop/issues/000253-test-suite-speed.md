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
flow: {kind: full, provenance: inferred}
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

(Revised twice, see ## Revisions; moved items are struck through.)

- A timing report script exists, and its first output is recorded in the Log.
- ~~`go test -short ./cmd/sdlc/...` finishes in about 60 s or less on the
  development machine.~~ → #261 (in-memory fake makes the fast tier)
- The full suite finishes in at most 5 minutes on the same machine (from
  about 20), via `make test`.
- ~~A guard enforces the budget (short-tier total and per-test cap for non-e2e
  tests), and the e2e marker is checked.~~ → #261
- ~~The sdlc binary is built once per test process.~~ → dropped for now (the
  sharded run meets the target without it)
- Development guidance names how to run the suite and why it is sharded.

## Plan

Scope after the second revision: the two short-term speed-ups, measured.

- [x] Timing report (`go test -json`), recorded in the Log
- [x] Real git on PATH: `testfix.PreferRealGit()` puts `$(git --exec-path)`
  first on PATH; called from a `TestMain` in every package that drives real
  git (cmd/sdlc, gitx, tracker, fleet, activetime)
- [x] Sharded runner `scripts/test-shard.py`: compile the `cmd/sdlc` test binary once, split its
  top-level tests over N processes balanced by the last run's timings (LPT),
  run the other packages alongside; `test2json` output, so
  `scripts/test-timing.py` reads it; non-zero exit on any failure; `make test`
- [x] Fixes the sharded run surfaced: `TestBrainDefaultsExplicitEstimatorOverride`
  passes `SetArgs(nil)` (cobra then reads the test binary's own os.Args);
  the hermeticity guard counts another slot's live `sdlc.lock` as a leak —
  flag it only when this test process holds it; investigate the load-sensitive
  5 s waits and the silent exit-1 shard
- [x] Verify: sharded run green (bar #210) on an idle machine; record wall
  times (serial vs. sharded) in the Log; guidance line in AGENTS.local.md

## Revisions

### 2026-09-28 — scope narrowed to short-term, incremental speed-ups

- **Reason:** the operator split the structural work out. #253 stays a
  short-term incremental speed-up of the real-git tests.
- **Delta:**
  - Spec item 4 (parallel-safe e2e by threading cwd and package seams, then
    `t.Parallel()`) moves to #262 (route all git through one
    directory-explicit seam). The stateful git fake is #261.
  - In: constant-factor wins (direct git binary instead of the macOS xcrun
    shim; one shared sdlc build per test process), the `-short` tier with a
    checked e2e marker, an explicit `-timeout`, the budget guard, and
    `t.Parallel()` only where a test is already parallel-safe without
    refactoring production code.
  - The "full suite ≤ 5 min" done-when is re-set from the measurement after
    the constant-factor wins, not assumed.

### 2026-09-28 — second narrowing: two speed-ups only

- **Reason:** the operator wants the two measured ideas shipped as the
  short-term speed-up before the in-memory fake (#261): real git on PATH
  (1,531 s → 867 s for `cmd/sdlc`) and process sharding (867 s → about 194 s
  over 10 processes in a first, not-yet-green experiment).
- **Delta:** out of #253: the `-short` tier and its e2e marker, the shared
  binary build, and the budget guard. The `-short` tier belongs with #261,
  whose in-memory fake is what makes a fast tier (per ARCH-MOCK: in memory,
  timing under test control, fidelity covered by conformance runs against
  real git). "Full suite ≤ 5 min" becomes: the sharded runner's wall time,
  measured and logged.

## Log

### 2026-09-27

- Filed from #252 (issue-cards tracker), whose e2e tests pushed the full
  suite to about 20 minutes. Numbers above are from #252's full runs at
  4f7260ae (18 packages ok; cmd/sdlc 1,140 s).

### 2026-09-28
- 2026-09-28: closed — make test: 181 s wall balanced (1,531 s serial before); only known failures #210 and sandbox-only processgroup ps; close-review Minors fixed in 3f11b26f (guard red without a TestMain, green with); review verdict: SHIP
- 2026-09-28: closed — make test: 181 s wall balanced (1,531 s serial before); only known failures #210 and sandbox-only processgroup ps; review verdict: SHIP
- 2026-09-28: flow upgraded quick → full — 321 added lines in code files (limit 100)

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
- Implemented (754ba3d4, 1fb3c843): `testfix.PreferRealGit()` from a
  `TestMain` in cmd/sdlc, gitx, tracker, fleet, activetime;
  `scripts/test-shard.py` / `make test`; the three fixes the experiment
  surfaced (cobra reading the test binary's os.Args via `SetArgs(nil)`; the
  guard now flags a lock only when this process holds it or its holder is dead
  — `lockMayBeOurs`; a shared one-minute hang guard in place of 5 s / 2 s
  bounds). The silent exit-1 shard from the experiment did not recur.
- Results, `cmd/sdlc/...` full suite on this machine (12 cores, sandbox):

  | run | wall |
  |---|---|
  | serial `go test`, before | 1,531 s for `cmd/sdlc` alone (needs `-timeout 60m`) |
  | serial, real git on PATH | 867 s for `cmd/sdlc` |
  | `make test`, first run (no timings, round-robin) | 246 s, then 240 s |
  | `make test`, LPT-balanced (shards 156–180 s) | **181 s** |

  Failures in every run are the two known ones only: #210
  (TestFleetPlanHas… reads an archived plan) and processgroup's
  TestCancellationKillsDescendants (sandbox blocks `/bin/ps`).
- One balanced run tripped the hermeticity guard because a commit landed in
  the checkout mid-run (HEAD moved): the guard working as designed. Run the
  suite in a checkout nobody commits to during the run; a `:1+` slot is the
  natural place.
- The operator switched `:0` to main mid-session (the hazard of sharing `:0`);
  switched back with no loss.

