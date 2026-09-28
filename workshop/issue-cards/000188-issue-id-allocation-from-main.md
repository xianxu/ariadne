---
id: '000188'
status: done
created: 2026-07-28
updated: 2026-09-11
actual_hours: 1.46
---

# allocate issue IDs from origin/main with retry-on-reject

## Problem

`sdlc issue new` allocates the next ID from the **current checkout's working
tree**, never from main, and never over the network:

- `issue.NextID` (`cmd/sdlc/internal/issue/scaffold.go:31`) is `os.ReadDir` over
  `workshop/issues/`, `workshop/history/`, and the archive subdir.
- There is **no `git fetch`** anywhere in the create path.
- The publish to main (`syncIssuesToMain`, `issue.go:307`) happens *after* the ID
  is chosen and the file is named — and it is best-effort, downgraded to a
  warning on failure (`issue.go:313`).
- That publish is also **conditional on a main worktree existing**:
  `findMainWorktree` (`claim.go:382`) parses `git worktree list --porcelain` and
  errors with *"could not find a worktree on branch 'main'"* when there is none.
  With no main worktree, `sdlc issue new` silently degrades to "write a file on
  whatever branch you are standing on."

So the tracker's ID space is only as correct as the branch you happen to be on.
Observed live in two repos this session (pair and ariadne), where the sync was
skipped for exactly this reason and the new issue file stayed branch-local.

**The existing conflict check cannot catch the resulting collision.**
`syncOnBranch` (`claim.go:247-266`) diffs `merge-base..main` for changed **paths**
and intersects with the paths you changed. `000188-my-slug.md` and
`000188-their-slug.md` are different paths, so the intersection is empty, it
prints *"No conflicts detected"*, copies both onto main, and you have two issues
sharing an ID with no warning anywhere.

**Why renumbering after the fact is expensive.** The ID leaks into at least four
places, all confirmed in this codebase:

- the **branch name** — `change-code --issue N` derives it from
  `issues/NNNNNN-*.md`
- **commit subjects** (`#127: …`), which agents grep via `git log --grep "^#127"`
- **`deps:`** in sibling issues, and `repo#id` refs in project files
- **review sidecar filenames** — `sidecarPath` (`reviewsidecar.go`) reuses the
  issue filename stem

That asymmetry is the whole design argument: a collision is cheap to *prevent* at
creation (the file is still a fresh template, nothing references it) and
expensive to *repair* later.
