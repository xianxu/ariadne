---
id: '000115'
status: done
started: 2026-06-18T13:22:13-07:00
created: 2026-06-17
updated: 2026-06-18
estimate_hours: 12
actual_hours: 1.79
---

# DAG-merged dynamic skills: per-repo datatype enumeration across the layer graph

## Problem

#111 shipped the datatype dynamic skill **repo-agnostic**: `cmd/datatype`
enumerates ariadne's `construct/datatype/*.md`, the SKILL.md is generated +
committed *once* in the owner, and consumers **symlink ariadne's copy**. But the
effective datatype set is **per-repo** — the union over a repo's layer DAG
(ariadne's shared set + each intermediate layer's + the leaf's local
`datatype/`), local-shadows-shared. So a derivative's eager skill *description*
lists ariadne's nouns, not its own local datatypes — a derivative-defined type
won't trigger `xx-datatype` until the skill body's apply-time enumeration runs.
#111 scoped this out ("per-repo local datatype lists in the description") as a
known limitation; this issue closes it.

It is also the first concrete instance of a broader pattern (see the pensive
referenced below): **DAG-awareness as a capability available to subsystems
*alongside* weave, not a weave monopoly** — the layer graph is a platform
primitive, not weave's private state.
