---
id: 000602
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '80e9602d3f39a661ae4b296e6e2fe89fdea70761' # card fields mirrored from issue-cards; edit via sdlc
---

# Split suite priority from landing priority (a suite grant must not hold the repo's landings)

## Problem

## Spec

**Split suite priority from landing priority in `sdlc landing grant` (operator request, 10-24).** Today one grant does
two things: (1) the granted issue's queued suites jump the machine-wide test queue (aging still protects waiters over
an hour), and (2) the repo's normal landing admissions are held until that issue lands or the grant is revoked. Part 2
is right when the granted issue is ready to merge, but it blocks ready work whenever the TL only wants to speed up a
suite: overnight 10-23/24, pair#849's grant held pair#841, #812 and #825 while #849 still needed its check, and
ariadne#593's grant held #594 after #593 was parked behind a new bug. The TL ended up granting and revoking repeatedly.

Proposal: two separate controls.
- **Suite priority** (for example `sdlc queue prioritize --issue N`): N's suites jump the test queue, with the same
  aging protection; no effect on landings.
- **Landing hold** (the current `sdlc landing grant`): holds normal admissions so N lands next; meant for an issue that
  is ready to merge now.

## Done when

- A suite-priority grant moves the issue's suites ahead in the queue without holding any repo's landings; tested with
  a ready waiter in the same repo admitted while the prioritized suite runs.
- The landing hold keeps today's behaviour; help text says when to use each.
- Both expire and revoke the same way grants do today.

## Plan

- [ ]

## Log

### 2026-10-24
