---
id: 000063
status: working
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

- [x] M1 — `die`-injectable seam + `tempRepo(t)` harness; the two e2e tests
  (dirty→refuse, resume→cleanup). Concretely:
  - `term.go`: `die` → `var die = func(...)`; shared `expectDie` test helper
    (swap die → `panic(&dieSignal{msg})`, recover, return the message).
  - `fetch.go`: `detectRepo` → injectable `var detectRepo` — the real one
    demands a github.com origin URL we can't push a local bare origin to;
    the slug it returns is only handed to the (stubbed) `ghClient`, so the
    test swaps it for a dummy.
  - `merge.go` step 5: route `runPreflightJudges` through a package var
    `runPreflightJudgesFn` — this is the dirtying seam (the dirty test injects
    a "judge" that writes a file into the worktree after step 2's clean check
    and before 9b's re-check).
  - Consider a follow-up to route detectRepo/RepoTopLevel through `gitx.run`
    only if a future test needs pure-in-memory (deferred — not needed here).

## Log


- 2026-06-02: closed M1 — go test ./cmd/sdlc/... green; both e2e tests pass; mutation-check: disabling 9b reddens the dirty test (proves teeth); fresh-eyes review found no Critical/Important; review verdict: SHIP
### 2026-06-02

- Filed (deferral from #62 made trackable).
- Implemented M1 (in-place topology):
  - **Three seams**, all minimal `func`→`var` flips (callers unchanged):
    `die` (term.go), `detectRepo` (fetch.go), `runPreflightJudgesFn` (merge.go
    step 5). detectRepo needed the seam because the real one demands a
    github.com origin URL — the harness uses a *local bare* origin so push/pull/
    archive run for real; the slug only feeds the stubbed ghClient.
  - `die_test.go`: `expectDie(t, fn)` → swaps die for `panic(&dieSignal{msg})`,
    recovers, returns `(msg, died)`. Re-raises non-dieSignal panics so genuine
    bugs surface. This is the reusable unlock for any `run*` refusal-path test.
  - `merge_e2e_test.go`: `tempRepo(t)` (git init -b main → seed done issue +
    history/ → push main w/ upstream → branch feature + push) + `e2eGH`
    recorder + `swapMergeDeps`. Two tests:
    - dirty-after-judge → refuses **pre**-merge (asserts 9b message AND
      `PRMerge` call-count == 0).
    - resume (open PR "", merged exists) → no PRMerge, ends on main, feature
      deleted, archive moved 000999-done.md to history/ (real cleanup).
  - **Honored judge findings**: pushed main with upstream so resume's `git pull`
    has origin/main; dirty test keeps `NoJudge=false` so the step-5 injection
    fires; topology is in-place (named in the test header; worktree path left
    untested by this harness).
- **Verification**:
  - `go test ./cmd/sdlc/...` — all green; `go vet` clean.
  - **Mutation check**: disabling the 9b guard (`redirty != "" && false`) makes
    the dirty test go red ("expected merge to refuse") — confirms the test has
    teeth, not just passes by construction. Restored → green.
- gofmt: `gofmt -l` flags merge.go/fetch.go + several untouched files — a
  pre-existing repo-wide drift (newer gofmt reflows doc-comment list indent +
  a pre-existing trailing blank in fetch.go). My added lines are gofmt-clean;
  did NOT reformat untouched files (unrelated churn). Tools: `perl -0pi` for
  the throwaway mutation; `$TMPDIR` (not /tmp) for the backup under sandbox.
- **Milestone-review M1: SHIP** (no Critical/Important). Addressed the minors
  worth keeping: documented the `expectDie` defer-vs-os.Exit semantic gap, the
  no-parallel/no-synchronization caution on `swapMergeDeps`, and a comment-
  accuracy fix on the dirtying judge. Noted **coverage gaps left for a future
  adopter of the harness** (not #63 scope): worktree topology
  (`findMainWorktree` / `worktree remove` / `.goto`) and the `actionNoPR` branch
  are still e2e-untested — natural next targets reusing `tempRepo`/`expectDie`.
