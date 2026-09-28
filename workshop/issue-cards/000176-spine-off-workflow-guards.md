---
id: '000176'
status: done
started: 2026-07-14T17:44:20-07:00
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.65
actual_hours: 1.53
---

# spine guards for off-workflow invocations: change-code on done issues + non-SDLC repos

## Problem

#172's friction audit found the spine gets invoked where the workflow state
says it shouldn't be, with no guard until close-time:

1. **Working on a `done` issue is un-gated.** `change-code` (and every other
   verb) runs silently against a done issue; only re-close is guarded.
   Firing-order measured 11 change-code-after-close inversions (8 brain,
   2 pair, 1 ariadne-dogfood).
2. **brain runs the spine against its own charter.** brain is a Drive-like
   capture repo, not an SDLC repo, yet it concentrates 19 bypasses and 8/11
   firing-order anomalies — every gate becomes noise to route around there,
   and it pollutes cross-repo friction measurement.
