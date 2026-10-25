---
id: 000617
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '8d421a02c4db6b3dd2c12334e2dfe49ee1607c11' # card fields mirrored from issue-cards; edit via sdlc
---

# close and milestone-close hold the repo lock through the whole review, blocking other slots' bookkeeping

## Problem

## Spec

**`sdlc close` and `sdlc milestone-close` hold the repo's shared transaction lock for their whole run, including the
boundary review, blocking every other slot's bookkeeping in that repo (TL observation, 10-24).** pair:5's
`sdlc milestone-close --issue 859 --milestone M1` (pid 77562) held pair's repo lock for 25+ minutes while its review
ran; meanwhile pair:4's `sdlc stage --clear` and other slots' doc-only pushes and issue publishes waited on it. The same
happened earlier with pair#851 waiting on another close's lock. With 8 to 9 active slots per repo, one long review
serializes everyone's bookkeeping.

Fix: take the lock only around the short critical sections (reading and writing tracker or card state, publishing
commits), not across the review dispatch and wait; revalidate state after the review before writing the verdict.

## Done when

- A close or milestone-close holds the repo lock only for its bookkeeping writes; a concurrent `stage --clear`,
  `issue publish` or another close's bookkeeping proceeds while a review runs; tested with a slow fake review.
- A state change during the review is detected at the post-review write (refuse or re-run), never overwritten.

## Plan

- [ ]

## Log

### 2026-10-24
