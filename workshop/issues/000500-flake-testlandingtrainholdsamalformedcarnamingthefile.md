---
id: 000500
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd677d654a54abf335b763b4a2725219083621efe' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestLandingTrainHoldsAMalformedCarNamingTheFile

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainHoldsAMalformedCarNamingTheFile: Same exact head 18d51cce1ff21c7977bf70093f3065510482f9f7: recorded integration attempt fd7235f612af84bd4ddadf2187e50169 failed only TestLandingTrainHoldsAMalformedCarNamingTheFile at landingtrain_test.go:970 (merge did not finish, 128.90s); all other shards/packages passed. Failed log: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fd7235f612af84bd4ddadf2187e50169.log. Unchanged same-head targeted go test ./cmd/sdlc -run ^TestLandingTrainHoldsAMalformedCarNamingTheFile$ -count=1 -timeout=5m -v then PASS (test 6.11s, package 6.499s), captured in ariadne:3 terminal/session transcript on 2026-10-20. No code changes between failure and pass. Investigate the fixture merge ordering; no timeout increase.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-20
