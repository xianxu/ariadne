---
id: 000522
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1d3d1b13b268305b3ca041297005f617e4765c42' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead shaping: mock feedback rounds are commits; a new mock-N+1 only for a new direction

## Problem

## Spec

Add the operator-approved (10-20) mock feedback-round convention to tl-lead's SHAPING section (landed in #518):

- **Each round of operator feedback refines the current mock version in place and is one commit** to that version
  (e.g. `mock-3: round 4: …`). Undo/redo of a round is git: "undo the last round" means the agent reverts that commit
  and the mock server reloads. The operator never touches git.
- **A new `mock-N+1` only for a different direction worth comparing side by side** (the operator says "try it another
  way", or a rework big enough that the old one should stay beside it). Older versions stay servable by URL.
- So versions stay few and meaningful (alternatives), rounds stay cheap (commits).
- Versions are separate `mock-N/` folders committed to git (not branches), sharing one root `fixtures.json`
  (pair#755's `couch mock serve` layout).

Source: ops `session.side/2026-10-20-34-mock-server-plan.md`; product-shaping project Log 10-20 23:05.

## Done when

- tl-lead's SHAPING section states the round/version rule above, concisely, next to the mock conventions.
- `tl-lead --help` (or its shaping subsection) shows it; existing tests for the help text pass.

## Plan

- [ ]

## Log

### 2026-10-21
