---
id: 000492
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '531f905d43554970cbf7138b8b7163163a2d7328' # card fields mirrored from issue-cards; edit via sdlc
---

# Landing-train suite attempts omit slot provenance, so overlapping issues' sdlc actual refuses

## Problem

Landing-train suite attempts are saved without a `slot` field. pair's train 174 suite (attempt
f6a4337408d36adfd9d00ddb9d298d99, head c2c6b927, conductor pair:7) was recorded as version 2 with no `slot`, so
`sdlc actual` for an unrelated issue whose window overlapped it (pair#725) refused: "suite attempt … has unknown
slot provenance … ask the TL/hub to restore the named measurement evidence". The TL restored it by hand on 10-20
(added `"slot": "pair:7"`; original kept). Every train will do this to every overlapping issue's actuals.

## Spec

The train conductor records its suite attempt with `slot` (the conductor's workspace) like any local-check suite,
plus whatever identifies it as a train run (train id) if the schema allows, so measurement attributes it correctly.

## Done when

- A train-suite attempt written by a fixture train carries `slot` (test), and `sdlc actual` for an overlapping issue
  measures without refusal (test).

## Plan

- [ ]

## Log

### 2026-10-20
