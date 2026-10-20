---
id: 000503
status: codecomplete
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestVouchingStartsWhenTrainsTurnOn]
started: 2026-10-20T13:10:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.05
tracker:
    version: 1
    completion:
        token: close-67830a6c6c49
        repository: github.com/xianxu/ariadne
        reviewed_head: a99969209d1387ede6d2d71e297bff9c4b218abb
        evidence_commit: 3857b567dbc4ecc5d68d71ca949b7c703e4c9405
---

# flake: TestVouchingStartsWhenTrainsTurnOn

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestVouchingStartsWhenTrainsTurnOn: Recorded suite attempt bf7ab6a535a92236feeee763b72f9002 failed this test on exact head 12aff78f6515173afc4aabd5f28ee7f47b4306e4 (0.01s, no vouch after trains turned on); all other shards and packages passed. Targeted go test ./cmd/sdlc -run ^TestVouchingStartsWhenTrainsTurnOn$ -count=1 -v in a detached worktree at that same exact head passed (test 0.01s, package 0.482s). Fixture landingcar_test.go waits for gateReads, incremented inside readiness before the callback returns; it cancels immediately after seeing that counter. vouchWhileWaiting checks ctx.Err after readiness and before line.Vouch, so cancellation can win the publication race. No source change or deadline bump. Fix the fixture to wait for the persisted vouch, with adversarial scheduling coverage.
