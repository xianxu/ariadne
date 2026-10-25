---
id: 000614
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '01ac3d2e16a812fe8f6fd518902009c5347cb49e' # card fields mirrored from issue-cards; edit via sdlc
---

# A granted merge waiting behind an active landing exits on old-head readiness instead of waiting

## Problem

## Spec

**A granted merge that has to wait behind an active landing exits on the readiness check that a granted admission is
supposed to skip (pair:9, pair#869, 10-24).** `sdlc merge` help says a fleet-blocking grant departs alone, and after
granted admission the merge integrates main and runs the missing checks. But when the granted merge has to wait behind
an already-active landing, the waiting path applies the normal train-boarding readiness check, which demands a recorded
check on the old head. pair#869 (granted; its old head lacked another issue's fix, so an old-head suite would fail) was
refused and its runner exited while waiting behind pair#878, leaving the grant active and nothing in the line. It had
to re-enter by hand after the active landing finished.

Fix: while waiting, a granted merge applies the same rule as at its granted admission (missing checks run after
integration), and stays in the line.

## Done when

- A granted merge waiting behind an active landing stays queued without demanding old-head readiness, then integrates
  main and runs its checks at admission; tested.

## Plan

- [ ]

## Log

### 2026-10-24
