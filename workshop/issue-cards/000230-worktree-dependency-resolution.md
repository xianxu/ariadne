---
id: 000230
status: working
created: 2026-09-15
updated: 2026-10-12
estimate_hours:
github_issue:
started: 2026-10-12T18:45:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Make Weave dependency resolution worktree-aware

## Problem

During publication of parley.nvim#254, `weave compile` in a linked worktree
could not resolve `substrate ../ariadne` from `construct/deps`. The dependency
exists next to the primary checkout, but not next to the linked worktree:

```text
~/workspace/parley.nvim/../ariadne                         -> exists
~/workspace/worktree/parley.nvim/000254-.../../ariadne     -> absent
```

The first compile silently skipped the dependency, applied only four actions,
and pruned generated vocabulary. SDLC then refused the merge because the issue
schema was unavailable. Creating a sibling symlink to the real Ariadne checkout
allowed the next compile to apply99 actions and validation to pass. Worktrees
should not require this manual filesystem setup.
