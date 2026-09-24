---
id: 000249
status: open
deps: []
github_issue:
created: 2026-09-24
updated: 2026-09-24
estimate_hours:
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

## Spec

A resting branch should never stay diverged from `origin/main` because of an
`sdlc` verb.

- **After any publish** (`change-code`, `issue publish`, `claim`, `close`,
  `issue new`): when the current or source checkout is on a resting branch and
  every local-only commit's content is now on `origin/main` (patch-id or tree
  equivalence, not SHA), fast-forward/reset the resting branch to `origin/main`.
  Never discard a commit whose content is not on origin.
- **`issue sync` on a resting branch:** keep the local, no-network commit (the
  crash-safety guarantee from #206 stands), but reconcile before committing when
  origin is reachable and cheap to check — rebase local-only issue commits onto
  `origin/main` — or print the one next action (`sdlc issue publish --commit
  SHA`) instead of silently leaving the branch ahead.
- **Detection surface:** `sdlc state` reports `resting branch diverged
  (ahead N, behind M; K ahead commits already on origin)` so the condition is
  visible before the operator hits it in git.

## Done when

- Publishing from a resting branch leaves it equal to `origin/main` when all its
  local commits were the ones published (fixture: three sync commits then
  `change-code` → resting branch not ahead).
- A peer publish between a local `issue sync` and the next `sdlc` verb does not
  leave the resting branch diverged after that verb (fixture with two clones).
- A local commit whose content is NOT on origin is never dropped — the verb
  refuses or rebases, and says which.
- `sdlc state` names the diverged/duplicate condition.
- #240's feature-branch case is either covered by the same mechanism or
  explicitly left to #240 with the boundary written down.

## Plan

- [ ]

## Log

### 2026-09-24

Filed from pair slot 2 after the operator asked what keeps diverging
`main-slot2` from `origin/main`. Evidence: pair reflog for `main-slot2`
(`15f5220f`/`6aa7d819`/`24199071` #311 syncs left behind after publish;
`76e677e1` #320 sync vs origin `d8b521fb`/`e52a7e2b` #319).
