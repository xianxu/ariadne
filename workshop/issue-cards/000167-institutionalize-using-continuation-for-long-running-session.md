---
id: '000167'
status: done
started: 2026-07-12T22:02:15-07:00
created: 2026-07-07
updated: 2026-07-12
estimate_hours: 1.08
actual_hours: 2.71
---

# institutionalize using continuation for long running session

## Problem

Ariadne has an agent-neutral `continuation` datatype for preserving the human-
meaningful state of a long-running session, and Pair's continuation writer can
durably write it and restart the session from it. Agents currently initiate that
flow only after an explicit operator request. They are not told to use it
proactively as context fills, so long sessions can reach compaction with less
room to create a good handoff. In metis-v2, an agent learned this behavior after
the operator demonstrated it once; the behavior should be institutionalized for
every Ariadne consumer instead of relearned per project.
