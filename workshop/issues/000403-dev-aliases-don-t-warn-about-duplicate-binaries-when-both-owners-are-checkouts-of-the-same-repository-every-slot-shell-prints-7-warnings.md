---
id: 000403
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '0b4295fe45d1ee657a065881bcd69f0cdd725d63' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

- Treat checkouts of the same repository (same origin remote URL, normalized) as one owner: no warning; pick the
  current repo's own substrate clone deterministically (as now), and say so only under a verbose flag.
- Keep the warning, and `--strict`'s non-zero exit, for genuinely different repositories owning one binary name.

## Done when

- A fixture test: a slot layout (workspace sibling plus the slot's substrate clone of the same origin) prints nothing
  and maps binaries to the slot's clone; two different repos owning one name still warn, and fail under `--strict`.

## Plan

- [ ]

## Log

### 2026-10-15
