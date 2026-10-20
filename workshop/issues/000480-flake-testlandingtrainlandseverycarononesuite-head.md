---
id: 000480
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6488a03b1b1166345eb24ce438faa0f038b43edb' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/head

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/head: Same exact head 184cd3a42a421c9eea1309871204b7d05aeeb890: recorded suite attempt fc26a661e6e90915454776817ced8e68 completed all 12 shards and 72 other packages, with only TestLandingTrainLandsEveryCarOnOneSuite/head failing at landingtrain_test.go:410 (finished helper two-minute merge timeout; head variant 139.59s). Saved failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fc26a661e6e90915454776817ced8e68.log, sha256 prefix fad0a6acdb5f. Without any source or HEAD change, focused go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 21.16s, null 20.01s, head 22.64s, total test 63.81s; package exit 0 in 64.342s. HEAD verified by git rev-parse before and after. Focused run is outside test-stats; this is explicit same-head fail-to-pass evidence. Register only /head, not its passing siblings.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-19
