---
id: 000339
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'cb46d6266bb8264dafff6bc81a3cbd5effb06650' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T13:20:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "622ef2a9", done: "1a2ed4c2"}
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

- [ ] Pure pieces: outside-write snapshot diff, git shim (refuses init/clone outside the reviewer's roots)
- [ ] Confinement: base dir (inside the checkout, else temp), TMPDIR/CLAUDE_CODE_TMPDIR, GIT_CEILING_DIRECTORIES, shim on PATH; setup failure aborts
- [ ] Dispatch wiring for every judge + incident repro test (forced mktemp failure)
- [ ] Atlas/helptext; make test

## Log

### 2026-10-12
- Cause: Claude Code's Bash sets TMPDIR from CLAUDE_CODE_TMPDIR, inherited by the reviewer from the worker's session (/tmp/claude-501); in the nested sandbox mktemp failed there, the reviewer's `cd "$(mktemp -d)"` setup landed in $HOME (cd with an empty argument), and its git init / add -A / commit ran in home. The dispatch gave the reviewer no temp dir of its own, nothing stopped git creating a repository above the checkout, and nothing checked afterward. Plan-quality, estimate-quality and `sdlc judge` share the same dispatch and the same exposure (they run unisolated, in the operator tree).
