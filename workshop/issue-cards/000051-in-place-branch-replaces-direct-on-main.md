---
id: '000051'
status: done
created: 2026-05-30
updated: 2026-06-03
estimate_hours: 3
actual_hours: 4
---

# In-place branch workflow replaces direct-on-main

## Problem

There are three publish modes today, and they overlap awkwardly:

1. **Direct-on-main** — work + commit on `main`, `sdlc push` (`make push`) ships it.
2. **Worktree** — `sdlc change-code --worktree=yes` (separate checkout dir) → `sdlc pr` → `sdlc merge`.
3. **In-place branch** — `sdlc change-code --worktree=no` already creates a branch that carries the working tree forward in the *same* directory. But it's a half-citizen: the **merge-back path is worktree-shaped** (`sdlc merge` / `make merge` locate and clean a worktree — `git worktree list | grep '[main]'`), so there's no clean "merge this in-place branch and switch back to main" verb.

We want to **retire direct-on-main and make in-place branch the default**, with worktree as the opt-in heavyweight (true isolation, parallel work). Rationale from the 2026-05-30 design discussion (you-decide review-gate thread):

- **Commit is a sync/handoff primitive, not a readiness claim.** Committing preserves state and hands off between agents; it doesn't mean "ready." So commits (and branch pushes) should be free, and the readiness boundary is *publishing to main*.
- A **branch is the transient staging lane** that the publish step evaluates — `main` is the published truth. Direct-on-main collapses the lane and the truth into one ref, which is why "commit ≠ ready" had nowhere to live.
- An **in-place branch in the same dir** is ergonomically ~identical to working on main (one `switch -c` up front, one merge at the end — both agent-automatable), so this is a *simplification*, not added ceremony. Worktrees are only worth their setup cost for genuine parallelism.

**Branch protection is explicitly out of scope here.** The operator does not run agents fully autonomously and will notice whether work is on a branch, so the value is *workflow consistency*, not enforcement. (Server-side enforcement — branch protection + a required publish-gate check — is a separate, optional layer tracked in you-decide#4 / its M3.)
