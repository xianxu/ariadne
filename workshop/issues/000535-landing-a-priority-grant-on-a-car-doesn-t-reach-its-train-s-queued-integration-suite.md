---
id: 000535
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '26405f5431573eb602d282f44d90333ef99d6895' # card fields mirrored from issue-cards; edit via sdlc
---

# landing: a priority grant on a car doesn't reach its train's queued integration suite

## Problem

## Spec

pair train 254 (cars #775, #774, #771), 10-21 ~17:25: the TL granted fleet-blocking priority to #775 so the train's
already-queued integration suite would jump the machine queue (the train held pair's landing turn ahead of an
operator-accepted car). The grant showed in `sdlc landing status`, but the train's retained suite (lease 57913550, on
the composed train prefix 9ff56af4, not an issue head) stayed in normal FIFO at position 7 across queue ticks.
Grants bind to an issue's own suite lease; a train suite carrying that issue as a car doesn't inherit it.

Fix: a fleet-blocking grant on any car of a composed train applies to the train's suite lease (queued or about to
queue), re-evaluated on each queue tick; revoking or expiring the grant drops it. No priority for unrelated trains.

## Done when

- A train whose car holds a grant gets suite-queue priority for its integration suite (a test with a queued train lease
  and a later grant), and loses it on revoke or expiry.

## Plan

- [ ]

## Log

### 2026-10-21
