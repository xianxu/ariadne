---
type: pensive
date: 2026-04-30
topic: parley-datatype-duality
mode: eureka
description: workshop/ is parley-authoritative + agent-readable; data/ is agent-authoritative + parley-readable. The frontmatter `type:` field is the bridge, and each side needs an escape hatch into the other.
references: [/Users/xianxu/workspace/ariadne/construct/datatype/pensive.md, /Users/xianxu/workspace/brain/data/project/charon-launch-push.md]
---

# Pensive: parley/datatype duality

There's a duality in how structured markdown gets produced and maintained in this system. `workshop/` (issues, plans, history) is parley-authoritative — deterministic code in parley.nvim manipulates the structure, agents read the artifacts. `data/` (projects, products, pensives, the rest of the datatype zoo) is the inverse — agents are authoritative on the structure, parley reads the artifacts. Both sides emit the same shape: markdown with `type:` frontmatter in well-known locations.

The `type:` field is what makes the bridge cheap. Parley doesn't need to know about every datatype; it needs one type-aware picker (a thin wrap of `rg "^type: <X>"`) and the escape hatch I already have at `<C-g>m` for the long tail. I filed parley.nvim#115 for the structural-manipulation side — beyond navigation, things like "flip a project task to `[x]` and append `actual:`" are micro-edits where invoking an agent is overkill but doing it by hand is annoying. That's the next round of investment on the parley side.

The deeper symmetry is that *both* directions need an escape hatch. Agent reading parley state is already implicit (agents grep markdown). Parley reading agent state is what `<C-g>m` covers today and what #115 will sharpen. Neither side owns the file format — the convention does. That's the load-bearing piece: as long as we hold the line on `type:` + well-known location + greppable frontmatter, the two authorities can evolve in parallel without negotiating.

## Open questions

- Where exactly does the boundary fall for editing? An agent updating a task checkbox is fine; a human-in-parley updating a task checkbox is fine; both at once on the same file is the kind of race that conventions don't solve. Do we need any sync discipline beyond "don't have both windows open"?
- Should parley grow a generic "datatype prototype reader" so its commands stay in lockstep with new types added under `construct/datatype/`, or is per-type bespoke commands (`<C-g>p` for projects, `<C-g>v` for pensives) actually fine since the set turns over slowly?
- Is `<C-g>m` the right escape hatch in the long run, or does it want to be type-aware by default (sort by mtime within a chosen type)?
