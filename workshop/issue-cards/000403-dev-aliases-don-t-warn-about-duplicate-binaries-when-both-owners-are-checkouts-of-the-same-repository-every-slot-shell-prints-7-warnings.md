---
id: 000403
status: done
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T15:20:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.52
tracker:
    version: 1
    completion:
        token: close-20e9579f196f
        repository: github.com/xianxu/ariadne
        reviewed_head: c4815824a6b8a59fcc673bf7a7e6e5cdd2550109
        evidence_commit: 81885d820f5111e98f1aa539a4869f3982d9b486
        landed_commit: a5105cbd44468be2b4a90bd495d679bd18930b1c
---

# dev-aliases: don't warn about duplicate binaries when both owners are checkouts of the same repository (every slot shell prints 7 warnings)

## Problem

Every shell that starts inside a slot worktree prints seven warnings from `construct/dev-aliases.sh`, e.g.
`dev-aliases: duplicate binary 'sdlc' owned by: /Users/xianxu/workspace/ariadne
/Users/xianxu/workspace/worktree/parley.nvim-slot4/ariadne (last in sorted order wins)`. Reproduced from
parley.nvim-slot3 and -slot4 (2026-10-15); the operator saw it when parley.nvim:4's pane started and asked whether a
`.zshrc` change caused it.

The two owners are the **same repository**: the slot's own ariadne clone (its substrate) and the workspace sibling
`~/workspace/ariadne`. That's the designed slot layout, not a conflict, and last-in-sorted-order picks the slot's clone,
which is the intended owner. The warning is noise that trains people to ignore real duplicates (two *different*
repos owning one binary name).
