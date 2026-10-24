---
id: 000603
status: working
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'bfbc050d1ccb4a92f57c892697ec75bd8489038d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T15:56:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

Durable plan: [000603-measured-actuals-never-block-close-plan.md](../plans/000603-measured-actuals-never-block-close-record-exact-partial-or-unmeasured-instead-of-plan.md).

- [ ] M1 — close never refuses on measurement: confidence classifier, unknown-branch and per-source notes, oversize summary, queue degradation, close adopts every outcome, replay fixtures (#856/#812/#863/#830/#851/#838)
- [ ] M2 — calibration weights by confidence (ledger columns, weighted forecast/converge/drift); close/actual help and atlas

## Log

### 2026-10-24
- Planned after #593 landed (TL: take #603, plan to TL). Code survey: (b)/(e) oversize and source cap already hold exact hours at refusal; (d) unknown branch, (f) per-source attribution and (g) queue evidence fail inside the input resolver and discard partial windows. Open decisions D1-D5 in the durable plan go to the TL.
