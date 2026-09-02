---
id: 000207
status: open
deps: [ariadne#206]
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# Publish issue files without a main worktree

## Problem

Publishing an issue file to `origin/main` from a feature branch requires that
*some worktree is checked out on main*. `syncViaMainWorktree` locates it via
`git worktree list --porcelain -z`, and when none exists it dead-ends:

```
could not find a worktree on branch 'main'. Is main checked out somewhere?
```

Hit for real on 2026-09-02: an agent ran `sdlc issue new` in `pair`, whose only
checkout had been moved to a feature branch by another actor. The issue file
was written, the ID was never reserved on `origin`, and the stub was left
untracked inside the other actor's working tree.

The route is elaborate because it drives someone else's checkout — find main,
assert it has no uncommitted issue changes, `pull --rebase` it, detect files
changed on both branches since the merge-base, copy the issue file across,
commit, push. Every one of those steps exists to make a *shared working
directory* safe, and each is a way to fail: main can be dirty, main can be
mid-rebase, main can be another actor's active tree, or main can simply not be
checked out anywhere.

Adjacent, and fixed incidentally by the same change: `syncInPlace` pushes
without fetching first, so an ID allocated by `issue.NextID` from a stale local
`main` is only discovered when the push is rejected.

## Spec

**Build the commit in the object database and push it. Never touch a checkout.**

Publishing one small markdown file needs no working directory at all:

```
git fetch origin main
GIT_INDEX_FILE=$tmp git read-tree FETCH_HEAD
blob=$(git hash-object -w --path <relpath> <file>)
GIT_INDEX_FILE=$tmp git update-index --add --cacheinfo 100644,$blob,<relpath>
tree=$(GIT_INDEX_FILE=$tmp git write-tree)
commit=$(git commit-tree $tree -p FETCH_HEAD -m "<subject>")
git push origin $commit:main
```

To be clear about what this is and is not: **it is not remote-side.** Only
`fetch` and `push` touch the network; `origin/main` is a local remote-tracking
ref and every object step reads and writes the local object database. What it
avoids is a *working tree*, which is the thing that can be dirty, absent, or
owned by another actor.

**This deletes machinery rather than adding a fallback.** The clean-main check
and the merge-base conflict detection exist only because the current route
edits a shared checkout. Here, `push $commit:main` is a compare-and-swap: if
`origin/main` moved since the fetch, the push is rejected non-fast-forward.
That is the correct concurrency primitive and it is stronger than the check it
replaces, which can only see *local* divergence.

**Retry is trivial and bounded.** The input is "the current content of this one
file", not a diff, so on rejection: re-fetch, rebuild, re-push. Bound it (3
attempts) and surface the last rejection.

**Scope: replace `syncViaMainWorktree` only.** Leave `syncInPlace` alone. When
the caller is already on main, "here" and "main" coincide and an ordinary
add/commit/push is both simpler and correct — and building out-of-tree there
would leave the caller's own branch behind the commit they just pushed, which
is friction for no gain. (`syncInPlace`'s missing fetch is a separate one-line
fix, not this issue.)

**Details that will bite if unnamed:**

- `hash-object -w --path <relpath>` (not bare `hash-object`) so any
  `.gitattributes` filter or EOL normalization for that path is applied — a
  blob written without them produces a commit whose checkout differs from the
  file.
- File mode is `100644`; issue files are never executable.
- If the repo signs commits, `commit-tree` needs `-S`; a signing repo that
  silently produces unsigned commits is a regression.
- The temp index must be a real temp file, removed on every exit path, and must
  never be `$GIT_DIR/index` — writing that would corrupt whatever checkout
  shares the git dir.

## Done when

- `sdlc issue new` and `sdlc issue sync` publish from a feature branch with
  **no worktree on main anywhere** in the repo.
- They publish while another worktree on main is dirty, mid-rebase, or owned by
  another actor, without reading or writing that worktree.
- A push rejected by a concurrent publisher retries and succeeds; a test drives
  two publishers against one bare origin and asserts both issue files land.
- `syncViaMainWorktree` and its clean-main and merge-base conflict checks are
  deleted, not left as a second path (`ARCH-DRY`) — a shadow sweep confirms no
  caller reaches them.
- Tests run against a real throwaway repo with a local bare `origin`
  (`ARCH-MOCK`: git is the external binary, and a temp repo is its portable
  stateful fake — a function-call mock cannot exercise a non-fast-forward
  rejection).
- The published blob round-trips: checking out the pushed commit yields a file
  byte-identical to the local one, with attributes applied.

## Plan

- [ ] Out-of-tree publish helper behind the existing `gitRunner` seam; pure
      decision logic (paths, subject, retry policy) separated from the git
      calls (`ARCH-PURE`).
- [ ] Bounded retry on non-fast-forward, with the rejection surfaced.
- [ ] Repoint the publish-from-elsewhere arm at it; delete
      `syncViaMainWorktree`, `mainHasUncommittedIssueChanges`, and the
      merge-base conflict detection.
- [ ] Two-publisher concurrency test against a bare origin.
- [ ] Round-trip test for attributes/EOL and for a signing repo.

## Log

### 2026-09-02

Blocked on `ariadne#206`, which is rewriting `syncIssuesToMain` right now —
this issue changes the same dispatch, so it lands after #206 merges rather than
racing it.

Grep confirms sdlc uses **no** git plumbing today: `hash-object`, `commit-tree`,
`write-tree`, `read-tree`, `update-index`, `mktree` and `GIT_INDEX_FILE` appear
nowhere in `cmd/` or `pkg/`. Both sync arms are porcelain over a checkout. So
this introduces the first plumbing use in the tree, which is worth weighing
against the machinery it removes — the argument for it is that every deleted
check exists solely to make a shared working directory safe.

Note sdlc *can* already create worktrees (`branchcreate.go:150,155` do
`worktree add` / `worktree add -b`; `merge.go:577` removes them), so "create a
temp worktree on main when none exists" is the available alternative. Rejected:
it checks out an entire repository to add one markdown file, needs
cleanup-on-failure, and races `git worktree prune` — while still leaving the
clean-main and conflict machinery in place. It solves the missing-worktree case
and none of the others.
