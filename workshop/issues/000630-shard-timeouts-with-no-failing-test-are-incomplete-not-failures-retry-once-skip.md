---
id: 000630
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'cecdb40c8c7221289acf9fbd9dca5282e9afca00' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T09:51:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis

## Problem

## Spec

**An all-shard timeout under load is treated as an unattributable failure and triggers a full main-baseline run
(ariadne:5 on ariadne#616, 10-25).** #616's integrated merge suite (04:31–05:01) hit the 1800 s shard timeout on every
shard while the machine's load average was 30–50; 764 `cmd/sdlc` tests never ran (the same suite took 19–20 min
earlier). sdlc classified it as an unattributable failure and started a main-baseline diagnosis: another full ariadne
suite (both machine slots), queued behind everything, holding ariadne's landing line for an hour or more, to diagnose
a run that didn't fail any test.

A run where shards time out without a failing test isn't evidence against the code or main. Classify it as
"incomplete (timeout)", distinct from a failure: retry the same head's suite once (in its normal queue turn), and
diagnose against main only if a test actually fails. Report the load average at the timeout in the outcome.

## Done when

- A suite whose shards time out with no failing test records "incomplete (timeout)" with the load at the time,
  retries once, and runs no main-baseline diagnosis; a real test failure still diagnoses as now. Tests for both.
- **The retry runs only the packages that didn't finish** (operator, 10-25): Go reports results per package and sdlc records which passed in the attempt, so the retry selects the unfinished packages through the affected-run machinery instead of rerunning the whole suite; test-level granularity is a non-goal. Test: a timed-out attempt with some packages passed retries only the rest.

## Plan

- [ ]

## Log

### 2026-10-25
