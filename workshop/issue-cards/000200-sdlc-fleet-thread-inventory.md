---
id: '000200'
status: done
started: 2026-08-24T13:24:43-07:00
created: 2026-08-21
updated: 2026-08-27
estimate_hours: 4.22
actual_hours: 49.90
---

# sdlc: fleet thread inventory

## Problem

`sdlc state` answers "where am I" for **one** repo. Nothing answers "what work is
open across the fleet" — which is the operator-facing failure that `pair#145`
(couch) exists to fix: threads are forgotten, not mis-ranked, and the forgetting
happens across repos over multiple days.

A cold-revival experiment on 2026-08-20 produced the constraint that makes this
non-trivial. `kbench#24` read `status: working, estimate_hours: 4.98` for a
month while 256 commits of the real work happened elsewhere; git said
`0 ahead, last touched 2026-07-23` and was correct. So an inventory built on
issue frontmatter would hand its caller a **confident lie**, in exactly the case
that matters most.

Design context:
`brain/workshop/pensive/2026-08-20-01-pensive-couch-agent-switcher.md`.
