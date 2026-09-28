---
id: '000248'
status: done
started: 2026-09-23T23:34:13-07:00
created: 2026-09-23
updated: 2026-09-24
actual_hours: 0.83
---

# Agent procedure: move this branch to :N

## Problem

Operators say "move this branch to :0" (or `:N`): take the issue branch
checked out in the current slot and check it out in slot `:N` instead, usually
`:0`, because that is the checkout whose built binary and runtime files the
operator's live sessions run from, so a smoke test needs the branch there.
Git allows a branch in only one worktree at a time, so the source slot must
first go back to its resting branch.

No agent procedure names this. `atlas/workflow/workspace-branching.md` covers
"In :2, create an issue branch from :1", refreshing a resting slot, and
independent work, but not moving an existing branch. On 2026-09-23 (pair#316)
an agent did not recognise the phrase until the operator spelled out the
steps, then improvised them: it guessed checkout paths instead of using
`sdlc workspace :N`, and it switched a `:0` holding untracked operator files
without a stated rule.
