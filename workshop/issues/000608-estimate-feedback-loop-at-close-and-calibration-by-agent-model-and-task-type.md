---
id: 000608
status: open
deps: [ariadne#603]
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '4f3885badb65157f2031c7ba7dd67bae9fb1dbac' # card fields mirrored from issue-cards; edit via sdlc
---

# Estimate feedback loop at close and calibration by agent, model and task type

## Problem

## Spec

**Close the estimate feedback loop and calibrate by agent, model and task type (operator, 10-24).** Goal two of the
estimation system: estimates that improve over time. Depends on ariadne#603 (measured actuals never block; each close
records exact, partial or unmeasured hours).

Change:
1. **At close, show the estimate against the measured hours**, per part where attributable, and ask the agent for one
   line on the biggest miss; record it with the issue.
2. **Calibration:** from exact and partial results (partial weighted by its attributed share), compute per-agent,
   per-model, per-task-type multipliers (for example "Codex medium underestimates UI work by 1.6x") and show the
   relevant one to the planner at estimate time.
3. **A weekly estimate-versus-actual summary** by agent, model and task type, with the trend, readable by tl-lead (for
   the ops weekly report) and by model-bench.

Side doc: ops `session.side/2026-10-24-11-estimation-system-proposal.md` (operator approved).

## Done when

- Close displays estimate versus actual and records the one-line miss note; tested.
- A calibration command outputs the multipliers from recorded closes, honouring confidence; tested against fixtures.
- change-code shows the relevant multiplier when an estimate is requested.

## Plan

- [ ]

## Log

### 2026-10-24
