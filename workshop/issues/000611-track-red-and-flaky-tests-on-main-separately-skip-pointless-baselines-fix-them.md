---
id: 000611
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '9a5b58bfe172fafbe3bc0ab5a959f2764c1114c6' # card fields mirrored from issue-cards; edit via sdlc
---

# Track red and flaky tests on main separately, skip pointless baselines, fix them on a schedule

## Problem

## Spec

**Track red and flaky tests on main separately, skip pointless baselines, and fix them on a schedule (operator
approved, 10-24).** When a merge's integrated suite fails, sdlc automatically runs a full main-baseline suite to see
whether main is red. That costs a full suite even when the failing tests are already registered or already known red
on main (pair#877 on 10-24 was 17 of 20 failing, and was a bug, not a flake). Source: ops session.side/2026-10-24-09-test-inventory-across-lifecycle.md (operator approved 10-24).

Change:
1. **Two registries:** flaky (intermittent; known-flake reruns apply) and red-on-main (fails deterministically or
   nearly so; a real bug with an owner). Registration records which, with the evidence (fail rate).
2. **Skip the baseline** when every failing test is already registered in either list.
3. **Scheduled fixing:** a recurring task (a TL procedure plus a tl-lead or scheduled-task hook) lists both registries
   by age and failure count and dispatches the oldest or most frequent to a free slot.

## Done when

- Registration distinguishes flaky from red-on-main, with evidence; a red-on-main entry doesn't get flake reruns.
- Merge skips the baseline when all failures are registered; tested.
- A command lists both registries by age and frequency for the scheduled fix pass.

## Plan

- [ ]

## Log

### 2026-10-24
