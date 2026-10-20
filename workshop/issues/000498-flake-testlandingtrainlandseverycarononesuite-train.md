---
id: 000498
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '05bc11323517a730bfd0774c2d9c3d792158018e' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/train

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/train: Same exact code head 68bfefbde7c363252ba67ab2ddcfd05ca1aaa48b: required recorded full suite attempt 11cab56f89656e3bab43c5a95e196394 completed all eight shards, 1854 cmd/sdlc tests and 75 other packages; only /train failed at landingtrain_test.go:414 finished() two-minute merge ceiling, variant 121.77s, parent 138.62s. Failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/11cab56f89656e3bab43c5a95e196394.log, SHA256 prefix 8930df16c1d9. In a clean detached checkout of that exact commit, go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 6.94s, null 7.12s, head 6.90s, parent 20.96s, package 21.500s. Git rev-parse before and after printed the same exact SHA. Focused log persisted at /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000494/train-parent-same-head.log. This is an explicit same-head fail-to-pass classification; only /train is registered, not siblings. Prior #480 registered /head and landed the eight-shard envelope; this new observation shows the train wait can still intermittently exceed its ceiling at eight shards.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-20
