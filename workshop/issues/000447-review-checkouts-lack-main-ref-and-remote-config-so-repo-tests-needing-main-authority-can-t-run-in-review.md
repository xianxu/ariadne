---
id: 000447
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f1cfabfb26259454df3003f7ca3420417301f497' # card fields mirrored from issue-cards; edit via sdlc
---

# review checkouts lack main ref and remote config, so repo tests needing main authority can't run in review

## Problem

The boundary review runs in a disposable checkout that lacks the main branch ref, its upstream and remote configuration. Repository tests that need main authority (pair's Go lint catalog finds tracked files and the merge base against main) correctly refuse there, so the reviewer cannot run them and raises an "unverifiable" finding. It hit pair#665 M1 (BR-4) and pair#666 (BR-1) on 2026-10-18; each slot had to capture host evidence and dispute it.

## Spec

Give the review checkout the same main authority a slot checkout has: the main ref at the reviewed base, and the remote and upstream configuration (read-only; the reviewer never pushes). Any test that runs in the slot then runs the same way in review.

## Done when

- A repository test that resolves main and the merge base passes inside the review checkout; a regression test reproduces the pair#665 shape.
- The review checkout still cannot push or change the real repository.

## Plan

- [ ]

## Log

### 2026-10-18
