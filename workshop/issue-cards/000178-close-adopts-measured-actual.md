---
id: '000178'
status: done
started: 2026-07-14T16:42:01-07:00
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.65
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
