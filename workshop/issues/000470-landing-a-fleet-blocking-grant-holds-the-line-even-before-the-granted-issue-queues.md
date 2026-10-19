---
id: 000470
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9ab2c63119836f04300e00670f9e2956cc9c907e' # card fields mirrored from issue-cards; edit via sdlc
---

# landing: a fleet-blocking grant holds the line even before the granted issue queues

## Problem

2026-10-19 ~12:20: ariadne main was red (#469 fixes it) and #469 held a fleet-blocking landing grant, but #469 had not yet
entered `sdlc merge`. The landing line admitted #465's waiting merge anyway: it integrated main and pushed, about to spend a
two-slot recorded suite on a known-broken main. A fleet-blocking grant only reorders merges already waiting; while the granted
issue is still being worked, everyone else keeps landing (and failing).

## Spec

While an unexpired fleet-blocking grant exists for an issue that is not yet waiting in the line, the line admits no other
merge in that repo (they wait, with a message naming the grant, its reason and expiry). Grants that are not fleet-blocking
keep today's reordering. Expiry or revocation releases the hold. `sdlc landing status` shows the hold. State table in the plan.

## Done when

- A fixture: grant for #A (not queued) + waiting #B → #B is not admitted and reports the grant; #A enters → #A admitted first;
  grant expires with #A never queued → #B admitted.
- Existing grant-ordering tests unchanged.

## Plan

- [ ]

## Log

### 2026-10-19

## Log

- 2026-10-19 ~15:00 (hub): a second case. ariadne:3 held the turn with #459 (granted) and would switch the same checkout to
  #465 (granted, fixes the flake hitting every ariadne suite). Between #459 exiting and #465's ticket existing, the normal
  waiters #466 and #467 could be admitted first. A grant also needs to cover "my next granted issue in this checkout", or a
  short hand-off hold. Same fix shape: a fleet-blocking grant holds the line until its issue queues or the grant expires.
