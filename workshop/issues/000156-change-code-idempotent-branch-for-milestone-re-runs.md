---
id: 000156
status: open
deps: []
github_issue:
created: 2026-07-01
updated: 2026-07-01
estimate_hours:
---

# change-code idempotent branch for milestone re-runs

## Problem

`sdlc change-code --issue N` errors at its branching step when re-run for a
**subsequent milestone** on an issue whose feature branch already exists and is
already checked out:

```
==> branching: in-place (default; --worktree=yes for an isolated worktree)
Error: git checkout -b 000153-sdlc-retro-process-manual: exit status 128
fatal: a branch named '000153-sdlc-retro-process-manual' already exists
```

Multi-milestone issues legitimately re-run `change-code` per milestone to get that
milestone's **plan-quality + estimate-quality gates** on the (re-)designed plan. In
the failing run the gates ran fine (both INFO) — only the final branch-creation step
failed, because it unconditionally does `git checkout -b <branch>` instead of
detecting that the branch already exists / we're already on it.

Encountered on ariadne#153 M2 (re-running `change-code` for the M2 milestone on the
existing `000153-*` branch).

## Spec

`change-code`'s branching (the in-place `git checkout -b` path in
`cmd/sdlc/changecode.go`, and the worktree path in `createWorktreeBranch`) should be
**idempotent**:

- If already on the target branch → info message ("already on branch X"), skip
  creation, proceed (exit 0).
- If the branch exists but isn't checked out → `git checkout <branch>` (switch), don't
  `-b`.
- Only `git checkout -b` when the branch doesn't exist.

The plan-quality + estimate-quality gates should run regardless (they already do).

## Done when

- Re-running `sdlc change-code --issue N` for a 2nd milestone, while already on that
  issue's branch, runs the gates and exits 0 — no "branch already exists" error.
- A test covers the branch-already-exists / already-checked-out path.

## Plan

- [ ] Make `change-code` branching idempotent in `cmd/sdlc/changecode.go`: detect
      already-on-target-branch (skip, info) and branch-exists-but-not-checked-out
      (`git checkout`, not `-b`); only `-b` when absent. Cover the branch-exists path
      with a test.

## Log

### 2026-07-01

Filed from the ariadne#153 M2 flow (see that issue's Log, "change-code idempotency
bug"). Note: `sdlc issue new` could not auto-sync this to `main` (main not checked out
in a worktree while on the 000153 branch) — this file needs to land on main separately,
NOT bundled into #153's PR.
