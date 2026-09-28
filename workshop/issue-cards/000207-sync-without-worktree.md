---
id: '000207'
status: done
started: 2026-09-09T18:26:20-07:00
created: 2026-09-02
updated: 2026-09-11
estimate_hours: 3.64
actual_hours: 13.09
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

Adjacent but **NOT** fixed here: `syncInPlace` pushes without fetching first, so
an id allocated from a stale local `main` is only discovered when the push is
rejected. An earlier draft claimed this change fixed it incidentally; it does
not, because the Scope below deliberately leaves `syncInPlace` alone. It stays a
separate one-line fix, and saying otherwise would have let it fall through the
gap between two sections of one issue.
