---
id: '000043'
status: done
created: 2026-05-28
updated: 2026-05-28
actual_hours: 0.3
---

# Codify pensive home as workshop/pensive (was docs/vision, never scaffolded)

## Problem

Pensives are documented (AGENTS.md §1) to live in `docs/vision/`, but:

1. **`docs/vision/` is never scaffolded.** `construct/base.manifest` scaffolds `workshop/{issues,history,plans,parley,staging}` + `atlas`, but not `docs/vision`. So the convention names a home the bootstrap doesn't create — `workshop/parley/` (parley's home) exists after `make bootstrap`, but pensive's home doesn't.
2. **The convention drifted.** AGENTS.md itself calls pensives *"in a similar vein to `workshop/parley` but more focused on a topic"* — i.e. a sibling of parley — yet splits them: parley → `workshop/`, pensive → `docs/vision/`. Operator's call (2026-05-28): pensives started as vision docs in `docs/vision`, but the better home is `workshop/pensive`, alongside parley. Not previously codified.

Surfaced when a `you-decide` pensive had no scaffolded home; stray empty `workshop/vision/` + `workshop/notes/` dirs there were orphans of the same confusion.
