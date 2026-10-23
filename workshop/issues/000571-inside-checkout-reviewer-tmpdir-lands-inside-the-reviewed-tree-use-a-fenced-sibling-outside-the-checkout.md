---
id: 000571
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd9c1694aebaaec984f4eed355de551248c3f43b6' # card fields mirrored from issue-cards; edit via sdlc
---

# Inside-checkout reviewer TMPDIR lands inside the reviewed tree; use a fenced sibling outside the checkout

## Problem

## Spec

From parley.nvim#422 (root cause by parley.nvim:1, 10-23): an inside-checkout `sdlc` close review sets `TMPDIR` to
`<clone>/.sdlc-review/tmp` (reviewbox/confine.go, by design, to stay inside the write fence). Repos whose test
harness derives HOME/XDG/scratch from `TMPDIR` then put all test scratch **inside the reviewed tree**; parley's
zero-config chat walks up, finds the clone's tracked `.parley` marker, and writes into `<clone>/workshop/parley`,
breaking parley's placement invariant (about 10 scattered failures plus a stray untracked artifact in review clones).

Fix: give the reviewer a `TMPDIR` that is **inside the write fence but outside the reviewed checkout** (a sibling of the
review checkout under the review box), so no repo's scratch can land in the tree under review. Keep every other
confinement property.

## Done when

- In an inside-checkout review, `TMPDIR` resolves outside the reviewed checkout and inside the write fence (test).
- A parley test run inside a review clone leaves the reviewed tree clean (integration test or documented manual
  probe with parley.nvim#422's guard).

## Plan

- [ ]

## Log

### 2026-10-23
