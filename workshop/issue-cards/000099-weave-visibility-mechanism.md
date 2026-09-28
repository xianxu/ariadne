---
id: '000099'
status: done
created: 2026-06-14
updated: 2026-06-16
estimate_hours: 5
actual_hours: 0.56
---

# weave: export/internal visibility mechanism

## Problem

weave composes `prose` (and will compose `skill`/`settings`) across the layer DAG, but has **no visibility axis**: every `base.manifest` row is implicitly inherited, and `loadLayer` reads each `prose` fragment from the *declaring layer's* directory. So `prose AGENTS.local.md` (declared in ariadne's manifest) always pulls **ariadne's** local into every consumer, and a manifest-less leaf contributes nothing of its own. Empirically (ariadne#95 M5 parley tart pass): parley's composed `AGENTS.md` = ariadne-base + **ariadne's** local, with parley's own local missing — the `@AGENTS.local.md` bug reproduced one level down. `weave golden` missed it because it diffs the composed file against the live *symlink* (an intended divergence) and never inspects composed *content*.
