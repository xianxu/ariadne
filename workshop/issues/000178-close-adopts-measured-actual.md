---
id: 000178
status: codecomplete
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.65
started: 2026-07-14T16:42:01-07:00
actual_hours: 0.49
---

# close: adopt the measured actual when --actual is omitted (kill the compute-then-ask loop)

## Problem

`sdlc close` without `--actual` REFUSES, then computes the measured hours and
prints "→ close with: --actual N" — asking the agent to call it again with a
number sdlc itself just derived. That compute-then-ask loop produced ~48
no-actual refusals in the corpus (#172: the second-highest refusal volume in
the spine), and agents copy the suggested value verbatim ~45/48 times. The
gate's purpose is preventing GUESSED hours (they pollute velocity calibration);
a value sdlc measured cannot be a guess, so the explicit copy-back round-trip
adds no information.

## Spec

When `--actual` is omitted: close runs the measurement (as it already does in
the omit-path, `explainActual`) and ADOPTS the result — loud info line
("using measured actual: N h (window X → HEAD, attributed across #A, #B)") —
and proceeds. Keep the refusal ONLY when measurement fails or attribution is
genuinely ambiguous (e.g. zero active time in window). `--actual <n>` stays as
the explicit override; `--no-actual` keeps its meaning (genuinely nothing to
measure → actual_hours: N/A).

## Done when

- `sdlc close --issue N --verified '…'` with no `--actual` closes in ONE
  invocation using the measured value, and the info line shows the number +
  attribution before the calibration ledger records it.
- Measurement failure still refuses with the current next-action text.
- Re-measure with `--friction-report`: no-actual refusal volume collapses.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.2  impl=0.3
item: milestone-review     design=0.0  impl=0.12
design-buffer: 0.15
total: 0.65
```

Σdesign 0.2 × 1.15 + Σimpl 0.42 × 1.0 = 0.65. Single-pass atomic change
(smaller-go-module: reorder one gate block behind an existing Fn-seam pattern +
tests + shadow-doc sweep); milestone-review = the close-time boundary review.
*Derived against `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`.*

## Plan

Durable plan: `workshop/plans/000178-close-adopts-measured-actual-plan.md`.
Single review boundary (no Mx tags).

- [x] Task 1 — adopt-path in computeClose via `computeActualForCloseFn` seam (TDD: adopt / unmeasurable-refusal / N/A sentinel / deviation-skip-when-adopted)
- [x] Task 2 — wording sweep: warmup, explainActual, helptext/close.md, AGENTS(.base).md §5, atlas
- [x] Task 3 — live verify + close (the close itself dogfoods the adopt path)

## Log

### 2026-07-14 — built (single pass)
- 2026-07-14: closed — go test ./cmd/sdlc/... green incl the 4 new adopt suites (pure decision %.2f-pinned per status, adopt line incl milestone cumulative note, seam wiring with single-engine-run call count, gatesig no-collision); live real-engine verify: omit-path printed "using measured actual: 0.47h (window e4e06cbe -> HEAD)" then the verified gate refused pre-mutation; THIS close itself runs with --actual omitted, dogfooding the adoption; review verdict: FIX-THEN-SHIP

Omit-path now measures once (computeActualForCloseFn seam) and adopts on
actualMeasured (%.2f-pinned, info line with window + peer attribution +
milestone-mode cumulative note); unmeasurable statuses keep the explainActual
refusal, which reuses the same measurement (no re-run); --no-actual unchanged;
adopted values skip the #87 deviation check. Wording swept: warmup,
explainActual, helptext close.md + milestone-close.md, AGENTS.base.md §5,
atlas. Gatesig no-collision test guards the #172 instrument (adopt line
classifies to none; refusal anchors byte-identical). Live verify: real-engine
adoption observed ("using measured actual: 0.47h (window e4e06cbe → HEAD)")
with the verified gate stopping the run pre-mutation. go build/vet/test green.

### 2026-07-14 — close review: FIX-THEN-SHIP (both Importants fixed)

Sidecar `workshop/plans/000178-close-adopts-measured-actual-close-review.md`.
Important #1: milestone mode no longer adopts (issue-scoped window would write
cumulative hours into per-milestone project blocks — double-count; suggest
flow retained there, regression test added). Important #2: xx-issues SKILL.md
shadow doc updated. Minors: flag help, prefixHash reuse, `sdlc actual` advice
now adopt/increment-shaped. Ledger: est 0.65 / actual 0.49 (ratio 1.3x) — the
actual was ADOPTED by the feature itself (first production use).

### 2026-07-14

Filed from #172 M4 follow-on discussion (operator-approved: "sdlc shouldn't
fail the actuals, as it knows how to compute").
