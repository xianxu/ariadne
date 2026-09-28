---
id: 000220
status: open
created: 2026-09-11
updated: 2026-09-11
estimate_hours:
github_issue:
---

# Issue publishes from a feature branch: false success, empty subjects, add/add at merge

## Problem

Field evidence from parley.nvim#227 (branch `000227-…`, PR #172, 2026-09-10).
Every defect below is in the **trunk path that ariadne#207 adds and has not
merged**. parley ran it because the `sdlc` shell function rebuilds from
ariadne's *working tree* on every call, and that tree is on
`000207-sync-without-worktree`. Line numbers are against that branch at
`671f0f6`.

1. **Empty commit subjects.** `sdlc issue new` from a feature branch published
   blank templates to parley `main` as `1ead2e2`, `4905785` and `c955972`, all
   with empty subjects. The trunk path passes `msg` raw
   (`claim.go:158` → `synctrunk.go:107`), while the on-main path defaults it
   through `syncMessage` (`claim.go:313`).
2. **A publish that never happened, reported as a publish.**
   `sdlc claim --issue 234 --no-start` on the branch printed "Issue changes
   published to the trunk." and `synced`, but nothing reached `main`: its
   first-parent line has no commit between `4905785` (18:15:53) and `c955972`
   (18:32:36). There are two causes.
   - `changedIssueFiles` diffs the working tree against **HEAD**
     (`claim.go:336-341`). The filled #234 had already been committed on the
     branch, so it looked unchanged — but on the trunk path the question is
     what differs from the *trunk*.
   - `syncViaTrunkWithRealloc` returns `(nil, nil)` after "No issue changes to
     sync.", and `syncViaTrunk` then prints the success line and `synced`
     unconditionally (`synctrunk.go:59-62`).
3. **add/add at merge.** GitHub refused PR #172 as "not mergeable" on
   `000233`, `000234` and `000235`. The trunk held the blank stubs from (1); the
   branch held the filled files, committed independently. With no common
   ancestor for those paths, git cannot merge them. It was resolved by hand
   (`63390fe`).
4. **`sdlc pr` is not idempotent.** On a branch that already has an open PR it
   pushes, then `gh pr create` fails with "already exists", and it exits nonzero
   (`pr.go:107-115` never checks for an existing PR).

Two further observations from the same session:

- The `sdlc` function builds from ariadne's working tree, so an in-flight
  ariadne branch silently changes `sdlc` for every repo on the machine. That is
  how (1)–(3) reached parley before #207 merged.
- ariadne's local `main` is 8 commits ahead of `origin/main`: #207's
  `change-code` spec/plan syncs `8970d88`…`ff1b52b` and side-quest `59debf4`,
  all from 2026-09-09. Any on-main `sdlc` publish would push them wholesale —
  the exact hazard `PublishExisting`'s comment warns about. This issue was filed
  by hand from a detached `origin/main` worktree to avoid both that and (1).

### Root cause: two writers, unrelated histories

A trunk publish is a commit built on `origin/main`'s tip, and the publishing
branch never contains it. The branch keeps committing the same paths, because
the files are in its working tree. Git merges cleanly only where the two sides
share an ancestor for a path, or happen to agree byte-for-byte.

(3) is the add/add form of this. The branch's *own* issue file is the
modify/modify form. It merged cleanly on #227 only because the trunk stopped
writing it after the branch point (`555b81a`); any later trunk publish of that
file would have conflicted the same way. (2) hides the divergence instead of
closing it.
