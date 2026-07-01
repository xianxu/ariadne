---
id: 000148
status: open
deps: []
github_issue:
created: 2026-06-30
updated: 2026-06-30
estimate_hours:
---

# sdlc merge: guard against a reused branch name silently skipping unmerged commits

## Problem

When `sdlc merge` finds a **merged** (not open) PR for the head branch, it treats
the work as already shipped and "resumes post-merge cleanup" — switches to main,
pulls, archives, and **deletes the branch** — WITHOUT checking whether the branch
has commits *beyond* what that PR merged. So a reused branch name silently drops
new work: the commits stay on `origin` but never reach main, and the branch is
deleted out from under them, with a green exit code.

Concrete (parley.nvim#116): the M2/M3 work reused the branch name of M1, which had
shipped early via its own merged PR #95 (to unblock #128). `sdlc merge` found #95
(MERGED), resumed cleanup, and switched to main + deleted the branch **without
merging the 16 new M2/M3 commits**. `git rev-list --left-right --count
main...origin/<branch>` showed `0 16` — main never advanced — but nothing warned.
Recovery required re-pushing under a fresh name + `sdlc pr` + `sdlc merge`.

This is a "form gate defends against omission" gap: the merge silently did the
wrong thing instead of refusing with a next-action spec.

## Spec

In the `sdlc merge` post-merge-cleanup path (the branch where an existing PR is
**MERGED** rather than open):

- Before resuming cleanup, compute the unmerged-commit count:
  `git rev-list --count <base>..<head>` (base = the PR's base, e.g. `origin/main`;
  head = the branch tip, e.g. `origin/<branch>`).
- If the count is **0** → the branch is genuinely fully merged → proceed with
  cleanup as today (switch to main, pull, archive, delete branch).
- If the count is **> 0** → **abort** with an actionable error, e.g.:
  `branch '<b>' has N commit(s) not in main despite a merged PR (#<n>) — likely a
  reused branch name. Rename the branch (e.g. <issue>-<short-slug>) and run
  `sdlc pr`, then `sdlc merge`.` Do NOT switch branches, delete, or archive.

Keep this scoped to the merged-PR path; the open-PR and no-PR paths are unchanged.

## Done when

- `sdlc merge` refuses (with the actionable message, non-zero exit, no branch
  deletion) when a merged-PR head branch still has commits not in the base.
- A genuinely fully-merged branch (count 0) still cleans up exactly as before.
- Tests cover both: merged-PR-with-unmerged-commits → abort + tree untouched;
  merged-PR-fully-merged → cleanup proceeds. (The count computation is the pure
  seam to unit-test; the git IO is faked/injected.)
- `sdlc --help` (`cmd/sdlc/helptext/root.md`, PUBLISH) notes the "publish once at
  issue close, not per milestone; don't reuse a branch name with a merged PR"
  guidance (the doc half of this fix).

## Plan

- [ ]

## Log

### 2026-06-30

Filed from the parley.nvim#116 landing session — see parley `workshop/lessons.md`
(2026-06-30, lesson #4) for the incident. The doc half (the PUBLISH-block wording)
is drafted there too. Root-cause fix lives in `cmd/sdlc/merge.go`'s merged-PR
cleanup branch.
