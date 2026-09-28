---
id: '000144'
status: done
started: 2026-07-05T10:18:05-07:00
created: 2026-06-29
updated: 2026-07-05
estimate_hours: 1.8
actual_hours: 2.0
---

# sdlc resolve — read-only artifact-reference resolver

## Problem

ariadne artifacts (issues, plans, review sidecars, targets) refer to each other —
and across peer repos — with **symbolic** refs (`ariadne#11`, `#15 M4`, `pair#84`),
but there's no mechanism to turn a ref into the file it names. Consumers (a human in
parley, an agent, the CLI) each have to glob/guess the path.

This issue **was** framed as "files should carry stored cross-links" — rejected:
the id is stable but the path is not (slug renames; `issues/ → history/` on
close/merge, which ariadne#160 made happen on every merge), so stored links rot on
archive. The fix is **read-time resolution** — keep the symbolic ref canonical and
resolve the path on demand. This issue is the **ariadne slice** of that: a
read-only `sdlc resolve`. The editor UX is **parley#160** (`navigate ariadne
artifact references`), which shells to this.
