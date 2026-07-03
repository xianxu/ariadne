---
id: 000144
status: wontfix
deps: []
created: 2026-06-29
updated: 2026-07-03
---

# issue, plan, and review files should interlink

They should refer to each other, to make navigation among those files fast.

## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-06-29

### 2026-07-03 — wontfix (superseded by parley#160)

Design discussion with the operator concluded that this issue's premise — files
carrying **stored cross-links** — is the wrong approach: the issue *number* is
stable but the *path* is not (slug renames; `issues/ → history/` on close/merge,
which ariadne#160 made happen on every merge), so stored links rot on archive and
would need rewriting across every repo.

Chosen instead: **read-time resolution** in parley — the symbolic ref (`ariadne#11`)
stays canonical, and the path is resolved on demand via a read-only `sdlc resolve`
(deriving artifact locations from the vocab model; lock-free, so not subject to the
mutating-verb slowness). The feature is owned by **parley#160**
(`navigate ariadne artifact references`), with a small `sdlc resolve` ariadne
dependency tracked there. Closing this out as wontfix (premise rejected; work
relocated).

