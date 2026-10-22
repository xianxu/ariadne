---
id: 000538
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2ffd060342ed4742f771d5fe481cbf5ae932dbeb' # card fields mirrored from issue-cards; edit via sdlc
---

# stage --clear recognizes its landed holder after branch deletion

## Problem

## Spec

After pair#752 landed, landing deleted its issue branch and returned pair:9 to its resting branch. Its subsequent plain `sdlc stage --clear` refused the stage lease because the holder identity is currently derived from the caller branch; the holder had to use `--clear --abandoned 752` even though landing is the normal completion, not abandonment. Reported by ops:0 while dispatching ariadne#537.

Allow the original holder slot to clear its landed stage after normal branch deletion using confirmed durable landing evidence. Preserve refusal for other slots/issues, unlanded or unknown outcomes, replaced leases and explicit abandoned-holder acknowledgement. Do not treat a missing branch alone as landing proof.

## Done when

- A regression stages from a worker issue branch, completes a durable landing and branch cleanup, then clears from that original slot on its rest without --abandoned.
- Another slot, unfinished/unknown holder and replaced lease still refuse without clearing or rebuilding.

## Plan

- [ ]

## Log

2026-10-21: Split from #537 per hub permission if not small. Fresh-main restoration is independent of the holder-authentication policy. Existing checkClearLease only sees branch-derived holder strings; safe normal-completion recovery needs durable landed evidence and original-slot binding, rather than a branch-absence or slot-only bypass.
