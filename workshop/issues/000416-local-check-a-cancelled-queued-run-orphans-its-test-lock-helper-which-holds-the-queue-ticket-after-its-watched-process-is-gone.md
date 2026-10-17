---
id: 000416
status: codecomplete
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-17
estimate_hours:
card_mirror: 'dc678791e3a0d3df98006c333b899eb27133f0ad' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T19:36:21-07:00
actual_hours: N/A
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# local-check: a cancelled queued run orphans its test-lock helper, which holds the queue ticket after its watched process is gone

## Problem

## Spec

Reported by pair:3 (pair#629), 2026-10-16. Cancelling `sdlc local-check run` while it was still queued for the machine test line left its `test-lock.py` helper orphaned:
- PID 42255, PPID 1, running `--watch 42240`, but process 42240 no longer existed;
- it kept queue ticket 9 for the cancelled run `sdlc-local-1087195770`, so later runs waited behind a ticket nobody would use.

pair:3 stopped only that helper and deleted no gate files.

Two gaps:
1. Cancelling a queued run does not stop its queue helper.
2. The helper's `--watch` does not notice that the watched process is gone, so the ticket is never released.

Related to, but not covered by, #400 (cancel must stop only your own run).

## Done when

- Cancelling a queued `local-check run` stops its helper and releases its ticket; a test shows the next run admitted.
- A helper whose watched PID is gone releases its ticket and exits within a bounded time; a test kills the watched process with SIGKILL and checks that the ticket is released.

## Plan

- [ ]

## Log

### 2026-10-16
