---
id: 000403
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
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
