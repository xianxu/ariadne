---
id: 000503
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '905aa66fe8e36a90de1bb6ea4f1ce52f937b12d2' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestVouchingStartsWhenTrainsTurnOn

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestVouchingStartsWhenTrainsTurnOn: Recorded suite attempt bf7ab6a535a92236feeee763b72f9002 failed this test on exact head 12aff78f6515173afc4aabd5f28ee7f47b4306e4 (0.01s, no vouch after trains turned on); all other shards and packages passed. Targeted go test ./cmd/sdlc -run ^TestVouchingStartsWhenTrainsTurnOn$ -count=1 -v in a detached worktree at that same exact head passed (test 0.01s, package 0.482s). Fixture landingcar_test.go waits for gateReads, incremented inside readiness before the callback returns; it cancels immediately after seeing that counter. vouchWhileWaiting checks ctx.Err after readiness and before line.Vouch, so cancellation can win the publication race. No source change or deadline bump. Fix the fixture to wait for the persisted vouch, with adversarial scheduling coverage.

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-20
