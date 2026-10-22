---
id: 000560
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3f6f1e389a95dbabab5c0e658929ae5b0bb5deb9' # card fields mirrored from issue-cards; edit via sdlc
---

# Suite queue: landing grants starve old waiters; bound grants by waiter age

## Problem

## Spec

Landing grants (`sdlc landing grant`, fleet-blocking priority) jump the machine-wide suite queue with no bound, so a
run of grants plus ariadne's weight-2 suite can starve small suites indefinitely. Evidence (ops-echo:0, 10-22): the two
oldest tickets (88, parley.nvim-slot3 #409, queued 11:58; 90, echo-slot1 echo#4, queued 12:01) waited about 2.5 h while
newer granted tickets went first (pair#773 twice, ariadne#552, ariadne#555), each granted for a good local reason.

Fix: **aging bounds grants.** A granted ticket jumps only waiters younger than an age limit (e.g. 60 minutes); any
waiter older than that is admitted first, in FIFO order, before further granted tickets. `sdlc queue status` shows
each waiter's age and when aging overrode a grant. The grant still beats everything younger, so it keeps its purpose.

## Done when

- With a ticket older than the age limit waiting, a newly granted ticket is admitted after it, not before (test).
- Granted tickets still jump younger waiters (test).
- Queue status names the aging override.

## Plan

- [ ]

## Log

### 2026-10-22
