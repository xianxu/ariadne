---
id: 000324
status: codecomplete
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: '9199b9d356d2713dfcd1ea0759357a19d5595587' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T17:53:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.98
---

# Targeted go test runs share the machine-wide test gate and a CPU cap, not only make test

## Problem

ariadne#319's machine-wide lock covers `make test` only. Agents iterate with targeted `go test ./cmd/sdlc/ -run …`, which bypass it. On 10-10 at ~17:40 (pair:4's diagnosis, ops learnings kinks 53–54), two such runs from two ariadne slots ran at ~140% CPU each for 14+ minutes, forking git heavily; load hit 72–87 on 12 cores. Effects: Couch exited (its terminal stopped reading output for 5s, pair#430), and ariadne tests with 2s subprocess deadlines failed 4/4 (#325). `GOFLAGS=-p=3` limits packages only; in-package `-parallel` still defaults to GOMAXPROCS=12 per run, per slot.

## Spec

Every Go test run an agent starts in a woven repo goes through one machine-level budget, not only `make test`:
1. Cap per-run parallelism in the slot environment that weave or the launcher sets (e.g. `GOFLAGS='-p=2 -parallel=2'`, and GOMAXPROCS of about NCPU divided by active slots). Pick the mechanism that reaches agents' direct `go test` calls.
2. Make targeted runs take the #319 lock or a counting semaphore (N concurrent runs machine-wide), with the same loud waiting and escape hatches as #319.
3. Measure before and after: load average and test wall time with 4 slots iterating.

## Done when

- With 4 slots each running targeted `go test` in a loop, load stays under about 2× NCPU, and Couch and the tests stay up.
- A direct `go test` from an agent shell honors the cap (test: env inspection or a fake runner).
- The escape hatch is documented and loud.

## Plan

Mechanism (decided 10-10): gate **inside the test binary**, not in the slot env. Env is fixed at launch and differs per harness (pair#203 makes the same point). A `go -exec` shim needs an executable on PATH, and slots only have shell functions. A `TestMain` runs on every invocation of the package's tests, whatever the harness, shell or flags.

- [x] `pkg/testgate` (module-level, so cmd/sdlc, cmd/weave and later pair can import it): `Enter()` at the top of `TestMain`:
  - Takes one of N machine-wide slots: flock(2) on `test-gate.<k>.lock` beside #319's `test-suite.lock`. N comes from `WF_TEST_GATE_SLOTS`, default NCPU/4, at least 1. Each slot file holds the holder's JSON (repo, worktree, package, pid, started). The descriptor is close-on-exec, so the slot is released when the test binary exits, however it exits.
  - Waits loudly: `[test-gate] waiting Ns for a test slot; held by …`, repeated every 30s.
  - Caps `-test.parallel` at `WF_TEST_PARALLEL` (default 2) unless the run set it explicitly.
  - Uses #319's hatches: `WF_TEST_LOCK=off` skips with a loud line, and `WF_TEST_LOCK_TIMEOUT=<s>` gives up with "a lock wait, not a test failure".
  - `WF_TEST_LOCK_HELD_BY`, which `make test` exports, bypasses the gate: the full suite has its own lock and sharding.
- [x] Wire `testgate.Enter()` into the heavy packages' `TestMain`: cmd/sdlc, internal/{gitx,tracker,fleet,activetime}, cmd/weave, weave internal/{refresh,acquire}. Light packages stay ungated, so they don't queue for slots.
- [x] Tests (re-exec helper process): a held slot makes a second run wait and print the holder; the escape hatch skips loudly; the timeout gives up with its message; the parallel cap applies only when unset; the make-test bypass works.
- [x] Measure load and wall time with 4 concurrent targeted-run loops, gate off vs on.
- [x] Atlas (sdlc-binary.md "Running the test suite") and the hatch documented in the package doc.

ARCH-FUNERAL: N slot files of one JSON record each, rewritten on acquire and never appended. The count is bounded by `WF_TEST_GATE_SLOTS`, and they live in `~/.local/state/ariadne/` beside #319's lock file.

## Log

### 2026-10-10
- 2026-10-10: closed — Round 2: fixed BR-1..3 (hatches stand aside fully, with GOMAXPROCS-default helper assertions for both bypass and off; slot record without git fork; weave guard TestGitDrivingPackagesTakeTheTestGate). BR-4 kept with reason. pkg/testgate tests ok; go test fleet, gitx and all of cmd/weave/... green (gated). The round-1 full make test evidence stands: 12 shards green, processgroup fails only on the sandbox ps denial.; review verdict: SHIP
- 2026-10-10: closed — make test (waited 184s on #319 lock behind a pair suite): 12 sdlc shards 0 failed (1026 tests), packages green incl. pkg/testgate, except processgroup TestCancellationKillsDescendants which fails only on the sandbox fork/exec /bin/ps denial and passes outside it (untouched). pkg/testgate tests: free slot, loud wait naming holders then acquire, full-suite yield (names make test holder; probe never keeps suite lock), timeout as lock wait, budget go-ahead, off hatch loud + creates nothing, env config, re-exec helper proving a direct run of a gated binary takes a slot + parallel cap (explicit -parallel wins, make-test bypass silent) + shrinks -test.timeout by the wait. TestGitDrivingPackagesPreferRealGit now enforces testgate.Enter in git-driving packages. Measurements in Log: quiet machine 4 loops gate off load mean 7.7/max 8.4 (71s) vs on 8.6/10.6 (85s); with full suites the gate held targeted runs back (load 43 to ~12-16); gate-off under suites hit max 52.8 + a #325 deadline flake.; review verdict: SHIP
- 2026-10-10: flow upgraded quick → full — 263 added lines in code files (limit 100)
- Baseline load at claim (17:53): 43.96 / 55.24 / 52.41 on 12 cores. The agent env already has GOFLAGS=-p=3 GOMAXPROCS=3 from the operator's shell, so the env cap exists but doesn't bound the number of concurrent test binaries machine-wide, nor the git forks inside them.
- Implemented `pkg/testgate`. It is wired into the 5 git-driving cmd/sdlc packages (now enforced by `TestGitDrivingPackagesPreferRealGit`) and into cmd/weave plus weave's refresh/acquire.
- Ambient load (18:12) was mostly ariadne:2's full `make test` (12 `sdlc.test` shards) plus targeted runs, so a gate that only counted targeted runs would miss the biggest source. Added the spec's "take the #319 lock" half: gated binaries wait while a full suite holds `test-suite.lock`. They probe with a shared flock and drop it at once, so targeted runs can't starve a waiting suite.
- The measurement caught a real bug: the wait spends `go test`'s timeout (cmd/go kills the binary at timeout + 1m from its start). Two runs that queued 10+ min behind a pair suite were SIGQUIT-killed. Fix: wait at most timeout/4, then go ahead without a slot (loudly), and shrink `-test.timeout` by the wait. Both are tested.
- Limitation: in package-list mode `go test` drops a passing binary's output, so wait lines show only with -v, on a failure, or in local-directory mode. A binary can't reach the go command's terminal. Documented with how to see the holders (the slot files and test-suite.lock).
- Measurements, 4 concurrent loops x 2 runs of `go test ./cmd/sdlc/internal/fleet/ ./cmd/sdlc/internal/activetime/` (8 binaries at once), load sampled every 5s:
  - quiet machine, gate off (off3): 71s total, load mean 7.7 / max 8.4; earlier off2: 54s, 14.4 / 16.7.
  - quiet machine, gate on (on3): 85s total, load mean 8.6 / max 10.6, at most 3 binaries at once.
  - during full suites, gate on (on2): the targeted runs held back. Load fell 43 → 12 while ariadne:2's suite finished and stayed around 12–16 under pair:3's suite. This run surfaced the timeout bug above.
  - during full suites, gate off (first run): load mean 34.9 / max 52.8, and one #325-shape deadline flake (fleet `TestHangingRemoteDegradesOnlyItsRepository`).
  - Reading: targeted loops alone stay well under 2×NCPU (24). The spikes are full suites overlapping with targeted runs, which the gate now separates. The full suite's own 12 shards on 12 cores stay #319's design; reducing them is a follow-up for the TL to decide.
- Close review round 1: SHIP with 4 Minors.
  - BR-1 fixed: the make-test bypass and `WF_TEST_LOCK=off` now stand aside entirely, with no parallel cap. The helper test asserts go's own default (GOMAXPROCS=7 → parallel=7).
  - BR-2 fixed: the slot record no longer forks `git rev-parse`. The package dir names the repo.
  - BR-3 fixed: `TestGitDrivingPackagesTakeTheTestGate` in cmd/weave enforces the gate on git-driving weave packages. Today's set matches exactly.
  - BR-4 kept: the gofmt change in publishgate_test.go fixes my own #320 table, and reverting it would leave main misformatted.
