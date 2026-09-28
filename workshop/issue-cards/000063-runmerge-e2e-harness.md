---
id: '000063'
status: done
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 3
actual_hours: 1.5
---

# e2e test harness for runMerge — die-injectable seam + the two #62 regression tests

## Problem

#62 hardened `sdlc merge` (M1 re-check clean before merge, M2 read-only judges,
M3 resumable cleanup) but shipped only **decision-level** unit tests
(`worktreeDirty`, `decideMergeAction`, all-read-only `AllowedTools`). #62's
own Done-when asked for **end-to-end** regression tests — "judge dirties tree →
merge refuses pre-merge (not post-merge)" and "resume-after-merge finishes
cleanup" — which were deferred (operator-approved, fix-forward). This issue is
that deferral, made trackable instead of buried in #62's archived Log.

`runMerge` resists in-process testing for three reasons (verified):
1. `die()` → `os.Exit(1)` — refusal paths kill the test process. This is why
   *no* `run*` verb is e2e-tested today.
2. `detectRepo()` (cmd/sdlc/fetch.go) and `gitx.RepoTopLevel()`
   (internal/gitx/window.go) call `exec.Command("git"…)` **directly**, bypassing
   the injectable `gitx.run`.
3. Already-injectable seams: `mergeRunner`, `ghClient`, `mergePrompter`,
   `gitx.Capture` (via `gitx.run`).
