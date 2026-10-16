---
id: 000401
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '5621bc017662b772779c7f069a0a8d156f50b251' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead/tl-worker: hot-path plans state their expected cost per driver and the counted test that defends it

## Problem

Resource regressions (pair#579's 61× writes, pair#584's 15 s decode) passed plan review because nobody asked what an
operation should cost. Ops project resource-observability decided (10-15) that hot-path changes state their expected
cost and count it.

## Spec

- tl-lead's plan-review checklist gains one question for changes on a hot path (per-byte, per-event, per-launch or
  per-scan code): "What does this cost per unit of its driver (TTY bytes, events, launches, slots), and which counted
  test defends it?" Counts, not timings.
- tl-worker's planning rules gain the matching line, so plans arrive with it answered.
- Both phrase-guarded in the help tests.

## Done when

- `tl-lead --help` and `tl-worker --help` carry the question; help tests guard the phrases.

## Plan

- [ ]

## Log

### 2026-10-15
