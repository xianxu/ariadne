# Issue Sync

Syncs `workshop/issues/` changes to main and pushes to origin, even from a feature branch. This enables using issue files as a coordination/locking mechanism across branches and collaborators.

The canonical command is `sdlc claim`. The old `make issue-sync` target is a compatibility wrapper for repos that still enter workflow through Makefile targets.

## Usage

```
sdlc claim --issue <N>
```

## Behavior

### On main

Stages changed and untracked files in `workshop/issues/` (or just the selected issue with `--issue`), commits, and pushes to origin.

### On a feature branch (worktree)

1. Identifies changed + untracked files in `workshop/issues/` on the feature branch
2. Finds the main worktree and verifies it's on `main`
3. Pulls latest main from origin (`git pull --rebase`)
4. Computes the merge base and checks for conflicts (files changed on both sides)
5. **No conflicts**: copies files to main worktree, commits, pushes
6. **Conflicts detected**: stops and prints step-by-step resolution instructions

## Conflict detection

A conflict is when the same issue file was modified on both:
- The feature branch (since it diverged from main)
- Main itself (since the merge base)

When this happens, the script stops and tells the user exactly which files conflict and how to resolve them manually in the main worktree.

## Why

Issue state changes (status, assignment) need to be visible on main immediately, not deferred until a feature branch merges. This avoids two people picking up the same issue, and keeps the `workshop/issues/` folder on main as the single source of truth for coordination.

## `issue new` auto-syncs (#82 M1)

`sdlc issue new` also broadcasts the freshly-scaffolded file to origin/main, through the **same** branch-aware sync as `claim` (the shared `syncIssuesToMain` dispatch in `claim.go`, filtered to the new issue's `--issue`). Filing an issue therefore lands it on main as tracker state — not untracked working-tree residue that every symlinked derivative reads and that dirty-tree gates trip over. The filtered add (per #80) stages only the new file, so unrelated untracked WIP is left alone. On `main` the working tree is left clean; on a feature branch the file routes to the main worktree (any local copy left behind is non-blocking — see [base-layer.md](base-layer.md), #82 M2).

## Implementation

- Binary: `cmd/sdlc/claim.go` (`syncIssuesToMain` — shared by `claim` + `issue new`)
- Compatibility wrapper: `make issue-sync` in `Makefile.workflow`; when
  `bin/sdlc` is absent it builds Ariadne's `cmd/sdlc` source to a temporary
  binary, then runs that binary from the consumer repository cwd.
