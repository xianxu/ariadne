---
id: 000098
status: punt
deps: [ariadne#95]
github_issue:
created: 2026-06-14
updated: 2026-06-14
estimate_hours:
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

## Spec

- `weave unlink <path>` removes the matching `substrate <path>` row from the
  cwd repo's `construct/deps`, comparing the path **verbatim** (the same way
  `link` records it — so `unlink <path>` undoes `link <path>` exactly).
- **Idempotent:** if no matching `substrate <path>` row is present (or
  `construct/deps` is absent), it is a no-op and prints a clear message
  (e.g. `weave: substrate <path> not present in construct/deps`), exit 0.
- **Non-destructive to siblings:** every other row (`data` rows, other
  `substrate` rows) and the file's formatting are preserved untouched. Only
  the one matching row is dropped.
- **Reuse `layer.ParseDeps`** (ARCH-DRY) — the same grammar the walk + Apply +
  `link` read deps with — to identify the row, rather than a bespoke parser.
- Mirror `runLink`'s shape: pure-ish core over an injectable `weavefs.FS` +
  `io.Writer` for testability; read-only on everything but the one deps file.

## Done when

- `weave unlink <path>` removes the matching `substrate <path>` row from
  `construct/deps`.
- Idempotent: a no-op (with a clear message, exit 0) when the row is absent or
  the file does not exist.
- Never corrupts other rows or the file's formatting — only the matching row
  is removed.
- Tested: TDD covering remove-matching-row, idempotent-when-absent,
  preserve-other-rows, and the subcommand-wired assertion (mirroring the
  `TestLink*` set).

## Plan

- [ ]

## Log

### 2026-06-14

- Split out of the #95 `link` rename (`weave depend-on` → `weave link`): once
  `link` became the module-include verb, its inverse `unlink` (module-exclude)
  was the obvious symmetric gap. Filed as its own issue, dep on ariadne#95.
