---
id: 000063
status: open
deps: []
github_issue:
created: 2026-06-02
updated: 2026-06-02
estimate_hours: 3
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

## Spec

**Approach — real temp repo + stub only the network (gh) + injectable `die`.**
Rather than stub the dozen git calls (and fight #2), stand up a throwaway repo
and run `runMerge` against it:

- `t.TempDir()` → `git init`, commit on `main`, branch `feature`, a local **bare**
  origin, push `feature`. Then `detectRepo`, `RepoTopLevel`, `gitx.Capture`, and
  the real `execGitRunner` (switch/pull/archive/branch-delete) all just work — no
  git stubbing, and the cleanup is exercised *for real*.
- Swap two package vars: `ghClient` (stub `PRListForBranch`/`PRMergedForBranch`/
  `PRMerge`, record calls, never touch GitHub) + `mergePrompter` (or `--yes`).
- **Make `die` injectable**: `var die = func(...)`; a test helper swaps it to
  `panic(&dieSignal{msg})` + recovers, preserving die's halt semantics while
  letting the test assert the refusal message + that `PRMerge` was never called.
  This is the **reusable unlock** — it makes every `run*` verb's refusal path
  testable, so build it as shared test infra, not a one-off.
- For the dirty test, a small `preflightFn` package-var seam so the test can
  inject a "judge" that writes a file into the worktree between step 2 and 9b
  (M2 means the real Specs judge can't dirty anymore, so the test stands in a
  hypothetical dirtying gate to prove 9b's defense-in-depth).

Alternative considered + rejected for the default: subprocess + fake `gh` on
PATH (zero prod changes, tests the real binary) — but slower, asserts "PRMerge
not called" only indirectly, and adds a fake-gh script to maintain. Keep in
reserve as a single black-box smoke test if ever wanted.

## Done when

- `die` is an injectable package var with a test helper (panic+recover);
  documented as the seam for testing any `run*` refusal path.
- Test: **dirty introduced between step 2 and 9b → merge refuses pre-merge**,
  asserting the actionable message AND `ghClient.PRMerge` call-count == 0.
- Test: **resume** (`PRListForBranch=""` + `PRMergedForBranch=true`) → `PRMerge`
  NOT called, ends on `main`, branch deleted, archive ran (real cleanup).
- A reusable `tempRepo(t)` harness other `run*` verb tests can adopt.

## Plan

- [ ] M1 — `die`-injectable seam + `tempRepo(t)` harness; the two e2e tests
  (dirty→refuse, resume→cleanup). Consider a follow-up to route detectRepo/
  RepoTopLevel through `gitx.run` only if a future test needs pure-in-memory.

## Log

### 2026-06-02
