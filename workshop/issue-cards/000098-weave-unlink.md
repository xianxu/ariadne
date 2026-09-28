---
id: '000098'
status: punt
created: 2026-06-14
updated: 2026-06-14
---

# weave: unlink — remove a layer dependency

## Problem

`weave link <path>` records a `substrate <path>` row in `construct/deps` — the
**module-include** verb of weave-as-a-repo-composition dialect: the way a repo
declares which AI base layers it composes. There is no inverse. To stop
composing a layer today you hand-edit `construct/deps`, which is exactly the
kind of by-hand structural edit weave exists to remove (and risks corrupting
adjacent `data`/`substrate` rows or the file's formatting).

We need `weave unlink <path>` — the **module-exclude** verb: the way a repo
*retracts* a base-layer dependency it previously declared, the symmetric
counterpart to `link`.
