---
id: 000568
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'bdd4a98d544dedadd16fcf20111dfa7d1f809de6' # card fields mirrored from issue-cards; edit via sdlc
---

# Merge should wait on a pending ancestor suite across a docs-only window instead of requeueing

## Problem

## Spec

`sdlc merge` already accepts a **passed** recorded suite on a first-parent ancestor when every later commit is
docs/issue-only (`localRecordsOnAncestors`, verified by parley.nvim:3 on 10-22). But when that ancestor's suite is
still **queued or running**, merge doesn't adopt it: it queues a fresh exact-head suite at the back of the machine
queue, and the slot has to cancel one of the two by hand.

Seen twice on 10-22:
- echo#12 (ops-echo): suite queued on `6031373` at position 7; close added only workshop docs (5 files under
  `workshop/`, up to `7bde936`); merge queued a new suite on `7bde936` (lease-3680064347) and the worker cancelled the
  original, so a docs-only window cost a requeue in a saturated queue.
- parley.nvim#409: merge auto-queued a duplicate on the close-only head; the slot cancelled the duplicate to keep its
  original place, and only afterwards confirmed the ancestor pass would count.

Fix: when the window from an ancestor with a pending or running recorded suite to the merge head is docs/issue-only
(the same notion close uses for the atlas gate), merge **waits on that run** instead of queueing a new one, and uses
its result as it would a passed ancestor record. A code change anywhere in the window still requires the exact-head
suite.

## Done when

- With a queued ancestor suite and a docs-only window, merge queues nothing new, waits on it, and lands on its pass
  (test).
- A code change in the window still queues an exact-head suite (test).

## Plan

- [ ]

## Log

### 2026-10-22
