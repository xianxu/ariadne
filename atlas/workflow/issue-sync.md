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

### Anywhere else — a feature branch, a worktree, a detached HEAD

Publishes straight to the trunk; **no checkout is read or written** (#207).

1. Identifies changed, staged, untracked and deleted files in `workshop/issues/`
   (or just the selected issue with `--issue`), reading NUL-delimited paths so a
   non-ASCII filename is never mistaken for a missing one
2. Fetches `origin/main` and builds a tree in a temp index on top of it
3. `commit-tree`, then pushes `<commit>:refs/heads/main` as a compare-and-swap
4. On rejection: re-reads the trunk, re-derives the whole change set, retries
   (bounded). Deriving per attempt is what stops a retry from re-publishing an
   id a peer took in the meantime

This needs no worktree on `main`, which matters because `change-code` branches
**in place**: an actively-worked repo usually has none, and the previous route —
find the main worktree, refuse if dirty, `pull --rebase` it, copy files across,
commit and push there — was unavailable in exactly that case.

## Concurrency, and one guarantee that was LOST

The push is a compare-and-swap, so a peer landing between the read and the push
is detected by git rather than guessed at, and the retry rebuilds against what
they wrote. Two agents editing **different** issues therefore both land.

An **id collision** — a different slug already holding this issue's id — is
never auto-merged: `claim` and `issue sync` refuse and name both paths, because
a published id is already referenced by the branch name, commit subjects,
`deps:` and review sidecars (ariadne#188). Only `issue new` re-allocates, where
nothing references the id yet.

**Two agents editing the SAME issue file is now last-writer-wins.** The route
this replaced computed a merge base and refused when a file had changed on both
sides; the object-database route has no working tree to diff against, and the
naive replacement is worse than nothing — trunk commits are built outside the
branch, so a three-way check against the merge base would refuse every ordinary
republish. Stated here rather than left for someone to discover: a stale
double-claim silently overwrites the other agent's edit. Tracked as a
lost-update follow-up (ariadne#207 BR-5).

## Why

Issue state changes (status, assignment) need to be visible on main immediately, not deferred until a feature branch merges. This avoids two people picking up the same issue, and keeps the `workshop/issues/` folder on main as the single source of truth for coordination.

## `issue new` auto-syncs (#82 M1)

`sdlc issue new` also broadcasts the freshly-scaffolded file to origin/main, through the **same** branch-aware sync as `claim` (the shared `syncIssuesToMain` dispatch in `claim.go`, filtered to the new issue's `--issue`). Filing an issue therefore lands it on main as tracker state — not untracked working-tree residue that every symlinked derivative reads and that dirty-tree gates trip over. The filtered add (per #80) stages only the new file, so unrelated untracked WIP is left alone. On `main` the working tree is left clean; anywhere else the file is published straight to the trunk, and the local copy stays put. If the id was taken on the trunk, `issue new` re-allocates and prints the id it actually published.

## Implementation

- Binary: `cmd/sdlc/claim.go` (`syncIssuesToMain` — shared by `claim` + `issue new`);
  the publish arm is `cmd/sdlc/synctrunk.go` over `gitx.TrunkFile.UpdateMany`
- Compatibility wrapper: `make issue-sync` in `Makefile.workflow`; when
  `bin/sdlc` is absent it builds Ariadne's `cmd/sdlc` source to a temporary
  binary, then runs that binary from the consumer repository cwd.
