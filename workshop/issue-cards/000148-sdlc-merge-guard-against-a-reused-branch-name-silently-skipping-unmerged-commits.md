---
id: '000148'
status: done
started: 2026-07-05T20:31:25-07:00
created: 2026-06-30
updated: 2026-07-05
estimate_hours: 0.48
actual_hours: 0.51
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
