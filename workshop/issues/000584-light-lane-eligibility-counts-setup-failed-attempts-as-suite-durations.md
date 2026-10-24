---
id: 000584
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '48d4daabb77bff65ec5c7b0db56230dbdd18a7da' # card fields mirrored from issue-cards; edit via sdlc
---

# Light lane eligibility counts setup-failed attempts as suite durations

## Problem

## Spec

**The light lane counts failed setup runs as suite durations (metis:1, 10-23).** ariadne#559's short-suite lane decides
eligibility from the last 3 recorded suite attempts' durations. In metis, a 1-second attempt that failed during setup
was counted as a suite duration (the lane line showed 56s, 90s, 1s), which can make a repository look eligible on
evidence that isn't a suite run.

Fix: only completed suite runs (passed, or failed in tests after running) count toward lane eligibility; attempts that
ended in setup, preflight or could-not-run states are excluded and don't fill one of the three slots.

## Done when

- Setup-failed and could-not-run attempts are excluded from the eligibility window; completed runs count as today
  (tests).

## Plan

- [ ]

## Log

### 2026-10-23
