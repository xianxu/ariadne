---
id: '000087'
status: done
created: 2026-06-05
updated: 2026-06-05
estimate_hours: 1
actual_hours: 0.08
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
