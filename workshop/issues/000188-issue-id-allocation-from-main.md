---
id: 000188
status: open
deps: []
github_issue:
created: 2026-07-28
updated: 2026-07-28
estimate_hours:
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

## Spec

- **Allocate from the remote.** `git fetch origin main --quiet`, then derive the
  max from git, not the filesystem:
  `git ls-tree -r --name-only origin/main -- workshop/issues workshop/history`.
  Branch-independent; works from a stale branch, a detached HEAD, or a checkout
  with no main worktree.
- **Allocate + commit + push is ONE retryable unit.** On a non-fast-forward
  rejection: re-fetch, **re-derive the ID**, rename the file, recommit, push
  again. Bounded (~3 attempts), then refuse with the reason.
- **The subtle requirement, and the whole point of the issue:** never
  rebase-and-retry holding the ID already chosen. Push rejection is *ref*-based,
  so it fires correctly — but the naive recovery `git pull --rebase && git push`
  then succeeds, because two files with different names produce **no textual
  conflict**. The rebase silently lands the duplicate. Correctness comes from
  re-allocating inside the loop; the rejection is only the trigger that says to.
  The current code has the ingredients arranged exactly wrong for this: `NextID`
  is at `issue.go:259` while the `git pull --rebase origin main` is at
  `claim.go:243` — a different function, a different worktree, far downstream. A
  retry bolted onto today's shape would re-pull and push the duplicate.
- **Post-publish assertion: exactly-once, not max+1.** Re-list `origin/main` and
  assert this ID prefix appears exactly once. `max+1` arithmetic cannot detect
  the failure — if two agents both allocate 188 and both land, main's max is 188
  either way and "next expected = 189" passes for both. It infers the invariant
  from a maximum instead of observing it.
- **On a duplicate found: refuse loudly, do not auto-renumber.** Print both paths
  and the next free ID. Silently changing an ID under an agent that has already
  branched and committed against it is worse than the duplicate it fixes (see the
  four reference sites above).
- **No timestamp tiebreak.** Every candidate is unreliable: `created:` is
  day-granular and agent-writable, mtime does not survive git, commit time
  carries clock skew and is rewritten by rebase. None is needed: **whoever is
  already on `origin/main` is the incumbent, whoever is pushing loses.** Git's
  ref ordering settles it with no clocks involved.
- **Offline fallback.** With no reachable origin, fall back to the local scan,
  warn loudly that the ID is provisional, and let the retry-on-publish correct it
  when the push eventually happens. Refusing outright would make the tracker
  unusable on a plane.

### Deliberately rejected

- **A `.next-id` counter file** so two creates conflict textually and git
  enforces it hard. Rejected: it adds state already derivable from the filenames
  (which *are* the registry — `ARCH-DRY`), so it can drift; and it converts a
  silent-but-fixable race into a manual merge-conflict resolution every time two
  people file an issue the same day.
- **Delete-and-recreate the losing file.** Rejected: by detection time it may
  hold an authored Spec and Plan, and a recreate touches none of the four
  reference sites. The correct operation is `git mv` + rewrite `id:` + fix
  references — and note `git mv` does not stage subsequent edits, so it needs an
  explicit `git add` or the rename commits without the content change.

## Done when

- `sdlc issue new` fetches and derives the next ID from `origin/main`, verified by
  a test where the working tree's max and `origin/main`'s max differ and the
  remote wins.
- A simulated concurrent create (peer lands a different-slug file at the same ID
  between our fetch and our push) results in **our** file being re-allocated to
  the next ID, not in two files sharing one — the regression test for the
  rebase-papers-over-it hole.
- Running from a checkout with **no main worktree** allocates correctly.
- Post-publish exactly-once assertion runs; an injected duplicate makes the
  command refuse and name both paths.
- Offline creation still works, warns, and is corrected on the next successful
  publish.

## Plan

- [ ]

## Log

### 2026-07-28

- Filed from the pair#127 session postmortem, alongside **#187** (change-code gate
  tuning). #187 is about review *cost*; this is about tracker *correctness* —
  separate concerns, separate issues.
- Live evidence: `sdlc issue new` in both `pair` and `ariadne` warned "issue
  created but auto-sync to main did not complete: could not find a worktree on
  branch 'main'" and left the file branch-local. In ariadne it landed on the
  `000185` branch and had to be moved to main by hand.
- Correction to an earlier claim made in that session: `000186` *was* visible from
  the `000185` branch (it lives in `workshop/history/issues/`, which `NextID`
  scans), so that particular allocation was correct. The hazard is not history
  drift — it is a peer landing a new issue on main that a stale branch has never
  seen.
- **Operational follow-up, independent of any code change:** ariadne has no
  permanent `main` worktree, which is what disables the sync path today. Adding
  one restores it immediately.
- **Follow-up worth its own issue:** an explicit `sdlc issue renumber <id>` that
  moves an issue to a new ID *and* rewrites the branch name, commit-message
  references, `deps:`/project refs, and sidecar filenames. That is the honest home
  for repair; `issue new` should only ever prevent.
