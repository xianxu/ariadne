---
id: 000498
status: codecomplete
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
flake_tests: 
started: 2026-10-20T14:48:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.53
tracker:
    version: 1
    completion:
        token: close-ab7f12d92611
        repository: github.com/xianxu/ariadne
        reviewed_head: 18ef5d0ed97b172b1e1d02eb0b204b4adf76a475
        evidence_commit: aab0d01046e309b267e3f2d918b327a167946b01
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/train

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/train: Same exact code head 68bfefbde7c363252ba67ab2ddcfd05ca1aaa48b: required recorded full suite attempt 11cab56f89656e3bab43c5a95e196394 completed all eight shards, 1854 cmd/sdlc tests and 75 other packages; only /train failed at landingtrain_test.go:414 finished() two-minute merge ceiling, variant 121.77s, parent 138.62s. Failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/11cab56f89656e3bab43c5a95e196394.log, SHA256 prefix 8930df16c1d9. In a clean detached checkout of that exact commit, go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 6.94s, null 7.12s, head 6.90s, parent 20.96s, package 21.500s. Git rev-parse before and after printed the same exact SHA. Focused log persisted at /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000494/train-parent-same-head.log. This is an explicit same-head fail-to-pass classification; only /train is registered, not siblings. Prior #480 registered /head and landed the eight-shard envelope; this new observation shows the train wait can still intermittently exceed its ceiling at eight shards.
