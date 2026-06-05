---
id: 000087
status: working
deps: []
github_issue:
created: 2026-06-05
updated: 2026-06-05
estimate_hours: 1
---

# sdlc close: sanity-check a passed --actual against the computed active-time-v3 value (warn/refuse on large deviation)

## Problem

`sdlc close` (and `milestone-close`) compute the measured actual (active-time-v3) **only when
`--actual` is omitted** (`close.go:321` → `explainActual` → exit 1 with a suggestion). When
`--actual` **is** passed, it is parsed as a float and **trusted blindly** — no comparison to the
computed value (`close.go:310-313`).

That hole let nous#42 record `--actual 13.5` when the measured value was `0.30h` — a 45×
fabrication (a sum of per-milestone *estimates*) that sailed through and polluted velocity
calibration. The doc fix (#86) removes the *priming* that produced the bad number; this issue
adds the **backstop** so a fabricated/fat-fingered value can't pass silently even if a human or
agent types one. The "earned, not guessed" gate doing its job on the override path, not just the
omit-path.

## Spec

When `--actual <v>` is supplied to `close`/`milestone-close`, also run `computeActual` (the
engine already shared by `sdlc actual` and the missing-`--actual` explainer) and compare:

- If the engine measured a value `m` and `v` deviates beyond a threshold (proposal: ratio
  `max(v,m)/min(v,m) ≥ 3`, with a small absolute floor so tiny values don't trip it), **warn**
  loudly: e.g. `passed --actual 13.50 but active-time-v3 measured 0.30 (45× off) — confirm or
  re-run sdlc actual`.
- Decision to settle in design: **warn-only** (record the passed value, surface the discrepancy)
  vs **refuse without `--force`/a labeled override**. Lean toward warn-by-default + refuse-on-
  extreme (e.g. ≥10×), so the common "my measured value is a bit off" case isn't blocked but a
  45× fabrication is.
- Skip the check gracefully when the engine can't measure (empty window / 0-events / fallback) —
  never block a close because measurement was unavailable; that's the existing judgment path.

Design considerations: threshold + floor values; warn-vs-refuse policy; escape hatch (`--force`
already exists; consider a labeled `--actual-override <why>` for readability). Keep it in the
shared close path so `milestone-close` inherits it (it wraps `close`).

## Done when

- A passed `--actual` that deviates wildly from the active-time-v3 measurement is caught
  (warned, and refused beyond an extreme threshold without an explicit override).
- Near-measurement values close silently; engine-unavailable closes are unaffected.
- `milestone-close` inherits the check via the shared path. Tests cover the three cases.

## Plan

- [x] (design) settle threshold/floor + warn-vs-refuse policy + override affordance.
- [x] implement the compare in the shared close path (`close.go`), reusing `computeActual`.
- [x] tests: passed≈measured silent; passed≫measured warn/refuse; engine-unavailable skip.
- [x] update `helptext/close.md` + `milestone-close.md` to describe the check.

## Log

### 2026-06-05

Filed alongside #86 from the nous#42 retro. #86 removes the doc priming (cause); this is the
code backstop (defense in depth). Operator: "clever, and a clean way to manage the migration" —
even while downstream docs/habits catch up, the check stops a fabricated value at the gate.
Not scheduled for immediate implementation — filed for later.
