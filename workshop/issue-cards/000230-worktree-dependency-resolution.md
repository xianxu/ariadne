---
id: 000230
status: codecomplete
created: 2026-09-15
updated: 2026-10-13
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
actual_hours: 1.52
tracker:
    version: 1
    completion:
        token: close-e3010c8cdb3b
        repository: github.com/xianxu/ariadne
        reviewed_head: 69aa17f639a9ef681c2ddf8396cdbc3d9eadb6ea
        evidence_commit: ae36ec72e7d768da7e03801d484e31fc23cf9ae0
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
