---
id: 000213
status: open
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# Allocate issue IDs against origin/main

## Problem

`issue.NextID` scans **local files only** —
`cmd/sdlc/internal/issue/scaffold.go:31`, walking `issuesDir`, `historyDir` and
the archive subdir for the highest 6-digit prefix. A feature branch cut before
some issue landed on `main` has a `workshop/issues/` that never contained it, so
`sdlc issue new` on that branch reallocates ids that already exist.

**This is not a race.** Measured in `pair`, 2026-09-02:

```
12:10:10  branch 000170-… cut from main (88fe1de0)
12:42:03  #171, #172 land on origin/main
15:43:29  actor files its own #171 on the branch
17:07:08  actor files its own #172
```

Three hours after publication. Any `sdlc issue new` on that branch allocates
`171`, whether run then or next week — the branch simply does not contain the
newer files. The result is four files and two ids:

| id | on `origin/main` | on the branch |
|---|---|---|
| 171 | Always-on idle notification fallback | Reconcile stale incarnations left by a crashed couch |
| 172 | Clickable status bar switches to that actor | Parallelize the zellij session snapshot |

**The repo transaction lock is the wrong tool and demonstrably did not help.**
Both allocations ran in linked worktrees of the same repo, so they shared one
`.git/sdlc.lock` and *were* serialized. They still collided. The lock prevents
concurrent access to a shared view; it cannot reconcile two disjoint views.

**And nothing detects it.** Before `#206`, an on-main sync pushed, so a stale id
was caught by a rejected push. `#206`'s `syncInPlace` commits to the current
branch without pushing — correct for durability, and it removes the accident
that used to surface this. The collision now surfaces at merge, as two files
claiming `id: 000171`, or not at all.

## Spec

**Allocate against the published id space, not the checkout's.**

- Before allocating, `git fetch origin main` (or the configured trunk), then
  compute the next id from the **union** of the local scan and the ids present
  in `origin/main`'s issue directories. Union, not replacement: unpushed local
  issues on this branch are real and must still be excluded.
- Read the remote ref's tree directly — `git ls-tree <ref> <dir>` for the three
  directories `NextID` already scans. Filenames carry the id, so no file
  contents are needed and no checkout is required. This is the same "treat the
  fetched ref as truth" move `ariadne#207` makes for publishing.
- **Offline must not block creation.** A failed fetch falls back to the local
  scan with a **loud** warning naming the risk ("origin unreachable; id
  allocated from local files only — may collide"). Creating an issue is not
  something to refuse when the network is down, but a silent fallback here
  recreates the bug it is meant to fix.
- Keep `NextID` pure. It takes the id sets; the fetch and `ls-tree` live in the
  caller's IO shell (`ARCH-PURE`), so the union logic stays unit-testable
  without git.

**Detection, because allocation alone cannot fix what already exists.** Two
colliding files exist in `pair` right now, and more may be sitting on other
branches. Add a duplicate-id check to the merge gate: refuse a merge that would
land an id already present on the trunk, naming both files. Cheap — the same
`ls-tree` — and it is the last point where the collision is still repairable.

**Out of scope:** renumbering existing collisions (operator work, done by hand),
and any change to the repo transaction lock, which is working as designed.

## Done when

- `sdlc issue new` on a branch cut before a published issue allocates an id
  higher than that issue's — a test drives a bare origin, a branch cut before a
  push, and asserts no reuse.
- An unpushed local issue on the current branch is still counted, so two
  successive creations on one branch do not collide with each other.
- With origin unreachable, creation still succeeds and emits the warning; a
  test asserts the warning is present, not just that creation worked.
- `sdlc merge` refuses a branch introducing an id that already exists on the
  trunk, naming both paths.
- Tests run against a real throwaway repo with a local bare origin
  (`ARCH-MOCK`) — a function-call mock cannot express "a ref exists that the
  worktree does not contain", which is the entire bug.

## Plan

- [ ] Extract the id-set union as a pure function; move the scan behind it.
- [ ] `ls-tree` the trunk ref for the three issue directories; fetch first.
- [ ] Offline fallback with the loud warning.
- [ ] Duplicate-id refusal in the merge gate.
- [ ] Bare-origin tests: branch-cut-before-push, two-on-one-branch, offline.

## Log

### 2026-09-02

Found while looking for a published `pair#171` that was invisible from a
feature branch holding a different `#171`.

Related but distinct: `ariadne#207` makes *publishing* work without a main
worktree by building the commit against a fetched `origin/main`. This issue
makes *allocation* read the same source. Same principle — the fetched trunk is
the truth, not this checkout — applied at the other end of the lifecycle. They
share no code and neither blocks the other.

Worth stating plainly, since it was assumed otherwise earlier today: "creating
an issue is safe under concurrency" holds only within one branch. Every
long-lived feature branch is a separate id space, silently.
