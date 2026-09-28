---
id: '000157'
status: done
started: 2026-07-01T16:06:44-07:00
created: 2026-07-01
updated: 2026-07-01
estimate_hours: 2.6
actual_hours: 1.02
---

# process-manual dynamic session reconstruction

## Problem

The M1 process manual (`sdlc process-manual`) is the *static* catalog — what CAN be
injected. It can't say which injection points actually **fired** in a given session, in
what order. The parley's second output: mine a session transcript → the ordered stream
of fired injections, matched to the M1 catalog, so a human can see the process as it
actually ran (`workshop/parley/2026-07-01.10-47-05.962_agentic-process-documentation-strategy.md`).
