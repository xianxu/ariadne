---
id: '000224'
status: done
started: 2026-09-12T18:26:36-07:00
created: 2026-09-12
updated: 2026-09-12
estimate_hours: 0.89
actual_hours: 0.30
---

# ARCH-FUNERAL: everything created names its end — a funeral plan is part of the design

## Problem

The registry has a principle for how a component behaves under load
(`ARCH-CONSTRAINTS`) and one for state carried between events
(`ARCH-ORDER`), but none for **what a program creates and never destroys**.
Agents are systematically weak here: the payoff of a retention rule is
months out, the cost of omitting one is invisible at review, and "append a
record" always passes.

Field evidence, pair, 2026-09-12:

- Pair's data store is 13 GB. One event log is 4.8 GB. 143 per-thread
  ledgers hold 73 MB, 99% of it launch snapshots that stop mattering the
  moment the launch is superseded. Nothing in the tree prunes any family
  (pair#239).
- The same day, a thread became unresumable because its append-only ledger
  crossed an 8 MiB read cap on its 14th relaunch — a cliff the design
  reached by construction, since the file was defined to grow and no one
  had said how far (pair#237, pair#238).
- The agent's own store next door is bounded: Claude Code sweeps its
  transcripts after 30 days. Pair records more about the same sessions and
  keeps it forever.

The gap is not "we forgot to add GC". It is that the artifact's **end** was
never part of its design, so nothing at plan time asked for it and nothing
at review time could flag its absence.
