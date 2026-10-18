---
id: 000432
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ab311a744747af9c9f7e1be05540fbd3b599243b' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-worker: an early draft PR for CI overlap needs no hub approval

## Problem

Slots ask the hub for permission before opening an early draft PR so GitHub CI overlaps a queued local suite (pair:3 asked twice on 2026-10-17). The hub always says yes, because the suite, close review and merge gates still run before landing. Each ask costs a hub round trip while the suite queue is long.

## Spec

tl-worker help states it as a rule: an early draft PR for CI overlap needs no hub approval, provided the recorded suite, close review and `sdlc merge` gates still precede landing. Add it as a rules row like the other worker rules.

## Done when

- `tl-worker --help` carries the rule, and its rules test covers the phrase.

## Plan

- [ ]

## Log

### 2026-10-17
