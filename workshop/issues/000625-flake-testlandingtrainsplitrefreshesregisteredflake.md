---
id: 000625
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'fa84a834751e1ed8b92e88e62e5356256fefe282' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestLandingTrainSplitRefreshesRegisteredFlake

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainSplitRefreshesRegisteredFlake: Exact-head fail then pass on cdd00f6294a1b87f8d8a94548eabcbd7d24f6755. Recorded integrated suite attempt 7b97ba46047872efe0d2553087669993 failed only this test at landingtrain_attribution_test.go:29: the passing split prefix never reached the registration barrier (33.43s); saved log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/7b97ba46047872efe0d2553087669993.log, returned sha256 prefix cd673bb7fc1f. Targeted go test ./cmd/sdlc -run ^TestLandingTrainSplitRefreshesRegisteredFlake$ -count=1 -timeout=10m passed 28.291s, terminal53677 exit0. git rev-parse HEAD before and after returned the same full SHA and git status --short was empty. Fixture has a fixed30s registration-barrier deadline; related sweep ariadne#622. No code edits, timeout bumps or full-suite retry.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-25
