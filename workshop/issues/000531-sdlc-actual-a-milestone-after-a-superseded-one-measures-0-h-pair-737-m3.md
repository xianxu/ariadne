---
id: 000531
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c29a914a58aafc9c7a48c94bf5c634a5210749a3' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc actual: a milestone after a superseded one measures 0 h (pair#737 M3)

## Problem

## Spec

pair#737 (pair:1, 10-21): after the operator-approved re-scope (old M2 superseded, new work tagged M3 per the TL's
instruction), `sdlc` measures **0.00 h** active time for M3 despite 21 admitted sessions and substantive M3 commits
since the re-scope. The worker correctly refuses to hand-type hours. Likely cause: the milestone's measurement window
is anchored on a boundary (the superseded M2's start/close, or the M3 tag's first appearance) that excludes the
actual M3 work, or a superseded milestone isn't modelled at all.

Fix: measure a milestone that follows a superseded one from the supersession point (the last closed boundary or the
revision that superseded it) to the milestone's close, and make the measurement explain its window (start commit,
end, session count) so a 0 h result is visibly wrong rather than silently adopted.

## Done when

- A fixture with M1 closed, M2 superseded (never closed), M3 worked on measures M3 from after M1's close (or the
  supersession revision) and attributes the M3 sessions; a regression for the 0 h case.
- `sdlc actual` prints its window and session count.

## Plan

- [ ]

## Log

### 2026-10-21
