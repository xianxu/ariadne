---
id: '000134'
status: done
started: 2026-06-26T15:57:18-07:00
created: 2026-06-26
updated: 2026-06-26
estimate_hours: 3.9
actual_hours: 0.76
---

# Make actuals and estimates agent-robust

## Problem

`sdlc actual` and `sdlc`'s estimate guidance were originally dogfooded mostly
through Claude Code. That left two agent-portability gaps:

- actual measurement assumed Claude's transcript location/shape and a narrow
  issue-subject convention, so Codex work could look like "no measurable
  activity" even when the session transcript and commits existed.
- estimation logic is split correctly between shared method and repo-local/user
  calibration, but discovery is implicit. The shared `estimate-logic-v2` lives in
  brain, while the current operational grammar is in `cmd/sdlc/helptext/estimate.md`;
  an agent can satisfy the block syntax without realizing it should read the
  calibrated local estimator.

Commit `f62d099` fixed the immediate Codex transcript parser and `<area>: #N`
commit-window failure. This issue is for making that support robust instead of a
one-agent patch.
