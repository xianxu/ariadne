---
id: 000432
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '148ba2e0b06981f01e438d4f49dbe3cbe426e22d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T18:01:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-worker: an early draft PR for CI overlap needs no hub approval

## Problem

Slots ask the hub for permission before opening an early draft PR so GitHub CI overlaps a queued local suite (pair:3 asked twice on 2026-10-17). The hub always says yes, because the suite, close review and merge gates still run before landing. Each ask costs a hub round trip while the suite queue is long.

## Spec

tl-worker help states it as a rule: an early draft PR for CI overlap needs no hub approval, provided the recorded suite, close review and `sdlc merge` gates still precede landing. Add it as a rules row like the other worker rules.

## Done when

- `tl-worker --help` carries the rule, and its rules test covers the phrase.

## Plan

- [ ] Add a worker rules-table assertion for the early draft PR permission and confirm it fails against current help.
- [ ] Add one help rule beside the existing permissions, then run the focused tests and recorded suite.
- [ ] Close with the review gate and land through the normal PR and merge gates.

## Log

### 2026-10-17

The hub approved quick flow. The existing embedded help and `rules.tsv` remain the single source and coverage pattern (ARCH-DRY); no PR or landing behavior changes.
