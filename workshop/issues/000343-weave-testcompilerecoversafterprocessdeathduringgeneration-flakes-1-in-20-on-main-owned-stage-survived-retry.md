---
id: 000343
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: 'e9a1de54a60f080515fa0b732d8f2d7f77dc6b66' # card fields mirrored from issue-cards; edit via sdlc
---

# weave: TestCompileRecoversAfterProcessDeathDuringGeneration flakes ~1 in 20 on main ('owned stage survived retry')

## Problem

`cmd/weave` `TestCompileRecoversAfterProcessDeathDuringGeneration` fails about 1 in 20 runs on main with 'owned stage survived retry' (ariadne:1, 2026-10-12, while closing #265).

## Spec

Unverified hypothesis (ariadne:1): the killed helper's generator grandchild is still alive when the retry reclaims the stage. Reproduce with a targeted loop (never under machine load), confirm the cause, and fix the root (wait for or kill the process group before reclaiming), not the test's timing.

## Done when

- The cause is confirmed and fixed; 100 targeted runs pass; the fix fails the test when reverted.

## Plan

- [ ]

## Log

### 2026-10-12
