---
id: 000500
status: wontfix
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
flake_tests: 
---

# flake: TestLandingTrainHoldsAMalformedCarNamingTheFile

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainHoldsAMalformedCarNamingTheFile: Same exact head 18d51cce1ff21c7977bf70093f3065510482f9f7: recorded integration attempt fd7235f612af84bd4ddadf2187e50169 failed only TestLandingTrainHoldsAMalformedCarNamingTheFile at landingtrain_test.go:970 (merge did not finish, 128.90s); all other shards/packages passed. Failed log: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fd7235f612af84bd4ddadf2187e50169.log. Unchanged same-head targeted go test ./cmd/sdlc -run ^TestLandingTrainHoldsAMalformedCarNamingTheFile$ -count=1 -timeout=5m -v then PASS (test 6.11s, package 6.499s), captured in ariadne:3 terminal/session transcript on 2026-10-20. No code changes between failure and pass. Investigate the fixture merge ordering; no timeout increase.
