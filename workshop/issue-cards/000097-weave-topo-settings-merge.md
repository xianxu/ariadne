---
id: '000097'
status: done
started: 2026-07-07T22:21:33-07:00
created: 2026-06-14
updated: 2026-07-07
estimate_hours: 3
actual_hours: 0.69
---

# weave: topological multi-layer settings merge

## Problem

weave composes `prose` topologically across the whole layer DAG (a leaf gets
ariadne's `AGENTS.base.md` + every ancestor's `AGENTS.local.md`, foundation-first),
but `settings` does **not**. The `merge` verb lowers each `merge <src> <dst>` row to
an independent `MergeSettings{Source, Target}`, and `settingsx.Merge(base, local)` is
a **two-input** fold: `base` = the row's source (`settings.ariadne.json`), `local` =
the repo's `settings.local.json`. So "higher overrides lower" only ever means
"repo-local overrides ariadne-base" — a **middle** layer cannot contribute settings.

Concretely: brain (ariadne→nous→brain) merges ariadne-base + brain-local; nous is
skipped. The day metis (ML layer) or nous wants its own settings fragment — ML
permissions, layer-specific hooks — the current model can't express it. This is an
inconsistency with `prose`, not a `setup.sh` regression (`merge-settings.sh` was also
two-input), so it's an enhancement deferred out of the #95 cutover.

No consumer exists today: nous's `construct/base.manifest` has 0 `merge` rows and no
`settings.<layer>.json` file exists anywhere. The natural first consumer is metis.
