---
id: 000178
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours:
started: 2026-07-14T16:42:01-07:00
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

## Plan

- [ ] adopt-measured path in computeClose/close gates + info line + tests (incl. measurement-failure refusal kept)

## Log

### 2026-07-14

Filed from #172 M4 follow-on discussion (operator-approved: "sdlc shouldn't
fail the actuals, as it knows how to compute").
