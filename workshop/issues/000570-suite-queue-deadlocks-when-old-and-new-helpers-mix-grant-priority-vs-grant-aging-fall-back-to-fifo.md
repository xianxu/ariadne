---
id: 000570
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '45ed51532276e9c749f84573d856206eb180f934' # card fields mirrored from issue-cards; edit via sdlc
---

# Suite queue deadlocks when old and new helpers mix (grant priority vs grant aging); fall back to FIFO

## Problem

## Spec

**Suite-queue deadlock from mixed helper versions after ariadne#560 (grant aging), observed 10-23 ~01:30.** No suite
started for minutes with both machine slots free. Queue status (ariadne:0's sdlc) showed protected FIFO tickets 32–36
(all "age unknown", from older helpers) ahead of granted ticket 37 (pair#832, "aging overrode grant priority"). The
granted runner (new helper) waited for ticket 32's holder; ticket 33's runner (an older helper, from its checkout's
pinned ariadne) said the next was #832. Each side deferred to the other: older helpers apply plain grant priority, the
newer one applies aging. Revoking the grant restored a single order and suites started at once.

Root cause: admission order is computed by each waiter's own `test-lock.py`, and ariadne#563 deliberately keeps the
checkout-supplied helper, so during a rollout, helpers with different ordering rules share one lock root.

Fix options to weigh in the plan: (a) record the ordering-policy version in each ticket and have a newer helper defer
to the oldest policy present (or refuse grants) while mixed; (b) a single shared ordering authority (the newest
helper on the machine) that every waiter consults; (c) a guard: `sdlc landing grant` refuses (or warns) while any
waiting ticket comes from a helper older than the grant-aging policy. Must never deadlock: when waiters disagree,
fall back to plain FIFO.

## Done when

- A mixed-version queue (old-policy waiters plus a new-policy granted waiter) always admits someone (test with both
  helper versions against one lock root).
- Status names a mixed-policy queue explicitly.

## Plan

- [ ]

## Log

### 2026-10-23
