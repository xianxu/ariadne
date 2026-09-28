---
id: 000249
status: wontfix
created: 2026-09-24
updated: 2026-09-28
estimate_hours:
github_issue:
---

# Issue sync and publish leave a slot's resting branch diverged from origin/main

## Problem

In a multi-slot checkout (pair's `main-slot1`/`main-slot2`, each a *resting
branch* tracking `origin/main`), the operator had to reconcile a diverged resting
branch three times in one day. Measured in pair slot 2 on 2026-09-23/24:

1. **Local-only sync commits race peer publishes.** `sdlc issue sync` commits to
   the current branch and never pushes (by design, #206). On a resting branch
   that means `main-slot2` is ahead of `origin/main`, while every other slot's
   `claim`/`change-code`/`close` publishes straight to `origin/main`. Any peer
   publish then diverges the branch. Example: `76e677e1 #320: issue-sync` at
   14:13:14; slot 1's `#319: issue-sync: claim` reached origin at 14:13:32 and
   `change-code` at 14:14:11 → `ahead 1, behind 2`.
2. **Publishing copies the commit and leaves the original.** `change-code`
   reported `Source commit 15f5220f: published (6bc2340e)` — the issue commit
   landed on origin under a new SHA and the branch was cut from there, but the
   three original `#311: issue-sync` commits stayed on `main-slot2` as
   "ahead" duplicates of content already on origin. The resting branch diverged
   with no peer activity at all.

Each case needs a manual rebase (or reset once the content is confirmed on
origin), which the operator reasonably reads as "the workflow keeps breaking my
branch". Related: #240 (sync strands non-working issue updates on a feature
branch — same verb, feature-branch side), #248 (moving a branch between slots,
compares resting history with upstream).
