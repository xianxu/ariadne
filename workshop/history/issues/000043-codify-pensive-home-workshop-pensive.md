---
id: 000043
status: done
deps: []
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

## Spec

- Pensives (the per-topic `pensive` datatype) live in **`workshop/pensive/`**, a sibling of `workshop/parley/` — both thinking artifacts under `workshop/`.
- `docs/vision/` is retained for *broader* free-form vision/brainstorm notes (ariadne's own `AGENTS.local.md` uses it), but is no longer the pensive home and stays optional/unscaffolded.

## Plan

- [x] `construct/base.manifest`: add `scaffold workshop/pensive`.
- [x] `AGENTS.md` §1: pensives → `workshop/pensive` (sibling-of-parley bullet); reframe the `docs/vision` bullet as broader vision only.
- [x] `AGENTS.md` Directory Structure: add `workshop/pensive/`.
- [x] Verify scaffold resolves (a fresh `setup.sh` creates `workshop/pensive/`).

## Log

**2026-05-28 — done.** Codified workshop/pensive as the pensive home; manifest scaffolds it; AGENTS.md §1 + Directory Structure updated; `docs/vision` reframed as broader-vision-only. Downstream derivatives pick up the new scaffold on next `make refresh`/`bootstrap`. you-decide's pensive already relocated to `workshop/pensive/` and its stray `workshop/vision`+`workshop/notes` cleaned up.
