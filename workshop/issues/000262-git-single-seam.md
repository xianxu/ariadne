---
id: 000262
status: open
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '286ba17f96704f43175adc8fd9d767627aad10cd' # card fields mirrored from issue-cards; edit via sdlc
---

# Route all git calls through one directory-explicit seam

## Problem

sdlc reaches git through several unrelated paths, and many of them run in
the process's working directory rather than a named repository:

- the `gitRunner` interface (`cmd/sdlc/runner.go`): `Git` runs in cwd,
  `GitInDir` takes a directory;
- `gitx.Capture` / `gitx.RunGit` (about 27 call sites without `-C`),
  `gitx.RepoTopLevel()` (27 call sites) and `gitx.DiffBase`, all cwd-relative;
- the package-level `runGitIn`, `runGitInContext`, `runGitInputContext`
  variables in `internal/gitx/trunkfile.go` and `run` in `window.go`;
- `activetime.gitRun`, and fleet's `GitReader`;
- `os.Getwd()` in `repolock.go`, `propagatebase.go`, `internal/fleet/gitpaths.go`,
  and relative default directories (`workshop/issues`, …).

Consequences: tests must `t.Chdir`/`os.Chdir` (78 sites) and swap package
globals to exercise a verb, which makes them unable to run in parallel; and a
stateful fake (#261, ARCH-MOCK) cannot replace git where there is no seam.

## Spec

- One git seam, owned by `internal/gitx`, whose every call names its
  repository directory explicitly. No git invocation in production code
  outside it (a source guard enforces this).
- The repository root is resolved once at the command boundary (from cwd, or
  a future `-C`) and passed down; production code below the boundary does not
  call `os.Getwd()` or rely on relative paths.
- Package-level git seam variables are replaced by the injected runner, so a
  test supplies its runner per call instead of mutating a global.
- Behavior-preserving: the existing suite passes unchanged in meaning.

## Done when

- A source guard fails on any `exec.Command("git", …)` or cwd-relative git
  call outside the seam, and on `os.Getwd()` below the command boundary.
- The `runGitIn*` / `run` / `gitRun` package variables are gone.
- The real-git tests no longer need `Chdir` to exercise a verb (tests of the
  command boundary itself excepted).

## Plan

- [ ] Inventory call paths (rerun the #253 seam survey)
- [ ] Seam type + guard; migrate internal packages, then package main
- [ ] Drop cwd from tests; mark newly parallel-safe tests

## Log

### 2026-09-28

- Split out of #253 at the operator's direction: #253 stays a short-term
  incremental speed-up of the real-git tests (constant factors, tiers,
  shared binary); this issue is the structural consolidation that unblocks
  parallel tests and #261's stateful fake.
