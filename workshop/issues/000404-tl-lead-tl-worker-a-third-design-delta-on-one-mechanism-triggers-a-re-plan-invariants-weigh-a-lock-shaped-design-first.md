---
id: 000404
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '8cb487f21a93a3a6a3238f740ae9bcb074437b97' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead/tl-worker: a third design delta on one mechanism triggers a re-plan; invariants weigh a lock-shaped design first

## Problem

2026-10-15: the TL approved eight design deltas to pair#584, one at a time, each reasonable alone (argv for every OS
process, legacy rows, unbound launches, ledger caps, codec speed, concurrency), without asking why one guard needed
eight. The guard (pair#556) proves "no live writer" by forensics over historical state, about 1,700 lines with ~45
refusal reasons; a wrapper-held flock that dies with its process does the same in one step with two outcomes
(pair#595). The operator: "the same old over engineering." #556's plan never weighed a lock.

## Spec

- tl-lead's plan-review rules gain a **delta-count tripwire:** when a slot asks for a third design delta on the same
  mechanism, stop approving deltas; ask for a re-plan that names the simplest alternative and why it fails.
- tl-lead's and tl-worker's plan rules gain one question for any invariant ("no two X", "never while Y"): **what
  single resource, held for the owner's lifetime (a lock or lease), would make this a one-step check?** A plan that
  proves absence by observation must say why that resource can't work.
- Phrase-guarded in the help tests.

## Done when

- `tl-lead --help` and `tl-worker --help` carry both rules; help tests guard the phrases.

## Plan

- [ ]

## Log

### 2026-10-15
