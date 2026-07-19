---
id: 000182
status: working
deps: [ariadne#180]
github_issue:
created: 2026-07-16
updated: 2026-07-18
estimate_hours:
started: 2026-07-18T22:35:59-07:00
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

*(Settled at the 2026-07-18 brainstorm — supersedes the 2026-07-16 seed. One
material change from the seed: the reality-check INFORMS, it never BLOCKS —
"estimation is a funny business, often wrong; when behind we work overtime,
shift resources. Always track and inform the operator" (operator). See
`## Revisions`.)*

### Layering: issue facts, project math

Projects don't burn down directly — issues do. The forecast divides two
numbers, and issue-level data supplies both:

- **Numerator — this project's remaining hours:** Σ `estimate_hours` over
  unfinished breakdown rows (`computeBoard.RemainingHours`); before the
  breakdown resolves, fall back to the Phase-A total (#180) with a note in
  the statement. Issue estimates progressively replace the coarse Phase-A
  number as the project matures.
- **Denominator — throughput (hours/week):** measured from the calibration
  ledger (`brain/data/life/42shots/velocity/calibration-ledger.tsv`, one row
  per issue close): Σ `actual` per week over the blessed span.

**Unit identity (load-bearing):** both sides are SHIP WALL-CLOCK hours for
one engineer + AI (#118). Weekly Σ actual can exceed 168h because concurrent
agent sessions overlap wall-clock — the unit is *issue-hours produced*, not
human attention-hours. The division is sound because numerator and
denominator share the unit; we never convert to human hours. (Consequence:
parallelism is already priced into measured throughput, so the attention
ceiling is a warning, not arithmetic.) MVP uses issue estimates raw; the
est/actual ratio bias (~0.4–1.7× recently) is a future refinement once the
calendar ledger shows it matters.

### Blessed throughput baseline (new, in brain)

Trailing windows skew under vacations and life events (measured weekly
volatility: 14 → 56 → 153 → 91 → 139h across the last 6 active weeks). The
operator instead **blesses a representative span**; the machinery measures
the number — operator picks the SPAN, never types the RATE:

- `sdlc project throughput --bless 2026-06-16..2026-07-13` reads the ledger,
  computes Σ actual ÷ span-weeks, and APPENDS to
  `brain/data/life/42shots/velocity/throughput-baseline.tsv`:
  `blessed_date  span_start  span_end  hours_per_week  rows  ceiling`.
  Last row = current baseline; append-only = free history of re-blessings.
  `ceiling` defaults to 2 (#117), settable at bless time. Refuses an empty
  span (no ledger rows).
- Bare `sdlc project throughput` prints the current baseline + a trailing-4-
  week comparison, so staleness/divergence is visible on demand. The gate
  never auto-substitutes the trailing number.
- Same write-to-brain class as the calibration ledger (measurement — the
  #171 residency charter's brain).
- Test override: `WF_THROUGHPUT_BASELINE` env (mirrors `WF_CALIB_LEDGER`).

### Pure core

`internal/project/forecast.go` — `ComputeForecast(baseline, thisRemaining,
others []ProjectLoad, today) Forecast`, no IO:

- `n` = this project + others with status `committed`/`executing` (fleet-wide;
  enumerate via a small `ListFleetProjects` reusing #171's sibling walk; each
  project's remaining via its own `computeBoard` from its own repo vantage).
- share = `hours_per_week ÷ n`; projected finish = today + remaining ÷ share
  weeks. Auditable one-line arithmetic — no hidden knobs.
- **Ceiling = warning threshold, not arithmetic:** `n > ceiling` adds
  "n active projects exceed your ~2-session attention ceiling — forecast
  degrades".
- **Paused projects: weight 0, named risk lines** ("paused: metis-v2, 14h
  remaining — resuming it invalidates this forecast").
- Output struct: projected date, the arithmetic trail (n, share, remaining,
  source of remaining), divergence note, ceiling warning, risk lines.
  `RenderForecast` produces the one-paragraph statement every surface prints
  identically.

### Consumers — inform, never block

1. **`set-status →committed`:** compute → print → derive `planned_finish:` →
   record the rendered statement in the Log as the reality-check evidence.
   The guard passes on HAVING COMPUTED, never on the answer (feasible or
   not). Only when computation is impossible (no blessed baseline) does it
   fall back to requiring the legacy `--reality` prose — a process fallback,
   not a feasibility gate. Explicit `--planned-finish` overrides the derived
   date and is recorded as a manual override.
2. **`project show` / `project status`:** recompute live; render forecast-
   vs-deadline drift ("at baseline pace you land Oct 3; deadline Sep 1").
3. **`project close`:** append planned-vs-actual finish + slip days as a
   calendar calibration row beside the fog-factor row (exact file/columns
   decided at plan time after reading projectclose.go's ledger shape). This
   creates the project-level feedback loop that issue-level est/actual
   already has.

Future roadmap rollup (#15/#185 territory) = "call `ComputeForecast` per
project and sum" — deliberately nothing reserved for it.

### Errors

No baseline → legacy `--reality` fallback + bless hint. Empty bless span →
refuse. Unresolvable issue estimates → Phase-A fallback, noted in the
statement. Ledger unreadable → same fallback as no-baseline.

### Out of scope

Roadmap-level rollup, per-render forecast history, multi-operator
throughput, ratio-corrected numerator.

## Done when

- `set-status →committed` prints a computed feasibility statement
  (throughput ÷ contention → projected finish vs proposed deadline) and
  records it as the reality-check evidence; it INFORMS and never refuses on
  the answer (`--reality` prose survives only as the no-baseline fallback).
- `planned_finish:` derives from the computation (manual override recorded
  as such).
- Throughput comes from measured ledger data over an operator-BLESSED span
  (operator picks the span, never types the rate); baseline stored in brain
  with provenance.
- `project show`/`status` surface the live forecast-vs-deadline drift.
- Project close records planned-vs-actual finish as a calibration row.

## Revisions

### 2026-07-18 — brainstorm settles the four forks; inform-not-gate

Done-when amended from the seed: the computed statement no longer *replaces*
`--reality` as a gate — it never blocks at all (operator direction: track
and inform; slippage is recoverable by means the math can't see). Forks
settled: blessed-span baseline over trailing window (volatility evidence);
contention = committed+executing, paused listed at weight 0; ceiling =
constant 2 in the baseline record, warning-only; `--reality` survives as the
no-baseline process fallback. (Restored 2026-07-19: an editing slip dropped
this section when the Spec was rewritten; caught by the plan review.)

## Plan

- [ ] brainstorm: throughput window, paused-project weighting, ceiling as
      config vs constant, override UX (blocked on #180 M4 board machinery)

## Log

### 2026-07-16

Filed from the #180 plan-approval session. Operator: "we need a higher level
time estimator" — the effort→calendar gap identified while reviewing #180's
two-phase (hours-only) estimation design. Kept out of #180 scope deliberately:
needs its own design and consumes #180's model the same way #171 does.

Same day, operator placed this IN the project-management-primitive MVP
(initially proposed as post-MVP): "that is the key differentiator between a
project and an issue, the timeline aspect, of managing something higher
level, longer running." The taxonomy consequence for design: the calendar
computation isn't an optional refinement of the reality-check — it IS the
project noun's defining capability; without it a project degenerates to an
issue container with a hand-attested date.
