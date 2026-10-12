---
id: 000339
status: codecomplete
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'aba272b8db790a5193817a34f02dbc51cdf2e655' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T13:20:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 1.06
---

# boundary reviewer wrote outside its checkout (created ~/.git) when mktemp failed: no write fence under a nested sandbox

## Problem

10-12 13:09:46: during ariadne#309's M2 `sdlc milestone-close`, the dispatched boundary reviewer (a claude subprocess; isolation: checkout; write fence unavailable because sandbox-exec can't nest in the agent sandbox, the #204 decision B) had `mktemp` fail and ran its scratch git commands in the operator's home directory. It created `~/.git` (a new repository: one commit "a", ~7,700 loose objects from a partial `git add -A` over home, i.e. copies of home files) and `~/scripts/merge-checks.d/50-é.sh`. While `~/.git` exists, any git command in a non-repository directory under home sees a repository. The auto-mode classifier refused the worker's cleanup; the operator removes it by hand. A reviewer is supposed to be read-only (Review quality train: #204, #310).

## Spec

Fix at the root so a reviewer can't write outside its disposable checkout even without the sandbox-exec fence:
- the reviewer's working directory and TMPDIR are inside its own checkout (created before dispatch, verified writable), and a failure to create them aborts the review rather than falling back to anything outside;
- the reviewer's prompt and environment never default to the home directory (cwd, HOME for git, GIT_DIR/GIT_CEILING_DIRECTORIES set so git can't discover or create a repository above the checkout);
- after each review, a check that nothing outside the checkout changed (at least: no new `.git` above it, no new files in HOME) that fails loudly;
- reproduce: make mktemp fail in a test reviewer run and show nothing is written outside.
Check whether the same path affects the plan-quality reviewer and judges.

## Done when

- A test where the reviewer's temp creation fails shows no write outside the checkout and an aborted review with a clear error.
- The post-review outside-write check exists and is tested.
- The cause is in the Log.

## Plan

Durable plan: `workshop/plans/000339-reviewer-confinement-plan.md`. Single pass, one close.

- [x] Pure pieces: outside-write snapshot diff, git shim (refuses init/clone outside the reviewer's roots)
- [x] Confinement: base dir (inside the checkout, else temp), TMPDIR/CLAUDE_CODE_TMPDIR, GIT_CEILING_DIRECTORIES, shim on PATH; setup failure aborts
- [x] Dispatch wiring for every judge + incident repro test (forced mktemp failure)
- [x] Atlas/helptext; make test

## Log

### 2026-10-12
- 2026-10-12: closed — Repro: judge/confine_unix_test.go runs the incident script (mktemp failing, cd lands in HOME, git init/add/commit) through a real subprocess: no ~/.git, outside write ~/scripts fails Dispatch naming it; unwritable temp parent aborts before the reviewer runs; plain mktemp -d works in the reviewer TMPDIR (red on macOS before the mktemp shim). Mutations: shim off PATH -> repo in HOME; relative operands skipped -> shim table red. reviewbox/confine_test.go: Diff, git shim refuse/allow table (relative, ../ escapes, --git-dir/--work-tree, GIT_DIR), mktemp shim, env file re-asserts after inherited, Confine inside checkout + retry Prepare, Check catches ~/scripts and ancestor .git, unlistable home fails. Live: real claude -p through Dispatch: TMPDIR confined, shims via CLAUDE_ENV_FILE, outside git init refused rc 128. make test: only failures outside the diff (#338 TestPlanningReviewCompetingResults; TestVerbContextsReachTrackerReads on unchanged testfix file, red on main; repolock/weave load flakes pass alone); judge TestMain added for the realgit guard.; review verdict: SHIP
- 2026-10-12: flow upgraded quick → full — 494 added lines in code files (limit 100); an earlier round of this close already ran the full review
- Cause (reproduced): on macOS, `mktemp -d` / `mktemp -t` with no template write under `_CS_DARWIN_USER_TEMP_DIR` (`/var/folders/…/T`) and ignore TMPDIR (man mktemp; run here: "mkdtemp failed on /var/folders/…/T/tmp.…: Operation not permitted" even with TMPDIR set). The reviewer inherits the worker's agent sandbox, which denies that directory, so its `cd $(mktemp -d)` got an empty argument and landed in $HOME, where `git init`, `mkdir scripts/…`, `git add -A`, `git commit -m a` ran. Nothing gave the reviewer a usable temp dir, nothing stopped git creating a repository outside, and nothing checked afterward. Plan-quality, estimate-quality and `sdlc judge` share `judge.Dispatch` and the same exposure (they run in the operator tree, ContainNone, claude's Bash unrestricted).
- Fix: `reviewbox.Confine`, applied by `judge.Dispatch` to every reviewer. TMPDIR/TMP/TEMP/CLAUDE_CODE_TMPDIR point at a base `tmp/` (inside the review checkout, else `$TMPDIR/sdlc-review-*`), made and proven writable first; a failure aborts with "the review did not run". GIT_CEILING_DIRECTORIES is set; `bin/` on PATH holds a git shim (refuses init/clone outside root+base) and, on macOS, a mktemp shim (no-template calls get a TMPDIR template). Post-run `Check` diffs `.git` in every ancestor and $HOME's non-hidden entries; a difference is a dispatch error. Env goes through the run context, so argv and the `Run` seam are unchanged.
- Repro tests (`judge/confine_unix_test.go`, a real subprocess fake claude): the incident script (mktemp failing) gets no `~/.git` and the outside write `~/scripts` fails the dispatch, naming the path; an unwritable temp parent aborts before the reviewer runs; plain `mktemp -d` + a scratch repo work in the reviewer's own TMPDIR (red on macOS before the mktemp shim). Mutation: with the shim dropped from PATH, the incident test fails ("created a repository in HOME").
- Live check (real `claude -p` through `Dispatch`, throwaway test): claude's Bash rebuilds PATH from the shell snapshot, so a PATH shim in the process env never reached its commands (`command -v git` → /usr/bin/git); TMPDIR did reach them (via CLAUDE_CODE_TMPDIR). Fix: the confinement writes `env.sh` and sets CLAUDE_ENV_FILE (sourced by claude before each Bash command; an inherited env file is sourced first). After it: TMPDIR = the base's tmp, git and mktemp resolve to the shims, `mktemp -d` + `git init` inside worked, `git init` in an outside dir was refused (rc 128), and the no-verdict retry re-prepared the base. The nested claude's own sandbox still fails to start (EPERM on its socket), as in production.
- `TestDispatchParentDeadlineCannotBecomeVerdict` used a 20 ms parent deadline, which confinement setup now uses up before the fake reviewer starts; widened to 500 ms so the deadline still lands while the reviewer runs.
- make test (12 shards): every failure outside my diff. `TestPlanningReviewCompetingResults` is #338 (main red; ariadne:1 fixing). `TestVerbContextsReachTrackerReads` flags `internal/testfix/blockingorigin.go`, identical to origin/main, so a static guard red on main. `repolock` TestConcurrentAcquireReclaimsDeadHolderOnce and `weave` TestCompileCancellationStopsLateGeneratorChild failed under load and pass alone. Mine: `TestGitDrivingPackagesPreferRealGit` (judge tests now spawn git) is fixed with a judge TestMain (testgate.Enter + PreferRealGit) and now passes.
- Close review round 1, FIX-THEN-SHIP. BR-1 (Important): the git shim skipped relative operands and leading `--git-dir`. The unconfined reviewer showed both escaping, and it removed what it created. Now fixed: every init/clone operand is resolved, a `..` past the deepest existing dir counts as outside, and git-dir/work-tree options and GIT_DIR/GIT_WORK_TREE are checked. New table cases; mutation (relative operands skipped again) turns them red. Minors fixed in round: env.sh re-exports TMPDIR etc. after an inherited env file (tested); TakeSnapshot errors on an unlistable or unknown home (tested); Confine uses reviewbox's git helper; the plan Cause is marked superseded; the test comment is corrected.
