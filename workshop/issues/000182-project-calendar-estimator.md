---
id: 000182
status: open
deps: [ariadne#180]
github_issue:
created: 2026-07-16
updated: 2026-07-16
estimate_hours:
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

## Spec

*(Seeded from the #180 design session, 2026-07-16 — refine at brainstorm.)*

Mechanize the commit reality-check into a computed feasibility statement:

- **Throughput:** measured focused hours/week, derived from the same
  active-time machinery `sdlc actual` already uses (aggregate the velocity
  ledger over a trailing window) — no new data collection.
- **Contention:** Σ remaining hours across OTHER committed/executing projects —
  #180's board machinery (`computeBoard`) already computes per-project
  remaining hours; this is one roll-up away. Decide how paused projects weigh.
- **Ceiling:** the ~2-concurrent-sessions operator-attention constant (#117)
  bounds how parallelism converts hours to elapsed time; decide whether it is
  a per-operator config or stays a constant.
- **Output:** `sdlc project commit` computes "36h Phase-A at your measured
  ~12 focused h/wk with <other projects> already committed → lands ~Sep 20;
  proposed deadline Sep 1" — operator confirms or overrides (the override is
  the recorded evidence). `planned_finish:` derives from the same computation
  instead of being hand-typed.
- **Calibration:** planned vs actual finish at project close becomes a second
  project-level ledger row, analogous to #180's fog factor — slippage gets
  measured, not remembered.

Relationship to #180: consumes its model (deadline, planned_finish, board
remaining-hours) and its guard registry — the `reality-check` guard name is
already in project.cue's commit transition, so this issue upgrades the guard's
IMPLEMENTATION from evidence-flag to computed check; drop-in, no model change.

## Done when

- `sdlc project commit` (or set-status →committed) prints a computed
  feasibility statement (throughput × contention × ceiling → projected finish
  vs proposed deadline) instead of accepting bare `--reality` prose.
- `planned_finish:` derives from the computation (override recorded as such).
- Throughput comes from measured active-time data, not a typed constant.
- Project close records planned-vs-actual finish as a calibration row.

## Plan

- [ ] brainstorm: throughput window, paused-project weighting, ceiling as
      config vs constant, override UX (blocked on #180 M4 board machinery)

## Log

### 2026-07-16

Filed from the #180 plan-approval session. Operator: "we need a higher level
time estimator" — the effort→calendar gap identified while reviewing #180's
two-phase (hours-only) estimation design. Kept out of #180 scope deliberately:
needs its own design and consumes #180's model the same way #171 does.
