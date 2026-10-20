---
id: 000491
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
severity: low
collaboration-mode: collaboration
card_mirror: '7d198111156229a08e1bf607b34f13d5b6c6d3a9' # card fields mirrored from issue-cards; edit via sdlc
---

# Suite queue position becomes fictitious on allocation contention

## Problem

Observed while #481's recorded suite waited under real allocation contention: reports alternate between the actual position 3 (ahead contains the train and another waiter) and position 1 with the same next-in-line holder. No suite admission occurred. The new nonblocking secondary-lock callback returns its default empty ahead list when allocation cannot be inspected; acquire interprets len(ahead)+1 as an observed position. This produces false position updates and log churn.

## Spec

When allocation/global admission contention prevents observing FIFO order, report position as unknown or retain explicitly stale last-observed position. Never derive position 1 from a default empty ahead list. Keep admission, priority, weight and FIFO unchanged; do not make secondary locks blocking.

## Done when

- A real held-allocation fixture makes the admission attempt return promptly and never reports an observed position 1 from missing order evidence.
- Once allocation is available, the same ticket reports its actual position and retains FIFO admission.
- Stable contention does not generate alternating fake position updates.

## Plan

- [ ]

## Log

Filed from #481 final-suite observation, after its close review. Source: scripts/test-lock.py Tickets.admission default result and acquire position rendering. The #481 suite keeps its original waiter; this is diagnostic reporting, not an admission failure.
