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

The malformed-car train fixture hit the shared fixed two-minute finished() ceiling under suite load (128.90s), while its targeted control passed in 6.11s.

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainHoldsAMalformedCarNamingTheFile: Same exact head 18d51cce1ff21c7977bf70093f3065510482f9f7: recorded integration attempt fd7235f612af84bd4ddadf2187e50169 failed only TestLandingTrainHoldsAMalformedCarNamingTheFile at landingtrain_test.go:970 (merge did not finish, 128.90s); all other shards/packages passed. Failed log: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fd7235f612af84bd4ddadf2187e50169.log. Unchanged same-head targeted go test ./cmd/sdlc -run ^TestLandingTrainHoldsAMalformedCarNamingTheFile$ -count=1 -timeout=5m -v then PASS (test 6.11s, package 6.499s), captured in ariadne:3 terminal/session transcript on 2026-10-20. No code changes between failure and pass. Investigate the fixture merge ordering; no timeout increase.

## Spec

Folded into #498 by hub direction. Both incidents use the same completion helpers, so one shared remaining-supervisor-budget correction covers them. No independent implementation is needed.

## Done when

- Covered and verified by #498; this duplicate registration is removed after proof.

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [x] Fold into #498 and remove this registration after verification.

- [ ]

## Log

- 2026-10-20: Folded into #498 per hub direction (8345800b-3ef8-4886-9b48-cb04b823c141, reaffirmed 5f720053-8ab6-49f2-8b2c-68b962262d45). #498 shares the deadline-derived completion policy and goroutine diagnostic across train fixtures. The malformed-car control passed in 8.38s, and the exact implementation commit passed the recorded full suite: 1,877 cmd/sdlc tests, 75 other packages, all eight shards, zero failures/reruns, 878.6s. Evidence and raw logs are in #498. Explicitly unregistered this test after that pass; wontfix records duplicate triage, not an independent code close. The implementation lands with #498 / PR 355.

### 2026-10-20
