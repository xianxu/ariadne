---
id: 000452
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '56242d3540bce9ab9c4cf46e7951d42545eb8ab1' # card fields mirrored from issue-cards; edit via sdlc
---

# suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue

## Problem

A fleet-blocking landing grant (sdlc landing grant) orders the merge line but not the machine-wide suite queue (WF_TEST_SUITES, two slots). On 2026-10-18 the urgent pair#672 and pair#676 each waited behind ordinary suites for a machine slot even with the grant; the weekly snapshot shows a median queue wait of 11 to 20 minutes and a 90th percentile over 30 minutes.

## Spec

A suite started for an issue holding an active fleet-blocking grant (in that repo's landing line) is admitted to the machine suite queue ahead of every waiting normal suite. It never preempts or kills a running suite: it takes the next free slot. Ordering among several priority suites is FIFO. The wait message names the priority and what it waits behind. Weight (#346) is still respected. The grant's expiry ends the priority.

## Done when

- With two slots busy and normal suites waiting, a granted issue's suite starts on the next free slot before the earlier waiters (regression with the real lock helper).
- Without a grant, ordering is unchanged FIFO.
- An expired grant gets no priority.

## Plan

- [ ]

## Log

### 2026-10-18
