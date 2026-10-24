---
id: 000585
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: '77897b8090c2b77e4c1e77a9c233322b743f3f04' # card fields mirrored from issue-cards; edit via sdlc
---

# Boundary reviewer can't run headless Chrome in its fence; accept a recorded host browser run or permit Chrome

## Problem

## Spec

**The boundary reviewer can't run headless Chrome inside its read-only fence, so every browser change costs a review
round (ops-echo:0 and pair, 10-23).** Seen three times: echo#22 M2 (Chrome timed out at CDP Page.enable), echo#25 M2
(Chrome sandbox initialization failed), and pair#850 (the reviewer's Chrome GPU startup failed). Each slot then
answered with host-run browser evidence through `sdlc finding dispute`, which the reviewer accepted on echo#22; the
cost is a whole extra review round each time, and browser conformance tests are becoming common (cast notes, echo's
board).

Pick the simpler of two fixes in the plan:
1. **Recorded host evidence up front:** the worker runs the browser tests on the host against the exact reviewed head
   and records them (head, command, digest of the log, pass or fail) as a typed artifact the reviewer reads instead
   of trying to launch Chrome; the reviewer still reviews the code and the test itself, and a missing or mismatched
   record is a finding.
2. **A reviewer fence that permits Chrome** (the specific sandbox, GPU and process permissions headless Chrome needs),
   if that can be done without widening what the reviewer can write.

## Done when

- A browser-test change gets through review without a dispute round: either the reviewer accepts a matching host
  record, or Chrome runs inside its fence (tested with one of today's cases).

## Plan

- [ ]

## Log

### 2026-10-23
