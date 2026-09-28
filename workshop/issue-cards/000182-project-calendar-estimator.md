---
id: '000182'
status: done
started: 2026-07-18T22:35:59-07:00
created: 2026-07-16
updated: 2026-07-19
estimate_hours: 2.65
actual_hours: 3.21
---

# project calendar estimator: mechanize the commit reality-check (effort→calendar bridge)

## Problem

A project's load-bearing attribute is its `deadline:` — a DATE — but both
estimators produce HOURS: per-issue estimate-logic-v3.1 (Phase B) and #180's
Phase-A PRD-level estimator (workstreams × fog). Nothing bridges effort to
calendar. The bridge currently lives in exactly one place: the `reality-check`
guard at project commit (defined→committed), which as designed in #180 is an
evidence flag — the operator types `--reality "fits July"` and the gate takes
their word. A mandate, not a mechanism — off-brand for the vocabulary lift,
whose whole point is computed gates over attestation.
