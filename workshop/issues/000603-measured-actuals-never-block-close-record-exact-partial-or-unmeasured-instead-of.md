---
id: 000603
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'bb59cc83a8fedc5f1884e6fa61fe18a77a6dad7e' # card fields mirrored from issue-cards; edit via sdlc
---

# Measured actuals never block close: record exact, partial or unmeasured instead of refusing

## Problem

## Spec

**Measured actuals must never block a close; degrade instead of refusing (operator, 10-24).** The operator's framing:
"the goal of the estimation system is both a soft nudge for AI to break things down and reason about how complex the
parts are, and to be used for future estimates. It is auxiliary, and shouldn't be blocking landing, generally
speaking."

Since ariadne#580, close refuses measured actuals in several shapes (resting branch, fixed by #591; 64 KiB proof, fixed
by #593; stage branches, #594; detached checkouts; more than 128 sources in the window). Each refusal stops the close,
needs a TL authorization for `--no-actual`, and loses the hours entirely. Overnight 10-23/24 that happened to roughly a
dozen closes.

Change:
1. **Degrade, don't refuse.** Attribute what can be attributed; report unattributed minutes and why; record the result
   with a confidence: `exact`, `partial` (with the unattributed share) or `unmeasured` (with the reason). Close proceeds
   in every case; no TL authorization needed.
2. **No source-count refusal.** Past the source cap, merge or summarise sources rather than refuse.
3. **Calibration reads confidence.** Velocity and estimate calibration use `exact` fully, `partial` weighted by its
   attributed share, and skip `unmeasured`.
4. **Retire the `--no-actual` ritual** for measurement failures (keep the flag for genuinely inapplicable windows).

## Done when

- Every refusal shape listed above closes with a recorded confidence instead of refusing; tests replay the real
  pair#856, #812, #863, #830, #851 and #838 windows.
- Calibration output distinguishes and weights the three confidences.
- Close help and the atlas describe the confidence levels.

## Plan

- [ ]

## Log

### 2026-10-24
