---
id: 000253
status: open
deps: []
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
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
