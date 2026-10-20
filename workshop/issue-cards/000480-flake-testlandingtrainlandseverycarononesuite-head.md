---
id: 000480
status: done
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/head]
started: 2026-10-20T00:35:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 5.65
tracker:
    version: 1
    completion:
        token: close-ea1995e1c116
        repository: github.com/xianxu/ariadne
        reviewed_head: 740b851e5dbb5431f927352b93037c422fbd4666
        evidence_commit: bce36fbebd9c3dbe66f11f436d9e687e75257400
        landed_commit: e82c6e148cb7da99346422d083f04ab93a2360f0
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/head

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/head: Same exact head 184cd3a42a421c9eea1309871204b7d05aeeb890: recorded suite attempt fc26a661e6e90915454776817ced8e68 completed all 12 shards and 72 other packages, with only TestLandingTrainLandsEveryCarOnOneSuite/head failing at landingtrain_test.go:410 (finished helper two-minute merge timeout; head variant 139.59s). Saved failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fc26a661e6e90915454776817ced8e68.log, sha256 prefix fad0a6acdb5f. Without any source or HEAD change, focused go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 21.16s, null 20.01s, head 22.64s, total test 63.81s; package exit 0 in 64.342s. HEAD verified by git rev-parse before and after. Focused run is outside test-stats; this is explicit same-head fail-to-pass evidence. Register only /head, not its passing siblings.
