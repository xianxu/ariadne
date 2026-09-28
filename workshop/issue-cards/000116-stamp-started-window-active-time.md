---
id: '000116'
status: done
created: 2026-06-17
updated: 2026-06-18
estimate_hours: 1.4
actual_hours: 0.41
---

# Stamp started: at claim and window active-time from engagement start

## Problem

`sdlc actual` (active-time-v3) windows on commits: the window starts at the
first `#N`-subject commit's parent (`internal/gitx/window.go`). So all the
operator-attention-heavy DESIGN time — brainstorm, spec, plan, plan-reviews —
that happens *before the first code commit* is excluded from the measure.

Live evidence (the same data that motivated #112):
- **#111** actual measured **0.35h**, but the close note records the true figure
  as ~2h: *"design+brainstorm+4 reviews predate the first commit and both M1+M2
  were implemented by delegated subagents, so most effort is outside the operator
  transcript window."*
- **#110** close note: *"pre-first-commit planning/plan-review + background
  review-waits are outside the commit window, so true effort was higher."*

This is a systematic *under*-measurement of operator-attention. On its own it's a
fidelity bug; for #117 it is the upgrade that makes the calibration loop
trustworthy — #117's auto-calibration (mechanism 3) scores estimate against `sdlc
actual`, and a truncated actual is garbage data. #117 ships first and stamps such
rows `window-trusted: no`; **this issue is what flips new rows to trusted.** Not a
hard blocker for #117 (the trust flag handles its absence), but the reason #117's
data is worthless until this lands.
