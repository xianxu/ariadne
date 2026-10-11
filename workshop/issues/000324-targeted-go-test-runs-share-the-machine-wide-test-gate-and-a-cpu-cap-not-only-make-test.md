---
id: 000324
status: working
deps: []
github_issue:
created: 2026-10-10
updated: 2026-10-10
estimate_hours:
card_mirror: 'edfa164a5c3c8a8cd647f80700a791486584a490' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-10T17:53:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] `pkg/testgate` (module-level, so cmd/sdlc, cmd/weave and later pair can import it): `Enter()` at the top of `TestMain`:
  - Takes one of N machine-wide slots: flock(2) on `test-gate.<k>.lock` beside #319's `test-suite.lock`. N comes from `WF_TEST_GATE_SLOTS`, default NCPU/4, at least 1. Each slot file holds the holder's JSON (repo, worktree, package, pid, started). The descriptor is close-on-exec, so the slot is released when the test binary exits, however it exits.
  - Waits loudly: `[test-gate] waiting Ns for a test slot; held by …`, repeated every 30s.
  - Caps `-test.parallel` at `WF_TEST_PARALLEL` (default 2) unless the run set it explicitly.
  - Uses #319's hatches: `WF_TEST_LOCK=off` skips with a loud line, and `WF_TEST_LOCK_TIMEOUT=<s>` gives up with "a lock wait, not a test failure".
  - `WF_TEST_LOCK_HELD_BY`, which `make test` exports, bypasses the gate: the full suite has its own lock and sharding.
- [ ] Wire `testgate.Enter()` into the heavy packages' `TestMain`: cmd/sdlc, internal/{gitx,tracker,fleet,activetime}, cmd/weave, weave internal/{refresh,acquire}. Light packages stay ungated, so they don't queue for slots.
- [ ] Tests (re-exec helper process): a held slot makes a second run wait and print the holder; the escape hatch skips loudly; the timeout gives up with its message; the parallel cap applies only when unset; the make-test bypass works.
- [ ] Measure load and wall time with 4 concurrent targeted-run loops, gate off vs on.
- [ ] Atlas (sdlc-binary.md "Running the test suite") and the hatch documented in the package doc.

ARCH-FUNERAL: N slot files of one JSON record each, rewritten on acquire and never appended. The count is bounded by `WF_TEST_GATE_SLOTS`, and they live in `~/.local/state/ariadne/` beside #319's lock file.

## Log

### 2026-10-10
- Baseline load at claim (17:53): 43.96 / 55.24 / 52.41 on 12 cores. The agent env already has GOFLAGS=-p=3 GOMAXPROCS=3 from the operator's shell, so the env cap exists but doesn't bound the number of concurrent test binaries machine-wide, nor the git forks inside them.
