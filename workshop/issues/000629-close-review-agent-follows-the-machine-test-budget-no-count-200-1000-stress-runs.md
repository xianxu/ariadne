---
id: 000629
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'cbafd720961ca9b8e8ec2c9be9e2c148480e324f' # card fields mirrored from issue-cards; edit via sdlc
---

# Close-review agent follows the machine test budget (no -count 200/1000 stress runs under load)

## Problem

## Spec

**The close-review agent ignores the machine's test budget (pair:2 on pair#759, 10-25).** The fresh reviewer that
`sdlc close` dispatches reran a suspected flake at `-count=20` and `-count=200`, plus baselines at 100 and 1000, while
the fleet was under a load notice (targeted runs only, `-count` ≤ 10, one at a time; stress reps only with the TL's
OK). Worker slots get that rule from the hub's messages; the dispatched reviewer gets only its review prompt, so it
can't know. A 1000-rep stress run on a loaded machine slows every slot's suite and can itself create the timing
flakes it hunts.

Give the review dispatch a test budget it must follow: a default per-invocation cap (e.g. `-count` ≤ 10, no parallel
stress batches), stated in the review prompt and enforced where sdlc launches its runs if it can, overridable only by
an explicit flag a TL passes. Flake reproduction beyond the cap becomes a recommendation in the verdict ("rerun at
N under quiet load"), not something the reviewer runs.

## Done when

- The review prompt carries the budget; a review that would exceed it reports the recommendation instead, with a test
  over the prompt and any enforcement point.
- An explicit override exists and is logged when used.

## Plan

- [ ]

## Log

### 2026-10-25
