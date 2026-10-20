---
id: 000494
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b876efb01da25ecb47c677eb6d797b0dc866d0a5' # card fields mirrored from issue-cards; edit via sdlc
---

# Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it

## Problem

In pair's first multi-car train (176, 10-20 05:48), car pair#717's own `sdlc merge` (PR 483) ended with "active issue
missing without confirmed archive", although the train had already archived the issue (it's under history on main
and its card is done). Harmless, but it reads as an error at the end of a successful landing, and a worker can't tell.

## Spec

When a car's merge finishes after the train landed it, treat "archived by the train" as the confirmed outcome (the
train's `outcomes/<seq>` record says so), and print it as such, not as a missing-issue warning.

## Done when

- A 2-car train fixture: the second car's merge ends with a confirmed train-archive message, no warning (test).

## Plan

- [ ]

## Log

### 2026-10-20
