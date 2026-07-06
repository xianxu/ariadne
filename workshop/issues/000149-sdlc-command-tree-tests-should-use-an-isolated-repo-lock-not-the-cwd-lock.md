---
id: 000149
status: codecomplete
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-05
estimate_hours: 0.6
started: 2026-07-05T21:50:49-07:00
actual_hours: 0.35
---

# sdlc command-tree tests should use an isolated repo lock, not the cwd lock

## Problem

Command-tree tests that drive `buildRoot().Execute()` (e.g.
`TestSetStatusAlias_BothPathsMutate` and peers) acquire the **real cwd-based
repo transaction lock** (`.git/sdlc.lock`) rather than an isolated, per-test
lock. Consequences:

- `go test ./cmd/sdlc/...` **hangs to timeout** whenever a live `sdlc` command
  holds the lock concurrently — surfaced during the #140 close boundary review:
  the suite blocked on the lock held by `pid …: sdlc close --issue 140`.
- The affected tests are **non-hermetic and un-parallelizable across processes**:
  they contend on one machine-global lock keyed off the checkout's cwd.

This is pre-existing (outside the #140 window) — flagged by the #140 close
boundary review under "test-hygiene backlog".

## Spec

Make command-tree tests hermetic w.r.t. the repo lock: a test invoking
`buildRoot().Execute()` should serialize (if at all) on a lock rooted in the
test's own temp dir, not the developer's real checkout. Two candidate shapes:

- inject the lock directory/root through the command wiring so tests point it at
  `t.TempDir()`; or
- have the lock key off the resolved repo root the command operates on (already
  temp in these tests) rather than process cwd.

Prefer the option that keeps production behavior identical (the lock must still
serialize real mutating verbs on the real checkout) while letting tests isolate.

## Done when

- [ ] `go test ./cmd/sdlc/...` runs green while an unrelated live `sdlc`
      mutating command holds `.git/sdlc.lock` (no hang).
- [ ] The affected command-tree tests acquire a lock under a test-local temp
      dir, not the real checkout's `.git/sdlc.lock`.
- [ ] Production lock behavior unchanged — real mutating verbs still serialize on
      the common-dir lock (`internal/repolock` semantics intact).
- [ ] **(folds #165)** A package-level `TestMain` in `cmd/sdlc` FAILS the run
      (non-zero) if the package's tests changed the REAL repo's HEAD, branch,
      tracked/untracked set, or created `.git/sdlc.lock` — the backstop for BOTH
      the lock-grab (#149) and tree-mutation (#165, the session incident) classes.
      Pre-existing untracked files must not trip it; a proving test confirms it fires.

## Fold note (#165)

This issue now also delivers #165 (test-pollution guard). Same root cause: a
`cmd/sdlc` test invokes sdlc code that resolves the repo from cwd (`git rev-parse
--git-common-dir`), and cwd (`cmd/sdlc/`) is inside the real repo — so it grabs the
real lock (#149) and/or mutates the real tree (#165, which corrupted `main` this
session). One fix (isolate offenders into a temp repo) + one backstop (the TestMain
guard, whose `snapshotDiff` covers both symptom classes) — see the durable plan.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module   design=0.15 impl=0.2
item: smaller-go-module   design=0.1  impl=0.15
design-buffer: 0.15
total: 0.6
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*
Two focused test-infra chunks: (1) the pure `snapshotDiff` + `TestMain` guard;
(2) isolating the lock offenders into a temp repo + the proving/acceptance tests.
`impl=` at v3.1's 40%; familiarity 1.0 (just spent the session deep in `cmd/sdlc`
close/merge/lock code + diagnosed this exact non-hermeticity).

## Plan

Durable design: [`workshop/plans/000149-cmd-sdlc-test-hermeticity-plan.md`](../plans/000149-cmd-sdlc-test-hermeticity-plan.md).
Root cause (verified): `repoLockGitCommonDir()` (`repolock.go:115`) resolves the
common dir via `git rev-parse` from cwd → the REAL repo when a test doesn't isolate.
Single-pass. Fixes #149 + folds #165.

- [x] Pure guard core: `repoSnapshot` + `snapshotDiff` (new-mutations-only) + `guardVerdict` — unit-tested (`TestSnapshotDiff`, `TestGuardVerdict`).
- [x] `TestMain(m)` in `cmd/sdlc`: snapshots the real repo (HEAD/branch/porcelain/`.git/sdlc.lock`) from the initial cwd, before+after `m.Run()`; fails a passing run that left durable damage.
- [x] Isolate the lock offender: `hermeticRepo(t)` + chdir in `TestSetStatusAlias_BothPathsMutate`. (Sole offender — inspection: other `buildRoot().Execute()` tests are read-only or chdir'd; `issue new` tests call `runIssueNew` directly; `NewCloseCmd` isn't lock-wrapped.)
- [x] Guard-fires proof: `TestGuardVerdict` (pure decision) — fires on passing+mutation, doesn't mask a failing run.
- [x] `go build/vet/test ./...` green + guard 0 false-fires; acceptance `TestRepoLock_IsolatedFromRealRepo` (lock never resolves under the real checkout); atlas; close (#165 delivered).

## Log

### 2026-07-01

- Filed from the #140 close boundary review (SHIP), "test-hygiene backlog" note:
  `buildRoot().Execute()` tests grab the real cwd repo lock, so the suite hangs
  when a live `sdlc` (here `sdlc close --issue 140`, pid 82204) holds it.

### 2026-07-05 (implemented)
- 2026-07-05: closed — go build/vet/test ./... — all packages green; new cmd/sdlc TestMain hermeticity guard 0 false-fires + real repo unchanged across the run. #149: TestSetStatusAlias_BothPathsMutate (sole lock offender — others verified read-only/chdir-d/direct-run/unwrapped) now chdirs into hermeticRepo(t) so its lock resolves to the temp .git; acceptance TestRepoLock_IsolatedFromRealRepo asserts repoLockGitCommonDir never resolves under the real checkout. #165 (folded): TestMain snapshots the real repo (HEAD/branch/porcelain/.git/sdlc.lock) before+after m.Run() and fails a passing run that left durable damage — the backstop that would have caught the #148-session main corruption; snapshotDiff (new-mutations-only, pre-existing untracked ignored) + guardVerdict (does not mask a failing run) unit-tested. Production lock behavior unchanged.; review verdict: SHIP

change-code plan-quality **CLEAN**; estimate INFO (advisory). Implemented + verified:
`go build/vet/test ./...` — all packages green; the new `cmd/sdlc` `TestMain` guard
0 false-fires and the real repo is unchanged across the run. Root cause confirmed
(`repoLockGitCommonDir` resolves the common dir from cwd → the real repo). Sole lock
offender was `TestSetStatusAlias_BothPathsMutate` (the issue's "and peers" was
speculative — verified the other `buildRoot().Execute()` tests are read-only or
already chdir'd, `issue new` tests bypass the wrapper, `NewCloseCmd` isn't wrapped).
**#165 delivered here**: its Done-when (a TestMain that fails on real-repo mutation)
is met by the guard, whose `snapshotDiff` also flags a leaked `.git/sdlc.lock`.
