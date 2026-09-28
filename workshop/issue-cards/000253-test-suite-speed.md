---
id: 000253
status: codecomplete
created: 2026-09-27
updated: 2026-09-28
estimate_hours:
github_issue:
started: 2026-09-28T11:24:11-07:00
actual_hours: 1.08
tracker:
    version: 1
    completion:
        token: close-cbc89b0bdac0
        repository: github.com/xianxu/ariadne
        reviewed_head: 3f11b26fe786c88616d165056e42e9813f15172c
        evidence_commit: 34b0bc003e9686a1f2ea73a6c7615686d3d03a99
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
